package db

import (
	"context"
	"database/sql"
	"fmt"

	migrations "nighttroop.t1d.nz/data/sql/migrations"
	usersql "nighttroop.t1d.nz/data/sql/user"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

// //
// TenentDB
// //

type UserDB struct {
	serverConfig config.ServerConfig

	SQLiteDB  *sql.DB
	Subdomain string

	// Prepared statements
	DevicesGet    *usersql.DevicesGetStmt
	DevicesCreate *usersql.DevicesCreateStmt
	DevicesAll    *usersql.DevicesAllStmt
}

func NewUserDB(ctx context.Context, subdomain string, mustExist bool, cfg config.ServerConfig) (*UserDB, error) {
	path, _ := utils.UserDBPaths(subdomain)
	if mustExist && !utils.FileExists(path) {
		return nil, fmt.Errorf("The User DB %s Must exist and doesnt", subdomain)
	}

	sqldb, err := OpenSqliteDB(path, 100)
	if err != nil {
		return nil, fmt.Errorf("cannot open user db: %s %w", subdomain, err)
	}

	// Create a new control db
	userDB := &UserDB{
		serverConfig: cfg,
		SQLiteDB:     sqldb,
		Subdomain:    subdomain,
	}

	err = migrations.MigrateUser(ctx, sqldb)
	if err != nil {
		sqldb.Close()
		return userDB, fmt.Errorf("Migration Error Occured user db: %s %w", subdomain, err)
	}

	var errs []error

	userDB.DevicesGet = usersql.NewDevicesGet(sqldb, &errs)
	userDB.DevicesCreate = usersql.NewDevicesCreate(sqldb, &errs)
	userDB.DevicesAll = usersql.NewDevicesAll(sqldb, &errs)

	if len(errs) > 0 {
		sqldb.Close()
		return userDB, fmt.Errorf("Error creating prepared statements: %v", errs)
	}

	return userDB, nil
}

// WARNING THIS IS A BEST TRY FUNCTION ALL ERRORS ARE SWOLLOWED
func (db *UserDB) TryCloseAndDelete() {
	if db == nil {
		return
	}
	if db.SQLiteDB != nil {
		db.SQLiteDB.Close()
	}
	if db.Subdomain == "" {
		return
	}
	path, _ := utils.UserDBPaths(db.Subdomain)
	_ = utils.DeleteFile(path)
}

func (db *UserDB) Close() {
	db.SQLiteDB.Close()
}

func (db *UserDB) URL() string {
	return fmt.Sprintf("%s://%s.%s:%d", db.serverConfig.HostProtocol, db.Subdomain, db.serverConfig.HostName, db.serverConfig.HostPort)
}
