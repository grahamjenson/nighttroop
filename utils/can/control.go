package can

// control.go: what a control admin may do that a control reader may not. All control identities see all of control.

import (
	"nighttroop.t1d.nz/utils/consts"
)

// ControlChangeServer is whether the control role changes the server: only an admin manages sites, admins, rate limits
// and push credentials, and sends test pushes. A reader sees the same pages and changes nothing.
func ControlChangeServer(role consts.ControlRole) bool {
	return role == consts.ControlRoleAdmin
}
