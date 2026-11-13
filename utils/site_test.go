package utils

import (
	"strconv"
	"testing"
	"time"

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

// A record's ID is shaped as a MongoDB ObjectId, as xDrip+ only deletes a treatment whose _id is: 24 lower case hex
// characters, starting with the unix seconds it was made in, and different each time
func TestNextIDIsAnObjectID(t *testing.T) {
	before := time.Now().Unix()
	seen := map[string]bool{}
	for range 1000 {
		id := NextID()
		require.Regexp(t, `^[0-9a-f]{24}$`, id)
		seconds, err := strconv.ParseInt(id[:8], 16, 64)
		require.NoError(t, err)
		require.InDelta(t, before, seconds, 2)
		require.False(t, seen[id], "an ID was made twice")
		seen[id] = true
	}
}
