package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyntacticallyValidateSiteID(t *testing.T) {
	for _, test := range []struct {
		siteID string
		valid  bool
	}{
		{"alicezzz", true},
		{"abcdefgh", true},
		{"", false},
		{"alicezz", false},
		{"alicezzzz", false},
		{"Alicezzz", false},
		{"alicezz1", false},
		{"alice-zz", false},
		{"alice.zz", false},
		{"alicézz", false}, // é is two bytes, so this is eight bytes long
		{"ns.alice", false},
	} {
		t.Run(test.siteID, func(t *testing.T) {
			require.Equal(t, test.valid, SyntacticallyValidateSiteID(test.siteID))
		})
	}
}

// Every request to a subdomain checks its site ID, so the check allocates nothing
func TestSyntacticallyValidateSiteIDAllocatesNothing(t *testing.T) {
	require.Zero(t, testing.AllocsPerRun(100, func() { SyntacticallyValidateSiteID("alicezzz") }))
}

func BenchmarkSyntacticallyValidateSiteID(b *testing.B) {
	for b.Loop() {
		SyntacticallyValidateSiteID("alicezzz")
	}
}
