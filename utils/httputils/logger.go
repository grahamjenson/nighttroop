// ///
// Copied from https://github.com/samber/slog-chi
package httputils

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type customAttributesCtxKeyType struct{}

var customAttributesCtxKey = customAttributesCtxKeyType{}

var (
	TraceIDKey   = "trace_id"
	SpanIDKey    = "span_id"
	RequestIDKey = "id"

	RequestBodyMaxSize  = 64 * 1024 // 64KB
	ResponseBodyMaxSize = 64 * 1024 // 64KB

	HiddenRequestHeaders = map[string]struct{}{
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
	WithRequestID      bool
	WithRequestBody    bool
	WithRequestHeader  bool
	WithResponseBody   bool
	WithResponseHeader bool
	WithSpanID         bool
	WithTraceID        bool
	WithClientIP       bool
	WithCustomMessage  func(w http.ResponseWriter, r *http.Request) string

	Filters []Filter
}

// credentialParams are query and route params that carry tokens or secrets
var credentialParams = []string{"token", "secret"}

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

// New returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func NewLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return NewLoggerWithConfig(logger, DefaultConfig())
}

// NewWithFilters returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func NewLoggerWithFilters(logger *slog.Logger, filters ...Filter) func(http.Handler) http.Handler {
	config := DefaultConfig()
	config.Filters = filters
	return NewLoggerWithConfig(logger, config)
}

// DefaultConfig returns the default configuration for the request logger.
func DefaultConfig() Config {
	return Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		WithUserAgent:      false,
		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,
		WithSpanID:         false,
		WithTraceID:        false,
		WithClientIP:       true,
		WithCustomMessage:  nil,

		Filters: []Filter{},
	}
}

// NewWithConfig returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
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

			// Make sure we create a map only once per request (in case we have multiple middleware instances)
			if v := r.Context().Value(customAttributesCtxKey); v == nil {
				r = r.WithContext(context.WithValue(r.Context(), customAttributesCtxKey, &sync.Map{}))
			}

			defer func() {
				// Pass thru filters and skip early the code below, to prevent unnecessary processing.
				for _, filter := range config.Filters {
					if !filter(ww, r) {
						return
					}
				}

				params := map[string]string{}
				for i, k := range chi.RouteContext(r.Context()).URLParams.Keys {
					params[k] = chi.RouteContext(r.Context()).URLParams.Values[i]
				}
				path, query := redactCredentials(path, query, params)

				status := ww.Status()
				method := r.Method
				host := r.Host
				route := chi.RouteContext(r.Context()).RoutePattern()
				end := time.Now()
				latency := end.Sub(start)
				userAgent := r.UserAgent()
				ip := r.RemoteAddr
				referer := r.Referer()

				baseAttributes := make([]slog.Attr, 0, 3)
				requestAttributes := make([]slog.Attr, 0, 13)
				responseAttributes := make([]slog.Attr, 0, 6)

				requestAttributes = append(requestAttributes,
					slog.Time("time", start.UTC()),
					slog.String("method", method),
					slog.String("host", host),
					slog.String("path", path),
					slog.String("query", query),
					slog.Any("params", params),
					slog.String("route", route),
					slog.String("referer", referer),
				)

				if config.WithClientIP {
					requestAttributes = append(requestAttributes, slog.String("ip", ip))
				}

				responseAttributes = append(responseAttributes,
					slog.Time("time", end.UTC()),
					slog.Duration("latency", latency),
					slog.Int("status", status),
				)

				if config.WithRequestID {
					baseAttributes = append(baseAttributes, slog.String(RequestIDKey, middleware.GetReqID(r.Context())))
				}

				// request body
				requestAttributes = append(requestAttributes, slog.Int("length", br.bytes))
				if config.WithRequestBody {
					requestAttributes = append(requestAttributes, slog.String("body", br.body.String()))
				}

				// request headers
				if config.WithRequestHeader {
					kv := []any{}

					for k, v := range r.Header {
						if _, found := HiddenRequestHeaders[strings.ToLower(k)]; found {
							continue
						}
						kv = append(kv, slog.Any(k, v))
					}

					requestAttributes = append(requestAttributes, slog.Group("header", kv...))
				}

				if config.WithUserAgent {
					requestAttributes = append(requestAttributes, slog.String("user-agent", userAgent))
				}

				// response body
				responseAttributes = append(responseAttributes, slog.Int("length", ww.BytesWritten()))
				if config.WithResponseBody {
					responseAttributes = append(responseAttributes, slog.String("body", bw.body.String()))
				}

				// response headers
				if config.WithResponseHeader {
					kv := []any{}

					for k, v := range w.Header() {
						if _, found := HiddenResponseHeaders[strings.ToLower(k)]; found {
							continue
						}
						kv = append(kv, slog.Any(k, v))
					}

					responseAttributes = append(responseAttributes, slog.Group("header", kv...))
				}

				attributes := append(
					[]slog.Attr{
						{
							Key:   "request",
							Value: slog.GroupValue(requestAttributes...),
						},
						{
							Key:   "response",
							Value: slog.GroupValue(responseAttributes...),
						},
					},
					baseAttributes...,
				)

				// custom context values
				if v := r.Context().Value(customAttributesCtxKey); v != nil {
					if m, ok := v.(*sync.Map); ok {
						m.Range(func(key, value any) bool {
							attributes = append(attributes, slog.Attr{Key: key.(string), Value: value.(slog.Value)})
							return true
						})
					}
				}

				level := config.DefaultLevel
				if status >= http.StatusInternalServerError {
					level = config.ServerErrorLevel
				} else if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
					level = config.ClientErrorLevel
				}

				msg := strconv.Itoa(status) + ": " + http.StatusText(status)
				if config.WithCustomMessage != nil {
					msg = config.WithCustomMessage(ww, r)
				}

				logger.LogAttrs(r.Context(), level, msg, attributes...)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// AddCustomAttributes adds custom attributes to the request context. This func can be called from any handler or middleware, as long as the slog-chi middleware is already mounted.
func AddCustomAttributes(r *http.Request, attrs ...slog.Attr) {
	AddContextAttributes(r.Context(), attrs...)
}

// AddContextAttributes is the same as AddCustomAttributes, but it doesn't need access to the request struct.
func AddContextAttributes(ctx context.Context, attrs ...slog.Attr) {
	if v := ctx.Value(customAttributesCtxKey); v != nil {
		if m, ok := v.(*sync.Map); ok {
			for _, attr := range attrs {
				m.Store(attr.Key, attr.Value)
			}
		}
	}
}

/////
// filter.go
////

type Filter func(ww middleware.WrapResponseWriter, r *http.Request) bool

// Basic
func Accept(filter Filter) Filter { return filter }
func Ignore(filter Filter) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool { return !filter(ww, r) }
}

// Method
func AcceptMethod(methods ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		reqMethod := strings.ToLower(r.Method)

		for _, method := range methods {
			if strings.ToLower(method) == reqMethod {
				return true
			}
		}

		return false
	}
}

func IgnoreMethod(methods ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		reqMethod := strings.ToLower(r.Method)

		for _, method := range methods {
			if strings.ToLower(method) == reqMethod {
				return false
			}
		}

		return true
	}
}

// Status
func AcceptStatus(statuses ...int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return slices.Contains(statuses, ww.Status())
	}
}

func IgnoreStatus(statuses ...int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return !slices.Contains(statuses, ww.Status())
	}
}

func AcceptStatusGreaterThan(status int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return ww.Status() > status
	}
}

func AcceptStatusGreaterThanOrEqual(status int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return ww.Status() >= status
	}
}

func AcceptStatusLessThan(status int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return ww.Status() < status
	}
}

func AcceptStatusLessThanOrEqual(status int) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return ww.Status() <= status
	}
}

func IgnoreStatusGreaterThan(status int) Filter {
	return AcceptStatusLessThanOrEqual(status)
}

func IgnoreStatusGreaterThanOrEqual(status int) Filter {
	return AcceptStatusLessThan(status)
}

func IgnoreStatusLessThan(status int) Filter {
	return AcceptStatusGreaterThanOrEqual(status)
}

func IgnoreStatusLessThanOrEqual(status int) Filter {
	return AcceptStatusGreaterThan(status)
}

// Path
func AcceptPath(urls ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return slices.Contains(urls, r.URL.Path)
	}
}

func IgnorePath(urls ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return !slices.Contains(urls, r.URL.Path)
	}
}

func AcceptPathContains(parts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, part := range parts {
			if strings.Contains(r.URL.Path, part) {
				return true
			}
		}

		return false
	}
}

func IgnorePathContains(parts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, part := range parts {
			if strings.Contains(r.URL.Path, part) {
				return false
			}
		}

		return true
	}
}

func AcceptPathPrefix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return true
			}
		}

		return false
	}
}

func IgnorePathPrefix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return false
			}
		}

		return true
	}
}

func AcceptPathSuffix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return true
			}
		}

		return false
	}
}

func IgnorePathSuffix(suffixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, suffix := range suffixs {
			if strings.HasSuffix(r.URL.Path, suffix) {
				return false
			}
		}

		return true
	}
}

func AcceptPathMatch(regs ...regexp.Regexp) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Path)) {
				return true
			}
		}

		return false
	}
}

func IgnorePathMatch(regs ...regexp.Regexp) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Path)) {
				return false
			}
		}

		return true
	}
}

// Host
func AcceptHost(hosts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return slices.Contains(hosts, r.URL.Host)
	}
}

func IgnoreHost(hosts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		return !slices.Contains(hosts, r.URL.Host)
	}
}

func AcceptHostContains(parts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, part := range parts {
			if strings.Contains(r.URL.Host, part) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostContains(parts ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, part := range parts {
			if strings.Contains(r.URL.Host, part) {
				return false
			}
		}

		return true
	}
}

func AcceptHostPrefix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Host, prefix) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostPrefix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Host, prefix) {
				return false
			}
		}

		return true
	}
}

func AcceptHostSuffix(prefixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, prefix := range prefixs {
			if strings.HasPrefix(r.URL.Host, prefix) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostSuffix(suffixs ...string) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, suffix := range suffixs {
			if strings.HasSuffix(r.URL.Host, suffix) {
				return false
			}
		}

		return true
	}
}

func AcceptHostMatch(regs ...regexp.Regexp) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Host)) {
				return true
			}
		}

		return false
	}
}

func IgnoreHostMatch(regs ...regexp.Regexp) Filter {
	return func(ww middleware.WrapResponseWriter, r *http.Request) bool {
		for _, reg := range regs {
			if reg.Match([]byte(r.URL.Host)) {
				return false
			}
		}

		return true
	}
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
