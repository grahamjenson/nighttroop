package site_settings_templ

import (
	"context"
	"testing"
	"time"

	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/consts"
)

func TestIdentityStateSaysHowEachIdentityIsUsed(t *testing.T) {
	now := time.Now()
	ago := func(d time.Duration) *int64 { ms := now.Add(-d).UnixMilli(); return &ms }
	device, person := string(consts.IdentityRoleDeviceRead), string(consts.IdentityRoleUser)
	for _, tc := range []struct {
		name     string
		role     string
		lastSeen *int64
		want     string
	}{
		{"a device no app has used", device, nil, "unused"},
		{"a person who has not signed in", person, nil, "unused"},
		{"a device seen a minute ago", device, ago(time.Minute), "active"},
		{"a person seen 29 minutes ago", person, ago(29 * time.Minute), "active"},
		{"a device seen an hour ago", device, ago(time.Hour), ""},
		{"a device seen two days ago", device, ago(48 * time.Hour), "quiet"},
		{"a person seen two days ago", person, ago(48 * time.Hour), ""},
	} {
		got, label, _ := identityStateOf(context.Background(), &sitemodels.UserIdentity{Role: tc.role, LastSeenAt: tc.lastSeen}, now)
		if got != tc.want || (got == "") != (label == "") {
			t.Errorf("%s: state %q with label %q, want %q", tc.name, got, label, tc.want)
		}
	}
}
