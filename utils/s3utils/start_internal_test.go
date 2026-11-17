package s3utils

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/config"
)

// Without S3_URL S3 is off and never tested; with it, the server only starts once a test works, tried again after each
// wait, and is told why it never did
func TestStart(t *testing.T) {
	waits := []time.Duration{0, 0}
	tests := []struct {
		name     string
		endpoint string
		failures int
		wantErr  string
		wantRuns int
	}{
		{name: "without S3_URL S3 is off and never tested", endpoint: "", failures: 99, wantRuns: 0},
		{name: "a working S3 is on", endpoint: "http://s3.test", wantRuns: 1},
		{name: "an S3 that works on a later try is on", endpoint: "http://s3.test", failures: 2, wantRuns: 3},
		{name: "an S3 that never works stops the server", endpoint: "http://s3.test", failures: 99, wantRuns: 3, wantErr: "S3_URL is set but S3 is not working: refused"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewS3Client(&config.ServerConfig{S3URL: test.endpoint})
			runs := 0
			err := c.start(context.Background(), func(context.Context) error {
				runs++
				if runs <= test.failures {
					return errors.New("refused")
				}
				// Working S3 has versioning set up next, which fails here against no server, and is only logged
				return nil
			}, waits)
			require.Equal(t, test.wantRuns, runs)
			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
				require.False(t, c.ServerConfig.Status.S3Working)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.endpoint != "", c.ServerConfig.Status.S3Working)
		})
	}
}
