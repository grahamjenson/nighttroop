package control_connectors_templ

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/controlmodels"
)

func TestOrDash(t *testing.T) {
	tests := []struct {
		value *int64
		want  string
	}{
		{value: nil, want: "-"},
		{value: new(int64(0)), want: "0"},
		{value: new(int64(9999)), want: "9999"},
		{value: new(int64(10000)), want: "10000"},
		{value: new(int64(12345)), want: "12345"},
		{value: new(int64(1234567890)), want: "1234567890"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, orDash(tt.value))
	}
}

func TestAverageOrDash(t *testing.T) {
	tests := []struct {
		value *float64
		want  string
	}{
		{value: nil, want: "-"},
		{value: new(0.0), want: "0.0"},
		{value: new(12.34), want: "12.3"},
		{value: new(10000.0), want: "10000.0"},
		{value: new(12345.67), want: "12345.7"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, averageOrDash(tt.value))
	}
}

func TestState(t *testing.T) {
	tests := []struct {
		name    string
		source  controlmodels.Connector
		state   string
		classes string
	}{
		{name: "syncing", source: controlmodels.Connector{Enabled: true, ConsecutiveFailedSyncs: new(int64(0))}, state: "syncing"},
		{name: "failing", source: controlmodels.Connector{Enabled: true, ConsecutiveFailedSyncs: new(int64(2))}, state: "failing", classes: "font-medium text-red-700"},
		{name: "turned off", source: controlmodels.Connector{ConsecutiveFailedSyncs: new(int64(0))}, state: "disabled", classes: "text-gray-500"},
		{name: "turned off, never synced", source: controlmodels.Connector{}, state: "disabled", classes: "text-gray-500"},
		{name: "turned off after failing", source: controlmodels.Connector{ConsecutiveFailedSyncs: new(int64(3))}, state: "disabled after failing", classes: "font-medium text-amber-800"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.state, state(&tt.source))
			require.Equal(t, tt.classes, stateClass(&tt.source))
		})
	}
}

// The table and its details show counts of 10000 or more whole, not as 1e+04
func TestConnectorCountsRender(t *testing.T) {
	source := &controlmodels.Connector{
		SiteID:               "abcdefgh",
		Enabled:              true,
		TotalSyncs:           new(int64(23456)),
		TotalSuccessfulSyncs: new(int64(12345)),
		TotalFailedSyncs:     new(int64(11111)),
		LastSyncDuration:     new(int64(10000)),
		TotalEma:             new(12345.67),
	}
	var table, details bytes.Buffer
	require.NoError(t, ConnectorsTable([]*controlmodels.Connector{source}, 0, 10).Render(t.Context(), &table))
	require.NoError(t, ConnectorModal(source).Render(t.Context(), &details))
	for _, html := range []string{table.String(), details.String()} {
		require.Contains(t, html, "12345")
		require.Contains(t, html, "11111")
		require.NotContains(t, html, "e+04")
	}
	require.Contains(t, details.String(), "23456")
	require.Contains(t, details.String(), "10000")
	require.Contains(t, details.String(), "12345.7")
}
