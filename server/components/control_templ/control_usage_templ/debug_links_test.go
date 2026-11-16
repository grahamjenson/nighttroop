package control_usage_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

// The usage page links to Go's profiles and runtime stats only for an admin, the only one who can open them
func TestDebugLinksOnlyForAdmins(t *testing.T) {
	for _, tt := range []struct {
		role  consts.ControlRole
		shown bool
	}{{consts.ControlRoleAdmin, true}, {consts.ControlRoleRead, false}} {
		ctx := httputils.CreateControlContext(context.Background(), nil, "", httputils.HTTPActionRead)
		ctx = httputils.SetAdminControlContext(ctx, &controldb.ControlIdentity{Role: tt.role}, config.ServerStatus{})
		var page bytes.Buffer
		require.NoError(t, debugLinks().Render(ctx, &page))
		require.Equal(t, tt.shown, bytes.Contains(page.Bytes(), []byte(`href="/debug/pprof/"`)), tt.role)
		require.Equal(t, tt.shown, bytes.Contains(page.Bytes(), []byte(`href="/debug/vars"`)), tt.role)
	}
}
