package nighttroop_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/clients/nighttroop"
	"nighttroop.t1d.nz/utils/config"
)

// The served schema puts each endpoint's server on the serving server's host: the public host, a site's host, or a
// site's Nightscout host, with the site's ID left for the reader
func TestOpenAPIOn(t *testing.T) {
	type schema struct {
		Servers []struct{ URL string }
		Paths   map[string]struct{ Servers []struct{ URL string } }
	}
	tests := []struct {
		name                     string
		config                   config.ServerConfig
		public, site, nightscout string
	}{
		{"staging", config.ServerConfig{HostName: "staging.nighttroop.nz", HostProtocol: "https", HostPort: 3000},
			"https://staging.nighttroop.nz", "https://{site_id}.staging.nighttroop.nz", "https://{site_id}.ns.staging.nighttroop.nz"},
		{"development", config.ServerConfig{HostName: "localhost", HostProtocol: "http", HostPort: 3000},
			"http://localhost:3000", "http://{site_id}.localhost:3000", "http://{site_id}.ns.localhost:3000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			served := nighttroop.OpenAPIOn(tt.config.LinkURL)
			require.NotContains(t, string(served), nighttroop.SchemaHost)
			var s schema
			require.NoError(t, json.Unmarshal(served, &s))

			require.Equal(t, tt.site, s.Servers[0].URL, "an endpoint with no servers of its own is on a site's host")
			require.Equal(t, tt.public, s.Paths["/site"].Servers[0].URL)
			require.Empty(t, s.Paths["/login"].Servers)
			for path, item := range s.Paths {
				if strings.HasPrefix(path, "/api/") {
					require.Len(t, item.Servers, 1, path)
					require.Equal(t, tt.nightscout, item.Servers[0].URL, path)
				}
			}
		})
	}
}
