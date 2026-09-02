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

func (r Request) ExecuteAsserts(t *testing.T, statusCode int, output any, headers http.Header, ctx map[string]any) {
	code := r.AssertStatus.Code

	outputRaw, _ := MarshalJsonWithoutEscapeHTML(output)

	if code != 0 {
		message := fmt.Sprintf("request %s expected status %d but got %d with body %v", r.Name, code, statusCode, string(outputRaw))
		AssertEqual(t, code, statusCode, message, r.AssertStatus.Path, r.AssertStatus.Line)
	}

	for _, assertStruct := range r.Assert {
		_, err := assertTemplateCheck(assertStruct, ctx, output, headers, t)
		if err != nil {
			t.Fatalf("assert Template Error %s assert %s output %v render failed: %v", r.Name, assertStruct.Raw, string(outputRaw), err)
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
			"site": func() string {
				if v, ok := ctx["site"]; ok && v != nil {
					return ctx["site"].(string)
				}
				return ctx["default_site"].(string)
			},
			"token": func() string {
				if v, ok := ctx["token"]; ok && v != nil {
					return ctx["token"].(string)
				}
				return ctx["default_token"].(string)
			},
			"read_token": func() string {
				return ctx["read_token"].(string)
			},
			"jwt": func() string {
				if v, ok := ctx["jwt"]; ok && v != nil {
					return ctx["jwt"].(string)
				}
				return ctx["default_jwt"].(string)
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

func assertTemplateCheck(assertStruct AssertType, ctx map[string]any, localResult any, headers map[string][]string, t *testing.T) (string, error) {
	// Need to forse the headers to be json-esk so types work
	b, err := json.Marshal(headers)
	if err != nil {
		t.Fatalf("Some weird Marshalerror happened in templateDig: %v", err)
		return "false", err
	}
	var headersConverted map[string]any
	err = json.Unmarshal(b, &headersConverted)
	if err != nil {
		t.Fatalf("Some weird Unmarshal error happened in templateDig: %v", err)
		return "false", err
	}

	prettyOutput, _ := MarshalJsonWithoutEscapeHTML(localResult)
	tmpl, err := template.New("req").
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"dig": func(path string) any {
				return templateDig(ctx, path)
			},
			// TESTING FUNCTIONS
			"assertEquals": func(rawExpected, rawActual any) string {
				expected := normalizeStringSearch(rawExpected, localResult, headersConverted, ctx)
				actual := normalizeStringSearch(rawActual, localResult, headersConverted, ctx)
				message := fmt.Sprintf("assertEquals failed: expected %v:%T got %v:%T from [%v] output: %s", expected, expected, actual, actual, rawActual, prettyOutput)
				AssertEqual(t, expected, actual, message, assertStruct.Path, assertStruct.Line)
				return "true"
			},
			"assertContains": func(rawExpected, rawActual any) string {
				expected := normalizeStringSearch(rawExpected, localResult, headersConverted, ctx)
				actual := normalizeStringSearch(rawActual, localResult, headersConverted, ctx)
				message := fmt.Sprintf("assertContains failed: expected %v:%T got %v:%T from [%v] output: %s", expected, expected, actual, actual, rawActual, prettyOutput)
				AssertContains(t, formatAnyAsString(expected), actual, message, assertStruct.Path, assertStruct.Line)
				return "true"
			},
			"assertExists": func(rawActual any) string {
				actual := normalizeStringSearch(rawActual, localResult, headersConverted, ctx)
				message := fmt.Sprintf("assertExists failed: expected value to exist, got nil from [%v], output: %s", rawActual, prettyOutput)
				AssertNotNil(t, actual, message, assertStruct.Path, assertStruct.Line)
				return "true"
			},
			"assertLength": func(expectedLength int, rawActual any) string {
				actual := normalizeStringSearch(rawActual, localResult, headersConverted, ctx)
				actualLength := -1
				if arr, ok := actual.([]any); ok {
					actualLength = len(arr)
				}
				message := fmt.Sprintf("assertLength failed: expected length %d, got %d from [%v], output: %s", expectedLength, actualLength, rawActual, prettyOutput)
				AssertEqual(t, expectedLength, actualLength, message, assertStruct.Path, assertStruct.Line)
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

func MarshalJsonWithoutEscapeHTML(v any) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func normalizeStringSearch(expected, localResult any, headers map[string]any, ctx map[string]any) any {

	// Allow the expected to look through all ctx
	if s, ok := expected.(string); ok {
		if strings.HasPrefix(s, ".") {
			// look up in the context
			expected = templateDig(ctx, s[1:])
		} else if strings.HasPrefix(s, "@") {
			// look up in localResult
			expected = templateDig(localResult, s[1:])
		} else if strings.HasPrefix(s, "#") {
			// look up in headers
			expected = templateDig(headers, s[1:])
		}
	}

	expected = normalizeValue(expected)

	return expected
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

func formatAnyAsString(v any) string {
	switch v := v.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

func AssertContains(t *testing.T, expected, actual any, message string, file string, line int) bool {
	t.Helper()

	contains := false

	switch a := actual.(type) {
	case string:
		e, ok := expected.(string)
		contains = ok && strings.Contains(a, e)

	case []string:
		e, ok := expected.(string)
		if ok {
			for _, v := range a {
				if v == e {
					contains = true
					break
				}
			}
		}

	case map[string][]string:
		e, ok := expected.(string)
		contains = ok
		if contains {
			_, contains = a[e]
		}

	default:
		// Fall back to equality if you don't want unsupported types
		// to silently pass.
		contains = reflect.DeepEqual(expected, actual)
	}

	if contains {
		return true
	}

	t.Fatalf(
		"\n ASSERT FAILED \n %s:%d: %s\nexpected to contain: %#v\nactual:              %#v",
		file,
		line,
		message,
		expected,
		actual,
	)

	return false
}

func AssertEqual(t *testing.T, expected, actual any, message string, file string, line int) bool {
	t.Helper()

	if reflect.DeepEqual(expected, actual) {
		return true
	}

	t.Fatalf(
		"\n ASSERT EQUALS FAILED \n %s:%d: %s\nexpected: %#v\nactual:   %#v",
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

	t.Fatalf(
		"\n ASSERT NOT NIL FAILED \n %s:%d: %s\nactual:   %#v",
		file,
		line,
		message,
		actual,
	)

	return false
}
