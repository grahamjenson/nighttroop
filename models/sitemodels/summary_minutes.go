package sitemodels

import (
	"slices"
	"time"

	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/settings"
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

func newProfile(start int64, minutes int64, timezone *string, low, high, basal *string, location *time.Location, schedules schedules) *profile {
	if timezone != nil {
		if tz, err := settings.LoadLocation(*timezone); err == nil {
			location = tz
		}
	}
	p := &profile{start: start, low: schedules.parse(low), high: schedules.parse(high), basal: schedules.parse(basal), location: location, basalScale: 1, permanent: minutes == 0}
	p.end = start + minutes*60_000
	return p
}

// schedules are the schedules parsed by their JSON, as each profile switch most often repeats the same profile's
type schedules map[string][]*SiteScheduleValue

// parse is the schedule of the JSON, parsed the first time it is seen. Its values are only read, so profiles share it.
func (s schedules) parse(jsonStr *string) []*SiteScheduleValue {
	if jsonStr == nil {
		return ScheduleFromJSON(nil)
	}
	schedule, ok := s[*jsonStr]
	if !ok {
		schedule = ScheduleFromJSON(jsonStr)
		s[*jsonStr] = schedule
	}
	return schedule
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

// Profiles are the default profiles and the profile switches in force through a time, each parsed once, to read what
// they have in force at each minute
type Profiles struct {
	defaults, switches []*profile
}

// NewProfiles are the defaults and switches, each oldest first. One without its own timezone is read in location's
// time.
func NewProfiles(location *time.Location, defaults []*sitedb.ProfilePreset, switches []*sitedb.ProfileSwitch) *Profiles {
	p, parsed := &Profiles{}, schedules{}
	for _, d := range defaults {
		p.defaults = append(p.defaults, newProfile(d.OccurredAt, 0, d.Timezone, d.TargetsLowJson, d.TargetsHighJson, d.BasalRatesJson, location, parsed))
	}
	for _, s := range switches {
		sp := newProfile(s.OccurredAt, s.DurationMinutes, s.Timezone, s.TargetsLowJson, s.TargetsHighJson, s.BasalRatesJson, location, parsed)
		if s.Multiplier != nil {
			sp.basalScale = *s.Multiplier
		}
		if s.Timeshift != nil {
			sp.shift = *s.Timeshift * 60
		}
		p.switches = append(p.switches, sp)
	}
	return p
}

// BasalAt is the profiles' scheduled basal in U/hr at ms, with a switch's percent and timeshift, as Nightscout reads
// one for a temp basal then. It is nil when no profile has a basal then.
func (p *Profiles) BasalAt(ms int64) *float64 {
	if p == nil {
		return nil
	}
	return p.Minutes(ms, ms, nil, nil, nil)[0].Basal
}

// Minutes is each minute from since until until, oldest first, with the targets and basal in force: the default
// profile as of that minute, overridden by a profile switch, overridden by a temp target or temp basal. A switch lasts
// its duration, or until the next one when it has none, and one that ends hands back to the last switch without a
// duration. A newer temp target or temp basal replaces an older one, and a temp basal's basal is its RateAt the
// scheduled basal of each minute, so one sent as a percent changes with the schedule and the switches. A suspend of 0
// minutes is a basal of 0 until the next.
func (p *Profiles) Minutes(since, until int64, units *consts.Units, targets []*SiteTempTarget, basals []*SiteTempBasal) []SummaryMinute {
	defaultProfiles, switchProfiles := p.defaults, p.switches
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
		if basal != nil && basal.Running(ms) {
			minute.Basal = basal.RateAt(minute.Basal)
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
