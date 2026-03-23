package config

// Control Roles

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
