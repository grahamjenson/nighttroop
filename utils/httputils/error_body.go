package httputils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/getsentry/sentry-go"
)

// errorRequestBodyBytes is how much of a request's body is kept to debug it by, should it fail
const errorRequestBodyBytes = 8 * 1024

type errorRequestBodyKey struct{}

// errorRequestBody is a request's body as it was read, kept to debug the request by should it fail
type errorRequestBody struct {
	buffer      cappedBuffer
	contentType string
}

// KeepErrorRequestBodies keeps the first errorRequestBodyBytes of each request's body as its handler reads it, but for
// a long lived request's, so a request that fails can be recorded and reported with what it sent
func KeepErrorRequestBodies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if LongLived(r) || r.Body == nil || r.Body == http.NoBody {
			next.ServeHTTP(w, r)
			return
		}
		kept := &errorRequestBody{buffer: cappedBuffer{limit: errorRequestBodyBytes}, contentType: r.Header.Get("Content-Type")}
		r.Body = struct {
			io.Reader
			io.Closer
		}{io.TeeReader(r.Body, &kept.buffer), r.Body}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), errorRequestBodyKey{}, kept)))
	})
}

// ErrorRequestBody is the body of the request ctx is for, as KeepErrorRequestBodies kept it, with its credentials
// redacted: its JSON, a form, or its text, such as JSON that did not parse; nil when it had none
func ErrorRequestBody(ctx context.Context) map[string]any {
	kept, _ := ctx.Value(errorRequestBodyKey{}).(*errorRequestBody)
	if kept == nil {
		return nil
	}
	contentType := kept.contentType
	// A client may send JSON without saying so; it is redacted as JSON all the same
	if json.Valid(kept.buffer.Bytes()) {
		contentType = "application/json"
	}
	return body(&kept.buffer, contentType)
}

// reportErrorRequestBodies sends the body of a request whose handler reports an error to Sentry with it, under
// request_body, as it was read by the time of the error
func reportErrorRequestBodies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
			ctx := r.Context()
			hub.Scope().AddEventProcessor(func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
				if requestBody := ErrorRequestBody(ctx); requestBody != nil {
					if event.Contexts == nil {
						event.Contexts = map[string]sentry.Context{}
					}
					event.Contexts["request_body"] = requestBody
				}
				return event
			})
		}
		next.ServeHTTP(w, r)
	})
}
