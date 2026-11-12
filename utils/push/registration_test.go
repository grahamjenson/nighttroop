package push

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

func TestCheckRegistration(t *testing.T) {
	apns := Registration{Provider: consts.PushProviderAPNs, App: "loop", BundleID: "com.TEAM123456.loopkit.Loop", Token: strings.Repeat("ab", 32)}
	fcm := Registration{Provider: consts.PushProviderFCM, App: "t1dcarer", BundleID: "nz.t1d.carer", Token: "dQw4w9WgXcQ:APA91b" + strings.Repeat("x", 140)}
	with := func(r Registration, change func(*Registration)) Registration {
		change(&r)
		return r
	}
	tests := []struct {
		name         string
		registration Registration
		err          string
	}{
		{"an APNs device token", apns, ""},
		{"a Firebase registration token", fcm, ""},
		{"an unknown app", with(apns, func(r *Registration) { r.App = "nope" }), "unknown app"},
		{"no bundle ID", with(apns, func(r *Registration) { r.BundleID = "" }), "bad bundle ID"},
		{"a bundle ID with a quote", with(apns, func(r *Registration) { r.BundleID = `com."a` }), "bad bundle ID"},
		{"an APNs token that is not hex", with(apns, func(r *Registration) { r.Token = strings.Repeat("zz", 32) }), "in hex"},
		{"a short APNs token", with(apns, func(r *Registration) { r.Token = "abcd" }), "in hex"},
		{"Trio's 160 character APNs token", with(apns, func(r *Registration) { r.Token = strings.Repeat("ab", 80) }), ""},
		{"a 256 character APNs token", with(apns, func(r *Registration) { r.Token = strings.Repeat("ab", 128) }), ""},
		{"a too long APNs token", with(apns, func(r *Registration) { r.Token = strings.Repeat("ab", 128) + "a" }), "in hex"},
		{"a Firebase token with spaces", with(fcm, func(r *Registration) { r.Token = strings.Repeat("a b", 20) }), "not a Firebase"},
		{"a too long Firebase token", with(fcm, func(r *Registration) { r.Token = strings.Repeat("a", 1025) }), "not a Firebase"},
		{"an unknown provider", with(apns, func(r *Registration) { r.Provider = "sms" }), "unknown provider"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckRegistration(tt.registration)
			if tt.err == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.err)
		})
	}
}

// A token's hash is the same each time, and differs for another token
func TestTokenHash(t *testing.T) {
	require.Len(t, TokenHash("a"), 64)
	require.Equal(t, TokenHash("a"), TokenHash("a"))
	require.NotEqual(t, TokenHash("a"), TokenHash("b"))
}
