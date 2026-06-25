package testutils

import (
	"fmt"
	"log/slog"
	"os"

	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/utils/config"
)

var storedServerConfig *config.ServerConfig

func GetTestServerConfig() config.ServerConfig {
	if storedServerConfig != nil {
		return *storedServerConfig
	}
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, nil),
	)
	cfg, err := config.NewServerConfig(logger)
	if err != nil {
		panic(err)
	}
	storedServerConfig = &cfg
	return *storedServerConfig
}

func GetConfigControlDBAndUserDBForTest() (config.ServerConfig, *db.ControlDB, *db.UserDB) {
	cfg := GetTestServerConfig()

	controlDB, err := db.NewControlDB(cfg)
	if err != nil {
		panic(fmt.Errorf("cannot create control db: %w", err))
	}

	identityValues, err := controlDB.CreateUser()
	if err != nil {
		panic(fmt.Errorf("cannot create user db: %w", err))
	}

	return cfg, controlDB, identityValues.UserDB
}
