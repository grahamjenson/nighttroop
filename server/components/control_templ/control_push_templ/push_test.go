package control_push_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPushDevicesEmptySpansTheTable checks the empty row spans every column, as only an admin has a column for the
// test buttons
func TestPushDevicesEmptySpansTheTable(t *testing.T) {
	tests := []struct {
		admin bool
		want  string
	}{
		{admin: true, want: `colspan="7"`},
		{admin: false, want: `colspan="6"`},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		require.NoError(t, PushDevices(nil, tt.admin).Render(context.Background(), &buf))
		require.Contains(t, buf.String(), tt.want, "admin %v", tt.admin)
	}
}
