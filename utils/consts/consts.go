package consts

import (
	"fmt"
	"log/slog"
)

type ServerMode int

const (
	Production ServerMode = iota
	Development
	Test
)

const FIRST_ADMIN_NAME = "admin"

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
type DeviceRole string

const (
	DeviceRoleAdmin    DeviceRole = "admin"
	DeviceRolePWDPhone DeviceRole = "pwd_phone" // PWD (person with diabetes) phone that controls pump/cgm...
	DeviceRoleRead     DeviceRole = "read"
)

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
