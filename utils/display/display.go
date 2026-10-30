// Package display shows stored values the way people read them
package display

import (
	"fmt"
	"strings"
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

// TimeAgo is how long before now the time in ms was, in whole minutes, then hours and minutes, then days, hours and
// minutes: "5m ago", "1h5m ago" or "2d1h5m ago". A time after now, as from a phone's clock running fast, is "0m ago".
func TimeAgo(ms int64, now time.Time) string {
	minutes := max(int(now.Sub(time.UnixMilli(ms)).Minutes()), 0)
	days, hours := minutes/(24*60), minutes/60%24
	switch {
	case days > 0:
		return fmt.Sprintf("%dd%dh%dm ago", days, hours, minutes%60)
	case hours > 0:
		return fmt.Sprintf("%dh%dm ago", hours, minutes%60)
	}
	return fmt.Sprintf("%dm ago", minutes)
}

// Duration is how long ms milliseconds is, as "250 ms", "1.5 s", "2 min 5 s" or "3 h 20 min", leaving off a
// second part of 0, as in "2 min" or "3 h"
func Duration(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	switch {
	case d < time.Second:
		return fmt.Sprintf("%d ms", ms)
	case d < time.Minute:
		return fmt.Sprintf("%.1f s", d.Seconds())
	case d < time.Hour:
		return withPart(int(d.Minutes()), "min", int(d.Seconds())%60, "s")
	}
	return withPart(int(d.Hours()), "h", int(d.Minutes())%60, "min")
}

// withPart is a count of a unit, then of the next smaller unit unless it is 0
func withPart(count int, unit string, part int, partUnit string) string {
	if part == 0 {
		return fmt.Sprintf("%d %s", count, unit)
	}
	return fmt.Sprintf("%d %s %d %s", count, unit, part, partUnit)
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

// Name is a stored snake_case name as words, e.g. "site_created" as "Site created"
func Name(name string) string {
	words := strings.ReplaceAll(name, "_", " ")
	if words == "" {
		return words
	}
	return strings.ToUpper(words[:1]) + words[1:]
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
