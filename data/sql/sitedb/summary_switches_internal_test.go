package sitedb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A switch by name reads the profile of its name in the store in force when it started, and no other store's
func TestStoreProfile(t *testing.T) {
	profile := func(store, name string, at int64) *ProfilePreset {
		return &ProfilePreset{ProfileStoreID: store, Name: name, OccurredAt: at}
	}
	// Newest first, each store's together
	profiles := []*ProfilePreset{
		profile("new", "Default", 300),
		profile("old", "Default", 100), profile("old", "Weekend", 100),
		profile("first", "Holiday", 50),
	}
	tests := []struct {
		name  string
		at    int64
		named string
		want  *ProfilePreset
	}{
		{"in the store in force", 200, "Weekend", profiles[2]},
		{"not in the newer store in force, though an older has it", 400, "Weekend", nil},
		{"not in the store in force, though the first has it", 200, "Holiday", nil},
		{"before any store", 10, "Holiday", nil},
		{"the store's default", 400, "Default", profiles[0]},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Same(t, test.want, storeProfile(profiles, &ProfileSwitch{Name: test.named, OccurredAt: test.at}))
		})
	}
}
