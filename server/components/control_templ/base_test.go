package control_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
)

func TestStatusBanner(t *testing.T) {
	tests := []struct {
		name   string
		status config.ServerStatus
		banner bool
		s3     bool
		sentry bool
	}{
		{name: "both working", status: config.ServerStatus{S3Working: true, SentryWorking: true}},
		{name: "s3 missing", status: config.ServerStatus{SentryWorking: true}, banner: true, s3: true},
		{name: "sentry missing", status: config.ServerStatus{S3Working: true}, banner: true, sentry: true},
		{name: "both missing", status: config.ServerStatus{}, banner: true, s3: true, sentry: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, StatusBanner(tt.status).Render(context.Background(), &buf))
			html := buf.String()
			require.Equal(t, tt.banner, bytes.Contains(buf.Bytes(), []byte(`id="control--status-banner"`)), html)
			require.Equal(t, tt.s3, bytes.Contains(buf.Bytes(), []byte(`id="control--status-banner--s3"`)), html)
			require.Equal(t, tt.sentry, bytes.Contains(buf.Bytes(), []byte(`id="control--status-banner--sentry"`)), html)
		})
	}
}

// The banner says how much memory is used and disk is free in bytes, where their totals are known, and the CPU's seconds
func TestResourceWarning(t *testing.T) {
	total := float64(954 << 30)
	tests := []struct {
		level config.ResourceLevel
		want  string
	}{
		{config.ResourceLevel{Resource: "disk", Value: 87, Limit: 60, Total: &total}, "124.0 GB free of 954.0 GB."},
		{config.ResourceLevel{Resource: "memory", Value: 75, Limit: 60, Total: new(float64(2 << 30))}, "1.5 GB used of 2.0 GB."},
		{config.ResourceLevel{Resource: "memory", Value: 75, Limit: 60}, "75% used, and 60% or more warns."},
		{config.ResourceLevel{Resource: "cpu", Value: 25, Limit: 20}, "at 100% for 25 seconds, and 20 or more warns."},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, resourceWarning(tt.level))
	}
}
