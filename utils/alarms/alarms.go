// Package alarms decides, from a person's alarms and recent glucose readings, which alarms are active, which are
// alerting, and when to notify about them. It does no I/O: its caller gives it the readings, the time and the state
// it returned last, keeps the new state and sends the notifications.
package alarms

import (
	"cmp"
	"slices"
)

// Alarm is one of a person's alarms with its settings, each kind its own struct, as LowAlarm
type Alarm interface {
	// Type names the kind of alarm, as low
	Type() Type
	// RepeatMs is how long it waits, while it stays alerting, to notify again
	RepeatMs() int64
	// QuietedBy is the louder alarm that says all this one would, so this one does not alert while that one is
	// active, or "" when there is none
	QuietedBy() Type
	// IsActive is whether its condition is met at now by the readings, sorted oldest first
	IsActive(readings []Reading, now int64) bool
}

// Type names a kind of alarm
type Type string

const (
	UrgentLow  Type = "urgent_low"
	Low        Type = "low"
	UrgentHigh Type = "urgent_high"
	High       Type = "high"
	FastRise   Type = "fast_rise"
	FastDrop   Type = "fast_drop"
	NoData     Type = "no_data"
	// DeviceBattery is a listed device's battery being low, or the device not reporting
	DeviceBattery Type = "device_battery"
)

// Types are every kind of alarm, each louder one before the one it quiets, the order results list them in
var Types = []Type{UrgentLow, Low, UrgentHigh, High, FastRise, FastDrop, NoData, DeviceBattery}

// Level is how urgent an alarm is, numbered as Nightscout's are
type Level int

const (
	Warn   Level = 1
	Urgent Level = 2
)

// Level is Urgent for the urgent low and high, and Warn for the others
func (t Type) Level() Level {
	if t == UrgentLow || t == UrgentHigh {
		return Urgent
	}
	return Warn
}

const minute = int64(60_000)

// UrgentLowAlarm is active while the newest reading is under Mgdl, and alerts every 5 minutes
type UrgentLowAlarm struct{ Mgdl float64 }

func (UrgentLowAlarm) Type() Type      { return UrgentLow }
func (UrgentLowAlarm) RepeatMs() int64 { return 5 * minute }
func (UrgentLowAlarm) QuietedBy() Type { return "" }
func (a UrgentLowAlarm) IsActive(readings []Reading, now int64) bool {
	return lasted(readings, 0, func(r Reading) bool { return r.Mgdl < a.Mgdl })
}

// LowAlarm is active once readings have been under Mgdl for Minutes, alerts every 15 minutes, and is quiet while the
// urgent low is active
type LowAlarm struct {
	Mgdl    float64
	Minutes int64
}

func (LowAlarm) Type() Type      { return Low }
func (LowAlarm) RepeatMs() int64 { return 15 * minute }
func (LowAlarm) QuietedBy() Type { return UrgentLow }
func (a LowAlarm) IsActive(readings []Reading, now int64) bool {
	return lasted(readings, a.Minutes, func(r Reading) bool { return r.Mgdl < a.Mgdl })
}

// UrgentHighAlarm is active while the newest reading is over Mgdl, and alerts every 5 minutes
type UrgentHighAlarm struct{ Mgdl float64 }

func (UrgentHighAlarm) Type() Type      { return UrgentHigh }
func (UrgentHighAlarm) RepeatMs() int64 { return 5 * minute }
func (UrgentHighAlarm) QuietedBy() Type { return "" }
func (a UrgentHighAlarm) IsActive(readings []Reading, now int64) bool {
	return lasted(readings, 0, func(r Reading) bool { return r.Mgdl > a.Mgdl })
}

// HighAlarm is active once readings have been over Mgdl for Minutes, alerts every 15 minutes, and is quiet while the
// urgent high is active
type HighAlarm struct {
	Mgdl    float64
	Minutes int64
}

func (HighAlarm) Type() Type      { return High }
func (HighAlarm) RepeatMs() int64 { return 15 * minute }
func (HighAlarm) QuietedBy() Type { return UrgentHigh }
func (a HighAlarm) IsActive(readings []Reading, now int64) bool {
	return lasted(readings, a.Minutes, func(r Reading) bool { return r.Mgdl > a.Mgdl })
}

// FastRiseAlarm is active once glucose has risen by Mgdl or more over the newest readings within Minutes, and alerts
// every 15 minutes
type FastRiseAlarm struct {
	Mgdl    float64
	Minutes int64
}

func (FastRiseAlarm) Type() Type      { return FastRise }
func (FastRiseAlarm) RepeatMs() int64 { return 15 * minute }
func (FastRiseAlarm) QuietedBy() Type { return "" }
func (a FastRiseAlarm) IsActive(readings []Reading, now int64) bool {
	return changed(readings, a.Minutes) >= a.Mgdl
}

// FastDropAlarm is active once glucose has fallen by Mgdl or more over the newest readings within Minutes, and alerts
// every 15 minutes
type FastDropAlarm struct {
	Mgdl    float64
	Minutes int64
}

func (FastDropAlarm) Type() Type      { return FastDrop }
func (FastDropAlarm) RepeatMs() int64 { return 15 * minute }
func (FastDropAlarm) QuietedBy() Type { return "" }
func (a FastDropAlarm) IsActive(readings []Reading, now int64) bool {
	return -changed(readings, a.Minutes) >= a.Mgdl
}

// NoDataAlarm is active once there has been no reading for Minutes, after at least one, and alerts every 15 minutes
type NoDataAlarm struct{ Minutes int64 }

func (NoDataAlarm) Type() Type      { return NoData }
func (NoDataAlarm) RepeatMs() int64 { return 15 * minute }
func (NoDataAlarm) QuietedBy() Type { return "" }
func (a NoDataAlarm) IsActive(readings []Reading, now int64) bool {
	return len(readings) > 0 && now-readings[len(readings)-1].At >= a.Minutes*minute
}

// DeviceBatteryAlarm is active while any of its Devices, by its newest status, is under Percent battery and not
// charging, or has sent no status for Minutes. A device that has never sent one is not alarmed for, as no data is
// not. It alerts every 30 minutes.
type DeviceBatteryAlarm struct {
	Percent float64
	Minutes int64
	Devices []string
	// Statuses are the newest status of each of its devices that has sent one, which its checker gives it
	Statuses map[string]DeviceStatus
}

// DeviceStatus is what a device last said of itself, at a time in milliseconds
type DeviceStatus struct {
	At             int64
	BatteryPercent *float64
	Charging       bool
}

func (DeviceBatteryAlarm) Type() Type      { return DeviceBattery }
func (DeviceBatteryAlarm) RepeatMs() int64 { return 30 * minute }
func (DeviceBatteryAlarm) QuietedBy() Type { return "" }
func (a DeviceBatteryAlarm) IsActive(_ []Reading, now int64) bool {
	for _, device := range a.Devices {
		status, ok := a.Statuses[device]
		if !ok {
			continue
		}
		low := status.BatteryPercent != nil && *status.BatteryPercent < a.Percent && !status.Charging
		if low || now-status.At >= a.Minutes*minute {
			return true
		}
	}
	return false
}

// lasted is whether the newest readings, sorted oldest first, have all been beyond a threshold for minutes: from the
// oldest of the newest readings that are all beyond it to the newest. It is measured by readings, not the time now,
// so a check without a new reading cannot make it last.
func lasted(readings []Reading, minutes int64, beyond func(Reading) bool) bool {
	i := len(readings)
	for i > 0 && beyond(readings[i-1]) {
		i--
	}
	return i < len(readings) && readings[len(readings)-1].At-readings[i].At >= minutes*minute
}

// changed is how much glucose changed over the newest readings, sorted oldest first, within minutes of the newest:
// from the oldest of them to the newest, or 0 with only one. Like lasted, it is measured by readings, not the time now.
func changed(readings []Reading, minutes int64) float64 {
	if len(readings) == 0 {
		return 0
	}
	newest, i := readings[len(readings)-1], len(readings)-1
	for i > 0 && newest.At-readings[i-1].At <= minutes*minute {
		i--
	}
	return newest.Mgdl - readings[i].Mgdl
}

// Reading is a glucose reading in mg/dL at a time in milliseconds
type Reading struct {
	At   int64
	Mgdl float64
}

// AlarmState is one kind of alarm's state between checks, all times in milliseconds
type AlarmState struct {
	// ActiveSince is when the alarm became active, or 0 while it is not
	ActiveSince int64
	// NotifiedAt is when it last notified. It is kept after the alarm stops being active, so one that stops and
	// starts again is not notified sooner than it would have been had it stayed active.
	NotifiedAt int64
	// SnoozedUntil is when its snooze ends; until then it does not alert, even if it stops and starts again
	SnoozedUntil int64
}

// State is each kind of alarm's state, which a check returns for the caller to give the next one
type State map[Type]AlarmState

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

// Result is what a check found
type Result struct {
	// State is for the caller to give the next check
	State State
	// Active are the alarms whose condition is met, alerting or not
	Active []Alarm
	// Alerting are the active alarms that are not snoozed or quieted by a louder one, which should be heard
	Alerting []Alarm
	// Notifications are the news to send now: an alerting alarm starts, then repeats while it stays alerting, then
	// clears when it stops being active, if it notified since it became active
	Notifications []Notification
}

// Check is the alarms at now, from the alarms turned on, the last check's state, and the recent readings, which must
// include the newest if there is any. It is called for each new reading, and periodically, so an alarm that waits on
// time, like no data or a repeat, notifies without a new reading.
func Check(alarms []Alarm, state State, readings []Reading, now int64) Result {
	readings = slices.Clone(readings)
	slices.SortStableFunc(readings, func(a, b Reading) int { return cmp.Compare(a.At, b.At) })
	turnedOn := map[Type]Alarm{}
	active := map[Type]bool{}
	for _, alarm := range alarms {
		turnedOn[alarm.Type()] = alarm
		active[alarm.Type()] = alarm.IsActive(readings, now)
	}

	result := Result{State: State{}}
	for _, t := range Types {
		alarm, s := turnedOn[t], state[t]
		notifiedSinceActive := s.ActiveSince != 0 && s.NotifiedAt >= s.ActiveSince
		if !active[t] {
			if notifiedSinceActive {
				result.Notifications = append(result.Notifications, Notification{Type: t, Kind: Cleared, At: now})
			}
			s.ActiveSince = 0
		} else {
			if s.ActiveSince == 0 {
				s.ActiveSince = now
			}
			result.Active = append(result.Active, alarm)
			if now >= s.SnoozedUntil && !active[alarm.QuietedBy()] {
				result.Alerting = append(result.Alerting, alarm)
				if now-s.NotifiedAt >= alarm.RepeatMs() {
					kind := Started
					if notifiedSinceActive {
						kind = Repeated
					}
					result.Notifications = append(result.Notifications, Notification{Type: t, Kind: kind, At: now})
					s.NotifiedAt = now
				}
			}
		}
		// An alarm that is not active, and whose last notification and snooze no longer matter, is forgotten
		if s.ActiveSince != 0 || now < s.SnoozedUntil || (alarm != nil && now-s.NotifiedAt < alarm.RepeatMs()) {
			result.State[t] = s
		}
	}
	return result
}
