package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Links behind HTTPS have no port, as a proxy takes HTTPS on 443 and forwards to the server's HOST_PORT; in development
// the server is reached directly, on HOST_PORT
func TestLinks(t *testing.T) {
	tests := []struct {
		name                          string
		protocol                      string
		site, recover, setup, hostURL string
	}{
		{"behind HTTPS", "https", "https://abcdefgh.nt.t1d.nz", "https://nt.t1d.nz/recover", "https://control.nt.t1d.nz/setup", "https://nt.t1d.nz"},
		{"in development", "http", "http://abcdefgh.nt.t1d.nz:3000", "http://nt.t1d.nz:3000/recover", "http://control.nt.t1d.nz:3000/setup", "http://nt.t1d.nz:3000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ServerConfig{HostName: "nt.t1d.nz", HostPort: 3000, HostProtocol: tt.protocol}
			config.setHostURLs()
			require.Equal(t, tt.site, config.SiteURL("abcdefgh"))
			require.Equal(t, tt.recover, config.RecoverURL())
			require.Equal(t, tt.setup, config.SetupURL())
			require.Equal(t, tt.hostURL, config.HostURLWithPort)
		})
	}
}
