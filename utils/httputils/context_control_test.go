package httputils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

// TestIsControlAdmin checks only a signed in control admin is one
func TestIsControlAdmin(t *testing.T) {
	signedOut := func() context.Context { return CreateControlContext(context.Background(), nil, "", HTTPActionRead) }
	signedIn := func(role consts.ControlRole) context.Context {
		return SetAdminControlContext(signedOut(), &controldb.ControlIdentity{Role: role}, config.ServerStatus{})
	}
	tests := []struct {
		name string
		ctx  context.Context
		want bool
	}{
		{"no control context", context.Background(), false},
		{"signed out", signedOut(), false},
		{"read only", signedIn(consts.ControlRoleRead), false},
		{"admin", signedIn(consts.ControlRoleAdmin), true},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, IsControlAdmin(tt.ctx), tt.name)
	}
}
