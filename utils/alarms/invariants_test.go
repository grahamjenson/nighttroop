package alarms

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

// TestInvariants runs thousands of random steps, with every alarm turned on, and checks what must always hold after
// each:
//
//   - an alarm is active exactly when oracle, a slower and simpler reading of its condition, says it is
//   - an alarm is alerting exactly when it is active, not snoozed, and its louder alarm is not active
//   - an alerting alarm not notified within its repeat notifies, and only alerting alarms start or repeat
//   - each alarm's notifications go started, then repeated, then cleared, and none is left started once inactive
//   - an alarm never notifies sooner than its repeat
//
// A failure prints the timeline that made it, which can be pasted into a test in alarms_test.go to debug it.
func TestInvariants(t *testing.T) {
	const configText = "urgent_low 56, low 72 5m, urgent_high 270, high 180 30m, no_data 20m"
	config := parseConfig(t, configText)
	turnedOn := map[Type]Alarm{}
	for _, alarm := range config {
		turnedOn[alarm.Type()] = alarm
	}
	// quietedBy is which alarms quiet which, written out again here as what the alarms must do
	quietedBy := map[Type]Type{Low: UrgentLow, High: UrgentHigh}

	for seed := range uint64(200) {
		r := rand.New(rand.NewPCG(seed, 1))
		state := State{}
		var readings []Reading
		var lines []string
		readingAt := map[int64]bool{}
		started := map[Type]bool{}
		notifiedAt := map[Type]int64{}
		snoozedUntil := map[Type]int64{}
		now, mgdl := start, float64(40+r.IntN(260))

		for range 200 {
			// Mostly every few minutes, sometimes a gap long enough for no data
			now += int64(1+r.IntN(10)) * minute / 2
			if r.IntN(30) == 0 {
				now += 25 * minute
			}
			line := since(now) + " "
			switch n := r.IntN(20); {
			case n < 12:
				mgdl = max(39, min(400, mgdl+float64(r.IntN(41)-20)))
				if readingAt[now] {
					line += "check"
					break
				}
				readingAt[now] = true
				readings = append(readings, Reading{At: now, Mgdl: mgdl})
				line += fmt.Sprint(mgdl)
			case n < 14:
				// A backfilled reading for an earlier time, as a sync catching up sends
				at := max(start, now-int64(1+r.IntN(30))*minute)
				if readingAt[at] {
					line += "check"
					break
				}
				readingAt[at] = true
				value := float64(40 + r.IntN(260))
				readings = append(readings, Reading{At: at, Mgdl: value})
				line += fmt.Sprintf("%v at %s", value, since(at))
			case n < 15:
				alarm := Types[r.IntN(len(Types))]
				until := now + int64(r.IntN(60))*minute
				state = snooze(state, alarm, until)
				snoozedUntil[alarm] = until
				line += fmt.Sprintf("snooze %s %s", alarm, since(until-now+start))
			default:
				line += "check"
			}
			lines = append(lines, line)

			result := Check(config, state, readings, now)
			state = result.State
			fail := func(format string, args ...any) {
				t.Helper()
				t.Fatalf("seed %d: %s\n\nstate: %s\n\n{\n\tname:   \"seed %d\",\n\tconfig: %q,\n\ttimeline: `\n\t\t%s\n\t`,\n},",
					seed, fmt.Sprintf(format, args...), formatState(state), seed, configText, strings.Join(lines, "\n\t\t"))
			}

			active, alerting := map[Type]bool{}, map[Type]bool{}
			for _, alarm := range result.Active {
				active[alarm.Type()] = true
			}
			for _, alarm := range result.Alerting {
				alerting[alarm.Type()] = true
			}
			for _, alarm := range Types {
				if want := oracle(turnedOn[alarm], readings, now); active[alarm] != want {
					fail("%s is active %v, but the oracle says %v", alarm, active[alarm], want)
				}
				if got := state[alarm].ActiveSince != 0; got != active[alarm] {
					fail("%s's state is active %v, but the result says %v", alarm, got, active[alarm])
				}
				if want := active[alarm] && now >= snoozedUntil[alarm] && !active[quietedBy[alarm]]; alerting[alarm] != want {
					fail("%s is alerting %v, want %v", alarm, alerting[alarm], want)
				}
			}

			notified, sounded := map[Type]bool{}, map[Type]bool{}
			for _, notification := range result.Notifications {
				alarm := notification.Type
				if notified[alarm] {
					fail("%s notified twice in one check", alarm)
				}
				notified[alarm] = true
				switch notification.Kind {
				case Started:
					if started[alarm] {
						fail("%s started again without clearing", alarm)
					}
					started[alarm] = true
				case Repeated, Cleared:
					if !started[alarm] {
						fail("%s %s without starting", alarm, notification.Kind)
					}
					started[alarm] = notification.Kind == Repeated
				}
				if notification.Kind == Cleared {
					continue
				}
				sounded[alarm] = true
				if !alerting[alarm] {
					fail("%s %s while not alerting", alarm, notification.Kind)
				}
				if notifiedAt[alarm] != 0 && now-notifiedAt[alarm] < turnedOn[alarm].RepeatMs() {
					fail("%s %s %s after it last notified", alarm, notification.Kind, since(now-notifiedAt[alarm]+start))
				}
			}

			for _, alarm := range Types {
				due := alerting[alarm] && (notifiedAt[alarm] == 0 || now-notifiedAt[alarm] >= turnedOn[alarm].RepeatMs())
				if due && !sounded[alarm] {
					fail("%s is alerting and due, but did not notify", alarm)
				}
				if sounded[alarm] {
					notifiedAt[alarm] = now
				}
				if started[alarm] && !active[alarm] {
					fail("%s is not active, but was not cleared", alarm)
				}
			}
		}
	}
}

// oracle is whether an alarm's condition holds, worked out the slow way: for no data, whether the newest reading is
// at least its minutes old; for the others, whether some reading at least its minutes before the newest has only
// readings beyond the threshold from it on
func oracle(alarm Alarm, readings []Reading, now int64) bool {
	if alarm == nil || len(readings) == 0 {
		return false
	}
	newest := readings[0]
	for _, reading := range readings {
		if reading.At > newest.At {
			newest = reading
		}
	}
	var mgdl float64
	var minutes int64
	under := false
	switch a := alarm.(type) {
	case NoDataAlarm:
		return now-newest.At >= a.Minutes*minute
	case UrgentLowAlarm:
		mgdl, under = a.Mgdl, true
	case LowAlarm:
		mgdl, minutes, under = a.Mgdl, a.Minutes, true
	case UrgentHighAlarm:
		mgdl = a.Mgdl
	case HighAlarm:
		mgdl, minutes = a.Mgdl, a.Minutes
	}
	beyond := func(reading Reading) bool {
		if under {
			return reading.Mgdl < mgdl
		}
		return reading.Mgdl > mgdl
	}
	for _, from := range readings {
		if newest.At-from.At < minutes*minute {
			continue
		}
		all := true
		for _, reading := range readings {
			if reading.At >= from.At && !beyond(reading) {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	return false
}

// since is a time as a duration since the timeline started, as the rig writes it
func since(ms int64) string {
	return (time.Duration(ms-start) * time.Millisecond).String()
}
