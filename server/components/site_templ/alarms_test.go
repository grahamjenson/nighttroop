package site_templ

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/settings"
)

// A snoozed alarm's row says when its snooze ends, in the settings' timezone and time format with no seconds, and an
// alarm that is quiet with no snooze, or alerting, says no time
func TestAlarmRowSaysWhenASnoozeEnds(t *testing.T) {
	until := new(time.Date(2026, 1, 15, 1, 35, 20, 0, time.UTC).UnixMilli())
	site := func(s settings.Settings) context.Context {
		return httputils.CreateSiteContext(context.Background(), nil, nil, nil, &controldb.Site{SettingsJson: settings.StoredOf(s)}, "site", "", 0, 0, "")
	}
	tests := []struct {
		name  string
		ctx   context.Context
		alarm sitemodels.SiteAlarm
		want  string
	}{
		{"the default settings are UTC on a 24 hour clock", site(settings.Settings{}), sitemodels.SiteAlarm{SnoozedUntil: until}, "Snoozed: Urgent high until 01:35<"},
		{"the site's timezone", site(settings.Settings{Timezone: new("Pacific/Auckland")}), sitemodels.SiteAlarm{SnoozedUntil: until}, "Snoozed: Urgent high until 14:35<"},
		{"a 12 hour clock", site(settings.Settings{Timezone: new("Pacific/Auckland"), TimeFormat: new("hh:mm:ss a")}), sitemodels.SiteAlarm{SnoozedUntil: until}, "Snoozed: Urgent high until 02:35 PM<"},
		{"with no site, as when rendered alone", context.Background(), sitemodels.SiteAlarm{SnoozedUntil: until}, "Snoozed: Urgent high until 01:35<"},
		{"quiet with no snooze", site(settings.Settings{}), sitemodels.SiteAlarm{}, "Snoozed: Urgent high<"},
		{"alerting", site(settings.Settings{}), sitemodels.SiteAlarm{Alerting: true}, "Alarm: Urgent high<"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.alarm.ID = "urgent_high"
			var html bytes.Buffer
			require.NoError(t, AlarmRow(&tt.alarm).Render(tt.ctx, &html))
			require.Contains(t, html.String(), tt.want)
		})
	}
}
