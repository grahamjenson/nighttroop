package flow

import (
	"net/http"
	"strings"
)

// RemoteClient sends a script's requests to a server running elsewhere, as staging: a URL on localhost, or on a host
// under it such as a site's abcdefgh.localhost or abcdefgh.ns.localhost, goes to the same name under host, e.g.
// abcdefgh.ns.staging.nighttroop.nz, over scheme, and any other URL as it is. Like Client, it follows no redirect and
// asks for no compression.
func RemoteClient(scheme, host string) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableCompression = true
	return &http.Client{
		Transport:     &remote{scheme: scheme, host: host, next: transport},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

type remote struct {
	scheme, host string
	next         http.RoundTripper
}

func (r *remote) RoundTrip(req *http.Request) (*http.Response, error) {
	name := req.URL.Hostname()
	if name == "localhost" || strings.HasSuffix(name, ".localhost") {
		req = req.Clone(req.Context())
		req.URL.Scheme = r.scheme
		req.URL.Host = strings.TrimSuffix(name, "localhost") + r.host
		req.Host = ""
	}
	return r.next.RoundTrip(req)
}
