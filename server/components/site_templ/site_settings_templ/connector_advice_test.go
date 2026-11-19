package site_settings_templ

import (
	"context"
	"strings"
	"testing"
)

func TestConnectorAdviceSaysWhatToDoAboutEachKindOfFailure(t *testing.T) {
	for _, tc := range []struct {
		err  string
		want string
	}{
		{"Dexcom Error: Dexcom Login Error: dexcom authentication failed: Dexcom refused the username or password", "refused the sign in"},
		{"Libre Error: libre authentication failed: invalid email or password", "refused the sign in"},
		{"Nightscout connector error: NSClient Get Last Modified Error: error refreshing JWT: JWT server returned an error: HTTP 401", "refused the sign in"},
		{"Nightscout connector error: NSClient Get Last Modified Error: error refreshing JWT: JWT host unavailable", "cannot reach"},
		{"Nightscout connector error: NSClient Get Last Modified Error: error refreshing JWT: JWT request timed out", "tries again"},
		{"Nightscout connector error: error refreshing JWT: JWT server returned an error: HTTP 502", "tries again"},
		{"Libre Error: libre account follows no one: accept a LibreLinkUp invite in the app first", "LibreLinkUp app"},
		{"Libre Error: libre login needs you to accept the terms of use in the LibreLinkUp app", "LibreLinkUp app"},
		{"nightscout cannot be on a private or local network: 10.0.0.1", "public address"},
		{"Nightscout: Failed to create entries: something new", ""},
	} {
		got := connectorAdvice(context.Background(), []string{"The first sync failed, so syncing has been turned off. " + tc.err})
		if (tc.want == "") != (got == "") || !strings.Contains(got, tc.want) {
			t.Errorf("advice for %q = %q, want it to contain %q", tc.err, got, tc.want)
		}
	}
}
