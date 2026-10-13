package siteapi

import "nighttroop.t1d.nz/data/sql/controldb"

func HasPermission(storeName string, identity *controldb.SiteIdentity) bool {
	// TODO
	return true
}
