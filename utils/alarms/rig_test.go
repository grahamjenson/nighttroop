package alarms

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The rig runs timelines. A timeline is a config and steps, one a line, each followed by a check, as a new reading or
// the periodic check would be. A step is its time since the timeline started, what happens then, and, after "->",
// the notifications that check must send, where nothing means it must send none. "on:" at the end lists the alarms
// that must be on after it, notifying or not, and "on:" alone means none. "//" starts a comment.
//
// What can happen at a step:
//
//	70                     a reading of 70 mg/dL at the step's time
//	70 at 3m               a reading of 70 mg/dL for another time, as a backfilled, or clock skewed, one would be
//	check                  nothing new, as the periodic check
//	snooze low 30m         the low alarm is snoozed for 30 minutes from the step's time; 0m ends a snooze
//	config low 72 5m       the config changes to this, written as a timeline's config is; config alone turns all off
//
// A config is each alarm turned on, separated by commas, with its mg/dL and its minutes, as
// "urgent_low 56, low 72 5m, no_data 20m". Notifications are each "<alarm> <kind>", separated by commas, in any order,
// as "urgent_low started, low cleared".
//
// For example:
//
//	{
//		name:   "a low starts once it has lasted, repeats, then clears",
//		config: "low 72 5m",
//		timeline: `
//			0m   70
//			5m   70     -> low started
//			20m  check  -> low repeated
//			25m  80     -> low cleared   on:
//		`,
//	},
type timeline struct {
	name     string
	config   string
	timeline string
}

// start is when timelines start, as an alarm whose times are 0 is off
var start = time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC).UnixMilli()

func runTimelines(t *testing.T, timelines []timeline) {
	t.Helper()
	for _, tl := range timelines {
		t.Run(tl.name, func(t *testing.T) {
			runTimeline(t, parseConfig(t, tl.config), tl.timeline)
		})
	}
}

func runTimeline(t *testing.T, config Config, text string) {
	t.Helper()
	state := State{}
	var readings []Reading
	for n, line := range strings.Split(text, "\n") {
		step, _, _ := strings.Cut(line, "//")
		step, onText, checkOn := strings.Cut(step, "on:")
		step, wantText, _ := strings.Cut(step, "->")
		fields := strings.Fields(step)
		if len(fields) == 0 {
			continue
		}
		fail := func(format string, args ...any) {
			t.Helper()
			t.Fatalf("line %d %q: %s\n\nstate: %s", n, strings.TrimSpace(line), fmt.Sprintf(format, args...), formatState(state))
		}
		if len(fields) < 2 {
			fail("a step needs a time and what happens")
		}
		now := start + parseDuration(t, fields[0]).Milliseconds()
		switch fields[1] {
		case "check":
		case "snooze":
			if len(fields) != 4 {
				fail("snooze needs an alarm and a duration")
			}
			state = state.Snooze(parseType(t, fields[2]), now+parseDuration(t, fields[3]).Milliseconds())
		case "config":
			config = parseConfig(t, strings.Join(fields[2:], " "))
		default:
			mgdl, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				fail("%q is not a reading, check, snooze or config", fields[1])
			}
			at := now
			if len(fields) == 4 && fields[2] == "at" {
				at = start + parseDuration(t, fields[3]).Milliseconds()
			} else if len(fields) != 2 {
				fail("a reading is its mg/dL, then optionally at and a time")
			}
			readings = append(readings, Reading{At: at, Mgdl: mgdl})
		}

		var notifications []Notification
		state, notifications = Check(config, state, readings, now)
		for _, notification := range notifications {
			if notification.At != now {
				fail("%s %s is at %d, not the check's time %d", notification.Type, notification.Kind, notification.At, now)
			}
		}
		if got, want := formatNotifications(notifications), sortedList(wantText); got != want {
			fail("sent %s, want %s", got, want)
		}
		if checkOn {
			if got, want := formatOn(state), sortedList(onText); got != want {
				fail("on: %s, want %s", got, want)
			}
		}
	}
}

// parseConfig is a config written as "urgent_low 56, low 72 5m, no_data 20m"
func parseConfig(t *testing.T, text string) Config {
	t.Helper()
	config := Config{}
	for part := range strings.SplitSeq(text, ",") {
		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}
		threshold := Threshold{}
		for _, field := range fields[1:] {
			if mgdl, err := strconv.ParseFloat(field, 64); err == nil {
				threshold.Mgdl = mgdl
			} else {
				threshold.Minutes = int64(parseDuration(t, field) / time.Minute)
			}
		}
		config[parseType(t, fields[0])] = threshold
	}
	return config
}

func parseType(t *testing.T, text string) Type {
	t.Helper()
	if !slices.Contains(Types, Type(text)) {
		t.Fatalf("%q is not an alarm, which are %v", text, Types)
	}
	return Type(text)
}

func parseDuration(t *testing.T, text string) time.Duration {
	t.Helper()
	d, err := time.ParseDuration(text)
	if err != nil {
		t.Fatalf("%q is not a duration, as 5m or 1h30m: %v", text, err)
	}
	return d
}

// sortedList is a list separated by commas, sorted and joined again, or nothing when it is empty
func sortedList(text string) string {
	var items []string
	for item := range strings.SplitSeq(text, ",") {
		if item = strings.Join(strings.Fields(item), " "); item != "" {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return "nothing"
	}
	slices.Sort(items)
	return strings.Join(items, ", ")
}

func formatNotifications(notifications []Notification) string {
	items := make([]string, len(notifications))
	for i, notification := range notifications {
		items[i] = string(notification.Type) + " " + string(notification.Kind)
	}
	return sortedList(strings.Join(items, ","))
}

func formatOn(state State) string {
	var on []string
	for t, alarm := range state {
		if alarm.OnSince != 0 {
			on = append(on, string(t))
		}
	}
	return sortedList(strings.Join(on, ","))
}

// formatState is each alarm's times since the timeline started, for a failure
func formatState(state State) string {
	since := func(ms int64) string {
		if ms == 0 {
			return "-"
		}
		return (time.Duration(ms-start) * time.Millisecond).String()
	}
	var items []string
	for _, t := range Types {
		if alarm, ok := state[t]; ok {
			items = append(items, fmt.Sprintf("%s{on since %s, notified at %s, snoozed until %s}", t, since(alarm.OnSince), since(alarm.NotifiedAt), since(alarm.SnoozedUntil)))
		}
	}
	if len(items) == 0 {
		return "nothing"
	}
	return strings.Join(items, "; ")
}
