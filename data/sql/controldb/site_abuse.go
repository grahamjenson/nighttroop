package controldb

// AbuseLevel is how much a site may abuse the server, from its abuse score, so the server limits it harder
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
	// abuseRefusalsHalf is the refusals at which the refusals signal is a half
	abuseRefusalsHalf = 50
	// The caps keep refusals under Blocked, connectors under OverUse, and both together under Blocked
	abuseRefusalsCap   = 0.79
	abuseConnectorsCap = 0.39
	abuseOtherCap      = 0.79
)

// AbuseSignals are what a site's abuse score is from
type AbuseSignals struct {
	// SizeKB is its DB's size and LimitKB its size limit, nil if not known
	SizeKB, LimitKB *int64
	// Refusals is how many of its requests its rate limits refused, halving every hour
	Refusals float64
	// ConnectorFailures is the highest failure rate of its connectors, from 0 to 1
	ConnectorFailures float64
}

// Score is the site's abuse score from 0 to 1. Only its size can make it Blocked, as its other signals stay under.
func (s AbuseSignals) Score() float64 {
	other := min(abuseOtherCap, 1-(1-s.refusals())*(1-s.connectors()))
	return 1 - (1-s.size())*(1-other)
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

func (s AbuseSignals) refusals() float64 {
	if s.Refusals <= 0 {
		return 0
	}
	return min(abuseRefusalsCap, s.Refusals/(s.Refusals+abuseRefusalsHalf))
}

func (s AbuseSignals) connectors() float64 { return clamp(s.ConnectorFailures, abuseConnectorsCap) }

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
