package integration_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

type Request struct {
	Name     string
	Method   string
	Path     string
	URL      string
	Headers  http.Header
	Body     string
	IsMethod bool
}

func (r Request) PrintReqeust() {
	fmt.Printf("NAME %s\n", r.Name)
	fmt.Printf("%s %s\n", r.Method, r.Path)
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
	Name      string
	Requests  []Request
	Responses []Response
	CTX       map[string]any
}

type Response struct {
	Request    Request
	Name       string
	StatusCode int
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
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
		CTX:      ctx,
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
			current.URL = parts[1]
			// We split the URL into / and just take the first three parts

			u, err := url.Parse(parts[1])
			if err != nil {
				fmt.Println("invalid URL in line", line, ":", err)
			} else {
				cleanPath := strings.TrimSuffix(strings.TrimSuffix(u.Path, "/"), ".json")
				pathParts := strings.Split(cleanPath, "/")
				if len(pathParts) >= 4 {
					current.Path = strings.Join(pathParts[:4], "/")
				} else {
					current.Path = u.Path
				}
			}
		case strings.HasPrefix(line, "SET "):
			// THIS IS A SPECIAL CASE FOR SETTING CONTEXT VARIABLES WITHOUT MAKING A REQUEST
			if current == nil {
				return nil, fmt.Errorf("request line before NAME")
			}

			parts := strings.SplitN(line, " ", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid request line: %q", line)
			}

			current.Method = parts[0]
			current.Path = parts[1]
			current.URL = parts[1]
			current.IsMethod = true
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

func (rf *RequestFile) nonMethodRequestLength() int {
	count := 0
	for _, r := range rf.Requests {
		if !r.IsMethod {
			count += 1
		}
	}
	return count
}

type savedResponse struct {
	Request    savedRequest `json:"request"`
	StatusCode int          `json:"statusCode"`
	Headers    http.Header  `json:"headers"`
	Body       any          `json:"body"`
	DurationMs int64        `json:"durationMs"`
}

type savedRequest struct {
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	URL     string      `json:"url"`
	Headers http.Header `json:"headers"`
	Body    any         `json:"body"`
}

func (rf *RequestFile) WriteResponsesToFolder(name string) error {
	folder := fmt.Sprintf("results/%s/%s", name, rf.Name)
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return err
	}

	for i, resp := range rf.Responses {

		out := savedResponse{
			Request:    createSavedRequest(resp.Request),
			StatusCode: resp.StatusCode,
			Headers:    resp.Headers,
			DurationMs: resp.Duration.Milliseconds(),
		}

		// If the body is JSON, embed it as JSON.
		// Otherwise store it as a string.
		if len(resp.Body) > 0 {
			var body any
			if err := json.Unmarshal(resp.Body, &body); err == nil {
				out.Body = body
			} else {
				out.Body = string(resp.Body)
			}
		}

		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}

		filename := filepath.Join(
			folder,
			fmt.Sprintf("%03d.%s.json", i+1, sanitizeFilename(resp.Name)),
		)

		data = []byte(strings.ReplaceAll(string(data), `\u0026`, "&"))
		if err := os.WriteFile(filename, data, 0o644); err != nil {
			return err
		}
	}

	return nil
}
func createSavedRequest(req Request) savedRequest {

	var body any
	if len(req.Body) > 0 {
		if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
			body = string(req.Body)
		}
	}

	return savedRequest{
		Method:  req.Method,
		Path:    req.Path,
		URL:     req.URL,
		Headers: req.Headers,
		Body:    body,
	}
}

func sanitizeFilename(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	var b strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) ||
			unicode.IsDigit(r) ||
			r == '-' ||
			r == '_' {
			b.WriteRune(r)
		}
	}

	return b.String()
}
