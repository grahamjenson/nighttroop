// ///
// Copied from https://github.com/samber/slog-chi
package httputils

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	TraceIDKey = "trace_id"
	SpanIDKey  = "span_id"

	RequestBodyMaxSize  = 64 * 1024 // 64KB
	ResponseBodyMaxSize = 64 * 1024 // 64KB

	HiddenRequestHeaders = map[string]struct{}{
		"api-secret":    {}, // the Nightscout API takes a token in it
		"authorization": {},
		"cookie":        {},
		"set-cookie":    {},
		"x-auth-token":  {},
		"x-csrf-token":  {},
		"x-xsrf-token":  {},
	}
	HiddenResponseHeaders = map[string]struct{}{
		"set-cookie": {},
	}
)

type Config struct {
	DefaultLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level

	WithUserAgent      bool
	WithRequestBody    bool
	WithRequestHeader  bool
	WithResponseBody   bool
	WithResponseHeader bool
	WithSpanID         bool
	WithTraceID        bool
	WithClientIP       bool
	WithCustomMessage  func(w http.ResponseWriter, r *http.Request) string
}

// credentialParams are query and route params that carry tokens; the Nightscout API takes a token as secret too, and
// Engine.IO's sid is all a Socket.IO session's polls and posts carry once it is authorized
var credentialParams = []string{"token", "secret", "sid"}

// authorizationRequestPath is where a Nightscout authorization request's path ends in its token
const authorizationRequestPath = "/authorization/request/"

// redactCredentials keeps tokens out of logs, and so out of Sentry
func redactCredentials(path, query string, params map[string]string) (string, string) {
	for _, k := range credentialParams {
		if v := params[k]; v != "" {
			path = strings.ReplaceAll(path, v, "REDACTED")
			params[k] = "REDACTED"
		}
	}
	return path, RedactQuery(query)
}

// RedactQuery replaces credential values in a raw query string
func RedactQuery(query string) string {
	values, _ := url.ParseQuery(query)
	for _, k := range credentialParams {
		if values.Has(k) {
			values.Set(k, "REDACTED")
		}
	}
	return values.Encode()
}

// RedactSentryEvent keeps tokens out of what is sent to Sentry: its request's query, its hidden headers, which Sentry
// only partly leaves out, and the token a Nightscout authorization request's path ends in
func RedactSentryEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	if event.Request == nil {
		return event
	}
	event.Request.QueryString = RedactQuery(event.Request.QueryString)
	for k := range event.Request.Headers {
		if _, hidden := HiddenRequestHeaders[strings.ToLower(k)]; hidden {
			delete(event.Request.Headers, k)
		}
	}
	if before, _, found := strings.Cut(event.Request.URL, authorizationRequestPath); found {
		event.Request.URL = before + authorizationRequestPath + "REDACTED"
	}
	return event
}

// New returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
// A nil logger logs to slog.Default() as it is when each request is logged.
func NewLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return NewLoggerWithConfig(logger, DefaultConfig())
}

// DefaultConfig returns the default configuration for the request logger.
func DefaultConfig() Config {
	return Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		WithUserAgent:      false,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,
		WithSpanID:         false,
		WithTraceID:        false,
		WithClientIP:       true,
		WithCustomMessage:  nil,
	}
}

// NewLoggerWithConfig returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
func NewLoggerWithConfig(logger *slog.Logger, config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			query := r.URL.RawQuery

			// dump request body
			br := newBodyReader(r.Body, RequestBodyMaxSize, config.WithRequestBody)
			r.Body = br

			// dump response body
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			var bw *bodyWriter
			if config.WithResponseBody {
				bw = newBodyWriter(ResponseBodyMaxSize)
				ww.Tee(bw)
			}

			defer func() {
				params := routeParams(r)
				path, query := redactCredentials(path, query, params)

				status := ww.Status()
				attributes := []slog.Attr{
					{Key: "request", Value: slog.GroupValue(config.requestAttributes(r, start, path, query, params, br)...)},
					{Key: "response", Value: slog.GroupValue(config.responseAttributes(ww, start, time.Now(), bw)...)},
				}
				level := config.level(status)

				msg := strconv.Itoa(status) + ": " + http.StatusText(status)
				if config.WithCustomMessage != nil {
					msg = config.WithCustomMessage(ww, r)
				}

				requestLogger := logger
				if requestLogger == nil {
					requestLogger = slog.Default()
				}
				requestLogger.LogAttrs(r.Context(), level, msg, attributes...)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// requestAttributes are what is logged of a request: where it went, from start, and its body's length, then its client
// IP, body, headers and user agent when config asks for them. path, query and params have had credentials redacted.
// routeParams is the request's chi route params by name
func routeParams(r *http.Request) map[string]string {
	params := map[string]string{}
	routeContext := chi.RouteContext(r.Context())
	for i, k := range routeContext.URLParams.Keys {
		params[k] = routeContext.URLParams.Values[i]
	}
	return params
}

// level is the level a response with status is logged at
func (config Config) level(status int) slog.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return config.ServerErrorLevel
	case status >= http.StatusBadRequest:
		return config.ClientErrorLevel
	}
	return config.DefaultLevel
}

func (config Config) requestAttributes(r *http.Request, start time.Time, path, query string, params map[string]string, br *bodyReader) []slog.Attr {
	attributes := []slog.Attr{
		slog.Time("time", start.UTC()),
		slog.String("method", r.Method),
		slog.String("host", r.Host),
		slog.String("path", path),
		slog.String("query", query),
		slog.Any("params", params),
		slog.String("route", chi.RouteContext(r.Context()).RoutePattern()),
		slog.String("referer", r.Referer()),
	}
	if config.WithClientIP {
		attributes = append(attributes, slog.String("ip", r.RemoteAddr))
	}
	attributes = append(attributes, slog.Int("length", br.bytes))
	if config.WithRequestBody {
		attributes = append(attributes, slog.String("body", br.body.String()))
	}
	if config.WithRequestHeader {
		attributes = append(attributes, headerGroup(r.Header, HiddenRequestHeaders))
	}
	if config.WithUserAgent {
		attributes = append(attributes, slog.String("user-agent", r.UserAgent()))
	}
	return attributes
}

// responseAttributes are what is logged of a response: when it ended, how long after start, its status and its
// body's length, then its body and headers when config asks for them
func (config Config) responseAttributes(ww middleware.WrapResponseWriter, start, end time.Time, bw *bodyWriter) []slog.Attr {
	attributes := []slog.Attr{
		slog.Time("time", end.UTC()),
		slog.Duration("latency", end.Sub(start)),
		slog.Int("status", ww.Status()),
		slog.Int("length", ww.BytesWritten()),
	}
	if config.WithResponseBody {
		attributes = append(attributes, slog.String("body", bw.body.String()))
	}
	if config.WithResponseHeader {
		attributes = append(attributes, headerGroup(ww.Header(), HiddenResponseHeaders))
	}
	return attributes
}

// headerGroup is the headers as a log group, leaving out the hidden ones, which carry credentials
func headerGroup(header http.Header, hidden map[string]struct{}) slog.Attr {
	kv := []any{}
	for k, v := range header {
		if _, found := hidden[strings.ToLower(k)]; !found {
			kv = append(kv, slog.Any(k, v))
		}
	}
	return slog.Group("header", kv...)
}

/////
// dump.go
/////

type bodyWriter struct {
	body    *bytes.Buffer
	maxSize int
}

// implements io.Writer
func (w *bodyWriter) Write(b []byte) (int, error) {
	length := len(b)

	if w.body.Len()+length > w.maxSize {
		w.body.Truncate(min(w.maxSize, length, w.body.Len()))
		return w.body.Write(b[:min(w.maxSize-w.body.Len(), length)])
	}
	return w.body.Write(b)
}

func newBodyWriter(maxSize int) *bodyWriter {
	return &bodyWriter{
		body:    bytes.NewBufferString(""),
		maxSize: maxSize,
	}
}

type bodyReader struct {
	io.ReadCloser
	body    *bytes.Buffer
	maxSize int
	bytes   int
}

// implements io.Reader
func (r *bodyReader) Read(b []byte) (int, error) {
	n, err := r.ReadCloser.Read(b)
	if r.body != nil && r.body.Len() < r.maxSize {
		if r.body.Len()+n > r.maxSize {
			r.body.Write(b[:min(r.maxSize-r.body.Len(), n)])
		} else {
			r.body.Write(b[:n])
		}
	}
	r.bytes += n
	return n, err
}

func newBodyReader(reader io.ReadCloser, maxSize int, recordBody bool) *bodyReader {
	var body *bytes.Buffer
	if recordBody {
		body = bytes.NewBufferString("")
	}
	return &bodyReader{
		ReadCloser: reader,
		body:       body,
		maxSize:    maxSize,
		bytes:      0,
	}
}
