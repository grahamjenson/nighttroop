// Package socketio runs .sio files: Socket.IO tests written like .http ones. Each NAME starts a step, which is
// one of the verbs CONNECT <url>, EMIT <event> (with ACK to wait for its acknowledgement), WAIT <event>,
// COLLECT <event>, DISCONNECT, an HTTP request (GET <url> and so on) or RUN_BACKGROUND_JOBS.
// NS is the namespace CONNECT joins (/ by default), which the steps after it use.
// B is the step's body (CONNECT's auth, EMIT's argument, a request's JSON), T how long it may take (e.g. 5s),
// and SET and ASSERT_* work on what the step got as they do in .http files.
package socketio

import (
	"bufio"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"nighttroop.t1d.nz/tests/dsl/jsonapi"
)

// File is a parsed .sio file
type File struct {
	Path  string
	Steps []*Step
}

// Step is one NAME block
type Step struct {
	Name    string
	Line    int
	Verb    string // CONNECT, EMIT, WAIT, COLLECT, DISCONNECT, RUN_BACKGROUND_JOBS or an HTTP method
	Target  string // the URL for CONNECT and HTTP, the event for EMIT, WAIT and COLLECT
	NS      string // the namespace CONNECT joins
	Body    *string
	Headers http.Header
	Ack     bool
	Timeout time.Duration
	Asserts []jsonapi.Assert
	Sets    []jsonapi.Set
}

var socketVerbs = []string{"CONNECT", "EMIT", "WAIT", "COLLECT", "DISCONNECT"}
var httpVerbs = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

// Events are the events the file waits for or collects, which a connection must listen for
func (f *File) Events() []string {
	var events []string
	for _, step := range f.Steps {
		if (step.Verb == "WAIT" || step.Verb == "COLLECT") && !slices.Contains(events, step.Target) {
			events = append(events, step.Target)
		}
	}
	return events
}

func ParseFile(contents string, path string) (*File, error) {
	file := &File{Path: path}
	var current *Step
	inBody := false
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		first, second, _ := strings.Cut(line, " ")
		second = strings.TrimSpace(second)
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

		assert, isAssert, err := jsonapi.ParseAssert(first, second, path, lineNumber)
		switch {
		case err != nil:
			return nil, err
		case isAssert:
			current.Asserts = append(current.Asserts, assert)
		case first == "NAME":
			if second == "" {
				return nil, fail("empty step name")
			}
			current = &Step{Name: second, Line: lineNumber, Headers: http.Header{}}
			file.Steps = append(file.Steps, current)
		case first == "RUN_BACKGROUND_JOBS":
			file.Steps = append(file.Steps, &Step{Name: first, Line: lineNumber, Verb: first})
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
		case first == "B":
			current.Body = &second
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
			current.Sets = append(current.Sets, jsonapi.Set{Key: key, Value: strings.TrimSpace(value)})
		default:
			return nil, fail("unknown line %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	for _, step := range file.Steps {
		if step.Verb == "" {
			return nil, fmt.Errorf("%s:%d: step %s has no verb", path, step.Line, step.Name)
		}
		if step.Ack && step.Verb != "EMIT" {
			return nil, fmt.Errorf("%s:%d: ACK is only for EMIT", path, step.Line)
		}
		if step.NS != "" && step.Verb != "CONNECT" {
			return nil, fmt.Errorf("%s:%d: NS is only for CONNECT", path, step.Line)
		}
	}
	return file, nil
}

// isKeyword is whether a line starts with something other than more of a body
func isKeyword(first string) bool {
	_, isAssert, _ := jsonapi.ParseAssert(first, "", "", 0)
	return isAssert || slices.Contains(socketVerbs, first) || slices.Contains(httpVerbs, first) ||
		slices.Contains([]string{"NAME", "RUN_BACKGROUND_JOBS", "ACK", "NS", "T", "B", "H", "SET"}, first)
}
