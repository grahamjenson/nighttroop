package httputils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// WithoutURL is the error of an http.Client's request without its URL, which can hold a secret, such as a session in
// its query or a device token in its path, or a user's own Nightscout. What went wrong is kept, with the URL's host
// left out of it too, as a DNS error names it, and errors.Is and errors.As still find its cause.
func WithoutURL(err error) error {
	urlErr, ok := errors.AsType[*url.Error](err)
	if !ok {
		return err
	}
	message := urlErr.Err.Error()
	if u, parseErr := url.Parse(urlErr.URL); parseErr == nil && u.Hostname() != "" {
		message = strings.ReplaceAll(message, u.Hostname(), "the host")
	}
	return &withoutURLError{message: urlErr.Op + ": " + message, err: urlErr.Err}
}

type withoutURLError struct {
	message string
	err     error
}

func (e *withoutURLError) Error() string { return e.message }
func (e *withoutURLError) Unwrap() error { return e.err }

// ReadLimitedBody reads an upstream response body of at most max bytes and fails
// when it is longer, so a large (or gzip-inflated) body cannot exhaust memory.
func ReadLimitedBody(r io.Reader, max int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > max {
		return nil, fmt.Errorf("response body is larger than %d bytes", max)
	}
	return body, nil
}

// LimitRequestBody caps how much of r's body handlers can read, so a huge upload cannot exhaust memory while it is
// decoded. A body that says it is larger than max is refused with 413; one that does not is cut off at max, failing
// its decode. It is false when the request was refused.
func LimitRequestBody(w http.ResponseWriter, r *http.Request, max int64) bool {
	if r.ContentLength > max {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, max)
	return true
}

// ////
// Accept Type
// ////

type AcceptType int

const (
	AcceptTypeHTML AcceptType = iota
	AcceptTypeJSON
	AcceptTypeHTMX
)

func GetAcceptType(r *http.Request) AcceptType {
	accept := r.Header.Get("Accept")
	hxRequest := r.Header.Get("HX-Request")

	switch {
	case strings.Contains(accept, "application/json"):
		return AcceptTypeJSON
	// The browser's back and forward buttons swap the whole body, so they need the full page
	case r.Header.Get("HX-History-Restore-Request") == "true":
		return AcceptTypeHTML
	case strings.Contains(accept, "text/htmx") || hxRequest == "true":
		return AcceptTypeHTMX
	case strings.Contains(accept, "text/html"):
		return AcceptTypeHTML
	default:
		// If no Accept header use ContentType
		contentType := GetContentType(r)
		switch contentType {
		case ContentTypeForm:
			return AcceptTypeHTML
		case ContentTypeJSON:
			return AcceptTypeJSON

		default:
			return AcceptTypeJSON
		}
	}
}

func (at AcceptType) API() bool {
	return at == AcceptTypeJSON
}

// RedirectToLogin sends a browser request that is not signed in to path. A page is sent there with a 303. An htmx
// request or an event stream is answered 401 with HX-Redirect instead: fetch follows a 303 without showing it, so htmx
// never saw the redirect, and a stream reconnecting got the login page as a working response and asked again at once.
func RedirectToLogin(w http.ResponseWriter, r *http.Request, path string) {
	w.Header().Set("HX-Redirect", path)
	if GetAcceptType(r) == AcceptTypeHTMX || strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}

// ////
// Content Type
// ////

type ContentType string

const (
	ContentTypeForm ContentType = "application/x-www-form-urlencoded"
	ContentTypeJSON ContentType = "application/json"
)

func GetContentType(r *http.Request) ContentType {
	contentType := r.Header.Get("Content-Type")
	// Defaults to JSON
	switch {
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return ContentTypeForm
	case strings.Contains(contentType, "application/json"):
		return ContentTypeJSON
	default:
		return ContentTypeJSON
	}
}

// ////
// HTTPAction
// ////

type HTTPAction int

const (
	HTTPActionRead HTTPAction = iota
	HTTPActionWrite
)

func GetHTTPAction(r *http.Request) HTTPAction {
	switch strings.ToUpper(r.Method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return HTTPActionRead
	default:
		return HTTPActionWrite
	}
}

//////
// Writer
//////

// THIS IS SO STUPID THAT I HAVE TO DO THIS TO CHECK THE STATUS OF THE RETURNED RESPONSE.
type StatusResponseWriter struct {
	http.ResponseWriter
	Status int
}

func (w *StatusResponseWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}

// Unwrap lets http.ResponseController reach the writer underneath, to flush a stream and lift its write deadline
func (w *StatusResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
