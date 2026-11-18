package nsutils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
)

// Only an identity added for a client with overrides gets them: every other identity, and a broadcast with none,
// gets what every client gets
func TestOverridesFor(t *testing.T) {
	client := func(c string) *controldb.SiteIdentity { return &controldb.SiteIdentity{Client: &c} }
	tests := []struct {
		name     string
		identity *controldb.SiteIdentity
		want     ClientOverrides
	}{
		{"no identity, as a broadcast", nil, ClientOverrides{}},
		{"no client", &controldb.SiteIdentity{}, ClientOverrides{}},
		{"Loop", client("loop"), ClientOverrides{FractionalDurations: true, SuspendReason: true}},
		{"Trio", client("trio"), ClientOverrides{}},
		{"AAPS", client("aaps"), ClientOverrides{}},
		{"xDrip4iOS", client("xdrip4ios"), ClientOverrides{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, OverridesFor(test.identity))
		})
	}
}
