package sitemodels

// Settings are one level of settings, the site's or an identity's. A nil value is not set at
// this level, so the level above applies: the site's for an identity, the defaults for the site.
// Units is a string, not consts.Units, so an empty value can clear it.
type Settings struct {
	Units        *string `json:"units"`
	Timezone     *string `json:"timezone"`
	Language     *string `json:"language"`
	DateFormat   *string `json:"date_format"`
	TimeFormat   *string `json:"time_format"`
	WeekStartDay *string `json:"week_start_day"`

	// Inherited are the values from the level above, by setting name
	Inherited map[string]string `json:"inherited,omitempty"`
	// Site is the site's settings, shown beside an admin's own on the settings page
	Site *Settings `json:"-"`
}
