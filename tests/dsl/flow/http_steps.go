// HTTP steps are written as .sio files are: Socket.IO tests written like .http ones. Each NAME starts a step, which is
// one of the verbs CONNECT <url>, EMIT <event> (with ACK to wait for its acknowledgement), WAIT <event>,
// COLLECT <event>, DISCONNECT, STREAM <url>, an HTTP request (GET <url> and so on) or RUN_BACKGROUND_JOBS.
// STREAM opens a page's live updates, server-sent events, with its H headers, and WAIT and COLLECT take its events
// by name, such as created, as they do the socket's.
// NS is the namespace CONNECT joins (/ by default), which the steps after it use.
// B is the step's body (CONNECT's auth, EMIT's argument, a request's JSON), and ARGS a JSON list of EMIT's
// arguments, for an event that takes more than one, as Nightscout's ack does. T is how long it may take (e.g. 5s),
// and SET and ASSERT_* work on what the step got as they do in .http files.

package flow

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
)

// HTTPFile is a parsed .sio file
type HTTPFile struct {
	Path  string
	Steps []*HTTPStep
}

// HTTPStep is one NAME block
type HTTPStep struct {
	Name string
	// Path and Line are where the step is, which is an INCLUDEd file for its steps
	Path   string
	Line   int
	Verb   string // CONNECT, EMIT, WAIT, COLLECT, DISCONNECT, RUN_BACKGROUND_JOBS or an HTTP method
	Target string // the URL for CONNECT and HTTP, the event for EMIT, WAIT and COLLECT
	NS     string // the namespace CONNECT joins
	Body   *string
	// Args is set when the body is ARGS, a list of EMIT's arguments
	Args    bool
	Headers http.Header
	Ack     bool
	Timeout time.Duration
	Asserts []HTTPAssert
	Sets    []Set
}

var socketVerbs = []string{"CONNECT", "EMIT", "WAIT", "COLLECT", "DISCONNECT", "STREAM"}
var httpVerbs = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

// Events are the events the file waits for or collects, which a connection must listen for
func (f *HTTPFile) Events() []string {
	var events []string
	for _, step := range f.Steps {
		if (step.Verb == "WAIT" || step.Verb == "COLLECT") && !slices.Contains(events, step.Target) {
			events = append(events, step.Target)
		}
	}
	return events
}

// ParseHTTPLines is the steps of a file's lines, which a .http file's are too, as each is a NAME with an HTTP verb
func ParseHTTPLines(lines []Line, path string) (*HTTPFile, error) {
	file := &HTTPFile{Path: path}
	var current *HTTPStep
	inBody := false
	for _, l := range lines {
		line, first, second, lineNumber, path := l.Text, l.First, l.Rest, l.Number, l.Path
		fail := func(format string, args ...any) error {
			return fmt.Errorf("%s:%d: %s", path, lineNumber, fmt.Sprintf(format, args...))
		}
		if first != "NAME" && first != "RUN_BACKGROUND_JOBS" && current == nil && !inBody {
			return nil, fail("%s before NAME", first)
		}
		if inBody && !isKeyword(first) {
			*current.Body += "\n" + line
			continue
		}
		inBody = false

		assert, isAssert, err := ParseAssert(first, second, path, lineNumber)
		switch {
		case err != nil:
			return nil, err
		case isAssert:
			current.Asserts = append(current.Asserts, assert)
		case first == "NAME":
			if second == "" {
				return nil, fail("empty step name")
			}
			current = &HTTPStep{Name: second, Path: path, Line: lineNumber, Headers: http.Header{}}
			file.Steps = append(file.Steps, current)
		case first == "RUN_BACKGROUND_JOBS":
			file.Steps = append(file.Steps, &HTTPStep{Name: first, Path: path, Line: lineNumber, Verb: first})
			current = nil
		case slices.Contains(socketVerbs, first) || slices.Contains(httpVerbs, first):
			if current.Verb != "" {
				return nil, fail("step %s already has %s", current.Name, current.Verb)
			}
			if (first == "DISCONNECT") != (second == "") {
				return nil, fail("%s needs a target, except DISCONNECT which takes none", first)
			}
			current.Verb, current.Target = first, second
		case first == "ACK":
			current.Ack = true
		case first == "NS":
			if !strings.HasPrefix(second, "/") {
				return nil, fail("NS %q must start with /", second)
			}
			current.NS = second
		case first == "T":
			timeout, err := time.ParseDuration(second)
			if err != nil || timeout <= 0 {
				return nil, fail("T %q must be a duration like 5s", second)
			}
			current.Timeout = timeout
		case first == "B", first == "ARGS":
			current.Body, current.Args = &second, first == "ARGS"
			inBody = true
		case first == "H":
			key, value, found := strings.Cut(second, ":")
			if !found {
				return nil, fail("invalid header %q", second)
			}
			current.Headers.Add(strings.TrimSpace(key), strings.TrimSpace(value))
		case first == "SET":
			key, value, found := strings.Cut(second, " ")
			if !found {
				return nil, fail("invalid SET %q", second)
			}
			current.Sets = append(current.Sets, Set{Key: key, Value: strings.TrimSpace(value)})
		default:
			return nil, fail("unknown line %q", line)
		}
	}
	for _, step := range file.Steps {
		if step.Verb == "" {
			return nil, fmt.Errorf("%s:%d: step %s has no verb", step.Path, step.Line, step.Name)
		}
		if step.Ack && step.Verb != "EMIT" {
			return nil, fmt.Errorf("%s:%d: ACK is only for EMIT", step.Path, step.Line)
		}
		if step.Args && step.Verb != "EMIT" {
			return nil, fmt.Errorf("%s:%d: ARGS is only for EMIT", step.Path, step.Line)
		}
		if step.NS != "" && step.Verb != "CONNECT" {
			return nil, fmt.Errorf("%s:%d: NS is only for CONNECT", step.Path, step.Line)
		}
	}
	return file, nil
}

// isKeyword is whether a line starts with something other than more of a body
func isKeyword(first string) bool {
	_, isAssert, _ := ParseAssert(first, "", "", 0)
	return isAssert || slices.Contains(socketVerbs, first) || slices.Contains(httpVerbs, first) ||
		slices.Contains([]string{"NAME", "RUN_BACKGROUND_JOBS", "ACK", "NS", "T", "B", "ARGS", "H", "SET"}, first)
}
