package httputils

import "net/http"

// FrameAncestorsNone is the Content-Security-Policy that lets no page frame this server's: every site is a subdomain of
// the one host, so a page on one site could otherwise frame another's and trick its identity into clicking on it. A
// handler that sets a Content-Security-Policy of its own keeps this in it.
const FrameAncestorsNone = "frame-ancestors 'none'"

// SecurityHeaders sets on every response what keeps a browser from misusing it: nothing sniffs a type it was not sent
// as, no page frames it, no other origin is sent its URL as the referrer, which would tell it the site's subdomain, and
// behind HTTPS the host and its subdomains are only reached over HTTPS. A handler can still set its own, stricter ones.
func SecurityHeaders(https bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()
			header.Set("X-Content-Type-Options", "nosniff")
			header.Set("X-Frame-Options", "DENY")
			header.Set("Content-Security-Policy", FrameAncestorsNone)
			header.Set("Referrer-Policy", "same-origin")
			if https {
				header.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}
