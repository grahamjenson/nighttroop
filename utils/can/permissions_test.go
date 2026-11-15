package can

// permissions_test.go: the permission matrix. Each table has one row for each role and one column for each check. A
// change to what a role may do must change a cell here, so a reviewer sees each change in this file.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
)

var identityRoles = []consts.IdentityRole{consts.IdentityRoleAdmin, consts.IdentityRoleUser, consts.IdentityRoleUserReadOnly,
	consts.IdentityRoleClinician, consts.IdentityRoleDevicePWDPhone, consts.IdentityRoleDeviceRead, consts.IdentityRoleDeviceReadWrite}

var controlRoles = []consts.ControlRole{consts.ControlRoleAdmin, consts.ControlRoleRead}

// identity is a signed in identity of the role
func identity(role consts.IdentityRole) *controldb.SiteIdentity {
	return &controldb.SiteIdentity{Role: role}
}

// Columns: login.go, nightscout.go and device_tokens.go
//   - login: IdentityLoginToWebsite
//   - totp, totp+secret: IdentityMustLoginWithTotp, without and with a TOTP secret
//   - nsToken: IdentityHaveNightscoutToken
//   - useToken, useToken+secret: IdentityUseNightscoutToken, without and with a TOTP secret
//   - pairAAPS: IdentityAuthorizeDevices
//   - deviceTokens: IdentitySeeDeviceTokens
const signInMatrix = `
role               login  totp  totp+secret  nsToken  useToken  useToken+secret  pairAAPS  deviceTokens
admin              yes    yes   yes          no       no        no               no        yes
user               yes    no    yes          yes      yes       no               no        yes
user_readonly      no     no    yes          yes      yes       no               no        no
clinician          yes    no    yes          no       yes       no               no        no
device_pwd_phone   no     no    yes          yes      yes       no               yes       yes
device_read        no     no    yes          yes      yes       no               no        no
device_read_write  no     no    yes          yes      yes       no               no        yes
`

var signInChecks = map[string]func(consts.IdentityRole) bool{
	"login":           IdentityLoginToWebsite,
	"totp":            func(role consts.IdentityRole) bool { return IdentityMustLoginWithTotp(role, false) },
	"totp+secret":     func(role consts.IdentityRole) bool { return IdentityMustLoginWithTotp(role, true) },
	"nsToken":         IdentityHaveNightscoutToken,
	"useToken":        func(role consts.IdentityRole) bool { return IdentityUseNightscoutToken(role, false) },
	"useToken+secret": func(role consts.IdentityRole) bool { return IdentityUseNightscoutToken(role, true) },
	"pairAAPS":        IdentityAuthorizeDevices,
	"deviceTokens":    func(role consts.IdentityRole) bool { return IdentitySeeDeviceTokens(identity(role)) },
}

// Columns: site_admin.go and site_report.go
//   - administer: IdentityAdministerSite
//   - siteSettings: IdentityReadSiteSettings
//   - regenerate: IdentityRegenerateSummaries
//   - allPush: IdentitySeeAllPushRegistrations
//   - resetOthers: IdentityResetOthersCredentials
//   - onlyReport: IdentityOnlySeeReport
const siteMatrix = `
role               administer  siteSettings  regenerate  allPush  resetOthers  onlyReport
admin              yes         yes           yes         yes      yes          no
user               no          no            no          no       no           no
user_readonly      no          no            no          no       no           no
clinician          no          no            no          no       no           yes
device_pwd_phone   no          no            no          no       no           no
device_read        no          no            no          no       no           no
device_read_write  no          no            no          no       no           no
`

var siteChecks = map[string]func(consts.IdentityRole) bool{
	"administer":   IdentityAdministerSite,
	"siteSettings": IdentityReadSiteSettings,
	"regenerate":   IdentityRegenerateSummaries,
	"allPush":      IdentitySeeAllPushRegistrations,
	"resetOthers":  IdentityResetOthersCredentials,
	"onlyReport":   IdentityOnlySeeReport,
}

// Columns: site_stores.go, IdentityAccessStore for each store of a group and each action
//   - none: no action. read: search and get only. all: search, get, create, update and delete.
const storesMatrix = `
role               admin  own   treatments  summaries  data
admin              all    all   all         all        all
user               none   all   all         all        all
user_readonly      none   all   read        read       read
clinician          none   none  none        read       none
device_pwd_phone   none   all   all         all        all
device_read        none   all   read        read       read
device_read_write  none   all   all         read       read
`

// storeGroups are the stores of each column of storesMatrix. "data" are stores in no group, the site's data
var storeGroups = map[string]map[string]bool{
	"admin": adminStores, "own": ownStores, "treatments": caregiverStores, "summaries": clinicianStores,
	"data": {"glucose": true, "device_statuses": true, "profile_presets": true, "food_presets": true,
		"loop_settings": true, "summary_daily": true},
}

// Columns: control.go
//   - changeServer: ControlChangeServer
const controlMatrix = `
role   changeServer
admin  yes
read   no
`

var controlChecks = map[string]func(consts.ControlRole) bool{"changeServer": ControlChangeServer}

// nightscoutMatrix is NightscoutPermissions for each role. nil is an error: the role cannot use the Nightscout API
var nightscoutMatrix = map[consts.IdentityRole][]string{
	consts.IdentityRoleAdmin:           nil,
	consts.IdentityRoleUser:            {"api:*:*"},
	consts.IdentityRoleUserReadOnly:    {"api:*:read"},
	consts.IdentityRoleClinician:       nil,
	consts.IdentityRoleDevicePWDPhone:  {"api:*:*"},
	consts.IdentityRoleDeviceRead:      {"api:*:read"},
	consts.IdentityRoleDeviceReadWrite: CaregiverNightscoutPermissions,
}

// cells are the matrix's values by role and column, after it checks that the matrix has a row for each role and the
// columns given
func cells[R ~string](t *testing.T, matrix string, roles []R, columns []string) map[R]map[string]string {
	lines := strings.Split(strings.TrimSpace(matrix), "\n")
	header := strings.Fields(lines[0])[1:]
	require.ElementsMatch(t, columns, header, "the matrix's columns")
	values := map[R]map[string]string{}
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		require.Len(t, fields, len(header)+1, line)
		row := map[string]string{}
		for i, column := range header {
			row[column] = fields[i+1]
		}
		values[R(fields[0])] = row
	}
	require.Len(t, values, len(roles), "a row for each role")
	for _, role := range roles {
		require.Contains(t, values, role)
	}
	return values
}

// checkMatrix checks that each check gives the matrix's yes or no for each role
func checkMatrix[R ~string](t *testing.T, matrix string, roles []R, checks map[string]func(R) bool) {
	columns := make([]string, 0, len(checks))
	for column := range checks {
		columns = append(columns, column)
	}
	for role, row := range cells(t, matrix, roles, columns) {
		for column, value := range row {
			require.Contains(t, []string{"yes", "no"}, value, "%s %s", role, column)
			require.Equal(t, value == "yes", checks[column](role), "%s %s", role, column)
		}
	}
}

func TestSignInMatrix(t *testing.T) {
	checkMatrix(t, signInMatrix, identityRoles, signInChecks)
}

func TestSiteMatrix(t *testing.T) {
	checkMatrix(t, siteMatrix, identityRoles, siteChecks)
}

func TestControlMatrix(t *testing.T) {
	checkMatrix(t, controlMatrix, controlRoles, controlChecks)
}

func TestStoresMatrix(t *testing.T) {
	groups := make([]string, 0, len(storeGroups))
	for group := range storeGroups {
		groups = append(groups, group)
	}
	allowed := map[string]map[Action]bool{
		"none": {},
		"read": {ActionSearch: true, ActionGet: true},
		"all":  {ActionSearch: true, ActionGet: true, ActionCreate: true, ActionUpdate: true, ActionDelete: true},
	}
	for role, row := range cells(t, storesMatrix, identityRoles, groups) {
		for group, value := range row {
			require.Contains(t, allowed, value, "%s %s", role, group)
			for store := range storeGroups[group] {
				for _, action := range []Action{ActionSearch, ActionGet, ActionCreate, ActionUpdate, ActionDelete} {
					require.Equal(t, allowed[value][action], IdentityAccessStore(identity(role), store, action), "%s %s %s", role, store, action)
				}
			}
		}
	}
}

// A role it does not know is given nothing: no sign in, no Nightscout token or permissions, and no store. The control
// DB refuses to store such a role, so this is only a second guard.
func TestUnknownRole(t *testing.T) {
	unknown := consts.IdentityRole("owner")
	require.False(t, IdentityLoginToWebsite(unknown))
	require.False(t, IdentityHaveNightscoutToken(unknown))
	_, err := NightscoutPermissions(unknown, nil)
	require.Error(t, err)
	for column, check := range siteChecks {
		require.False(t, check(unknown), column)
	}
	for _, stores := range storeGroups {
		for store := range stores {
			require.False(t, IdentityAccessStore(identity(unknown), store, ActionSearch), store)
		}
	}
	require.False(t, ControlChangeServer(consts.ControlRole("owner")))
}

// A signed out request (no identity) may use no store and is not given device tokens
func TestSignedOut(t *testing.T) {
	for _, stores := range storeGroups {
		for store := range stores {
			require.False(t, IdentityAccessStore(nil, store, ActionSearch), store)
		}
	}
	require.False(t, IdentitySeeDeviceTokens(nil))
}

// A device AAPS pairs as a client reads everything and writes only settings, whatever its role, unless its role cannot
// use the Nightscout API at all. Any other client keeps its role's permissions
func TestNightscoutMatrix(t *testing.T) {
	aapsClient, xdrip := consts.AAPSClient, "xdrip"
	require.Len(t, nightscoutMatrix, len(identityRoles))
	for _, role := range identityRoles {
		want, ok := nightscoutMatrix[role]
		require.True(t, ok, role)
		for _, client := range []*string{nil, &xdrip, &aapsClient} {
			permissions, err := NightscoutPermissions(role, client)
			switch {
			case want == nil:
				require.Error(t, err, role)
			case client == &aapsClient:
				require.Equal(t, []string{"api:*:read", "api:settings:create", "api:settings:update"}, permissions, role)
			default:
				require.Equal(t, want, permissions, role)
			}
		}
	}
}
