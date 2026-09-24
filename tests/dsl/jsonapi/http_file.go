package jsonapi

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func ParseRequestFile(fileContents string, ctx map[string]any, path string) (*RequestFile, error) {
	requests, err := parseFile(strings.NewReader(fileContents), path)
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

func parseFile(r io.Reader, path string) ([]*Request, error) {
	scanner := bufio.NewScanner(r)

	var requests []*Request
	var current *Request

	fmt.Println("Parsing request file:", path)
	flush := func() {
		if current != nil {
			requests = append(requests, current)
			current = nil
		}
	}
	lineNumber := 0
	inBody := false
	for scanner.Scan() {
		lineNumber += 1
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		if line == "" {
			continue
		}

		// Continuation of body

		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		split := strings.SplitN(line, " ", 2)
		if len(split) == 0 {
			continue
		}
		second := ""
		if len(split) > 1 {
			second = strings.TrimSpace(split[1])
		}
		first := strings.TrimSpace(split[0])

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

		case "ASSERT_STATUS":
			inBody = false
			if current == nil {
				return nil, fmt.Errorf("assert_status before NAME")
			}

			statusCode, err := strconv.Atoi(second)
			if err != nil {
				return nil, fmt.Errorf("invalid assert_status: %q", second)
			}
			current.AddAssert(AssertStatus{
				Path: path,
				Line: lineNumber,
				Code: statusCode,
			})

		case "ASSERT":
			inBody = false
			if current == nil {
				return nil, fmt.Errorf("assert before NAME")
			}

			current.AddAssert(AssertTemplate{
				Path: path,
				Line: lineNumber,
				Raw:  second,
			})
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

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func templateTime(offset int) string {
	// top of the minute
	t := time.Now().Add(-time.Duration(offset) * time.Minute)
	return t.Format(time.RFC3339)
}

func templateITime(offset int) int64 {
	// top of the minute
	t := time.Now().Add(-time.Duration(offset) * time.Minute)
	return t.UnixMilli()
}

func templateDig(root any, path string) any {

	parts := strings.Split(path, ".")
	var cur = root

	for _, p := range parts {
		switch v := cur.(type) {

		case map[string]any:
			next, ok := v[p]
			if !ok {
				return nil
			}
			cur = next

		case []any:
			// allow numeric indexing into slices
			i, err := strconv.Atoi(p)
			if err != nil || i < 0 || i >= len(v) {
				return nil
			}
			cur = v[i]

		default:
			return nil
		}
	}

	return cur
}
