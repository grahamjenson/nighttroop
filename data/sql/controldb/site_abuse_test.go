package controldb_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
)

func kb(n int64) *int64 { return &n }

// A site's abuse score and level from its signals: its size alone can reach Blocked, its refusals at most Abusive, its
// connectors at most Fair, and its refusals and connectors together at most Abusive
func TestAbuseSignalsScore(t *testing.T) {
	limit := kb(20480)
	times := func(f float64) *int64 { return kb(int64(f * 20480)) }
	tests := []struct {
		name    string
		signals controldb.AbuseSignals
		score   float64
		level   controldb.AbuseLevel
	}{
		{"nothing", controldb.AbuseSignals{}, 0, controldb.AbuseGood},
		{"no size", controldb.AbuseSignals{LimitKB: limit}, 0, controldb.AbuseGood},
		{"no limit", controldb.AbuseSignals{SizeKB: kb(1 << 30)}, 0, controldb.AbuseGood},
		{"under its limit", controldb.AbuseSignals{SizeKB: kb(10000), LimitKB: limit}, 0, controldb.AbuseGood},
		{"at its limit", controldb.AbuseSignals{SizeKB: limit, LimitKB: limit}, 0, controldb.AbuseGood},
		{"1.6 times its limit", controldb.AbuseSignals{SizeKB: times(1.6), LimitKB: limit}, 0.24, controldb.AbuseFair},
		{"2.2 times its limit", controldb.AbuseSignals{SizeKB: times(2.2), LimitKB: limit}, 0.48, controldb.AbuseOverUse},
		{"2.7 times its limit", controldb.AbuseSignals{SizeKB: times(2.7), LimitKB: limit}, 0.68, controldb.AbuseAbusive},
		{"3.2 times its limit", controldb.AbuseSignals{SizeKB: times(3.2), LimitKB: limit}, 0.88, controldb.AbuseBlocked},
		{"size alone reaches 1", controldb.AbuseSignals{SizeKB: times(10), LimitKB: limit}, 1, controldb.AbuseBlocked},
		{"connectors alone never above Fair", controldb.AbuseSignals{ConnectorFailures: 1}, 0.39, controldb.AbuseFair},
		{"connectors under their cap", controldb.AbuseSignals{ConnectorFailures: 0.1}, 0.1, controldb.AbuseGood},
		{"50 refusals", controldb.AbuseSignals{Refusals: 50}, 0.5, controldb.AbuseOverUse},
		{"refusals alone never Blocked", controldb.AbuseSignals{Refusals: 1e9}, 0.79, controldb.AbuseAbusive},
		{"no refusals below 0", controldb.AbuseSignals{Refusals: -5}, 0, controldb.AbuseGood},
		{"refusals and connectors never Blocked", controldb.AbuseSignals{Refusals: 1e9, ConnectorFailures: 1}, 0.79, controldb.AbuseAbusive},
		{"refusals and connectors together", controldb.AbuseSignals{Refusals: 50, ConnectorFailures: 0.3}, 0.65, controldb.AbuseAbusive},
		{"size and connectors together", controldb.AbuseSignals{SizeKB: times(1.6), LimitKB: limit, ConnectorFailures: 1}, 1 - 0.76*0.61, controldb.AbuseOverUse},
		{"size and every other signal", controldb.AbuseSignals{SizeKB: times(2.2), LimitKB: limit, Refusals: 1e9, ConnectorFailures: 1}, 1 - 0.52*0.21, controldb.AbuseBlocked},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.InDelta(t, tt.score, tt.signals.Score(), 0.001)
			require.Equal(t, tt.level, tt.signals.Level())
		})
	}
}

// Each level's band of scores starts at its lowest score
func TestAbuseLevelOf(t *testing.T) {
	tests := []struct {
		score float64
		level controldb.AbuseLevel
	}{
		{0, controldb.AbuseGood}, {0.199, controldb.AbuseGood},
		{0.2, controldb.AbuseFair}, {0.399, controldb.AbuseFair},
		{0.4, controldb.AbuseOverUse}, {0.599, controldb.AbuseOverUse},
		{0.6, controldb.AbuseAbusive}, {0.799, controldb.AbuseAbusive},
		{0.8, controldb.AbuseBlocked}, {1, controldb.AbuseBlocked},
	}
	for _, tt := range tests {
		require.Equal(t, tt.level, controldb.AbuseLevelOf(tt.score), "score %v", tt.score)
	}
	require.Equal(t, []string{"Good", "Fair", "OverUse", "Abusive", "Blocked"}, []string{controldb.AbuseGood.String(),
		controldb.AbuseFair.String(), controldb.AbuseOverUse.String(), controldb.AbuseAbusive.String(), controldb.AbuseBlocked.String()})
}

// A site's level is from the score last stored, not from its size
func TestSiteAbuseLevel(t *testing.T) {
	require.Equal(t, controldb.AbuseGood, (&controldb.Site{EstimatedDbSizeKb: kb(1 << 30), MaxDbSizeKbLimit: kb(20480)}).AbuseLevel())
	require.Equal(t, controldb.AbuseAbusive, (&controldb.Site{AbuseScore: 0.7}).AbuseLevel())
}
