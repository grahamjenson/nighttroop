package sitemodels

// SiteAlarm is one of the signed in identity's alarms, as the alarm checker keeps it in the live DB. Deleting it
// snoozes it.
type SiteAlarm struct {
	// ID is its kind, as urgent_low
	ID     string `json:"id"`
	Urgent bool   `json:"urgent"`
	// Active is whether its condition is met; one that is not is only kept while it is snoozed
	Active bool `json:"active"`
	// Alerting is false while it is snoozed, or quieted by a louder alarm
	Alerting bool `json:"alerting"`
	// SnoozedUntil is when its snooze ends in milliseconds, unset while it is not snoozed
	SnoozedUntil *int64 `json:"snoozed_until"`
}
