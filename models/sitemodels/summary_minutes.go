package sitemodels

import (
	"slices"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
)

// SummaryMinute is what was in force in a minute of the summary: its targets, in the settings' units, and its basal
// rate in U/hr, each null when nothing set it
type SummaryMinute struct {
	LowTarget  *float64 `json:"low_target"`
	HighTarget *float64 `json:"high_target"`
	Basal      *float64 `json:"basal"`
}

// profile is a default profile's or a profile switch's schedules, in mg/dL and U/hr, read in its timezone, shifted by
// shift seconds and with its basal rates scaled by basalScale
type profile struct {
	start, end       int64
	low, high, basal []*SiteScheduleValue
	location         *time.Location
	shift            int64
	basalScale       float64
	permanent        bool
}

func newProfile(start int64, minutes int64, timezone *string, low, high, basal *string, location *time.Location) *profile {
	if timezone != nil {
		if tz, err := time.LoadLocation(*timezone); err == nil {
			location = tz
		}
	}
	p := &profile{start: start, low: ScheduleFromJSON(low), high: ScheduleFromJSON(high), basal: ScheduleFromJSON(basal), location: location, basalScale: 1, permanent: minutes == 0}
	p.end = start + minutes*60_000
	return p
}

// at is the schedule's value at ms in the profile's time of day, or nil when it has no schedule
func (p *profile) at(schedule []*SiteScheduleValue, ms int64) *float64 {
	if p == nil || len(schedule) == 0 {
		return nil
	}
	t := time.UnixMilli(ms + p.shift*1000).In(p.location)
	seconds := int64(t.Hour()*3600 + t.Minute()*60 + t.Second())
	// Before the first entry is the end of the day before, so the last entry's value
	value := schedule[len(schedule)-1].Value
	for _, entry := range schedule {
		if entry.Seconds > seconds {
			break
		}
		value = entry.Value
	}
	return &value
}

// SummaryMinutes is each minute from since until until, oldest first, with the targets and basal in force: the
// default profile as of that minute, overridden by a profile switch, overridden by a temp target or temp basal.
// Defaults and switches are oldest first. A switch lasts its duration, or until the next one when it has none, and
// one that ends hands back to the last switch without a duration. A newer temp target or temp basal replaces an older
// one, and a temp basal's rate, the U/hr given, is the basal.
func SummaryMinutes(since, until int64, units *consts.Units, location *time.Location, defaults []*sitedb.ProfilePreset, switches []*sitedb.ProfileSwitch, targets []*SiteTempTarget, basals []*SiteTempBasal) []SummaryMinute {
	var defaultProfiles, switchProfiles []*profile
	for _, d := range defaults {
		defaultProfiles = append(defaultProfiles, newProfile(d.OccurredAt, 0, d.Timezone, d.TargetsLowJson, d.TargetsHighJson, d.BasalRatesJson, location))
	}
	for _, s := range switches {
		p := newProfile(s.OccurredAt, s.DurationMinutes, s.Timezone, s.TargetsLowJson, s.TargetsHighJson, s.BasalRatesJson, location)
		if s.Multiplier != nil {
			p.basalScale = *s.Multiplier
		}
		if s.Timeshift != nil {
			p.shift = *s.Timeshift * 60
		}
		switchProfiles = append(switchProfiles, p)
	}
	byStart := func(a, b *SiteTempTarget) int { return int(a.OccurredAt - b.OccurredAt) }
	targets = slices.SortedFunc(slices.Values(targets), byStart)
	basals = slices.SortedFunc(slices.Values(basals), func(a, b *SiteTempBasal) int { return int(a.OccurredAt - b.OccurredAt) })

	minutes := make([]SummaryMinute, 0, (until-since)/60_000+1)
	var current, permanent, temporary *profile
	var target *SiteTempTarget
	var basal *SiteTempBasal
	d, s, t, b := 0, 0, 0, 0
	for ms := since; ms <= until; ms += 60_000 {
		for ; d < len(defaultProfiles) && defaultProfiles[d].start <= ms; d++ {
			current = defaultProfiles[d]
		}
		for ; s < len(switchProfiles) && switchProfiles[s].start <= ms; s++ {
			if switchProfiles[s].permanent {
				permanent, temporary = switchProfiles[s], nil
			} else {
				temporary = switchProfiles[s]
			}
		}
		for ; t < len(targets) && targets[t].OccurredAt <= ms; t++ {
			target = targets[t]
		}
		for ; b < len(basals) && basals[b].OccurredAt <= ms; b++ {
			basal = basals[b]
		}

		switched := permanent
		if temporary != nil && ms < temporary.end {
			switched = temporary
		}
		// A switch without a schedule of its own keeps the default profile's
		from := func(schedule func(*profile) []*SiteScheduleValue) (*profile, []*SiteScheduleValue) {
			if switched != nil && len(schedule(switched)) > 0 {
				return switched, schedule(switched)
			}
			if current == nil {
				return nil, nil
			}
			return current, schedule(current)
		}
		var minute SummaryMinute
		if p, schedule := from(func(p *profile) []*SiteScheduleValue { return p.low }); p != nil {
			minute.LowTarget = fromMgdl(units, p.at(schedule, ms))
		}
		if p, schedule := from(func(p *profile) []*SiteScheduleValue { return p.high }); p != nil {
			minute.HighTarget = fromMgdl(units, p.at(schedule, ms))
		}
		if p, schedule := from(func(p *profile) []*SiteScheduleValue { return p.basal }); p != nil {
			if rate := p.at(schedule, ms); rate != nil {
				scale := 1.0
				if switched != nil {
					scale = switched.basalScale
				}
				minute.Basal = new(*rate * scale)
			}
		}

		if target != nil && ms < target.OccurredAt+target.DurationMinutes*60_000 {
			minute.LowTarget, minute.HighTarget = new(target.LowTarget), new(target.HighTarget)
		}
		if basal != nil && ms < basal.OccurredAt+basal.DurationMinutes*60_000 {
			if basal.Rate != nil {
				minute.Basal = new(*basal.Rate)
			}
		}
		minutes = append(minutes, minute)
	}
	return minutes
}

// fromMgdl is a glucose in mg/dL in the settings' units
func fromMgdl(units *consts.Units, mgdl *float64) *float64 {
	if mgdl == nil {
		return nil
	}
	return new(units.FromMGDL(*mgdl))
}

// BasalAt is the profiles' scheduled basal in U/hr at ms, with a switch's percent and timeshift, as Nightscout reads
// one for a temp basal starting then. It is nil when no profile has a basal then. Defaults and switches are oldest
// first.
func BasalAt(ms int64, location *time.Location, defaults []*sitedb.ProfilePreset, switches []*sitedb.ProfileSwitch) *float64 {
	return SummaryMinutes(ms, ms, nil, location, defaults, switches, nil, nil)[0].Basal
}
