package site_presets_templ

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/components/addmodal"
)

// TestFieldLimits checks the food preset form's fields name a column that is there, and only narrow its limits
func TestFieldLimits(t *testing.T) {
	require.NoError(t, addmodal.Check(foodPresetFields(context.Background())))
}
