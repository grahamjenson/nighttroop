package sitemodels

// SiteCleanGlucoseReading is a glucose reading kept over the same glucose read by other sources, with where it came
// from, its change since the reading kept before it, and the others
type SiteCleanGlucoseReading struct {
	SiteGlucoseReading
	// Source is direct, or the sync that wrote it: nightscout, dexcom, libre or virtual
	Source string `json:"source"`
	// Diff is its change in Units, and DiffMinutes the minutes, since the reading before it; unset for the first
	Diff        *float64                  `json:"diff"`
	DiffMinutes *float64                  `json:"diff_minutes"`
	Others      []SiteOtherGlucoseReading `json:"others"`
}

// SiteOtherGlucoseReading is a reading of the same glucose that was not kept, in Units
type SiteOtherGlucoseReading struct {
	ID      string  `json:"id"`
	Reading float64 `json:"reading"`
	Source  string  `json:"source"`
}
