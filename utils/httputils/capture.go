package httputils

import (
	"cmp"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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
// site and identity the request was for. Credentials in headers, the query and bodies are redacted, and a body that
// cannot be, as a page, is left out. A long lived request, as an event stream or Socket.IO connection, is captured
// without its bodies.
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

// cappedBuffer keeps the first limit bytes written to it, maxCapturedBody when it has none, whether there were more,
// and how many there were
type cappedBuffer struct {
	bytes.Buffer
	limit     int
	truncated bool
	size      int
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	b.size += len(p)
	if room := cmp.Or(b.limit, maxCapturedBody) - b.Len(); len(p) > room {
		b.truncated = true
		b.Buffer.Write(p[:max(room, 0)])
		return len(p), nil
	}
	return b.Buffer.Write(p)
}

// body is the captured body with its credentials redacted: its JSON, whatever its content type says, or its form. A
// body that cannot be parsed and redacted, as one cut short, a page or text, is left out, saying only its size and
// content type, so no credential or personal detail in it is kept.
func body(b *cappedBuffer, contentType string) map[string]any {
	if b.Len() == 0 {
		return nil
	}
	if !b.truncated {
		var value any
		if json.Unmarshal(b.Bytes(), &value) == nil {
			redactedJSON, _ := json.Marshal(redactJSON(value))
			return map[string]any{"truncated": false, "text": string(redactedJSON)}
		}
		if form, err := url.ParseQuery(b.String()); err == nil && strings.Contains(contentType, "x-www-form-urlencoded") {
			return map[string]any{"truncated": false, "text": url.Values(redactQuery(form)).Encode()}
		}
	}
	return map[string]any{"truncated": b.truncated, "omitted": true, "bytes": b.size, "content_type": contentType}
}

// secret is whether a header, query, form or JSON key names a credential or a connector's identity, such as a Dexcom
// username or a Nightscout URL in url_or_username, a backup or API key or a TOTP code
func secret(key string) bool {
	key = strings.ToLower(key)
	for _, word := range []string{"auth", "cookie", "secret", "token", "pass", "pwd", "jwt", "key", "totp", "username",
		"email", "url", "referer", "session", "credential"} {
		if strings.Contains(key, word) {
			return true
		}
	}
	// A TOTP code, or a Libre sensor's, but not a key only ending in code, as encode, and an Engine.IO session's ID
	return key == "code" || strings.HasSuffix(key, "_code") || key == "sid"
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
