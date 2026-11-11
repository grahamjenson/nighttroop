package httputils

import (
	"net/http"
	"net/url"
)

// RedactHeaders is header with the values of each credential, such as Authorization, Cookie or api-secret, redacted,
// by the same rules as captured requests
func RedactHeaders(header http.Header) map[string][]string {
	return redactHeaders(header)
}

// RedactValues is a query or form with the values of each credential, such as token, redacted, by the same rules as
// captured requests
func RedactValues(query url.Values) map[string][]string {
	return redactQuery(query)
}
