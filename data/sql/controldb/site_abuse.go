package controldb

import "math"

// AbuseSizeFactor is how many times its size limit a site's DB can be before the site is flagged for abuse. The server
// truncates a site near its limit, so a DB far over it is likely bloated on purpose, to hurt the server and its backups.
const AbuseSizeFactor = 3

// FlaggedForAbuse is whether the site may be abused, so the server limits it harder. Each check of abuse is added here,
// so all that limit an abused site use the same flag.
func (s *Site) FlaggedForAbuse() bool {
	return s.overAbuseSize()
}

// overAbuseSize is whether the site's DB is over AbuseSizeFactor times its size limit, false if either is not known
func (s *Site) overAbuseSize() bool {
	return s.EstimatedDbSizeKb != nil && s.MaxDbSizeKbLimit != nil && *s.EstimatedDbSizeKb > AbuseSizeFactor*(*s.MaxDbSizeKbLimit)
}

// AbuseLevel is how much a site may abuse the server, from its abuse score. It is only shown to control admins for now:
// the server does not yet treat a site differently by its level.
type AbuseLevel int

const (
	AbuseGood AbuseLevel = iota
	AbuseFair
	AbuseOverUse
	AbuseAbusive
	AbuseBlocked
)

var abuseLevelNames = []string{"Good", "Fair", "OverUse", "Abusive", "Blocked"}

func (l AbuseLevel) String() string { return abuseLevelNames[l] }

// The abuse score's tuning, all in one place: each level's lowest score, each signal's curve and the caps
const (
	abuseFairFrom    = 0.2
	abuseOverUseFrom = 0.4
	abuseAbusiveFrom = 0.6
	abuseBlockedFrom = 0.8

	// abuseSizeOver is how far over its limit, in limits, a site's DB is when its size signal reaches 1 (at 3.5 times)
	abuseSizeOver = 2.5
	// abuseConnectorFailuresHalf is the failed syncs in a row at which the connectors signal is a half (before its cap)
	abuseConnectorFailuresHalf = 5
	// abuseConnectorsCap keeps connector failures alone under OverUse, as an honest site's connector fails often (a
	// changed password, a server down)
	abuseConnectorsCap = 0.39
)

// AbuseSignals are what a site's abuse score is from
type AbuseSignals struct {
	// SizeKB is its DB's size and LimitKB its size limit, nil if not known
	SizeKB, LimitKB *int64
	// ConnectorFailures is the most failed syncs in a row now of its connectors that are on
	ConnectorFailures int64
}

// Score is the site's abuse score from 0 to 1: any one signal can raise it, and each adds to the others. Only its size
// can make it Blocked, as its connectors stay under OverUse. It is rounded to 2 places, as it is shown, so its level is
// the level of the score shown (1 − 0.8 is 0.19999… in floating point, shown as 0.20 but Good, not Fair).
func (s AbuseSignals) Score() float64 {
	return math.Round((1-(1-s.size())*(1-s.connectors()))*100) / 100
}

// Level is the level of the site's abuse score
func (s AbuseSignals) Level() AbuseLevel { return AbuseLevelOf(s.Score()) }

// size is 0 at or under its limit, rising to 1 at abuseSizeOver limits over it
func (s AbuseSignals) size() float64 {
	if s.SizeKB == nil || s.LimitKB == nil || *s.LimitKB <= 0 {
		return 0
	}
	return clamp((float64(*s.SizeKB)/float64(*s.LimitKB)-1)/abuseSizeOver, 1)
}

// connectors is 0 with no failed syncs in a row, a half at abuseConnectorFailuresHalf of them, and at most its cap
func (s AbuseSignals) connectors() float64 {
	failures := float64(max(0, s.ConnectorFailures))
	return clamp(failures/(failures+abuseConnectorFailuresHalf), abuseConnectorsCap)
}

func clamp(v float64, most float64) float64 { return max(0, min(most, v)) }

// AbuseLevelOf is the level of an abuse score
func AbuseLevelOf(score float64) AbuseLevel {
	switch {
	case score >= abuseBlockedFrom:
		return AbuseBlocked
	case score >= abuseAbusiveFrom:
		return AbuseAbusive
	case score >= abuseOverUseFrom:
		return AbuseOverUse
	case score >= abuseFairFrom:
		return AbuseFair
	}
	return AbuseGood
}

// AbuseLevel is the site's level from the abuse score the resource check last stored
func (s *Site) AbuseLevel() AbuseLevel { return AbuseLevelOf(s.AbuseScore) }
