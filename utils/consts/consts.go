package consts

import (
	"fmt"
	"log/slog"
	"strings"
)

type ServerMode string

const (
	Production  ServerMode = "prod"
	Development ServerMode = "dev"
	Test        ServerMode = "test"
)

func (mode ServerMode) Normalize() (ServerMode, error) {
	switch strings.ToLower(string(mode)) {
	case "prod", "production":
		return Production, nil
	case "dev", "development":
		return Development, nil
	case "test":
		return Test, nil
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
	case IdentityRoleUser, IdentityRoleUserReadOnly:
		return IdentityTypeUser
	case IdentityRoleDevicePWDPhone, IdentityRoleDeviceRead, IdentityRoleDeviceReadWrite:
		return IdentityTypeDevice
	default:
		slog.Warn("Unknown identity role type", "role", i)
		return IdentityTypeUser
	}
}

// TODO implement CanCan on the User and Admin Roles

/// Some reusable ENVARS

const KEY_SETS_ENV_KEY = "KEY_SETS"

/// Sync Source Types

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

type SyncSourceType string

const (
	SyncSourceTypeNightscout SyncSourceType = "nightscout"
	SyncSourceTypeDexcom     SyncSourceType = "dexcom"
	SyncSourceTypeVirtual    SyncSourceType = "virtual"
)

func NewSyncSourceTypeFromString(s string) (SyncSourceType, error) {
	switch s {
	case "nightscout":
		return SyncSourceTypeNightscout, nil
	case "dexcom":
		return SyncSourceTypeDexcom, nil
	case "virtual":
		return SyncSourceTypeVirtual, nil
	default:
		slog.Error("Invalid sync source type provided", "sync_source_type", s)
		return "", fmt.Errorf("invalid sync source type: %s", s)
	}
}

const PASSWORD_GARBAGE_VALUE = "********"

const DEFAULT_RATE_LIMITING_IDENTITY_NAME = "defaultIdentity"
