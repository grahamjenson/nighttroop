package consts

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

type ControlRole int

const (
	ControlRoleAdmin ControlRole = iota
	ControlRoleRead
)

func (r ControlRole) String() string {
	switch r {
	case ControlRoleAdmin:
		return "admin"
	case ControlRoleRead:
		return "read"
	default:
		return "unknown"
	}
}

// User Roles
type UserRole int

const (
	UserRoleAdmin UserRole = iota
	UserRoleRead
)

func (r UserRole) String() string {
	switch r {
	case UserRoleAdmin:
		return "admin"
	case UserRoleRead:
		return "read"
	default:
		return "unknown"
	}
}

// TODO implement CanCan on the User and Control Roles

/// Some reusable ENVARS

const KEY_SETS_ENV_KEY = "KEY_SETS"
