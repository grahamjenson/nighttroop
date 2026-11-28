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
// alerting alarm says no time
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

// An alarm that is not alerting and has no snooze is quieted by a louder one, so the banner has no row for it:
// beside the louder one alerting, and beside it snoozed, where it would say the quieted one is snoozed
func TestAlarmBannerLeavesOutAQuietedAlarm(t *testing.T) {
	until := new(time.Date(2026, 1, 15, 1, 35, 20, 0, time.UTC).UnixMilli())
	high := &sitemodels.SiteAlarm{ID: "high", Active: true}
	tests := []struct {
		name   string
		louder sitemodels.SiteAlarm
		want   string
	}{
		{"its louder alarm is alerting", sitemodels.SiteAlarm{Alerting: true}, "Alarm: Urgent high<"},
		{"its louder alarm is snoozed", sitemodels.SiteAlarm{SnoozedUntil: until}, "Snoozed: Urgent high until 01:35<"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.louder.ID, tt.louder.Urgent, tt.louder.Active = "urgent_high", true, true
			var html bytes.Buffer
			require.NoError(t, AlarmBanner([]*sitemodels.SiteAlarm{&tt.louder, high}).Render(context.Background(), &html))
			require.Contains(t, html.String(), tt.want)
			require.NotContains(t, html.String(), "site--alarms--alarm-high")
			require.NotContains(t, html.String(), "High")
		})
	}
	var html bytes.Buffer
	require.NoError(t, AlarmRow(high).Render(context.Background(), &html))
	require.Empty(t, html.String())
}
