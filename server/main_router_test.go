package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/httputils"
)

func TestGetRouteTargetFromSubdomain(t *testing.T) {
	const host = "nt.example.com"
	allowed := []string{host}

	tests := []struct {
		name      string
		host      string
		allowed   []string
		target    RouteTarget
		subdomain string
		wantErr   bool
	}{
		{name: "exact host", host: host, allowed: allowed, target: RouteTargetPublic},
		{name: "trailing dot", host: host + ".", allowed: allowed, target: RouteTargetPublic},
		{name: "site", host: "abcdefgh." + host, allowed: allowed, target: RouteTargetSite, subdomain: "abcdefgh"},
		{name: "nightscout", host: "abcdefgh.ns." + host, allowed: allowed, target: RouteTargetNightscout, subdomain: "abcdefgh"},
		{name: "control", host: "control." + host, allowed: allowed, target: RouteTargetControl},
		{name: "www", host: "www." + host, allowed: allowed, target: RouteTargetPublic},
		{name: "localhost in dev", host: "abcdefgh.localhost", allowed: []string{host, "localhost"}, target: RouteTargetSite, subdomain: "abcdefgh"},
		{name: "suffix inside label rejected", host: "abcdefgh.evil-" + host, allowed: allowed, wantErr: true},
		{name: "host as prefix rejected", host: host + ".evil.com", allowed: allowed, wantErr: true},
		{name: "localhost rejected in prod", host: "abcdefgh.localhost", allowed: allowed, wantErr: true},
		{name: "loopback rejected in prod", host: "127.0.0.1", allowed: allowed, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, subdomain, err := GetRouteTargetFromSubdomain(tt.host, tt.allowed)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected %q to be rejected, got target %d subdomain %q", tt.host, target, subdomain)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.host, err)
			}
			if target != tt.target {
				t.Fatalf("got target %d, want target %d", target, tt.target)
			}
			// public routes echo the first label, so only site routes carry a meaningful subdomain
			if target != RouteTargetPublic && subdomain != tt.subdomain {
				t.Fatalf("got subdomain %q, want subdomain %q", subdomain, tt.subdomain)
			}
		})
	}
}

// Only writes through a site's Nightscout API, a device's uploads, get the share of a busy server kept for them
func TestNightscoutWrite(t *testing.T) {
	allowed := []string{"nt.example.com"}
	tests := []struct {
		method string
		host   string
		want   bool
	}{
		{"POST", "abcdefgh.ns.nt.example.com", true},
		{"PUT", "abcdefgh.ns.nt.example.com:443", true},
		{"DELETE", "abcdefgh.ns.nt.example.com", true},
		{"GET", "abcdefgh.ns.nt.example.com", false},
		{"POST", "abcdefgh.nt.example.com", false},
		{"POST", "control.nt.example.com", false},
		{"POST", "abcdefgh.ns.evil.com", false},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(tt.method, "/api/v1/entries", nil)
		r.Host = tt.host
		require.Equal(t, tt.want, nightscoutWrite(r, allowed), "%s %s", tt.method, tt.host)
	}
}

// A Nightscout URL typed with a "/" at its end gives //api/v1/entries, which is served as /api/v1/entries, and as
// long lived when it is //socket.io/; other hosts' paths are as they were sent
func TestNightscoutOneSlash(t *testing.T) {
	allowed := []string{"nt.example.com"}
	tests := []struct {
		host      string
		path      string
		want      string
		longLived bool
	}{
		{"abcdefgh.ns.nt.example.com", "//api/v1/entries", "/api/v1/entries", false},
		{"abcdefgh.ns.nt.example.com", "///api//v3///entries.json", "/api/v3/entries.json", false},
		{"abcdefgh.ns.nt.example.com", "/api/v1/entries", "/api/v1/entries", false},
		{"abcdefgh.ns.nt.example.com", "//socket.io/?EIO=4", "/socket.io/", true},
		{"abcdefgh.ns.nt.example.com", "/socket.io/?EIO=4", "/socket.io/", true},
		{"abcdefgh.ns.nt.example.com", "//api/v1/entries/a%2Fb", "/api/v1/entries/a/b", false},
		{"abcdefgh.nt.example.com", "//static/app.js", "//static/app.js", false},
		{"control.nt.example.com", "//login", "//login", false},
	}
	for _, tt := range tests {
		var seen *http.Request
		handler := nightscoutOneSlash(allowed)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = r }))
		r := httptest.NewRequest(http.MethodGet, tt.path, nil)
		r.Host = tt.host
		handler.ServeHTTP(httptest.NewRecorder(), r)
		require.Equal(t, tt.want, seen.URL.Path, "%s %s", tt.host, tt.path)
		require.Equal(t, tt.longLived, httputils.LongLived(seen), "%s %s", tt.host, tt.path)
		require.Equal(t, tt.path, r.URL.RequestURI(), "the request sent is not changed")
	}
}
