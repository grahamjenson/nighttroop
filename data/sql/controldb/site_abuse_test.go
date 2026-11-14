package controldb_test

import (
	"testing"

	"nighttroop.t1d.nz/data/sql/controldb"
)

// A site is flagged for abuse only once its DB is over 3 times its size limit, and not while either is unknown
func TestSiteFlaggedForAbuse(t *testing.T) {
	kb := func(n int64) *int64 { return &n }
	tests := []struct {
		name        string
		size, limit *int64
		want        bool
	}{
		{"no size", nil, kb(20480), false},
		{"no limit", kb(1 << 30), nil, false},
		{"under its limit", kb(10000), kb(20480), false},
		{"over its limit, under 3 times it", kb(40960), kb(20480), false},
		{"exactly 3 times its limit", kb(3 * 20480), kb(20480), false},
		{"over 3 times its limit", kb(3*20480 + 1), kb(20480), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			site := &controldb.Site{EstimatedDbSizeKb: tt.size, MaxDbSizeKbLimit: tt.limit}
			if got := site.FlaggedForAbuse(); got != tt.want {
				t.Errorf("FlaggedForAbuse() = %v, want %v", got, tt.want)
			}
		})
	}
}
