package site_data_templ

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The glucose days are Today alone on a new site's first day, not "Last 1 days"
func TestGlucoseDaysTitle(t *testing.T) {
	require.Equal(t, "Today", glucoseDaysTitle(1))
	require.Equal(t, "Last 2 days", glucoseDaysTitle(2))
	require.Equal(t, "Last 8 days", glucoseDaysTitle(8))
}
