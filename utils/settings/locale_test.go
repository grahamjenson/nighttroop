package settings

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

func TestBrowserSettings(t *testing.T) {
	type want struct {
		timezone, language, date, clock, weekStart *string
		units                                      *consts.Units
	}
	mgdl, mmol := new(consts.UnitsMGDL), new(consts.UnitsMMOLL)
	tests := []struct {
		name               string
		timezone, language string
		want               want
	}{
		{"nothing from the browser is en-NZ", "", "", want{language: new("en-NZ")}},
		{"a time zone and language that are not valid are left out", "Mars/Olympus", "not a language!", want{language: new("en-NZ")}},
		{"a language with no country only sets the language", "Europe/Paris", "fr", want{timezone: new("Europe/Paris"), language: new("fr")}},
		{"the US reads mg/dL, month first, on a 12 hour clock from sunday", "America/New_York", "en-US",
			want{new("America/New_York"), new("en-US"), new("MM/DD/YYYY"), new("hh:mm a"), new("sunday"), mgdl}},
		{"France reads mg/dL", "Europe/Paris", "fr-FR", want{new("Europe/Paris"), new("fr-FR"), new("DD/MM/YYYY"), new("HH:mm"), new("monday"), mgdl}},
		{"Japan reads mg/dL, year first, from sunday", "Asia/Tokyo", "ja-JP", want{new("Asia/Tokyo"), new("ja-JP"), new("YYYY/MM/DD"), new("HH:mm"), new("sunday"), mgdl}},
		{"India reads mg/dL", "Asia/Kolkata", "hi-IN", want{new("Asia/Kolkata"), new("hi-IN"), new("DD/MM/YYYY"), new("hh:mm a"), new("sunday"), mgdl}},
		{"Israel reads mg/dL", "Asia/Jerusalem", "he-IL", want{new("Asia/Jerusalem"), new("he-IL"), new("DD/MM/YYYY"), new("HH:mm"), new("sunday"), mgdl}},
		{"Iran reads mg/dL", "Asia/Tehran", "fa-IR", want{new("Asia/Tehran"), new("fa-IR"), new("YYYY/MM/DD"), new("HH:mm"), new("monday"), mgdl}},
		{"New Zealand reads mmol/L on a 12 hour clock", "Pacific/Auckland", "en-NZ", want{new("Pacific/Auckland"), new("en-NZ"), new("DD/MM/YYYY"), new("hh:mm a"), new("monday"), mmol}},
		{"the UK reads mmol/L", "Europe/London", "en-GB", want{new("Europe/London"), new("en-GB"), new("DD/MM/YYYY"), new("HH:mm"), new("monday"), mmol}},
		{"Sweden writes dates as ISO", "Europe/Stockholm", "sv-SE", want{new("Europe/Stockholm"), new("sv-SE"), new("YYYY-MM-DD"), new("HH:mm"), new("monday"), mmol}},
		{"the country is from the language, not the time zone", "Pacific/Auckland", "en-US",
			want{new("Pacific/Auckland"), new("en-US"), new("MM/DD/YYYY"), new("hh:mm a"), new("sunday"), mgdl}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := BrowserSettings(tt.timezone, tt.language)
			got := want{m.GetTimezone(), m.GetLanguage(), m.GetDateFormat(), m.GetTimeFormat(), m.GetWeekStartDay(), m.GetUnits()}
			require.Equal(t, tt.want, got)
			// Every format set is one the settings accept
			if got.date != nil {
				require.NoError(t, DateFormat.ValidateFunc(*got.date))
				require.NoError(t, TimeFormat.ValidateFunc(*got.clock))
				require.NoError(t, WeekStartDay.ValidateFunc(*got.weekStart))
			}
		})
	}
}
