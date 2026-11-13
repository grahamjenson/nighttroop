package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type Set struct {
	Key   string
	Value string
}

func (r *Set) Set(ctx map[string]any) error {
	renderedValue, err := Render(r.Value, ctx)
	if err != nil {
		return fmt.Errorf("failed to render template for set %s: %v", r.Key, err)
	}
	ctx[r.Key] = renderedValue
	return nil
}

type Request struct {
	Name    string
	Method  string
	URL     string
	Headers http.Header
	Body    *string

	Asserts []HTTPAssert
	Sets    []Set
}

// Client sends a file's requests to the test server ts, whatever their host, as they are written: it follows no
// redirect, so a step sees the response it got, and asks for no compression, so the headers asserted are the server's
func Client(ts *httptest.Server) *http.Client {
	addr := ts.Listener.Addr().String()
	dial := func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	}
	client := &http.Client{Transport: &http.Transport{DialContext: dial, DisableCompression: true}}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return client
}

// Run sends the request with client, its values filled in from ctx, and keeps its response in ctx for the steps
// after it, failing the test if it cannot be sent. The request and its response are printed.
func (r *Request) Run(t *testing.T, client *http.Client, ctx map[string]any) {
	t.Helper()
	if sent := r.Describe(); sent != "" {
		logf(t, "%s", sent)
	}
	if _, err := r.ExecuteRequest(t, client, ctx); err != nil {
		t.Fatalf("failed to execute request: %v, method: %s, url: %s", err, r.Method, r.URL)
	}

	response, _ := MarshalJsonWithoutEscapeHTML(ctx[r.Name])
	logf(t, "  status %v: %s", ctx[r.Name+"_status"], response)
}

func (r *Request) ToFlatString() string {
	return fmt.Sprintf("NAME %s METHOD %s URL %s", r.Name, r.Method, r.URL)
}

// Describe is what the request sends, besides its method and URL, its Authorization left out
func (r *Request) Describe() string {
	out := ""
	if r.Body != nil && len(*r.Body) > 0 {
		out += "  body: " + strings.ReplaceAll(*r.Body, "\n", " ") + "\n"
	}
	for k, v := range r.Headers {
		if k == "Accept" || k == "Content-Type" {
			continue
		}
		if k == "Authorization" {
			v = []string{"<REDACTED>"}
		}
		out += fmt.Sprintf("  %s: %s\n", k, strings.Join(v, ", "))
	}
	return strings.TrimSuffix(out, "\n")
}

func (r *Request) AddAssert(assert HTTPAssert) {
	r.Asserts = append(r.Asserts, assert)
}

func (r *Request) AddSet(set Set) {
	r.Sets = append(r.Sets, set)
}

// Render is the request with its URL, body and headers filled in from ctx, sending JSON unless a header says not
func (r *Request) Render(ctx map[string]any) (*Request, error) {
	url, err := Render(r.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("request %s url render failed: %w", r.Name, err)
	}

	// ---- render body ----
	bodyStr := ""
	if r.Body != nil {
		bodyStr, err = Render(string(*r.Body), ctx)
		if err != nil {
			return nil, fmt.Errorf("request %s body render failed: %w", r.Name, err)
		}
	}

	// check body is valid JSON
	if bodyStr != "" {
		var js any
		if err := json.Unmarshal([]byte(bodyStr), &js); err != nil {
			return nil, fmt.Errorf("request %s body is not valid JSON: %w", r.Name, err)
		}
	}

	// -- render headers ----
	renderedHeaders := make(http.Header)
	for k, vv := range r.Headers {
		for _, v := range vv {
			renderedValue, err := Render(v, ctx)
			if err != nil {
				return nil, fmt.Errorf("request %s header %s render failed: %w", r.Name, k, err)
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

	return &Request{
		Name:    r.Name,
		Method:  r.Method,
		URL:     url,
		Headers: renderedHeaders,
		Body:    &bodyStr,
		Asserts: r.Asserts,
		Sets:    r.Sets,
	}, nil
}

func (r *Request) ExecuteRequest(t *testing.T, client *http.Client, context map[string]any) (outContext map[string]any, err error) {

	r, err = r.Render(context)
	if err != nil {
		return context, err
	}

	rawBody := ""
	if r.Body != nil {
		rawBody = *r.Body
	}

	req, err := http.NewRequestWithContext(t.Context(), r.Method, r.URL, strings.NewReader(rawBody))
	if err != nil {
		return context, fmt.Errorf("request %s is not a request: %w", r.Name, err)
	}
	for k, vv := range r.Headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	res, err := client.Do(req)
	if err != nil {
		return context, fmt.Errorf("request %s failed: %w", r.Name, err)
	}
	defer res.Body.Close()
	sent, err := io.ReadAll(res.Body)
	if err != nil {
		return context, fmt.Errorf("request %s response could not be read: %w", r.Name, err)
	}

	var out any
	bodyOut := string(sent)
	if err := json.NewDecoder(strings.NewReader(bodyOut)).Decode(&out); err != nil {
		logf(t, "  %s's response is not JSON: %v", r.Name, err)
	}
	if err != nil {
		return context, fmt.Errorf("request %s response body is not valid JSON: %v", r.Name, err)
	}

	headers := res.Header
	context[r.Name] = out
	context[r.Name+"_status"] = res.StatusCode
	context[r.Name+"_headers"] = headers

	// EXECUTE THE SETS
	for _, setStruct := range r.Sets {
		err := setStruct.Set(context)
		if err != nil {
			return context, fmt.Errorf("failed to execute set %s: %v", setStruct.Key, err)
		}
	}

	// EXECUTE THE ASSERTS
	for _, assert := range r.Asserts {
		assert.Execute(t, r.Name, res.StatusCode, out, headers, context)
	}

	return context, nil
}
