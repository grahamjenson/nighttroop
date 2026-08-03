package integration_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Request struct {
	Name       string
	Method     string
	URL        string
	Headers    http.Header
	Body       string
	IsFunction bool

	AssertStatus int
	Assert       string
}

func (r Request) ToFlatString() string {
	return fmt.Sprintf("NAME %s METHOD %s URL %s", r.Name, r.Method, r.URL)
}

func (r Request) PrintRequest() {
	fmt.Printf("NAME %s\n", r.Name)
	fmt.Printf("%s %s\n", r.Method, r.URL)
	for k, vv := range r.Headers {
		for _, v := range vv {
			fmt.Printf("H %s: %s\n", k, v)
		}
	}
	if len(r.Body) > 0 {
		fmt.Printf("B %s\n", string(r.Body))
	}
	fmt.Println()
}

type RequestFile struct {
	Name     string
	Requests []Request
}

func ParseRequestFile(file string, ctx map[string]any) (*RequestFile, error) {
	requests, err := parseFile(strings.NewReader(file))
	if err != nil {
		return nil, err
	}

	// name is the file name without folders or extension
	name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	return &RequestFile{
		Name:     name,
		Requests: requests,
	}, nil
}

func parseFile(r io.Reader) ([]Request, error) {
	scanner := bufio.NewScanner(r)

	var requests []Request
	var current *Request

	flush := func() {
		if current != nil {
			requests = append(requests, *current)
			current = nil
		}
	}

	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		if line == "" {
			continue
		}

		// Continuation of body
		if current != nil &&
			current.Body != "" &&
			(len(raw) > 0 && (raw[0] == ' ' || raw[0] == '\t')) {

			current.Body += "\n" + strings.TrimLeft(raw, " \t")
			continue
		}

		switch {
		case strings.HasPrefix(line, "#"):
			// comment ignore
		case strings.HasPrefix(line, "NAME "):
			flush()

			name := strings.TrimSpace(strings.TrimPrefix(line, "NAME "))
			if name == "" {
				return nil, fmt.Errorf("empty request name")
			}

			current = &Request{
				Name:    name,
				Headers: make(http.Header),
			}

		case strings.HasPrefix(line, "GET "),
			strings.HasPrefix(line, "POST "),
			strings.HasPrefix(line, "PUT "),
			strings.HasPrefix(line, "PATCH "),
			strings.HasPrefix(line, "DELETE "):

			if current == nil {
				return nil, fmt.Errorf("request line before NAME")
			}

			parts := strings.SplitN(line, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid request line: %q", line)
			}

			current.Method = parts[0]
			current.URL = strings.TrimSpace(parts[1])
		case strings.HasPrefix(line, "SET "):
			// THIS IS A SPECIAL CASE FOR SETTING CONTEXT VARIABLES WITHOUT MAKING A REQUEST
			if current == nil {
				return nil, fmt.Errorf("request line before NAME")
			}

			parts := strings.SplitN(line, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid request line: %q", line)
			}

			//
			current.Method = parts[0]
			current.URL = strings.TrimSpace(parts[1])
			current.IsFunction = true
		case strings.HasPrefix(line, "ASSERT_STATUS "):
			if current == nil {
				return nil, fmt.Errorf("assert_status before NAME")
			}

			statusStr := strings.TrimSpace(strings.TrimPrefix(line, "ASSERT_STATUS "))
			statusCode, err := strconv.Atoi(statusStr)
			if err != nil {
				return nil, fmt.Errorf("invalid assert_status: %q", statusStr)
			}
			current.AssertStatus = statusCode

		case strings.HasPrefix(line, "ASSERT "):
			if current == nil {
				return nil, fmt.Errorf("assert before NAME")
			}

			assert := strings.TrimSpace(strings.TrimPrefix(line, "ASSERT "))
			current.Assert = assert
		case strings.HasPrefix(line, "H "):
			if current == nil {
				return nil, fmt.Errorf("header before NAME")
			}

			header := strings.TrimSpace(strings.TrimPrefix(line, "H "))

			parts := strings.SplitN(header, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid header: %q", header)
			}

			current.Headers.Add(
				strings.TrimSpace(parts[0]),
				strings.TrimSpace(parts[1]),
			)

		case strings.HasPrefix(line, "B "):
			if current == nil {
				return nil, fmt.Errorf("body before NAME")
			}

			current.Body = strings.TrimPrefix(line, "B ")

		default:
			return nil, fmt.Errorf("unknown line: %q", line)
		}
	}

	flush()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r Request) BuildRealRequest(ctx map[string]any) (Request, error) {

	url, err := renderTemplate(r.URL, ctx)
	if err != nil {
		return Request{}, fmt.Errorf("request %s url render failed: %w", r.Name, err)
	}

	if r.IsFunction {
		return Request{
			Name:       r.Name,
			Method:     r.Method,
			URL:        url,
			Headers:    r.Headers.Clone(),
			Body:       r.Body,
			IsFunction: true,
		}, nil
	}

	// ---- render body ----
	bodyStr, err := renderTemplate(string(r.Body), ctx)
	if err != nil {
		return Request{}, fmt.Errorf("request %s body render failed: %w", r.Name, err)
	}

	// -- render headers ----
	renderedHeaders := make(http.Header)
	for k, vv := range r.Headers {
		for _, v := range vv {
			renderedValue, err := renderTemplate(v, ctx)
			if err != nil {
				return Request{}, fmt.Errorf("request %s header %s render failed: %w", r.Name, k, err)
			}
			renderedHeaders.Add(k, renderedValue)
		}
	}
	r.Headers = renderedHeaders

	// Default content and accept types
	if r.Headers.Get("Content-Type") == "" {
		r.Headers.Set("Content-Type", "application/json")
	}
	if r.Headers.Get("Accept") == "" {
		r.Headers.Set("Accept", "application/json")
	}
	return Request{
		Name:         r.Name,
		Method:       r.Method,
		URL:          url,
		Headers:      renderedHeaders,
		Body:         bodyStr,
		IsFunction:   r.IsFunction,
		AssertStatus: r.AssertStatus,
		Assert:       r.Assert,
	}, nil
}

func (r Request) ExecuteRequest(t *testing.T, handler http.Handler, context map[string]any) (int, map[string]any, http.Header, map[string]any) {
	var buf bytes.Buffer
	if r.Body != "" {
		if err := json.NewEncoder(&buf).Encode(r.Body); err != nil {
			t.Fatalf("failed to encode json: %v", err)
		}
	}

	req := httptest.NewRequest(r.Method, r.URL, &buf)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var out map[string]any

	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		fmt.Println("failed to decode json for", r.Name, "error:", err)
	}

	context[r.Name] = out
	context[r.Name+"_status"] = rr.Code

	fmt.Println("Executed request", r.Name, "status", rr.Code, "output", out, "fsasfsaf", r.AssertStatus)
	if r.AssertStatus != 0 && rr.Code != r.AssertStatus {
		t.Fatalf("request %s expected status %d but got %d", r.Name, r.AssertStatus, rr.Code)
	}

	if r.Assert != "" {
		assertValue, err := renderTemplate(r.Assert, context)
		if err != nil {
			t.Fatalf("request %s assert render failed: %v", r.Name, err)
		}
		if assertValue != "true" {
			t.Fatalf("request %s assert failed: %s", r.Name, r.Assert)
		}
	}

	return rr.Code, out, rr.Header(), context
}

func (r Request) ExecuteFunction(ctx map[string]any) (map[string]any, error) {
	if !r.IsFunction {
		return ctx, nil
	}

	if r.Method == "SET" {
		ctx[r.Name] = r.URL
		return ctx, nil
	}

	return ctx, fmt.Errorf("unsupported method %q", r.Method)
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
	var cur any = root

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

func renderTemplate(input string, ctx map[string]any) (string, error) {
	tmpl, err := template.New("req").
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"dig": func(path string) any {
				return templateDig(ctx, path)
			},
			"token": func() string {
				return ctx["token"].(string)
			},
			"read_token": func() string {
				return ctx["read_token"].(string)
			},
			"jwt": func() string {
				return ctx["jwt"].(string)
			},
			"read_jwt": func() string {
				return ctx["read_jwt"].(string)
			},
			"time":  templateTime,
			"itime": templateITime,
			"tokenByName": func(list []any, name string) string {
				for _, item := range list {
					m := item.(map[string]any)
					if m["name"] == name {
						return m["accessToken"].(string)
					}
				}
				return ""
			},
		}).
		Parse(input)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}
