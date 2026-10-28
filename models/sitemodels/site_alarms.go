package sitemodels

// SiteAlarm is one of the signed in identity's active alarms, held as the alarm checker last sent it rather than kept
// in a DB. Deleting it snoozes it.
type SiteAlarm struct {
	// ID is its kind, as urgent_low
	ID     string `json:"id"`
	Urgent bool   `json:"urgent"`
	// Alerting is false while it is snoozed, or quieted by a louder alarm
	Alerting bool `json:"alerting"`
}
