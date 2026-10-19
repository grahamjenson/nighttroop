package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"text/template"
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
	renderedValue, err := renderTemplate(r.Value, ctx)
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
	url, err := renderTemplate(r.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("request %s url render failed: %w", r.Name, err)
	}

	// ---- render body ----
	bodyStr := ""
	if r.Body != nil {
		bodyStr, err = renderTemplate(string(*r.Body), ctx)
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
			renderedValue, err := renderTemplate(v, ctx)
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

func (r *Request) ExecuteRequest(t *testing.T, handler http.Handler, context map[string]any) (outContext map[string]any, err error) {

	r, err = r.buildRealRequest(context)
	if err != nil {
		return context, err
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("httptest.NewRequest panicked: %v", recovered)
		}
	}()

	rawBody := ""
	if r.Body != nil {
		rawBody = *r.Body
	}

	req := httptest.NewRequest(r.Method, r.URL, strings.NewReader(rawBody))
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
		return context, fmt.Errorf("request %s response body is not valid JSON: %v", r.Name, err)
	}

	context[r.Name] = out
	context[r.Name+"_status"] = rr.Code
	context[r.Name+"_headers"] = rr.Header()

	// EXECUTE THE SETS
	for _, setStruct := range r.Sets {
		err := setStruct.Set(context)
		if err != nil {
			return context, fmt.Errorf("failed to execute set %s: %v", setStruct.Key, err)
		}
	}

	// EXECUTE THE ASSERTS
	for _, assert := range r.Asserts {
		assert.Execute(t, r.Name, rr.Code, out, rr.Header(), context)
	}

	return context, nil
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
