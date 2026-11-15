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

// AAPSClient is the client of a device AAPS pairs as a client, reading and writing, through Nightscout's subjects
const AAPSClient = "aapsclient"

type ControlRole string

const (
	ControlRoleAdmin ControlRole = "admin"
	ControlRoleRead  ControlRole = "read"
)

// User Roles
type IdentityRole string

// The identity roles of a site. What each may do is decided only in utils/can, whose permission matrix test
// (utils/can/permissions_test.go) lists every role against every check.
const (
	// IdentityRoleAdmin signs in to the website with a password and TOTP, and administers the site
	IdentityRoleAdmin IdentityRole = "admin"
	// IdentityRoleUser signs in to the website with a password, and a TOTP code if it has one set, and reads and changes
	// the site's data
	IdentityRoleUser IdentityRole = "user"
	// IdentityRoleUserReadOnly does not sign in to the website: it reads the site's data with its Nightscout token
	IdentityRoleUserReadOnly IdentityRole = "user_readonly"
	// IdentityRoleClinician signs in to the website with a password, and sees only the printable report
	IdentityRoleClinician IdentityRole = "clinician"
	// IdentityRoleDevicePWDPhone is the PWD's (person with diabetes) phone, which runs the loop and uploads the site's
	// data with its Nightscout token
	IdentityRoleDevicePWDPhone IdentityRole = "device_pwd_phone"
	// IdentityRoleDeviceRead is a follower's app, or an AAPS client, which reads the site's data with its Nightscout token
	IdentityRoleDeviceRead IdentityRole = "device_read"
	// IdentityRoleDeviceReadWrite is a carer's app (Caregiver), which reads the site's data and changes its treatments
	// with its Nightscout token
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

// What each role can do is in utils/can

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

// EntriesDefaultCount is how many readings a v1 search gives with no count, Nightscout's ENTRIES_DEFAULT_COUNT
const EntriesDefaultCount = 10

// DeviceStatusDefaultCount is how many device statuses a v1 search gives with no count, as Nightscout gives
const DeviceStatusDefaultCount = 10

// TreatmentsDefaultCount is how many treatments a v1 search with no count and no find gives, as Nightscout gives
const TreatmentsDefaultCount = 100
