package consts

import (
	"fmt"
	"log/slog"
)

type Units string

const (
	UNITS_MG_DL  Units = "mg/dL"
	UNITS_MMOL_L Units = "mmol/L"
)
const MMOL_L_TO_MG_DL = 18.018

type ServerMode int

const (
	Production ServerMode = iota
	Development
	Test
)

const FIRST_ADMIN_NAME = "admin"
const FIRST_PWD_NAME = "pwd_phone"

const DEXCOM_SYNC_IDENTITY_NAME = "dexcom_sync"
const NIGHTSCOUT_SYNC_IDENTITY_NAME = "nightscout_sync"
const LIBRE_SYNC_IDENTITY_NAME = "libre_sync"

// //
// DEVELOPMENT HARDCODED STUFF
// //
const DEV_ADMIN_PASSWORD = "password"
const DEV_ADMIN_TOTP_SECRET = "FakeTOTPSecretForDevOnly" // This is not a real secret, just for development. In production this should be generated and stored securely.

type AdminRole string

const (
	AdminRoleAdmin AdminRole = "admin"
	AdminRoleRead  AdminRole = "read"
)

// User Roles
type IdentityRole string

// Identity Roles are a little TYPE and a little PERMISSIONS.
// We are staying away from a formal permissions model, to be a bit simplier
const (
	IdentityRoleAdmin           IdentityRole = "admin"            // Login with password + totp
	IdentityRoleUser            IdentityRole = "user"             // Login with password with TOTP
	IdentityRoleUserReadOnly    IdentityRole = "user_readonly"    // Login with password with TOTP
	IdentityRoleSync            IdentityRole = "sync"             // Can Write Data
	IdentityRoleDevicePWDPhone  IdentityRole = "device_pwd_phone" // PWD (person with diabetes) phone that controls pump/cgm...
	IdentityRoleDeviceRead      IdentityRole = "device_read"
	IdentityRoleDeviceReadWrite IdentityRole = "device_read_write"
)

type IdentityType string

const (
	IdentityTypeAdmin  IdentityType = "admin"
	IdentityTypeUser   IdentityType = "user"
	IdentityTypeDevice IdentityType = "device"
	IdentityTypeSync   IdentityType = "sync"
)

func (i IdentityRole) GetType() IdentityType {
	switch i {
	case IdentityRoleAdmin:
		return IdentityTypeAdmin
	case IdentityRoleUser, IdentityRoleUserReadOnly:
		return IdentityTypeUser
	case IdentityRoleSync:
		return IdentityTypeSync
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
	SyncSourceTypeNightscout SyncSourceType = "NIGHTSCOUT"
	SyncSourceTypeDexcom     SyncSourceType = "DEXCOM"
)

func NewSyncSourceTypeFromString(s string) (SyncSourceType, error) {
	switch s {
	case "NIGHTSCOUT":
		return SyncSourceTypeNightscout, nil
	case "DEXCOM":
		return SyncSourceTypeDexcom, nil
	default:
		slog.Error("Invalid sync source type provided", "sync_source_type", s)
		return "", fmt.Errorf("invalid sync source type: %s", s)
	}
}

const PASSWORD_GARBAGE_VALUE = "********"

const DEFAULT_RATE_LIMITING_IDENTITY_NAME = "defaultIdentity"
