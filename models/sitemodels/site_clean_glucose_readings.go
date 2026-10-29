package sitemodels

// SiteCleanGlucoseReading is a glucose reading kept over the same glucose read by other sources, with where it came
// from, its change since the reading kept before it, and the others
type SiteCleanGlucoseReading struct {
	SiteGlucoseReading
	// Source is direct, or the sync that wrote it: nightscout, dexcom, libre or virtual
	Source string `json:"source"`
	// DiffPer5m is its change since the reading before it, in Units every 5 minutes; unset for the first
	DiffPer5m *float64                  `json:"diff_per_5m"`
	Others    []SiteOtherGlucoseReading `json:"others"`
}

// SiteOtherGlucoseReading is a reading of the same glucose that was not kept, in Units
type SiteOtherGlucoseReading struct {
	ID      string  `json:"id"`
	Reading float64 `json:"reading"`
	Source  string  `json:"source"`
}
