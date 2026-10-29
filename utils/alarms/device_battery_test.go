package alarms

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeviceBattery(t *testing.T) {
	now := int64(1_000 * minute)
	fresh, old := now-5*minute, now-30*minute
	battery := func(percent float64) *float64 { return &percent }
	tests := []struct {
		name     string
		statuses map[string]DeviceStatus
		active   bool
	}{
		{"a device that has never sent a status", nil, false},
		{"a fresh status over the percent", map[string]DeviceStatus{"phone": {At: fresh, BatteryPercent: battery(50)}}, false},
		{"a fresh status at the percent", map[string]DeviceStatus{"phone": {At: fresh, BatteryPercent: battery(20)}}, false},
		{"a fresh status under the percent", map[string]DeviceStatus{"phone": {At: fresh, BatteryPercent: battery(19)}}, true},
		{"under the percent while charging", map[string]DeviceStatus{"phone": {At: fresh, BatteryPercent: battery(5), Charging: true}}, false},
		{"a fresh status with no battery", map[string]DeviceStatus{"phone": {At: fresh}}, false},
		{"no status for its minutes", map[string]DeviceStatus{"phone": {At: old, BatteryPercent: battery(90)}}, true},
		{"no status for its minutes, charging when it was last heard from", map[string]DeviceStatus{"phone": {At: old, Charging: true}}, true},
		{"any one listed device", map[string]DeviceStatus{"phone": {At: fresh, BatteryPercent: battery(90)}, "pixel": {At: fresh, BatteryPercent: battery(10)}}, true},
		{"a device that is not listed", map[string]DeviceStatus{"other": {At: old, BatteryPercent: battery(1)}}, false},
	}
	for _, tt := range tests {
		alarm := DeviceBatteryAlarm{Percent: 20, Minutes: 30, Devices: []string{"phone", "pixel"}, Statuses: tt.statuses}
		require.Equal(t, tt.active, alarm.IsActive(nil, now), tt.name)
	}
}
