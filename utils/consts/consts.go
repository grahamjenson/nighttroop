package consts

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

type ServerMode string

const (
	Production  ServerMode = "prod"
	Development ServerMode = "dev"
	Test        ServerMode = "test"
	// Staging runs like production, but sets up its control admin from its config, so wiping its data needs no setup
	Staging ServerMode = "staging"
)

func (mode ServerMode) Normalize() (ServerMode, error) {
	switch strings.ToLower(string(mode)) {
	case "prod", "production":
		return Production, nil
	case "dev", "development":
		return Development, nil
	case "test":
		return Test, nil
	case "staging":
		return Staging, nil
	default:
		return "", fmt.Errorf("unknown server mode: %s", mode)
	}
}

const FIRST_ADMIN_NAME = "admin"
const FIRST_PWD_NAME = "pwd_phone"
const FIRST_READ_ONLY_NAME = "read_only"

const CONTROL_IDENTITY_SUBDOMAIN = "control"

type ControlRole string

const (
	ControlRoleAdmin ControlRole = "admin"
	ControlRoleRead  ControlRole = "read"
)

// User Roles
type IdentityRole string

// Identity Roles are a little TYPE and a little PERMISSIONS.
// We are staying away from a formal permissions model, to be a bit simplier
const (
	IdentityRoleAdmin           IdentityRole = "admin"            // Login with password + totp
	IdentityRoleUser            IdentityRole = "user"             // Login with password with TOTP
	IdentityRoleUserReadOnly    IdentityRole = "user_readonly"    // Login with password with TOTP
	IdentityRoleClinician       IdentityRole = "clinician"        // Login with password, sees only the printable report
	IdentityRoleDevicePWDPhone  IdentityRole = "device_pwd_phone" // PWD (person with diabetes) phone that controls pump/cgm...
	IdentityRoleDeviceRead      IdentityRole = "device_read"
	IdentityRoleDeviceReadWrite IdentityRole = "device_read_write"
)

type IdentityType string

const (
	IdentityTypeAdmin  IdentityType = "admin"
	IdentityTypeUser   IdentityType = "user"
	IdentityTypeDevice IdentityType = "device"
)

func (i IdentityRole) GetType() IdentityType {
	switch i {
	case IdentityRoleAdmin:
		return IdentityTypeAdmin
	case IdentityRoleUser, IdentityRoleUserReadOnly, IdentityRoleClinician:
		return IdentityTypeUser
	case IdentityRoleDevicePWDPhone, IdentityRoleDeviceRead, IdentityRoleDeviceReadWrite:
		return IdentityTypeDevice
	default:
		slog.Warn("Unknown identity role type", "role", i)
		return IdentityTypeUser
	}
}

// What each role can do is checked where it is done, not in one table of permissions

/// Some reusable ENVARS

const KEY_SETS_ENV_KEY = "KEY_SETS"

/// Connector Types

type DexcomRegion string

const (
	DexcomRegionUS    DexcomRegion = "USA"
	DexcomRegionOther DexcomRegion = "OTHER"
)

func NewDexcomRegionFromString(s string) DexcomRegion {
	switch s {
	case "US":
		return DexcomRegionUS
	case "OUS":
		return DexcomRegionOther
	case "USA":
		return DexcomRegionUS
	case "OTHER":
		return DexcomRegionOther
	case "INT":
		return DexcomRegionOther
	case "on":
		return DexcomRegionUS
	case "off":
		return DexcomRegionOther
	case "":
		return DexcomRegionOther
	default:
		slog.Error("Invalid Dexcom region provided", "dex_region", s)
		return DexcomRegionOther // Default to OTHER if invalid, but log the error
	}
}

// LibreRegion is a LibreLinkUp region, each with its own API host
type LibreRegion struct{ Code, Name string }

// LibreRegionGlobal is the default host, which tells a login to use the account's own region
const LibreRegionGlobal = "GLOBAL"

// LibreRegions are the only regions a Libre connector calls, so a sync never goes to a host a user or upstream chose
var LibreRegions = []LibreRegion{
	{LibreRegionGlobal, "Global"}, {"AE", "United Arab Emirates"}, {"AP", "Asia Pacific"}, {"AU", "Australia"},
	{"CA", "Canada"}, {"DE", "Germany"}, {"EU", "Europe"}, {"EU2", "Europe 2"}, {"FR", "France"},
	{"JP", "Japan"}, {"LA", "Latin America"}, {"US", "United States"},
}

// IsLibreRegion is whether the code is one of LibreRegions
func IsLibreRegion(code string) bool {
	return slices.ContainsFunc(LibreRegions, func(r LibreRegion) bool { return r.Code == code })
}

type ConnectorType string

const (
	ConnectorTypeNightscout ConnectorType = "nightscout"
	ConnectorTypeDexcom     ConnectorType = "dexcom"
	ConnectorTypeLibre      ConnectorType = "libre"
	ConnectorTypeVirtual    ConnectorType = "virtual"
)

func NewConnectorTypeFromString(s string) (ConnectorType, error) {
	switch s {
	case "nightscout":
		return ConnectorTypeNightscout, nil
	case "dexcom":
		return ConnectorTypeDexcom, nil
	case "libre":
		return ConnectorTypeLibre, nil
	case "virtual":
		return ConnectorTypeVirtual, nil
	default:
		slog.Error("Invalid connector type provided", "connector_type", s)
		return "", fmt.Errorf("invalid connector type: %s", s)
	}
}

// PushProvider is the service a push is sent through
type PushProvider string

const (
	PushProviderAPNs PushProvider = "apns" // Apple's, to iOS apps
	PushProviderFCM  PushProvider = "fcm"  // Firebase's, to Android apps
)

// PushApp is an app pushes can be sent to, which decides what a push to it says
type PushApp string

const (
	PushAppLoop       PushApp = "loop"
	PushAppTrio       PushApp = "trio"
	PushAppLoopFollow PushApp = "loopfollow"
	PushAppT1DCarer   PushApp = "t1dcarer"
)

// PushApps are every app pushes can be sent to
var PushApps = []PushApp{PushAppLoop, PushAppTrio, PushAppLoopFollow, PushAppT1DCarer}

const PASSWORD_GARBAGE_VALUE = "********"

const DEFAULT_RATE_LIMITING_IDENTITY_NAME = "defaultIdentity"

// MaxBatchRecords caps the records in one socket dbAdd, and is how many records of a v1 batch POST cost one write
const MaxBatchRecords = 100

// MaxV1BatchRecords caps the records in one v1 batch POST: Loop sends up to 1000 and Juggluco up to 5000 after a
// backlog, sending the same batch again until it is accepted
const MaxV1BatchRecords = 5000

// MaxProfilesPerStore caps the profiles in one profile store, far above any real set of presets
const MaxProfilesPerStore = 20

// ProfilesDefaultCount is how many profile stores a v1 profile search gives with no count, Nightscout's
// PROFILES_DEFAULT_COUNT
const ProfilesDefaultCount = 10
