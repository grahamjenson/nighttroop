package deps

import (
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/auth"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/config"
	"nighttroop.t1d.nz/server/nightscout_api/vendor_nightscout_api/store"
)

// Dependencies bundles the collaborators required to construct the Nightscout
// API routers.
type Dependencies struct {
	Config config.Config
	Store  store.Store
	Auth   auth.Manager
}
