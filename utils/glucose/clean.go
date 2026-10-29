// Package glucose cleans a site's glucose readings: the same glucose from more than one source, as when a site syncs
// both Dexcom and Nightscout, becomes one reading, and each reading says how much it changed since the one before
package glucose

import (
	"cmp"
	"slices"

	"nighttroop.t1d.nz/data/sql/sitedb"
)

const (
	// SameWithinMs and SameWithinMgdl are how close in time and value two readings are to be the same glucose
	SameWithinMs   = 60_000
	SameWithinMgdl = 10.0
	// fiveMinutesMs is the time a change is given over, a CGM's usual time between readings
	fiveMinutesMs = 5 * 60_000.0
)

// Where a reading came from, by the identity that wrote it
const (
	Direct     = "direct"
	Nightscout = "nightscout"
	Dexcom     = "dexcom"
	Libre      = "libre"
	Virtual    = "virtual"
)

// syncSources are the sources of the readings a sync wrote, by the name it writes them as
var syncSources = map[string]string{"nightscout_sync": Nightscout, "dexcom_sync": Dexcom, "libre_sync": Libre, "virtual_sync": Virtual}

// priority is the order a reading is kept over the same glucose from other sources in: a meter's, then one reported
// directly, then each sync's
var priority = []string{Direct, Nightscout, Dexcom, Libre, Virtual}

// Source is where a reading came from: Direct when a person or device wrote it, or the sync that did
func Source(row *sitedb.GlucoseReading) string {
	return cmp.Or(syncSources[row.CreatedByIdentityName], Direct)
}

// rank is how a reading is kept over others of the same glucose, lowest first
func rank(row *sitedb.GlucoseReading) int {
	if row.ReadingType == "mbg" {
		return -1
	}
	return slices.Index(priority, Source(row))
}

// Other is a reading of the same glucose that was not kept
type Other struct {
	ID     string
	Mgdl   float64
	Source string
}

// Clean is a reading kept for its glucose, its change since the reading kept before it, and the others it was kept over
type Clean struct {
	Row    *sitedb.GlucoseReading
	Source string
	// DiffPer5mMgdl is its change since the reading before it, as mg/dL every 5 minutes, so readings further apart
	// compare; unset for the first, and for one at the same time as the one before
	DiffPer5mMgdl *float64
	Others        []Other
}

// CleanReadings is the readings with each glucose read by more than one source kept once, by the order of priority,
// newest first as a search gives them. Readings are the same glucose when they are within SameWithinMs and
// SameWithinMgdl of the first of them; two CGMs that read differently are both kept.
func CleanReadings(rows []*sitedb.GlucoseReading) []*Clean {
	rows = slices.Clone(rows)
	slices.SortStableFunc(rows, func(a, b *sitedb.GlucoseReading) int { return cmp.Compare(a.OccurredAt, b.OccurredAt) })
	var groups [][]*sitedb.GlucoseReading
	for _, row := range rows {
		joined := false
		for i := len(groups) - 1; i >= 0 && row.OccurredAt-groups[i][0].OccurredAt <= SameWithinMs; i-- {
			if diff := row.Mgdl - groups[i][0].Mgdl; diff <= SameWithinMgdl && diff >= -SameWithinMgdl {
				groups[i], joined = append(groups[i], row), true
				break
			}
		}
		if !joined {
			groups = append(groups, []*sitedb.GlucoseReading{row})
		}
	}

	cleaned := make([]*Clean, len(groups))
	for i, group := range groups {
		slices.SortStableFunc(group, func(a, b *sitedb.GlucoseReading) int { return cmp.Compare(rank(a), rank(b)) })
		c := &Clean{Row: group[0], Source: Source(group[0])}
		for _, other := range group[1:] {
			c.Others = append(c.Others, Other{ID: other.ID, Mgdl: other.Mgdl, Source: Source(other)})
		}
		if i > 0 {
			if previous := cleaned[i-1].Row; c.Row.OccurredAt > previous.OccurredAt {
				c.DiffPer5mMgdl = new((c.Row.Mgdl - previous.Mgdl) * fiveMinutesMs / float64(c.Row.OccurredAt-previous.OccurredAt))
			}
		}
		cleaned[i] = c
	}
	slices.Reverse(cleaned)
	return cleaned
}
