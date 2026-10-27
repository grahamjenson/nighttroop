// Package display shows stored values the way people read them
package display

import (
	"fmt"
	"time"
)

// Ago is how long before now the time in ms was, as "just now", "5 min ago", "3 h ago" or "2 days ago",
// or "Never" when there is no time
func Ago(ms *int64, now time.Time) string {
	if ms == nil || *ms == 0 {
		return "Never"
	}
	since := now.Sub(time.UnixMilli(*ms))
	switch {
	case since < time.Minute:
		return "just now"
	case since < time.Hour:
		return fmt.Sprintf("%d min ago", int(since.Minutes()))
	case since < 48*time.Hour:
		return fmt.Sprintf("%d h ago", int(since.Hours()))
	}
	return fmt.Sprintf("%d days ago", int(since.Hours()/24))
}

// Duration is how long ms milliseconds is, as "250 ms", "1.5 s", "2 min 5 s" or "3 h 20 min"
func Duration(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	switch {
	case d < time.Second:
		return fmt.Sprintf("%d ms", ms)
	case d < time.Minute:
		return fmt.Sprintf("%.1f s", d.Seconds())
	case d < time.Hour:
		return fmt.Sprintf("%d min %d s", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%d h %d min", int(d.Hours()), int(d.Minutes())%60)
}

// ReadingType is what took a glucose reading: Sensor for sgv and Meter for mbg
func ReadingType(readingType string) string {
	switch readingType {
	case "sgv":
		return "Sensor"
	case "mbg":
		return "Meter"
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
