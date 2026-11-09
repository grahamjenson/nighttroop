package httputils

import (
	"cmp"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/aslog"
)

// maxCapturedBody is the most of a request's or response's body kept, so a captured request fits in one log
const maxCapturedBody = 64 * 1024

// redacted replaces each secret captured, so the logs hold no credentials
const redacted = "REDACTED"

// CaptureRequests logs each request and its response, with their headers and bodies, in the logs DB, along with the
// site and identity the request was for. Credentials in headers, the query and bodies are redacted. A long lived
// request, as an event stream or Socket.IO connection, is captured without its bodies.
func CaptureRequests(logs *db.LogsDB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started, longLived := time.Now(), LongLived(r)
			request := &cappedBuffer{}
			if !longLived && r.Body != nil {
				r.Body = struct {
					io.Reader
					io.Closer
				}{io.TeeReader(r.Body, request), r.Body}
			}
			ww, response := middleware.NewWrapResponseWriter(w, r.ProtoMajor), &cappedBuffer{}
			if !longLived {
				ww.Tee(response)
			}
			next.ServeHTTP(ww, r)
			if asset(ww.Header().Get("Content-Type")) {
				return
			}

			path := redactPath(r)
			attributes := map[string]any{"method": r.Method, "host": r.Host, "path": path, "query": redactQuery(r.URL.Query()),
				"remote_ip": GetIPFromHTTPContext(r.Context()), "request_headers": redactHeaders(r.Header),
				"request_body": body(request, r.Header.Get("Content-Type")), "status": ww.Status(),
				"response_headers": redactHeaders(ww.Header()), "response_body": body(response, ww.Header().Get("Content-Type")),
				"duration_ms": time.Since(started).Milliseconds()}
			var siteID *string
			for _, attr := range aslog.Attrs(r.Context()) {
				if attr.Key == "site_id" {
					siteID = new(attr.Value.String())
				} else if attr.Key != "request_id" {
					attributes[attr.Key] = attr.Value.Any()
				}
			}
			logs.Capture(context.WithoutCancel(r.Context()), siteID, "Captured "+r.Method+" "+path, attributes)
		})
	}
}

// asset is whether a response is a page's style, script, image or font, which no client sends data with
func asset(contentType string) bool {
	for _, kind := range []string{"text/css", "javascript", "image/", "font/"} {
		if strings.Contains(contentType, kind) {
			return true
		}
	}
	return false
}

// redactPath is the request's path with the values of its route's credentials, such as Nightscout's
// /api/v2/authorization/request/{token}, redacted
func redactPath(r *http.Request) string {
	path, routeContext := r.URL.Path, chi.RouteContext(r.Context())
	if routeContext == nil {
		return path
	}
	for i, key := range routeContext.URLParams.Keys {
		if value := routeContext.URLParams.Values[i]; secret(key) && value != "" {
			path = strings.ReplaceAll(path, value, redacted)
		}
	}
	return path
}

// cappedBuffer keeps the first limit bytes written to it, maxCapturedBody when it has none, and whether there were more
type cappedBuffer struct {
	bytes.Buffer
	limit     int
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if room := cmp.Or(b.limit, maxCapturedBody) - b.Len(); len(p) > room {
		b.truncated = true
		b.Buffer.Write(p[:max(room, 0)])
		return len(p), nil
	}
	return b.Buffer.Write(p)
}

// body is the captured body with its credentials redacted: as text, or base64 when it is not text, such as gzip
func body(b *cappedBuffer, contentType string) map[string]any {
	if b.Len() == 0 {
		return nil
	}
	raw, captured := b.Bytes(), map[string]any{"truncated": b.truncated}
	switch {
	case strings.Contains(contentType, "json") && !b.truncated:
		var value any
		if json.Unmarshal(raw, &value) == nil {
			raw, _ = json.Marshal(redactJSON(value))
		}
	case strings.Contains(contentType, "x-www-form-urlencoded") && !b.truncated:
		if form, err := url.ParseQuery(string(raw)); err == nil {
			raw = []byte(url.Values(redactQuery(form)).Encode())
		}
	}
	if utf8.Valid(raw) {
		captured["text"] = string(raw)
	} else {
		captured["base64"] = base64.StdEncoding.EncodeToString(raw)
	}
	return captured
}

// secret is whether a header, query, form or JSON key names a credential
func secret(key string) bool {
	key = strings.ToLower(key)
	for _, word := range []string{"authorization", "cookie", "secret", "token", "password", "jwt"} {
		if strings.Contains(key, word) {
			return true
		}
	}
	return false
}

func redactHeaders(header http.Header) map[string][]string {
	redactedHeader := map[string][]string{}
	for key, values := range header {
		if secret(key) {
			values = []string{redacted}
		}
		redactedHeader[key] = values
	}
	return redactedHeader
}

func redactQuery(query url.Values) map[string][]string {
	return redactHeaders(http.Header(query))
}

// redactJSON is value with every value of a key naming a credential redacted, however deep
func redactJSON(value any) any {
	switch v := value.(type) {
	case map[string]any:
		for key, inner := range v {
			if secret(key) {
				v[key] = redacted
			} else {
				v[key] = redactJSON(inner)
			}
		}
	case []any:
		for i, inner := range v {
			v[i] = redactJSON(inner)
		}
	}
	return value
}
