package flow

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	defaultTimeout = 5 * time.Second
	// defaultCollect is how long COLLECT listens without a T
	defaultCollect = time.Second
	// defaultNamespace is the namespace CONNECT joins without an NS
	defaultNamespace = "/"
	// defaultEIO is the Engine.IO version CONNECT speaks without an EIO in its URL
	defaultEIO = 4
)

// Socket is a Socket.IO client, which CONNECT opens and EMIT, WAIT and COLLECT use
type Socket interface {
	// On handles the namespace's event, registered before the connection opens
	On(namespace, event string, handler func(args []json.RawMessage))
	// Open opens the Engine.IO connection to the URL, sending its requests with the HTTP client
	Open(ctx context.Context, rawURL string, httpClient *http.Client) error
	// Connect joins the namespace, sending auth
	Connect(ctx context.Context, namespace string, auth any) error
	Emit(namespace, event string, args ...any) error
	// EmitWithAck sends the event and waits for its acknowledgement's arguments
	EmitWithAck(ctx context.Context, namespace, event string, args ...any) ([]json.RawMessage, error)
	Close() error
}

// httpRunner runs a file's steps one at a time with Step, keeping one Socket.IO connection open between them until Close
type httpRunner struct {
	// Client sends the HTTP steps' requests, and CONNECT's and STREAM's, as Client does to a test server
	Client *http.Client
	// RunBackgroundJobs runs every background job, for RUN_BACKGROUND_JOBS
	RunBackgroundJobs func()
	// NewSocket makes a Socket.IO client speaking that Engine.IO version, for CONNECT
	NewSocket func(eio int) Socket

	client    Socket
	namespace string // the namespace CONNECT joined, which EMIT and the events use
	mu        sync.Mutex
	received  map[string][]any // events, the socket's and the streams', not yet waited for or collected
	arrived   chan struct{}
	streams   []context.CancelFunc // the STREAMs open, each ended by Close
}

// Step runs one of the file's steps, keeping its connection open for the next, and fails the test if it cannot
// finish. The file says which events a connection listens for.
func (r *httpRunner) Step(t *testing.T, file *HTTPFile, step *HTTPStep, ctx map[string]any) {
	t.Helper()
	logf(t, "%s:%d: %s %s %s", step.Path, step.Line, step.Name, step.Verb, step.Target)
	output, err := r.run(t, file, step, ctx)
	if err != nil {
		t.Fatalf("%s:%d: %s %s: %v", step.Path, step.Line, step.Verb, step.Target, err)
	}
	if step.Verb == "RUN_BACKGROUND_JOBS" || step.Verb == "SLEEP" || isHTTP(step.Verb) {
		return
	}
	pretty, _ := MarshalJsonWithoutEscapeHTML(output)
	logf(t, "  got: %s", pretty)
	ctx[step.Name] = output
	for _, set := range step.Sets {
		if err := set.Set(ctx); err != nil {
			t.Fatalf("%s:%d: %v", step.Path, step.Line, err)
		}
	}
	for _, assert := range step.Asserts {
		assert.Execute(t, step.Name, 0, output, nil, ctx)
	}
}

// Close ends the connection, if there is one, and every STREAM
func (r *httpRunner) Close() {
	r.disconnect()
	for _, end := range r.streams {
		end()
	}
	r.streams = nil
}

// receive queues an event, from the socket or a stream, for a WAIT or COLLECT
func (r *httpRunner) receive(event string, value any) {
	r.mu.Lock()
	r.received[event] = append(r.received[event], value)
	r.mu.Unlock()
	select {
	case r.arrived <- struct{}{}:
	default:
	}
}

func isHTTP(verb string) bool { return slices.Contains(httpVerbs, verb) }

// Request is an HTTP step's request, or nil for any other step
func (s *HTTPStep) Request() *Request {
	if !isHTTP(s.Verb) {
		return nil
	}
	return &Request{Name: s.Name, Method: s.Verb, URL: s.Target, Headers: s.Headers, Body: s.Body, Asserts: s.Asserts, Sets: s.Sets}
}

func (r *httpRunner) run(t *testing.T, file *HTTPFile, step *HTTPStep, ctx map[string]any) (any, error) {
	timeout := step.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	switch step.Verb {
	case "RUN_BACKGROUND_JOBS":
		r.RunBackgroundJobs()
		return nil, nil
	case "SLEEP":
		time.Sleep(step.Timeout)
		return nil, nil
	case "CONNECT":
		return nil, r.connect(ctx, file.Events(), step, timeout)
	case "DISCONNECT":
		r.disconnect()
		return nil, nil
	case "STREAM":
		return nil, r.stream(ctx, step, timeout)
	}
	if isHTTP(step.Verb) {
		step.Request().Run(t, r.Client, ctx)
		return nil, nil
	}

	if r.received == nil {
		return nil, fmt.Errorf("not connected, and no stream open")
	}
	event, err := Render(step.Target, ctx)
	if err != nil {
		return nil, err
	}
	switch step.Verb {
	case "EMIT":
		if r.client == nil {
			return nil, fmt.Errorf("not connected")
		}
		args, err := r.body(step, ctx)
		if err != nil {
			return nil, err
		}
		if !step.Ack {
			return nil, r.client.Emit(r.namespace, event, args...)
		}
		ackCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		reply, err := r.client.EmitWithAck(ackCtx, r.namespace, event, args...)
		if err != nil {
			return nil, fmt.Errorf("no ACK within %s: %w", timeout, err)
		}
		return value(reply), nil
	case "WAIT":
		return r.wait(event, timeout)
	case "COLLECT":
		if step.Timeout == 0 {
			timeout = defaultCollect
		}
		time.Sleep(timeout)
		r.mu.Lock()
		defer r.mu.Unlock()
		collected := append([]any{}, r.received[event]...)
		r.received[event] = nil
		return collected, nil
	}
	return nil, fmt.Errorf("unknown verb")
}

// connect opens the URL's Socket.IO connection, speaking the EIO version in its query (4 by default) and sending
// the body as v5's auth, and listens for the events the file uses
func (r *httpRunner) connect(ctx map[string]any, events []string, step *HTTPStep, timeout time.Duration) error {
	r.disconnect()
	rawURL, err := Render(step.Target, ctx)
	if err != nil {
		return err
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if r.NewSocket == nil {
		return fmt.Errorf("no Socket.IO client: the runner has no NewSocket")
	}
	version := defaultEIO
	if eio := u.Query().Get("EIO"); eio != "" {
		if version, err = strconv.Atoi(eio); err != nil {
			return fmt.Errorf("bad EIO %q", eio)
		}
	}
	auth, err := r.body(step, ctx)
	if err != nil {
		return err
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	r.namespace = defaultNamespace
	if step.NS != "" {
		r.namespace = step.NS
	}
	client := r.NewSocket(version)
	r.listen()
	r.mu.Lock()
	// A new connection starts with none of the last one's events, though a stream's are kept
	for _, event := range events {
		delete(r.received, event)
	}
	r.mu.Unlock()
	for _, event := range events {
		client.On(r.namespace, event, func(args []json.RawMessage) { r.receive(event, value(args)) })
	}
	if err := client.Open(dialCtx, rawURL, r.Client); err != nil {
		return err
	}
	r.client = client
	var connectAuth any
	if len(auth) > 0 {
		connectAuth = auth[0]
	}
	return client.Connect(dialCtx, r.namespace, connectAuth)
}

func (r *httpRunner) disconnect() {
	if r.client != nil {
		r.client.Close()
		r.client = nil
	}
}

// wait takes the oldest event of that name not yet waited for or collected
func (r *httpRunner) wait(event string, timeout time.Duration) (any, error) {
	deadline := time.After(timeout)
	for {
		r.mu.Lock()
		if queued := r.received[event]; len(queued) > 0 {
			r.received[event] = queued[1:]
			r.mu.Unlock()
			return queued[0], nil
		}
		r.mu.Unlock()
		select {
		case <-r.arrived:
		case <-deadline:
			return nil, fmt.Errorf("no %s within %s", event, timeout)
		}
	}
}

// body is the step's B, rendered and parsed as the one JSON argument it sends
func (r *httpRunner) body(step *HTTPStep, ctx map[string]any) ([]any, error) {
	if step.Body == nil {
		return nil, nil
	}
	rendered, err := Render(*step.Body, ctx)
	if err != nil {
		return nil, err
	}
	if step.Args {
		var args []any
		if err := json.Unmarshal([]byte(rendered), &args); err != nil {
			return nil, fmt.Errorf("ARGS is not a JSON list: %w", err)
		}
		return args, nil
	}
	var arg any
	if err := json.Unmarshal([]byte(rendered), &arg); err != nil {
		return nil, fmt.Errorf("B is not valid JSON: %w", err)
	}
	return []any{arg}, nil
}

// value is what an event or ACK gives the asserts: its one argument, nil for none, or the list of them
func value(args []json.RawMessage) any {
	values := make([]any, len(args))
	for i, arg := range args {
		json.Unmarshal(arg, &values[i])
	}
	switch len(values) {
	case 0:
		return nil
	case 1:
		return values[0]
	}
	return values
}

// listen starts the queue of events, if it has not
func (r *httpRunner) listen() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.received == nil {
		r.received, r.arrived = map[string][]any{}, make(chan struct{}, 1)
	}
}

// stream opens the URL's server-sent events with the step's headers, and queues each event by its name, its data
// as JSON when it is, until Close. It fails unless the stream is open within the timeout.
func (r *httpRunner) stream(ctx map[string]any, step *HTTPStep, timeout time.Duration) error {
	request, err := (&Request{Name: step.Name, Method: http.MethodGet, URL: step.Target, Headers: step.Headers}).Render(ctx)
	if err != nil {
		return err
	}
	streamCtx, end := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(streamCtx, http.MethodGet, request.URL, nil)
	if err != nil {
		end()
		return err
	}
	req.Header = request.Headers
	req.Header.Set("Accept", "text/event-stream")
	opened := time.AfterFunc(timeout, end)
	res, err := r.Client.Do(req)
	opened.Stop()
	if err != nil {
		end()
		return err
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		end()
		return fmt.Errorf("stream answered %d", res.StatusCode)
	}
	r.listen()
	r.streams = append(r.streams, end)
	go func() {
		defer res.Body.Close()
		event, data := "message", ""
		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event:"):
				event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			case strings.HasPrefix(line, "data:"):
				data += strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			case line == "" && data != "":
				var parsed any = data
				_ = json.Unmarshal([]byte(data), &parsed)
				r.receive(event, parsed)
				event, data = "message", ""
			}
		}
	}()
	return nil
}
