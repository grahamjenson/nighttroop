// Package docs is the terms of service and privacy policy, as HTML generated from their Markdown by gen/docs
package docs

import (
	"html"
	"strings"
)

// WithOperator is a document's HTML with who runs the server in place of OPERATOR_NAME, and its support email address
// in place of SUPPORT_EMAIL, each escaped, in one pass so neither is replaced inside the other
func WithOperator(document string, operator string, email string) string {
	return strings.NewReplacer("OPERATOR_NAME", html.EscapeString(operator), "SUPPORT_EMAIL", html.EscapeString(email)).Replace(document)
}
