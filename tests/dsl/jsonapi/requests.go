package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"nighttroop.t1d.nz/utils/testutils"

	"nighttroop.t1d.nz/tests/dsl/script"
)

type RequestFile struct {
	Name     string
	Requests []*Request
}

type Set struct {
	Key   string
	Value string
}

func (r *Set) Set(ctx map[string]any) error {
	renderedValue, err := script.Render(r.Value, ctx)
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

	Asserts []Assert
	Sets    []Set

	// RunBackgroundJobs marks a RUN_BACKGROUND_JOBS step, which runs every background and cron job instead of a request
	RunBackgroundJobs bool
}

// Run runs the file's requests in order against handler, each seeing the values the ones before it SET in ctx,
// and calls runBackgroundJobs for each RUN_BACKGROUND_JOBS. Each request and its response are printed.
// Client sends a file's requests to the test server ts, whatever their host, as they are written: it follows no
// redirect, so a step sees the response it got, and asks for no compression, so the headers asserted are the server's
func Client(ts *httptest.Server) *http.Client {
	client := testutils.HostClient(ts)
	client.Transport.(*http.Transport).DisableCompression = true
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return client
}

// Run sends the file's requests in order with client, and runs runBackgroundJobs for each RUN_BACKGROUND_JOBS
func (f *RequestFile) Run(t *testing.T, client *http.Client, ctx map[string]any, runBackgroundJobs func()) {
	t.Helper()
	for _, request := range f.Requests {
		if request.RunBackgroundJobs {
			fmt.Println("=============== RUN_BACKGROUND_JOBS ================")
			runBackgroundJobs()
			continue
		}

		fmt.Printf("\n=============== %s  ================\n\n", request.Name)
		request.PrintRequest()
		var err error
		if ctx, err = request.ExecuteRequest(t, client, ctx); err != nil {
			t.Fatalf("failed to execute request: %v, method: %s, url: %s", err, request.Method, request.URL)
		}

		fmt.Printf("\nSTATUS CODE: %d\n", ctx[request.Name+"_status"])
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(ctx[request.Name]); err != nil {
			fmt.Printf("RESPONSE: %s and failed to format response: %v\n", ctx[request.Name], err)
		} else {
			fmt.Printf("RESPONSE:\n%s", buf.String())
		}
		fmt.Println("===============================")
	}
}

func (r *Request) ToFlatString() string {
	return fmt.Sprintf("NAME %s METHOD %s URL %s", r.Name, r.Method, r.URL)
}

func (r *Request) PrintRequest() {
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
	if r.Body != nil && len(*r.Body) > 0 {
		fmt.Printf("BODY %s\n", strings.ReplaceAll(*r.Body, "\n", " "))
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

func (r *Request) AddAssert(assert Assert) {
	r.Asserts = append(r.Asserts, assert)
}

func (r *Request) AddSet(set Set) {
	r.Sets = append(r.Sets, set)
}

func (r *Request) buildRealRequest(ctx map[string]any) (*Request, error) {
	url, err := script.Render(r.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("request %s url render failed: %w", r.Name, err)
	}

	// ---- render body ----
	bodyStr := ""
	if r.Body != nil {
		bodyStr, err = script.Render(string(*r.Body), ctx)
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
			renderedValue, err := script.Render(v, ctx)
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

	r, err = r.buildRealRequest(context)
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
		fmt.Println("failed to decode json for", r.Name, "error:", err)
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
