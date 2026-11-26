package sitemodels

import (
	"context"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
)

// PumpSuspendType is the type of a temp basal that is a suspended pump, as AAPS sends one
const PumpSuspendType = "PUMP_SUSPEND"

// maxMultiplier is the most times the scheduled basal a temp basal is, as Nightscout's 1000 percent, so a rate far
// over the basal, as one set over a basal of nearly nothing, has no multiplier
const maxMultiplier = 11

type SiteTempBasal struct {
	ID              string `json:"id"`
	DurationMinutes int64  `json:"duration_minutes"`
	// Rate is the U/hr given while it runs, sent as an absolute rate, a pointer so 0, a suspended basal, can be sent. It
	// is nil when it is a Multiplier.
	Rate *float64 `json:"rate"`
	// Multiplier is how many times the scheduled basal each minute it runs, sent as a percent, so 0.5 is half and 0 a zero
	// temp. It is nil when it is a Rate.
	Multiplier            *float64 `json:"multiplier"`
	Notes                 *string  `json:"notes"`
	Reason                *string  `json:"reason"`
	UtcOffsetMinutes      *int64   `json:"utc_offset_minutes"`
	OccurredAt            int64    `json:"occurred_at"`
	CreatedAt             int64    `json:"created_at"`
	UpdatedAt             int64    `json:"updated_at"`
	CreatedByAppName      *string  `json:"created_by_app_name"`
	CreatedByDeviceName   *string  `json:"created_by_device_name"`
	CreatedByIdentityName string   `json:"created_by_identity_name"`
	UpdatedByIdentityName *string  `json:"updated_by_identity_name"`
	// Type is the kind of temp basal, e.g. NORMAL, or PUMP_SUSPEND for a suspended pump
	Type *string `json:"type"`

	// ScheduledBasal is the profile's basal in U/hr at its start, read only: nil when no profile has a basal then
	ScheduledBasal *float64 `json:"scheduled_basal"`
}

// Suspend is whether it is a suspended pump, of the type PUMP_SUSPEND
func (b *SiteTempBasal) Suspend() bool {
	return b.Type != nil && *b.Type == PumpSuspendType
}

// Running is whether it is in force at ms, from its start for its duration. A suspended pump of 0 minutes has not
// resumed yet, so it runs until a newer temp basal replaces it. Any other temp basal of 0 minutes, such as a cancel,
// ends at once.
func (b *SiteTempBasal) Running(ms int64) bool {
	return ms < b.OccurredAt+b.DurationMinutes*60_000 || (b.DurationMinutes == 0 && b.Suspend())
}

// RateAt is the U/hr it gives at a time the profile's basal is scheduled: its rate, or its multiplier times the
// scheduled basal, as Nightscout (getTempBasal) and AAPS read a temp basal sent as a percent. It is nil for a
// multiplier when no profile has a basal then.
func (b *SiteTempBasal) RateAt(scheduled *float64) *float64 {
	switch {
	case b.Rate != nil:
		return new(*b.Rate)
	case b.Multiplier != nil && scheduled != nil:
		return new(consts.Round(*b.Multiplier**scheduled, 3))
	}
	return nil
}

// MultiplierAt is how many times the scheduled basal it is at a time the profile's basal is scheduled: its multiplier,
// or its rate over the scheduled basal. A rate has none when no profile has a basal then, over a basal of 0, or over
// 11 times the basal.
func (b *SiteTempBasal) MultiplierAt(scheduled *float64) *float64 {
	switch {
	case b.Multiplier != nil:
		return new(*b.Multiplier)
	case b.Rate != nil && scheduled != nil && *scheduled > 0 && *b.Rate <= maxMultiplier**scheduled:
		return new(consts.Round(*b.Rate / *scheduled, 5))
	}
	return nil
}

// StartRate is the U/hr it gives at its start, see RateAt and ScheduledBasal
func (b *SiteTempBasal) StartRate() *float64 {
	return b.RateAt(b.ScheduledBasal)
}

// StartMultiplier is how many times the scheduled basal it is at its start, see MultiplierAt and ScheduledBasal
func (b *SiteTempBasal) StartMultiplier() *float64 {
	return b.MultiplierAt(b.ScheduledBasal)
}

// profileRows is the most default profiles, and the most profile switches, read at once for the times of temp basals
const profileRows = 7 * 24 * 12

// ScheduledBasals are the profile's basal in U/hr at each of the times, see Profiles.BasalAt, from the profiles in force
// read once for all of them, or for each time on its own when more are in force between them than are read at once.
// A profile without its own timezone is read in location's.
func ScheduledBasals(ctx context.Context, q *sitedb.Queries, location *time.Location, times []int64) (map[int64]*float64, error) {
	basals := map[int64]*float64{}
	if len(times) == 0 {
		return basals, nil
	}
	defaults, switches, err := q.SummaryProfiles(ctx, sitedb.SummaryProfilesParams{Since: slices.Min(times), Until: slices.Max(times), Limit: profileRows})
	if err != nil {
		return nil, err
	}
	all := len(defaults) < profileRows && len(switches) < profileRows
	profiles := NewProfiles(location, defaults, switches)
	for _, at := range times {
		if !all {
			if defaults, switches, err = q.SummaryProfiles(ctx, sitedb.SummaryProfilesParams{Since: at, Until: at, Limit: profileRows}); err != nil {
				return nil, err
			}
			profiles = NewProfiles(location, defaults, switches)
		}
		basals[at] = profiles.BasalAt(at)
	}
	return basals, nil
}

// FillScheduledBasals sets each temp basal's ScheduledBasal, see ScheduledBasals
func FillScheduledBasals(ctx context.Context, q *sitedb.Queries, location *time.Location, tempBasals []*SiteTempBasal) error {
	times := []int64{}
	for _, b := range tempBasals {
		times = append(times, b.OccurredAt)
	}
	basals, err := ScheduledBasals(ctx, q, location, times)
	for _, b := range tempBasals {
		b.ScheduledBasal = basals[b.OccurredAt]
	}
	return err
}
