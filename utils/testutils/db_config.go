package testutils

import (
	"testing"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
)

var storedServerConfig *config.ServerConfig

func GetTestServerConfig() *config.ServerConfig {
	if storedServerConfig != nil {
		return storedServerConfig
	}

	cfg, err := config.NewServerConfig()
	if err != nil {
		panic(err)
	}
	// Force Test Mode
	cfg.Mode = consts.Test
	// Any site can be disabled, as a test cannot wait days for one to be inactive
	cfg.DisableAfterInactivityDays = 0
	// A day's backfill is enough to test, and keeps it from slowing the browser scenarios running beside it
	cfg.BackfillDays = 1

	storedServerConfig = cfg
	return storedServerConfig
}

// NewControlDB is the test config and a new control DB for it, failing the test if the DB cannot be made
func NewControlDB(t testing.TB) (*config.ServerConfig, *db.ControlDB) {
	t.Helper()
	cfg := GetTestServerConfig()
	controlDB, err := db.NewControlDB(t.Context(), cfg)
	if err != nil {
		t.Fatalf("cannot create control db: %v", err)
	}
	if err := controlDB.SeedDevAdmin(t.Context()); err != nil {
		t.Fatalf("cannot create the control admin: %v", err)
	}
	return cfg, controlDB
}

// NewControlDBWithSite is NewControlDB with a site created in it
func NewControlDBWithSite(t testing.TB) (*config.ServerConfig, *db.ControlDB, *db.SiteCreateResult) {
	t.Helper()
	cfg, controlDB := NewControlDB(t)
	site, err := controlDB.CreateSite(t.Context())
	if err != nil {
		t.Fatalf("cannot create site db: %v", err)
	}
	return cfg, controlDB, site
}
