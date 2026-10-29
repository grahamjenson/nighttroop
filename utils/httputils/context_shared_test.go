package httputils

import (
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIPAddressFromRequest(t *testing.T) {
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
		{"fly client ip on fly", true, "172.16.0.2:1234", map[string]string{"Fly-Client-IP": "198.51.100.7"}, "198.51.100.7"},
		{"fly client ip ignored off fly", false, "203.0.113.5:1234", map[string]string{"Fly-Client-IP": "198.51.100.7"}, "203.0.113.5"},
		{"forwarded from trusted proxy", false, "127.0.0.1:1234", map[string]string{"X-Forwarded-For": "6.6.6.6, 198.51.100.7"}, "198.51.100.7"},
		{"forwarded from untrusted peer", false, "203.0.113.5:1234", map[string]string{"X-Forwarded-For": "198.51.100.7"}, "203.0.113.5"},
		{"forwarded from trusted docker proxy", false, "172.30.0.2:1234", map[string]string{"X-Forwarded-For": "198.51.100.7"}, "198.51.100.7"},
		{"forwarded from another container", false, "172.30.0.3:1234", map[string]string{"X-Forwarded-For": "198.51.100.7"}, "172.30.0.3"},
		{"forwarded from trusted ipv6 proxy", false, "[::1]:1234", map[string]string{"X-Forwarded-For": "2001:db8::7"}, "2001:db8::7"},
		{"forwarded from ipv4 mapped trusted proxy", false, "[::ffff:127.0.0.1]:1234", map[string]string{"X-Forwarded-For": "198.51.100.7"}, "198.51.100.7"},
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
			assert.Equal(t, tt.want, getIPAddressFromRequest(r, trusted))
		})
	}
}
