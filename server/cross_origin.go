package server

import "net/http"

// WithCrossOriginProtection is the outermost layer: it refuses a POST, PUT or DELETE from another origin, by its
// Sec-Fetch-Site or Origin header, so no other site, nor another of this server's hosts, changes anything with an
// identity's cookies. GET, HEAD and OPTIONS are not checked, so they must never change anything. A page on one host
// that leads to another, as creating a site leads to adding its passkey, links to it rather than posting to it.
func WithCrossOriginProtection(handler http.Handler) http.Handler {
	return http.NewCrossOriginProtection().Handler(handler)
}
