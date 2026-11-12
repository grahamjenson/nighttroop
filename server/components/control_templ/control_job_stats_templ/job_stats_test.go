package control_job_stats_templ

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/controlmodels"
)

// TestDurations checks a day's mean and spread of durations read in the units that suit them, as a job's does
func TestDurations(t *testing.T) {
	tests := []struct {
		name      string
		stats     controlmodels.JobStatsDaily
		mean, std string
	}{
		{"no jobs", controlmodels.JobStatsDaily{}, "-", "-"},
		{"quick", controlmodels.JobStatsDaily{TotalJobs: 2, DurationSum: 500, DurationSquaredSum: 2 * 250 * 250}, "250 ms", "0 ms"},
		{"slow and spread", controlmodels.JobStatsDaily{TotalJobs: 2, DurationSum: 3000, DurationSquaredSum: 1000*1000 + 2000*2000}, "1.5 s", "500 ms"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.mean, meanDuration(context.Background(), &tt.stats))
			require.Equal(t, tt.std, stdDevDuration(context.Background(), &tt.stats))
		})
	}
}
