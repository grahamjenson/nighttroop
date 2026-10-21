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
