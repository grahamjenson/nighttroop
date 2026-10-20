package siteapiutils

import "nighttroop.t1d.nz/data/sql/controldb"

// Action is what a request does with a store, so an identity's permissions can differ by action
type Action string

const (
	ActionSearch Action = "search"
	ActionGet    Action = "get"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

func HasPermission(storeName string, action Action, identity *controldb.SiteIdentity) bool {
	// TODO
	return true
}
