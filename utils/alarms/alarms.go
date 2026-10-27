// Package alarms decides, from a person's alarm settings and recent glucose readings, which alarms are on and when to
// notify about them. It does no I/O: its caller gives it the readings, the time and the state it returned last, keeps
// the new state and sends the notifications.
package alarms

import (
	"cmp"
	"maps"
	"slices"
)

// Type is one kind of alarm
type Type string

const (
	UrgentLow  Type = "urgent_low"
	Low        Type = "low"
	UrgentHigh Type = "urgent_high"
	High       Type = "high"
	NoData     Type = "no_data"
)

// Types are every alarm, each urgent one before the one it quiets, in the order they notify
var Types = []Type{UrgentLow, Low, UrgentHigh, High, NoData}

// louder is the urgent alarm that quiets an alarm while it is on, as an urgent low says all a low would
var louder = map[Type]Type{Low: UrgentLow, High: UrgentHigh}

// Level is how urgent an alarm is, numbered as Nightscout's are
type Level int

const (
	Warn   Level = 1
	Urgent Level = 2
)

const minute = int64(60_000)

// Level is Urgent for the urgent alarms and Warn for the others
func (t Type) Level() Level {
	if t == UrgentLow || t == UrgentHigh {
		return Urgent
	}
	return Warn
}

// RepeatMs is how long an alarm that stays on waits to notify again: 5 minutes when urgent, 15 otherwise
func (t Type) RepeatMs() int64 {
	if t.Level() == Urgent {
		return 5 * minute
	}
	return 15 * minute
}

// Threshold is an alarm's settings: the glucose in mg/dL it is under, for the lows, or over, for the highs, and the
// minutes that must last. No data has only minutes, how long there has been no reading.
type Threshold struct {
	Mgdl    float64
	Minutes int64
}

// Config is the alarms that are turned on, each with its threshold
type Config map[Type]Threshold

// Reading is a glucose reading in mg/dL at a time in milliseconds
type Reading struct {
	At   int64
	Mgdl float64
}

// Alarm is one alarm's state between checks, all times in milliseconds
type Alarm struct {
	// OnSince is when the alarm came on, or 0 while it is off
	OnSince int64
	// NotifiedAt is when it last notified. It is kept after the alarm goes off, so one that goes off and on again is
	// not notified sooner than it would have been had it stayed on.
	NotifiedAt int64
	// SnoozedUntil is when its snooze ends; until then it does not notify, even if it goes off and on again
	SnoozedUntil int64
}

// State is each alarm's state, which a check returns for the caller to give the next one
type State map[Type]Alarm

// Kind is what a notification says about its alarm
type Kind string

const (
	Started  Kind = "started"
	Repeated Kind = "repeated"
	Cleared  Kind = "cleared"
)

// Notification is news of an alarm to send
type Notification struct {
	Type Type
	Kind Kind
	At   int64
}

// Check is the alarms' state at now and the notifications to send, from the config, the last check's state and the
// recent readings, which must include the newest if there is any. It is called for each new reading, and periodically,
// so an alarm that waits on time, like no data or a repeat, notifies without a new reading.
//
// An alarm starts once its condition is met, repeats while it stays on and is not snoozed, and clears when it goes off
// if it notified since it came on. While an urgent alarm is on, the alarm it quiets stays on without notifying.
func Check(config Config, state State, readings []Reading, now int64) (State, []Notification) {
	readings = slices.Clone(readings)
	slices.SortStableFunc(readings, func(a, b Reading) int { return cmp.Compare(a.At, b.At) })
	next := State{}
	on := map[Type]bool{}
	for _, t := range Types {
		threshold, enabled := config[t]
		on[t] = enabled && isOn(t, threshold, readings, now)
	}

	var notifications []Notification
	for _, t := range Types {
		alarm := state[t]
		notifiedSinceOn := alarm.OnSince != 0 && alarm.NotifiedAt >= alarm.OnSince
		switch {
		case !on[t]:
			if notifiedSinceOn {
				notifications = append(notifications, Notification{Type: t, Kind: Cleared, At: now})
			}
			alarm.OnSince = 0
		case alarm.OnSince == 0:
			alarm.OnSince = now
		}
		quiet := !on[t] || on[louder[t]] || now < alarm.SnoozedUntil || now-alarm.NotifiedAt < t.RepeatMs()
		if !quiet {
			kind := Started
			if alarm.NotifiedAt >= alarm.OnSince {
				kind = Repeated
			}
			notifications = append(notifications, Notification{Type: t, Kind: kind, At: now})
			alarm.NotifiedAt = now
		}
		// An alarm that is off, and whose last notification and snooze no longer matter, is forgotten
		if alarm.OnSince != 0 || now < alarm.SnoozedUntil || now-alarm.NotifiedAt < t.RepeatMs() {
			next[t] = alarm
		}
	}
	return next, notifications
}

// isOn is whether an alarm's condition is met at now by the readings, sorted oldest first
func isOn(t Type, threshold Threshold, readings []Reading, now int64) bool {
	if len(readings) == 0 {
		return false
	}
	newest := readings[len(readings)-1]
	if t == NoData {
		return now-newest.At >= threshold.Minutes*minute
	}
	beyond := func(r Reading) bool {
		if t == UrgentLow || t == Low {
			return r.Mgdl < threshold.Mgdl
		}
		return r.Mgdl > threshold.Mgdl
	}
	// It has lasted since the oldest of the newest readings that are all beyond the threshold
	i := len(readings)
	for i > 0 && beyond(readings[i-1]) {
		i--
	}
	return i < len(readings) && newest.At-readings[i].At >= threshold.Minutes*minute
}

// Snooze quiets an alarm until a time, whether it is on or not, so the snooze also covers it coming back on; a time
// already past ends a snooze
func (s State) Snooze(t Type, until int64) State {
	next := State{}
	maps.Copy(next, s)
	alarm := next[t]
	alarm.SnoozedUntil = until
	next[t] = alarm
	return next
}
