package sitemodels

// PumpSuspendType is the type of a temp basal that is a suspended pump, as AAPS sends one
const PumpSuspendType = "PUMP_SUSPEND"

type SiteTempBasal struct {
	ID              string `json:"id"`
	DurationMinutes int64  `json:"duration_minutes"`
	// Rate is the U/hr given while it runs, a pointer so 0, a suspended basal, can be sent
	Rate *float64 `json:"rate"`
	// Multiplier is how many times the scheduled basal at its start, so 0.5 is half and 0 a zero temp
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

	// FollowsBasal is whether a Nightscout client sent it as a percent with no absolute rate, as AAPS on a percent pump,
	// so each minute it runs it is Multiplier times the scheduled basal then, not Rate
	FollowsBasal *bool `json:"follows_basal"`
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
