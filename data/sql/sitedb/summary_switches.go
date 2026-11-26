package sitedb

import (
	"context"
	"slices"
)

// SwitchLookbackMs is how long before a time a profile switch with a duration is looked for, as one still in force then
const SwitchLookbackMs = 7 * 24 * 60 * 60_000

// SummaryProfilesParams are the times SummaryProfiles reads the profiles in force through, and the most it reads of each
type SummaryProfilesParams struct {
	Since, Until, Limit int64
}

// SummaryProfiles are the default profiles and the profile switches in force from Since to Until, each oldest first, at
// most Limit of each, see SearchSummaryDefaultProfiles and SummaryProfileSwitches
func (q *Queries) SummaryProfiles(ctx context.Context, arg SummaryProfilesParams) ([]*ProfilePreset, []*ProfileSwitch, error) {
	defaults, err := q.SearchSummaryDefaultProfiles(ctx, SearchSummaryDefaultProfilesParams{Since: arg.Since, Until: arg.Until, Limit: arg.Limit})
	if err != nil {
		return nil, nil, err
	}
	slices.Reverse(defaults)
	switches, err := q.SummaryProfileSwitches(ctx, SearchSummaryProfileSwitchesParams{Since: arg.Since, Until: arg.Until, Earliest: arg.Since - SwitchLookbackMs, Limit: arg.Limit})
	return defaults, switches, err
}

// SummaryProfileSwitches are the profile switches in force from Since to Until, oldest first, as
// SearchSummaryProfileSwitches finds them. A switch with no schedules of its own, as the careportal and T1DCarer send
// one by its profile's name, reads those of the profile of its name in the store in force when it started, as
// Nightscout's profile functions do, and keeps none when that store has no such profile. It is read from a copy, so the
// stored switch, and the Nightscout API, still have no profileJson.
func (q *Queries) SummaryProfileSwitches(ctx context.Context, arg SearchSummaryProfileSwitchesParams) ([]*ProfileSwitch, error) {
	switches, err := q.SearchSummaryProfileSwitches(ctx, arg)
	if err != nil {
		return nil, err
	}
	slices.Reverse(switches)
	first := slices.IndexFunc(switches, byName)
	if first < 0 {
		return switches, nil
	}
	profiles, err := q.SearchSummaryStoreProfiles(ctx, SearchSummaryStoreProfilesParams{Since: switches[first].OccurredAt, Until: arg.Until, Limit: arg.Limit})
	if err != nil {
		return nil, err
	}
	for i, s := range switches {
		if !byName(s) {
			continue
		}
		if p := storeProfile(profiles, s); p != nil {
			named := *s
			named.BasalRatesJson, named.TargetsLowJson, named.TargetsHighJson = p.BasalRatesJson, p.TargetsLowJson, p.TargetsHighJson
			if named.Timezone == nil {
				named.Timezone = p.Timezone
			}
			switches[i] = &named
		}
	}
	return switches, nil
}

// byName is whether the switch has none of the schedules the summary reads, so names its profile
func byName(s *ProfileSwitch) bool {
	return s.BasalRatesJson == nil && s.TargetsLowJson == nil && s.TargetsHighJson == nil
}

// storeProfile is the profile of the switch's name in the store in force when it started, from the profiles newest
// first with each store's together, or nil when there is none
func storeProfile(profiles []*ProfilePreset, s *ProfileSwitch) *ProfilePreset {
	store := ""
	for _, p := range profiles {
		if p.OccurredAt > s.OccurredAt {
			continue
		}
		if store == "" {
			store = p.ProfileStoreID
		}
		if p.ProfileStoreID != store {
			return nil
		}
		if p.Name == s.Name {
			return p
		}
	}
	return nil
}
