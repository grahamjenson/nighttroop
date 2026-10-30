// Package docs is the terms of service and privacy policy, as HTML generated from their Markdown by gen/docs
package docs

import (
	"html"
	"strings"
)

// WithSupportEmail is a document's HTML with the support email address in place of SUPPORT_EMAIL
func WithSupportEmail(document string, email string) string {
	return strings.ReplaceAll(document, "SUPPORT_EMAIL", html.EscapeString(email))
}
