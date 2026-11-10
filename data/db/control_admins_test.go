package db_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// The last control admin is never made read only or deleted, checked as each is written, so two admins removing each
// other at once cannot leave control with none
func TestLastControlAdminIsKept(t *testing.T) {
	_, controlDB, _ := testutils.NewControlDBWithSite(t)
	ctx, q := t.Context(), controlDB.Queries
	admin, err := q.GetControlIdentity(ctx, controldb.GetControlIdentityParams{Name: consts.FIRST_ADMIN_NAME})
	require.NoError(t, err)
	_, err = controlDB.CreateControlIdentity(ctx, "reader", consts.ControlRoleRead, admin)
	require.NoError(t, err)
	reader, err := q.GetControlIdentity(ctx, controldb.GetControlIdentityParams{Name: "reader"})
	require.NoError(t, err)
	require.Equal(t, consts.FIRST_ADMIN_NAME, reader.CreatedByIdentityName)

	setRole := func(identity *controldb.ControlIdentity, role consts.ControlRole) error {
		return controlDB.SetControlIdentityRole(ctx, identity, role, admin)
	}
	// The only admin stays an admin
	require.ErrorIs(t, setRole(admin, consts.ControlRoleRead), db.ErrLastControlAdmin)
	require.ErrorIs(t, controlDB.DeleteControlIdentity(ctx, admin), db.ErrLastControlAdmin)
	// With two, one can be made read only, leaving the other the last
	require.NoError(t, setRole(reader, consts.ControlRoleAdmin))
	require.NoError(t, setRole(admin, consts.ControlRoleRead))
	require.ErrorIs(t, setRole(reader, consts.ControlRoleRead), db.ErrLastControlAdmin)
	require.ErrorIs(t, controlDB.DeleteControlIdentity(ctx, reader), db.ErrLastControlAdmin)

	// A reset removes the admin's passkeys, so a lost device's no longer signs in, and so does deleting it
	require.NoError(t, q.CreateControlPasskey(ctx, controldb.CreateControlPasskeyParams{ControlIdentityID: admin.ID, CredentialID: "lost-key", CredentialJson: "{}"}))
	values, err := controlDB.ResetControlIdentityPassword(ctx, admin, reader)
	require.NoError(t, err)
	require.NotEmpty(t, values.Password)
	_, err = q.GetControlPasskey(ctx, controldb.GetControlPasskeyParams{CredentialID: "lost-key"})
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, q.CreateControlPasskey(ctx, controldb.CreateControlPasskeyParams{ControlIdentityID: admin.ID, CredentialID: "other-key", CredentialJson: "{}"}))
	require.NoError(t, controlDB.DeleteControlIdentity(ctx, admin))
	_, err = q.GetControlIdentity(ctx, controldb.GetControlIdentityParams{Name: consts.FIRST_ADMIN_NAME})
	require.ErrorIs(t, err, sql.ErrNoRows)
	_, err = q.GetControlPasskey(ctx, controldb.GetControlPasskeyParams{CredentialID: "other-key"})
	require.ErrorIs(t, err, sql.ErrNoRows)
}
