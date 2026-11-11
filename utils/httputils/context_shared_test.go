package httputils

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientIP(t *testing.T) {
	// Loopback, as by default, and Caddy's address on its Docker network
	trusted := []netip.Prefix{netip.MustParsePrefix("127.0.0.1/32"), netip.MustParsePrefix("::1/128"), netip.MustParsePrefix("172.30.0.2/32")}
	tests := []struct {
		name       string
		fly        bool
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{"socket peer", false, "203.0.113.5:1234", nil, "203.0.113.5"},
		{"fly client ip from trusted proxy on fly", true, "172.30.0.2:1234", map[string]string{"Fly-Client-IP": "198.51.100.7"}, "198.51.100.7"},
		// a peer that is not Fly's proxy, such as another machine on Fly's private network, can set it to anything
		{"fly client ip from untrusted peer on fly", true, "203.0.113.5:1234", map[string]string{"Fly-Client-IP": "198.51.100.7"}, "203.0.113.5"},
		{"fly client ip from private network peer on fly", true, "[fdaa:0:1::2]:1234", map[string]string{"Fly-Client-IP": "198.51.100.7"}, "fdaa:0:1::2"},
		{"fly client ip and forwarded from untrusted peer", true, "203.0.113.5:1234", map[string]string{"Fly-Client-IP": "198.51.100.7", "X-Forwarded-For": "198.51.100.8"}, "203.0.113.5"},
		{"fly client ip ignored off fly", false, "203.0.113.5:1234", map[string]string{"Fly-Client-IP": "198.51.100.7"}, "203.0.113.5"},
		{"fly client ip from trusted proxy ignored off fly", false, "172.30.0.2:1234", map[string]string{"Fly-Client-IP": "198.51.100.7"}, "172.30.0.2"},
		{"forwarded from trusted proxy", false, "127.0.0.1:1234", map[string]string{"X-Forwarded-For": "6.6.6.6, 198.51.100.7"}, "198.51.100.7"},
		{"forwarded from untrusted peer", false, "203.0.113.5:1234", map[string]string{"X-Forwarded-For": "198.51.100.7"}, "203.0.113.5"},
		{"forwarded from trusted docker proxy", false, "172.30.0.2:1234", map[string]string{"X-Forwarded-For": "198.51.100.7"}, "198.51.100.7"},
		{"forwarded from another container", false, "172.30.0.3:1234", map[string]string{"X-Forwarded-For": "198.51.100.7"}, "172.30.0.3"},
		{"forwarded from trusted ipv6 proxy", false, "[::1]:1234", map[string]string{"X-Forwarded-For": "2001:db8::7"}, "2001:db8::7"},
		{"forwarded from ipv4 mapped trusted proxy", false, "[::ffff:127.0.0.1]:1234", map[string]string{"X-Forwarded-For": "198.51.100.7"}, "198.51.100.7"},
		// chi skips each trusted hop, so a client behind two of the server's proxies is not taken for the first
		{"forwarded through two trusted proxies", false, "127.0.0.1:1234", map[string]string{"X-Forwarded-For": "198.51.100.7, 172.30.0.2"}, "198.51.100.7"},
		// and a header that does not parse is not believed
		{"fly client ip that is not an ip", true, "172.30.0.2:1234", map[string]string{"Fly-Client-IP": "not-an-ip"}, "172.30.0.2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fly {
				t.Setenv("FLY_APP_NAME", "nighttroop")
			} else {
				t.Setenv("FLY_APP_NAME", "")
			}
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}
			var got string
			ClientIP(trusted)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				got = GetIPFromHTTPContext(r.Context())
			})).ServeHTTP(httptest.NewRecorder(), r)
			assert.Equal(t, tt.want, got)
		})
	}
}

// With no trusted proxies, chi would believe a header from every peer, so on Fly.io Fly-Client-IP is never read
func TestClientIPFlyWithoutTrustedProxies(t *testing.T) {
	t.Setenv("FLY_APP_NAME", "nighttroop")
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.5:1234"
	r.Header.Set("Fly-Client-IP", "198.51.100.7")
	assert.Equal(t, "203.0.113.5", GetIPFromHTTPContext(WithClientIP(r)))
}
