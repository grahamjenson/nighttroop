package httputils

import (
	"net/http"
	"strings"
)

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
