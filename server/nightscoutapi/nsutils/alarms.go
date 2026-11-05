package nsutils

import (
	"cmp"
	"context"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/httputils"
)

const (
	// defaultSilence is how long an ack without a silence time snoozes for, Nightscout's shortest choice
	defaultSilence = 30 * time.Minute
	// maxSilence is the longest an ack snoozes for, so one cannot silence an alarm for good
	maxSilence = 24 * time.Hour
)

// AlarmGroup is Nightscout's group for an alarm: Time Ago for no data, and default for the glucose alarms
func AlarmGroup(t alarms.Type) string {
	if t == alarms.NoData {
		return "Time Ago"
	}
	return "default"
}

// SnoozeAlarms is Nightscout's ack of a level, group and silence time in milliseconds, over Socket.IO or the v1 API:
// it snoozes the identity's alarms in that group up to that level, which the alarm checker stops alerting on its
// next check, and records the snooze in the logs DB's events. It is a write, so within the site's write limit, as
// each records an event; it is false when over it.
func SnoozeAlarms(siteContext *httputils.SiteContext, identityID int64, level alarms.Level, group string, silenceMs int64) bool {
	controlDB, siteID := siteContext.ControlDB, siteContext.SiteID
	if allowed, _ := httputils.WithinRateLimitAll(siteContext.Config.RateLimiters.SitePOSTLimiter(siteID)); !allowed {
		return false
	}
	silence := min(cmp.Or(time.Duration(max(silenceMs, 0))*time.Millisecond, defaultSilence), maxSilence)
	until := time.Now().Add(silence).UnixMilli()
	var snoozed []string
	for _, t := range alarms.Types {
		if t.Level() <= level && AlarmGroup(t) == group {
			snooze := livedb.SnoozeAlarmParams{SiteID: siteID, IdentityID: identityID, Type: t, Until: until}
			if err := controlDB.Live.Queries.SnoozeAlarm(context.Background(), snooze); err != nil {
				slog.Error("cannot snooze alarm", "error", err)
				continue
			}
			snoozed = append(snoozed, string(t))
		}
	}
	if len(snoozed) > 0 {
		controlDB.Logs.AlarmSnoozedEvent(context.Background(), siteID, identityID, snoozed, until, "nightscout")
	}
	return true
}
