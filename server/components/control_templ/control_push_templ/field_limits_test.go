package control_push_templ

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/components/addmodal"
)

// TestFieldLimits checks the push credential form's fields, the site's form's and the site's, name a column that is
// there, and only narrow its limits
func TestFieldLimits(t *testing.T) {
	require.NoError(t, addmodal.Check(pushCredentialForm(context.Background()).Fields))
}
