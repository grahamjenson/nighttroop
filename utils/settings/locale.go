package settings

import (
	"slices"

	"golang.org/x/text/language"
	"nighttroop.t1d.nz/utils/consts"
)

// Countries by the formats most used in them, where they differ from DD/MM/YYYY, a 24 hour clock and a week from monday
var (
	mgdlCountries       = []string{"US", "FR", "JP", "IN", "IL", "IR"}
	monthFirstCountries = []string{"US", "PH", "FM", "MH", "PW", "GU", "AS", "PR", "UM", "VI"}
	yearFirstCountries  = []string{"JP", "CN", "TW", "HK", "MO", "ZA", "IR"}
	isoDateCountries    = []string{"CA", "SE", "LT", "HU", "KR", "MN", "BT"}
	twelveHourCountries = []string{"US", "CA", "AU", "NZ", "IN", "PH", "PK", "BD", "EG", "SA", "MY", "KR", "TW", "CO", "JO", "IQ", "GU", "PR", "AS"}
	sundayCountries     = []string{"AG", "AS", "BD", "BR", "BS", "BT", "BW", "BZ", "CA", "CO", "DM", "DO", "ET", "GT", "GU", "HK",
		"HN", "ID", "IL", "IN", "JM", "JP", "KE", "KH", "KR", "LA", "MH", "MM", "MO", "MT", "MX", "MZ", "NI", "NP", "PA", "PE",
		"PH", "PK", "PR", "PT", "PY", "SA", "SG", "SV", "TH", "TT", "TW", "UM", "US", "VE", "VI", "WS", "YE", "ZA", "ZW"}
)

// BrowserSettings are a new site's settings from the browser that created it: its time zone, its language, or en-NZ,
// and the units, date and time formats and week start day most used in the language's country. The country is only
// known from a language that names one, e.g. en-US but not en; a value that is not valid is left out.
func BrowserSettings(timezone string, browserLanguage string) *SettingMap {
	m := &SettingMap{}
	m.Timezone, _ = NewSettingAndParseValidate(Timezone, timezone, nil, nil)
	tag, err := language.Parse(browserLanguage)
	if err != nil {
		m.SetLanguage(new(Language.Default))
		return m
	}
	m.SetLanguage(new(tag.String()))
	region, confidence := tag.Region()
	if confidence != language.Exact {
		return m
	}
	country := region.String()

	units, date, clock, weekStart := consts.UnitsMMOLL, "DD/MM/YYYY", "HH:mm", "monday"
	if slices.Contains(mgdlCountries, country) {
		units = consts.UnitsMGDL
	}
	switch {
	case slices.Contains(monthFirstCountries, country):
		date = "MM/DD/YYYY"
	case slices.Contains(yearFirstCountries, country):
		date = "YYYY/MM/DD"
	case slices.Contains(isoDateCountries, country):
		date = "YYYY-MM-DD"
	}
	if slices.Contains(twelveHourCountries, country) {
		clock = "hh:mm a"
	}
	if slices.Contains(sundayCountries, country) {
		weekStart = "sunday"
	}
	m.SetUnits(&units)
	m.SetDateFormat(&date)
	m.SetTimeFormat(&clock)
	m.SetWeekStartDay(&weekStart)
	return m
}
