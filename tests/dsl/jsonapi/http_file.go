package jsonapi

import (
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"nighttroop.t1d.nz/tests/dsl/script"
)

// ParseRequestFile reads the requests of the file at path in fsys, and of the files it INCLUDEs
func ParseRequestFile(fsys fs.FS, path string) (*RequestFile, error) {
	requests, err := parseFile(fsys, path)
	if err != nil {
		return nil, err
	}

	// name is the file name without folders or extension
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return &RequestFile{
		Name:     name,
		Requests: requests,
	}, nil
}

func parseFile(fsys fs.FS, path string) ([]*Request, error) {
	lines, err := script.ReadLines(fsys, path)
	if err != nil {
		return nil, err
	}

	var requests []*Request
	var current *Request

	flush := func() {
		if current != nil {
			requests = append(requests, current)
			current = nil
		}
	}
	inBody := false
	for _, l := range lines {
		line, first, second, lineNumber, path := l.Text, l.First, l.Rest, l.Number, l.Path
		switch first {
		case "NAME":
			inBody = false
			flush()

			if second == "" {
				return nil, fmt.Errorf("empty request name")
			}

			current = &Request{
				Name:    second,
				Headers: make(http.Header),
				Asserts: []Assert{},
				Sets:    []Set{},
			}

		case "RUN_BACKGROUND_JOBS":
			inBody = false
			flush()
			requests = append(requests, &Request{Name: first, RunBackgroundJobs: true})

		case "GET", "POST", "PUT", "PATCH", "DELETE":
			inBody = false
			if current == nil {
				return nil, fmt.Errorf("request line before NAME")
			}

			current.Method = first
			current.URL = second
		case "SET":
			inBody = false
			if current == nil {
				return nil, fmt.Errorf("request line before NAME")
			}

			if second == "" {
				return nil, fmt.Errorf("invalid SET line: %q", line)
			}

			split := strings.SplitN(second, " ", 2)
			if len(split) != 2 {
				return nil, fmt.Errorf("invalid SET line: %q, split: %v", line, split)
			}

			current.AddSet(Set{
				Key:   strings.TrimSpace(split[0]),
				Value: strings.TrimSpace(split[1]),
			})

		case "ASSERT_STATUS", "ASSERT_EQUALS", "ASSERT_CONTAINS", "ASSERT_EXISTS", "ASSERT_LENGTH":
			inBody = false
			if current == nil {
				return nil, fmt.Errorf("%s before NAME", first)
			}
			assert, _, err := ParseAssert(first, second, path, lineNumber)
			if err != nil {
				return nil, err
			}
			current.AddAssert(assert)
		case "H":
			inBody = false
			if current == nil {
				return nil, fmt.Errorf("header before NAME")
			}

			parts := strings.SplitN(second, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid header: %q", second)
			}

			current.Headers.Add(
				strings.TrimSpace(parts[0]),
				strings.TrimSpace(parts[1]),
			)

		case "B":
			inBody = false
			if current == nil {
				return nil, fmt.Errorf("body before NAME")
			}
			inBody = true
			current.Body = &second
		default:
			if inBody {
				*current.Body += "\n" + line
				continue
			}
			return nil, fmt.Errorf("unknown line: %q", line)
		}
	}

	flush()

	return requests, nil
}

// ParseAssert is the assert on an ASSERT_* line, and false when the line is not one
func ParseAssert(first, second, path string, lineNumber int) (Assert, bool, error) {
	switch first {
	case "ASSERT_STATUS":
		statusCode, err := strconv.Atoi(second)
		if err != nil {
			return nil, true, fmt.Errorf("invalid assert_status: %q", second)
		}
		return AssertStatus{Path: path, Line: lineNumber, Code: statusCode}, true, nil

	case "ASSERT_EQUALS", "ASSERT_CONTAINS":
		expected, actual, err := parseAssertArgs(second)
		if err != nil {
			return nil, true, fmt.Errorf("%s:%d: invalid %s: %w", path, lineNumber, first, err)
		}
		if first == "ASSERT_EQUALS" {
			return AssertEquals{Path: path, Line: lineNumber, Expected: expected, Actual: actual}, true, nil
		}
		return AssertContains{Path: path, Line: lineNumber, Expected: expected, Actual: actual}, true, nil

	case "ASSERT_EXISTS":
		actual, err := strconv.Unquote(second)
		if err != nil {
			return nil, true, fmt.Errorf("%s:%d: invalid %s: value %q must be a quoted string", path, lineNumber, first, second)
		}
		return AssertExists{Path: path, Line: lineNumber, Actual: actual}, true, nil

	case "ASSERT_LENGTH":
		expected, actual, err := parseAssertArgs(second)
		length, isNumber := expected.(float64)
		if err == nil && (!isNumber || length != float64(int(length))) {
			err = fmt.Errorf("expected length %v must be a whole number", expected)
		}
		if err != nil {
			return nil, true, fmt.Errorf("%s:%d: invalid %s: %w", path, lineNumber, first, err)
		}
		return AssertLength{Path: path, Line: lineNumber, Expected: int(length), Actual: actual}, true, nil
	}
	return nil, false, nil
}
