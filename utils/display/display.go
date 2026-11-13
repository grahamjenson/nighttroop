// Package display shows stored values the way people read them
package display

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/i18n"
)

// Layouts a day is named with: in full, as "Monday 2 January 2006", short with its year, as "Mon 2 Jan 2006", and
// short without it, as "Mon 2 Jan"
const (
	LongDayLayout  = "Monday 2 January 2006"
	DayLayout      = "Mon 2 Jan 2006"
	ShortDayLayout = "Mon 2 Jan"
)

// Value is an optional number with its unit, rounded to 2 places in the context's number format, or a dash when it
// is not set
func Value(ctx context.Context, v *float64, unit string) string {
	if v == nil {
		return "–"
	}
	return Number(ctx, strconv.FormatFloat(consts.Round(*v, 2), 'f', -1, 64)) + unit
}

// PerDay is a period's daily dose or carbs with its unit, or a dash when none of its days had data, so a period with
// no data does not read as a real 0
func PerDay(ctx context.Context, days int64, v float64, unit string) string {
	if days == 0 {
		return "–"
	}
	return Value(ctx, &v, unit)
}

// Ago is how long before now the time in ms was, as "just now", "5 min ago", "3 h ago" or "2 days ago",
// or "Never" when there is no time, in the context's language
func Ago(ctx context.Context, ms *int64, now time.Time) string {
	if ms == nil || *ms == 0 {
		return i18n.T(ctx, "common.time.never")
	}
	since := now.Sub(time.UnixMilli(*ms))
	switch {
	case since < time.Minute:
		return i18n.T(ctx, "common.time.just_now")
	case since < time.Hour:
		return i18n.T(ctx, "common.time.min_ago", i18n.Data{"Count": int(since.Minutes())})
	case since < 48*time.Hour:
		return i18n.T(ctx, "common.time.h_ago", i18n.Data{"Count": int(since.Hours())})
	}
	return i18n.T(ctx, "common.time.days_ago", i18n.Data{"Count": int(since.Hours() / 24)})
}

// DateTime is the time in ms as its date and time in UTC, as "2026-10-26 12:00:00", or "-" when there is no time,
// as with the 0 a job has before it is fetched or finished
func DateTime(ms *int64) string {
	if ms == nil || *ms == 0 {
		return "-"
	}
	return time.UnixMilli(*ms).UTC().Format(time.DateTime)
}

// TimeAgo is how long before now the time in ms was, in whole minutes, then hours and minutes, then days, hours and
// minutes: "5m ago", "1h5m ago" or "2d1h5m ago". A time after now, as from a phone's clock running fast, is "0m ago".
func TimeAgo(ctx context.Context, ms int64, now time.Time) string {
	minutes := max(int(now.Sub(time.UnixMilli(ms)).Minutes()), 0)
	parts := i18n.Data{"Days": minutes / (24 * 60), "Hours": minutes / 60 % 24, "Minutes": minutes % 60}
	switch {
	case minutes >= 24*60:
		return i18n.T(ctx, "common.time.days_hours_minutes_ago", parts)
	case minutes >= 60:
		return i18n.T(ctx, "common.time.hours_minutes_ago", parts)
	}
	return i18n.T(ctx, "common.time.minutes_ago", parts)
}

// RoundedAgo is how long before now the time in ms was, in words, and to the nearest hour once it is an hour or more:
// "just now", "1 minute ago", "45 minutes ago", "3 hours ago", "1 day ago" or "2 days 1 hour ago". A time after now,
// as from a phone's clock running fast, is "just now".
func RoundedAgo(ctx context.Context, ms int64, now time.Time) string {
	since := now.Sub(time.UnixMilli(ms))
	ago := func(how string) string { return i18n.T(ctx, "common.time.ago", i18n.Data{"HowLong": how}) }
	switch {
	case since < time.Minute:
		return i18n.T(ctx, "common.time.just_now")
	case since < time.Hour:
		return ago(i18n.N(ctx, "common.time.minutes", int(since.Minutes())))
	}
	hours := int(since.Round(time.Hour).Hours())
	days := hours / 24
	switch {
	case days == 0:
		return ago(i18n.N(ctx, "common.time.hours", hours))
	case hours%24 == 0:
		return ago(i18n.N(ctx, "common.time.days", days))
	}
	return ago(i18n.T(ctx, "common.time.days_and_hours", i18n.Data{"Days": i18n.N(ctx, "common.time.days", days),
		"Hours": i18n.N(ctx, "common.time.hours", hours%24)}))
}

// Duration is how long ms milliseconds is, as "250 ms", "1.5 s", "2 min 5 s" or "3 h 20 min", leaving off a
// second part of 0, as in "2 min" or "3 h"
func Duration(ctx context.Context, ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	switch {
	case d < time.Second:
		return i18n.T(ctx, "common.duration.milliseconds", i18n.Data{"Count": ms})
	case d < time.Minute:
		return i18n.T(ctx, "common.duration.seconds", i18n.Data{"Count": fmt.Sprintf("%.1f", d.Seconds())})
	case d < time.Hour:
		return withPart(ctx, "common.duration.minutes", "common.duration.minutes_and_seconds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return withPart(ctx, "common.duration.hours", "common.duration.hours_and_minutes", int(d.Hours()), int(d.Minutes())%60)
}

// withPart is a count of a unit, then of the next smaller unit unless it is 0, by the message ids of the unit alone
// and of both
func withPart(ctx context.Context, alone, both string, count int, part int) string {
	if part == 0 {
		return i18n.T(ctx, alone, i18n.Data{"Count": count})
	}
	return i18n.T(ctx, both, i18n.Data{"Count": count, "Part": part})
}

// ReadingType is what took a glucose reading: Sensor for sgv and Meter for mbg
func ReadingType(ctx context.Context, readingType string) string {
	switch readingType {
	case "sgv":
		return i18n.T(ctx, "common.reading_type.sensor")
	case "mbg":
		return i18n.T(ctx, "common.reading_type.meter")
	}
	return readingType
}

// directions are Nightscout's trend directions as arrows
var directions = map[string]string{
	"DoubleUp":      "⇈",
	"SingleUp":      "↑",
	"FortyFiveUp":   "↗",
	"Flat":          "→",
	"FortyFiveDown": "↘",
	"SingleDown":    "↓",
	"DoubleDown":    "⇊",
}

// Direction is a trend direction as an arrow, a direction with no arrow as it is, and "-" for none
func Direction(direction *string) string {
	if direction == nil {
		return "-"
	}
	if arrow, ok := directions[*direction]; ok {
		return arrow
	}
	return *direction
}

// Arrow is a trend direction as an arrow, or "" for none or one with no arrow, for beside a reading
func Arrow(direction *string) string {
	if direction == nil {
		return ""
	}
	return directions[*direction]
}

// Bytes is a count of bytes as B, KB, MB or GB, in 1024s
func Bytes(bytes float64) string {
	for _, unit := range []string{"B", "KB", "MB"} {
		if bytes < 1024 {
			return fmt.Sprintf("%.1f %s", bytes, unit)
		}
		bytes /= 1024
	}
	return fmt.Sprintf("%.1f GB", bytes)
}
