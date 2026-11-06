package server

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
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
