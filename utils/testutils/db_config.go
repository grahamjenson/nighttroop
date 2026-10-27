package testutils

import (
	"context"
	"fmt"

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

func GetConfigControlDBForTest() (*config.ServerConfig, *db.ControlDB, error) {
	cfg := GetTestServerConfig()

	controlDB, err := db.NewControlDB(context.TODO(), cfg)
	if err != nil {
		return cfg, nil, fmt.Errorf("cannot create control db: %w", err)
	}

	return cfg, controlDB, nil
}

func GetConfigControlDBAndSiteDBForTest() (*config.ServerConfig, *db.ControlDB, *db.SiteCreateResult, error) {
	cfg := GetTestServerConfig()

	controlDB, err := db.NewControlDB(context.TODO(), cfg)
	if err != nil {
		return cfg, nil, nil, fmt.Errorf("cannot create control db: %w", err)
	}

	identityValues, err := controlDB.CreateSite()
	if err != nil {
		return cfg, nil, nil, fmt.Errorf("cannot create site db: %w", err)
	}

	return cfg, controlDB, identityValues, nil
}
