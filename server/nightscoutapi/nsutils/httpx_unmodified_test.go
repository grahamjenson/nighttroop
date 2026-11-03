package nsutils

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A record is modified since If-Unmodified-Since only when it changed in a later second, as the header has no
// milliseconds; without the header nothing is
func TestModifiedSince(t *testing.T) {
	since := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	r := httptest.NewRequest("PATCH", "/", nil)
	require.False(t, ModifiedSince(WithIfUnmodifiedSince(r), since.Add(time.Hour).UnixMilli()), "no header")

	r.Header.Set("If-Unmodified-Since", since.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	ctx := WithIfUnmodifiedSince(r)
	tests := []struct {
		modified time.Time
		want     bool
	}{
		{since.Add(-time.Second), false},
		{since, false},
		{since.Add(999 * time.Millisecond), false},
		{since.Add(time.Second), true},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, ModifiedSince(ctx, tt.modified.UnixMilli()), tt.modified)
	}
}
