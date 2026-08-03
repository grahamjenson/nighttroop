package testutils

import (
	"fmt"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
)

var storedServerConfig *config.ServerConfig

func GetTestServerConfig() config.ServerConfig {
	if storedServerConfig != nil {
		return *storedServerConfig
	}

	cfg, err := config.NewServerConfig()
	if err != nil {
		panic(err)
	}
	storedServerConfig = &cfg
	return *storedServerConfig
}

func GetConfigControlDBAndSiteDBForTest() (config.ServerConfig, *db.ControlDB, *db.SiteDB, error) {
	cfg := GetTestServerConfig()

	controlDB, err := db.NewControlDB(cfg)
	if err != nil {
		return cfg, nil, nil, fmt.Errorf("cannot create control db: %w", err)
	}

	identityValues, err := controlDB.CreateSite()
	if err != nil {
		return cfg, nil, nil, fmt.Errorf("cannot create site db: %w", err)
	}

	return cfg, controlDB, identityValues.SiteDB, nil
}
