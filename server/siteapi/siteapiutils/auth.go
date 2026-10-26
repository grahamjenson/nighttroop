package siteapiutils

import (
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
)

// Action is what a request does with a store, so an identity's permissions can differ by action
type Action string

const (
	ActionSearch Action = "search"
	ActionGet    Action = "get"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// adminStores change who can use the site and where its data comes from, so only admins touch them
var adminStores = map[string]bool{"identities": true, "sync_settings": true, "site_settings": true}

// ownSettingsStore is the signed in identity's own settings, which any identity may read and change
const ownSettingsStore = "identity_settings"

// readOnlyRoles may only read the site's data
var readOnlyRoles = map[consts.IdentityRole]bool{consts.IdentityRoleUserReadOnly: true, consts.IdentityRoleDeviceRead: true}

// readWriteRoles may read and change the site's data, but not administer the site
var readWriteRoles = map[consts.IdentityRole]bool{
	consts.IdentityRoleAdmin: true, consts.IdentityRoleUser: true,
	consts.IdentityRoleDevicePWDPhone: true, consts.IdentityRoleDeviceReadWrite: true,
}

// HasPermission is whether the identity may do the action on the store; every other store is the site's data
func HasPermission(collectionName string, action Action, identity *controldb.SiteIdentity) bool {
	if identity == nil {
		return false
	}
	if adminStores[collectionName] {
		return identity.Role == consts.IdentityRoleAdmin
	}
	if collectionName == ownSettingsStore {
		return readOnlyRoles[identity.Role] || readWriteRoles[identity.Role]
	}
	if readOnlyRoles[identity.Role] {
		return action == ActionSearch || action == ActionGet
	}
	return readWriteRoles[identity.Role]
}
