package controldb_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// The control DB's limits refuse texts far longer than real use, such as a name or an encrypted secret, so each
// column's longest is written, and one longer, through the queries that write it, so the limits values are checked by
// before they are written are its CHECKs
func TestControlDataLimits(t *testing.T) {
	_, control, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, id, by := control.Queries, t.Context(), site.SiteDB.Subdomain, strconv.Itoa, "test"
	admin, err := q.GetControlIdentity(ctx, controldb.GetControlIdentityParams{Name: consts.FIRST_ADMIN_NAME})
	require.NoError(t, err)
	require.NoError(t, q.CreateSiteIdentity(ctx, controldb.CreateSiteIdentityParams{SiteID: siteID, Name: "limits",
		PasswordHash: "hash", TotpSecretEnc: "secret", Role: consts.IdentityRoleUser, CreatedByIdentityName: by}))
	identity, err := q.GetSiteIdentity(ctx, controldb.GetSiteIdentityParams{SiteID: siteID, Name: "limits"})
	require.NoError(t, err)
	// A site's and an identity's settings, a site's abuse score and a connector's sync are written by their updates, not
	// checked here
	settings := []string{"Timezone", "Language", "DateFormat", "TimeFormat", "WeekStartDay", "NumberFormat",
		"AlarmDeviceBatteryLowDevicesJson"}
	testutils.CheckColumnLimits(t, controldb.ColumnLimits, map[string]testutils.LimitRecipe{
		"control_identities": testutils.Recipe(func(i int) controldb.CreateControlIdentityParams {
			return controldb.CreateControlIdentityParams{Name: id(i), PasswordHash: "hash", TotpSecretEnc: "secret",
				Role: consts.ControlRoleAdmin, CreatedByIdentityName: by}
		}, q.CreateControlIdentity, "UpdatedByIdentityName"),
		"passkeys": testutils.Recipe(func(i int) controldb.CreateControlPasskeyParams {
			return controldb.CreateControlPasskeyParams{ControlIdentityID: admin.ID, CredentialID: id(i), CredentialJson: "{}"}
		}, q.CreateControlPasskey),
		"push_credentials": testutils.Recipe(func(i int) controldb.CreatePushCredentialParams {
			return controldb.CreatePushCredentialParams{Name: id(i), Provider: "apns", TeamOrProjectID: "team",
				AppsJson: `["loop"]`, CreatedByIdentityName: by}
		}, testutils.NoResult(q.CreatePushCredential)),
		"push_registrations": testutils.Recipe(func(i int) controldb.UpsertPushRegistrationParams {
			return controldb.UpsertPushRegistrationParams{SiteID: siteID, IdentityID: identity.ID, App: "loop",
				Provider: "apns", BundleID: "bundle", TokenEnc: "token", TokenHash: fmt.Sprintf("%064d", i)}
		}, testutils.NoResult(q.UpsertPushRegistration), "LastError"),
		// There is one row of server settings, so its limits are written by its update
		"server_settings": testutils.Recipe(func(int) controldb.UpdateRateLimitsParams {
			return controldb.UpdateRateLimitsParams{RateLimitsJson: "{}"}
		}, q.UpdateRateLimits),
		// A site has one connector of a type, so each is of a site of its own
		"site_connectors": testutils.Recipe(func(i int) controldb.CreateConnectorParams {
			_, err := q.CreateSite(ctx, controldb.CreateSiteParams{ID: "connector" + id(i), BackupKeyHash: "hash"})
			require.NoError(t, err)
			return controldb.CreateConnectorParams{SiteID: "connector" + id(i), Type: consts.ConnectorTypeVirtual}
		}, q.CreateConnector, "LastBackfillError", "SyncIntervalSecondsReason", "PreviousSyncErrors", "FailureRateEma"),
		"site_identities": testutils.Recipe(func(i int) controldb.CreateSiteIdentityParams {
			return controldb.CreateSiteIdentityParams{SiteID: siteID, Name: id(i), PasswordHash: "hash",
				TotpSecretEnc: "secret", Role: consts.IdentityRoleUser, CreatedByIdentityName: by}
		}, q.CreateSiteIdentity, append(settings, "UpdatedByIdentityName")...),
		"sites": testutils.Recipe(func(i int) controldb.CreateSiteParams {
			return controldb.CreateSiteParams{ID: id(i), BackupKeyHash: "hash"}
		}, testutils.NoResult(q.CreateSite), append(settings, "TableSizesJson", "AbuseScore")...),
	})
}
