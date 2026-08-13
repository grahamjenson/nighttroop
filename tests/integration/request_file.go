package integration_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"
)

type AssertType struct {
	Path string
	Line int
	Raw  string
}

type AssertStatusType struct {
	Path string
	Line int
	Code int
}

type Request struct {
	Name       string
	Method     string
	URL        string
	Headers    http.Header
	Body       string
	IsFunction bool

	AssertStatus AssertStatusType
	Assert       []AssertType
}

func (r Request) ToFlatString() string {
	return fmt.Sprintf("NAME %s METHOD %s URL %s", r.Name, r.Method, r.URL)
}

func (r Request) PrintRequest() {
	// String URL of token param
	urlToPrint := r.URL
	u, err := url.Parse(r.URL)
	if err == nil {
		q := u.Query()
		q.Del("token")
		u.RawQuery = q.Encode()
		urlToPrint = u.String()
	}
	fmt.Printf("REQUEST %s: %s %s\n", r.Name, r.Method, urlToPrint)
	if len(r.Body) > 0 {
		fmt.Printf("BODY %s\n", strings.ReplaceAll(string(r.Body), "\n", " "))
	}
	for k, v := range r.Headers {
		if k == "Accept" || k == "Content-Type" {
			continue
		}
		if k == "Authorization" {
			v = []string{"<REDACTED>"}
		}
		fmt.Printf("H %s: %s\n", k, strings.Join(v, ", "))
	}
}

type RequestFile struct {
	Name     string
	Requests []Request
}

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

func parseFile(r io.Reader, path string) ([]Request, error) {
	scanner := bufio.NewScanner(r)

	var requests []Request
	var current *Request

	fmt.Println("Parsing request file:", path)
	flush := func() {
		if current != nil {
			requests = append(requests, *current)
			current = nil
		}
	}
	lineNumber := 0
	for scanner.Scan() {
		lineNumber += 1
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
		case strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//"):
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
			current.AssertStatus = AssertStatusType{
				Path: path,
				Line: lineNumber,
				Code: statusCode,
			}

		case strings.HasPrefix(line, "ASSERT "):
			if current == nil {
				return nil, fmt.Errorf("assert before NAME")
			}

			assert := strings.TrimSpace(strings.TrimPrefix(line, "ASSERT "))
			current.Assert = append(current.Assert, AssertType{
				Path: path,
				Line: lineNumber,
				Raw:  assert,
			})
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
	// check body is valid JSON
	if bodyStr != "" {
		var js any
		if err := json.Unmarshal([]byte(bodyStr), &js); err != nil {
			return Request{}, fmt.Errorf("request %s body is not valid JSON: %w", r.Name, err)
		}
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

func (r Request) ExecuteRequest(handler http.Handler, context map[string]any) (code int, body any, headers http.Header, outContext map[string]any, err error) {

	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("httptest.NewRequest panicked: %v", recovered)
		}
	}()

	if r.Body != "" {
		// Making Sure the body is JSON
		err = json.Unmarshal([]byte(r.Body), new(any))
		if err != nil {
			return 0, nil, nil, context, fmt.Errorf("request %s body %s is not valid JSON: %v", r.Name, r.Body, err)
		}

	}

	req := httptest.NewRequest(r.Method, r.URL, strings.NewReader(r.Body))
	for k, vv := range r.Headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var out any
	bodyOut := rr.Body.String()
	if err := json.NewDecoder(strings.NewReader(bodyOut)).Decode(&out); err != nil {
		fmt.Println("failed to decode json for", r.Name, "error:", err)
	}
	if err != nil {
		return rr.Code, nil, rr.Header(), context, fmt.Errorf("request %s response body is not valid JSON: %v", r.Name, err)
	}
	context[r.Name] = out
	context[r.Name+"_status"] = rr.Code

	return rr.Code, out, rr.Header(), context, nil
}

func (r Request) ExecuteAsserts(t *testing.T, statusCode int, output any, ctx map[string]any) {
	code := r.AssertStatus.Code
	if code != 0 {
		message := fmt.Sprintf("request %s expected status %d but got %d with body %v", r.Name, code, statusCode, output)
		AssertEqual(t, code, statusCode, message, r.AssertStatus.Path, r.AssertStatus.Line)
	}

	for _, assertStruct := range r.Assert {
		_, err := assertTemplateCheck(assertStruct, ctx, output, t)
		if err != nil {
			t.Fatalf("assert Template Error %s assert %s output %v render failed: %v", r.Name, assertStruct.Raw, output, err)
		}
	}
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
				dug := templateDig(ctx, path)
				switch dug.(type) {
				case nil:
					return "no_value"
				case string:
					return dug.(string)
				case float64:
					return fmt.Sprintf("%.0f", dug.(float64))
				case int:
					return dug.(int)
				case bool:
					return dug.(bool)
				default:
					return dug
				}
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

func assertTemplateCheck(assertStruct AssertType, ctx map[string]any, localResult any, t *testing.T) (string, error) {
	prettyOutput, _ := json.MarshalIndent(localResult, "", "  ")
	tmpl, err := template.New("req").
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"dig": func(path string) any {
				return templateDig(ctx, path)
			},
			// TESTING FUNCTIONS
			"assertEquals": func(rawExpected, rawActual any) string {
				expected, actual := normalizeExpectedAndAcutal(rawExpected, rawActual, localResult)
				message := fmt.Sprintf("assertEquals failed: expected %v:%T got %v:%T from [%v] output: %s", expected, expected, actual, actual, rawActual, prettyOutput)
				AssertEqual(t, expected, actual, message, assertStruct.Path, assertStruct.Line)
				return "true"
			},
			"assertExists": func(rawActual any) string {
				_, actual := normalizeExpectedAndAcutal(rawActual, rawActual, localResult)
				message := fmt.Sprintf("assertExists failed: expected value to exist, got nil from [%v], output: %s", rawActual, prettyOutput)
				AssertNotNil(t, actual, message, assertStruct.Path, assertStruct.Line)
				return "true"
			},
		}).
		Parse(assertStruct.Raw)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func normalizeExpectedAndAcutal(expected, actual, localResult any) (any, any) {
	if s, ok := actual.(string); ok && strings.HasPrefix(s, "@") {
		// handle template path/reference case
		actual = templateDig(localResult, s[1:])
	}

	if s, ok := expected.(string); ok && strings.HasPrefix(s, "@") {
		// handle template path/reference case
		expected = templateDig(localResult, s[1:])
	}
	expected = normalizeValue(expected)
	actual = normalizeValue(actual)
	return expected, actual
}
func normalizeValue(v any) any {
	switch x := v.(type) {
	case int:
		return float64(x)
	case int8:
		return float64(x)
	case int16:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)

	case uint:
		return float64(x)
	case uint8:
		return float64(x)
	case uint16:
		return float64(x)
	case uint32:
		return float64(x)
	case uint64:
		return float64(x)

	case float32:
		return float64(x)

	case []any:
		for i := range x {
			x[i] = normalizeValue(x[i])
		}
		return x

	case map[string]any:
		for k, v := range x {
			x[k] = normalizeValue(v)
		}
		return x

	default:
		return v
	}
}

func AssertEqual(t *testing.T, expected, actual any, message string, file string, line int) bool {
	t.Helper()

	if reflect.DeepEqual(expected, actual) {
		return true
	}

	t.Errorf(
		"\n ASSERT FAILED \n %s:%d: %s\nexpected: %#v\nactual:   %#v",
		file,
		line,
		message,
		expected,
		actual,
	)

	return false
}

func AssertNotNil(t *testing.T, actual any, message string, file string, line int) bool {
	t.Helper()

	if actual != nil {
		return true
	}

	t.Errorf(
		"\n ASSERT FAILED \n %s:%d: %s\nactual:   %#v",
		file,
		line,
		message,
		actual,
	)

	return false
}
