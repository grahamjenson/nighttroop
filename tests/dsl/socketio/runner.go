package socketio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	socket "nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/socketio/engineio"
	"nighttroop.t1d.nz/tests/dsl/jsonapi"
	"nighttroop.t1d.nz/utils/testutils"
)

const (
	defaultTimeout = 5 * time.Second
	// defaultCollect is how long COLLECT listens without a T
	defaultCollect = time.Second
)

// Runner runs a file's steps in order, with one Socket.IO connection at a time
type Runner struct {
	// Handler answers HTTP steps directly, as in .http tests
	Handler http.Handler
	// Server is what CONNECT dials, whatever its URL's host, so the Host header still picks the site
	Server *httptest.Server
	// RunBackgroundJobs runs every background job, for RUN_BACKGROUND_JOBS
	RunBackgroundJobs func()

	client    *socket.Client
	namespace string // the namespace CONNECT joined, which EMIT and the events use
	mu        sync.Mutex
	received  map[string][]any // events not yet waited for or collected
	arrived   chan struct{}
}

// Run runs the file, failing the test at the first step that cannot finish
func (r *Runner) Run(t *testing.T, file *File, ctx map[string]any) {
	defer r.disconnect()
	for _, step := range file.Steps {
		fmt.Printf("=============== %s: %s %s ================\n", step.Name, step.Verb, step.Target)
		output, err := r.run(t, file, step, ctx)
		if err != nil {
			t.Fatalf("%s:%d: %s %s: %v", file.Path, step.Line, step.Verb, step.Target, err)
		}
		if step.Verb == "RUN_BACKGROUND_JOBS" || isHTTP(step.Verb) {
			continue
		}
		pretty, _ := jsonapi.MarshalJsonWithoutEscapeHTML(output)
		fmt.Printf("GOT: %s\n", pretty)
		ctx[step.Name] = output
		for _, set := range step.Sets {
			if err := set.Set(ctx); err != nil {
				t.Fatalf("%s:%d: %v", file.Path, step.Line, err)
			}
		}
		for _, assert := range step.Asserts {
			assert.Execute(t, step.Name, 0, output, nil, ctx)
		}
	}
}

func isHTTP(verb string) bool { return slices.Contains(httpVerbs, verb) }

func (r *Runner) run(t *testing.T, file *File, step *Step, ctx map[string]any) (any, error) {
	timeout := step.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	switch step.Verb {
	case "RUN_BACKGROUND_JOBS":
		r.RunBackgroundJobs()
		return nil, nil
	case "CONNECT":
		return nil, r.connect(ctx, file.Events(), step, timeout)
	case "DISCONNECT":
		r.disconnect()
		return nil, nil
	}
	if isHTTP(step.Verb) {
		request := &jsonapi.Request{Name: step.Name, Method: step.Verb, URL: step.Target, Headers: step.Headers, Body: step.Body, Asserts: step.Asserts, Sets: step.Sets}
		_, err := request.ExecuteRequest(t, r.Handler, ctx)
		return nil, err
	}

	if r.client == nil {
		return nil, fmt.Errorf("not connected")
	}
	event, err := jsonapi.RenderTemplate(step.Target, ctx)
	if err != nil {
		return nil, err
	}
	switch step.Verb {
	case "EMIT":
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
func (r *Runner) connect(ctx map[string]any, events []string, step *Step, timeout time.Duration) error {
	r.disconnect()
	rawURL, err := jsonapi.RenderTemplate(step.Target, ctx)
	if err != nil {
		return err
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	version := engineio.V4
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
	r.namespace = socket.DefaultNamespace
	if step.NS != "" {
		r.namespace = step.NS
	}
	client := socket.NewClient(version)
	r.received, r.arrived = map[string][]any{}, make(chan struct{}, 1)
	for _, event := range events {
		client.On(r.namespace, event, func(args []json.RawMessage) {
			r.mu.Lock()
			r.received[event] = append(r.received[event], value(args))
			r.mu.Unlock()
			select {
			case r.arrived <- struct{}{}:
			default:
			}
		})
	}
	if err := client.Open(dialCtx, rawURL, testutils.HostClient(r.Server)); err != nil {
		return err
	}
	r.client = client
	var connectAuth any
	if len(auth) > 0 {
		connectAuth = auth[0]
	}
	return client.Connect(dialCtx, r.namespace, connectAuth)
}

func (r *Runner) disconnect() {
	if r.client != nil {
		r.client.Close()
		r.client = nil
	}
}

// wait takes the oldest event of that name not yet waited for or collected
func (r *Runner) wait(event string, timeout time.Duration) (any, error) {
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
func (r *Runner) body(step *Step, ctx map[string]any) ([]any, error) {
	if step.Body == nil {
		return nil, nil
	}
	rendered, err := jsonapi.RenderTemplate(*step.Body, ctx)
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
