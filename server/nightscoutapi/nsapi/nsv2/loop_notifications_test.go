package nsv2

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Each remote command is the push Nightscout sends Loop, with the fields Loop reads, and a command Loop could not act
// on, or one far outside real use, is refused before it is pushed
func TestLoopCommandAlert(t *testing.T) {
	now := time.Date(2026, 11, 12, 6, 0, 0, 0, time.UTC)
	sent := map[string]any{"remote-address": "carer", "sent-at": "2026-11-12T06:00:00.000Z", "expiration": "2026-11-12T06:05:00.000Z"}
	with := func(fields map[string]any) map[string]any {
		data := map[string]any{"entered-by": "carer"}
		for k, v := range sent {
			data[k] = v
		}
		for k, v := range fields {
			data[k] = v
		}
		return data
	}

	tests := []struct {
		name, body, title, alertBody, err string
		data                              map[string]any
	}{
		// As NightscoutKit sends them, every value a string
		{name: "bolus", body: `{"eventType": "Remote Bolus Entry", "remoteBolus": "1.5", "otp": "123456"}`,
			title: "Remote Bolus Entry: 1.5 U", alertBody: "carer", data: with(map[string]any{"bolus-entry": 1.5, "otp": "123456"})},
		{name: "carbs", body: `{"eventType": "Remote Carbs Entry", "remoteCarbs": "20.0", "remoteAbsorption": "2.5", "otp": "654321", "created_at": "2026-11-12T05:55:00.000Z"}`,
			title: "Remote Carbs Entry: 20 grams", alertBody: "Absorption Time: 2.5 hours - carer",
			data: with(map[string]any{"carbs-entry": 20.0, "absorption-time": 2.5, "otp": "654321", "start-time": "2026-11-12T05:55:00.000Z"})},
		{name: "carbs absorb in 3 hours unless said", body: `{"eventType": "Remote Carbs Entry", "remoteCarbs": 10, "remoteAbsorption": "0"}`,
			title: "Remote Carbs Entry: 10 grams", alertBody: "Absorption Time: 3 hours - carer", data: with(map[string]any{"carbs-entry": 10.0, "absorption-time": 3.0})},
		{name: "carbs with a food type", body: `{"eventType": "Remote Carbs Entry", "remoteCarbs": "15", "remoteAbsorption": "2", "foodType": "🍕"}`,
			title: "Remote Carbs Entry: 15 grams", alertBody: "Absorption Time: 2 hours - carer", data: with(map[string]any{"carbs-entry": 15.0, "absorption-time": 2.0, "food-type": "🍕"})},
		{name: "override", body: `{"eventType": "Temporary Override", "reason": "Running", "reasonDisplay": "🏃 Running", "duration": "60.0", "notes": ""}`,
			title: "🏃 Running Temporary Override", alertBody: "carer", data: with(map[string]any{"override-name": "Running", "override-duration-minutes": int64(60)})},
		{name: "override until cancelled", body: `{"eventType": "Temporary Override", "reason": "Running", "duration": "0"}`,
			title: "Running Temporary Override", alertBody: "carer", data: with(map[string]any{"override-name": "Running"})},
		{name: "cancel override", body: `{"eventType": "Temporary Override Cancel", "duration": "0"}`,
			title: "Cancel Temporary Override", alertBody: "carer", data: with(map[string]any{"cancel-temporary-override": "true"})},
		{name: "notes and who entered it", body: `{"eventType": "Remote Bolus Entry", "remoteBolus": 2, "notes": "lunch", "enteredBy": "LoopFollow"}`,
			title: "Remote Bolus Entry: 2 U", alertBody: "lunch - LoopFollow", data: with(map[string]any{"bolus-entry": 2.0, "notes": "lunch", "entered-by": "LoopFollow"})},

		{name: "no bolus", body: `{"eventType": "Remote Bolus Entry", "otp": "1"}`, err: "remoteBolus must be"},
		{name: "zero bolus", body: `{"eventType": "Remote Bolus Entry", "remoteBolus": "0"}`, err: "remoteBolus must be"},
		{name: "bolus over Loop's largest", body: `{"eventType": "Remote Bolus Entry", "remoteBolus": "30.5"}`, err: "remoteBolus must be"},
		{name: "bolus not a number", body: `{"eventType": "Remote Bolus Entry", "remoteBolus": "NaN"}`, err: "remoteBolus must be"},
		{name: "negative carbs", body: `{"eventType": "Remote Carbs Entry", "remoteCarbs": "-5"}`, err: "remoteCarbs must be"},
		{name: "carbs over Loop's largest", body: `{"eventType": "Remote Carbs Entry", "remoteCarbs": "251"}`, err: "remoteCarbs must be"},
		{name: "absorption too long", body: `{"eventType": "Remote Carbs Entry", "remoteCarbs": "20", "remoteAbsorption": "25"}`, err: "remoteAbsorption"},
		{name: "override with no name", body: `{"eventType": "Temporary Override", "duration": "60"}`, err: "reason"},
		{name: "override over a week", body: `{"eventType": "Temporary Override", "reason": "Running", "duration": "10081"}`, err: "duration"},
		{name: "override with a negative duration", body: `{"eventType": "Temporary Override", "reason": "Running", "duration": "-1"}`, err: "duration"},
		{name: "unknown command", body: `{"eventType": "Note"}`, err: "unknown eventType"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var command loopCommand
			require.NoError(t, json.Unmarshal([]byte(tt.body), &command))
			alert, err := loopCommandAlert(command, "carer", now)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.title, alert.Title)
			require.Equal(t, tt.alertBody, alert.Body)
			require.Equal(t, tt.data, alert.Data)
		})
	}
}
