package db

import (
	"context"
	"database/sql"
	"fmt"

	"nighttroop.t1d.nz/data/sql/user"

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
	Queries *user.Queries
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

	err = user.MigrateUp(sqldb)
	if err != nil {
		sqldb.Close()
		return userDB, fmt.Errorf("Migration Error Occured user db: %s %w", subdomain, err)
	}

	userDB.Queries, err = user.Prepare(context.TODO(), sqldb)
	if err != nil {
		sqldb.Close()
		return userDB, fmt.Errorf("cannot prepare statements for user db: %s %w", subdomain, err)
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

func (db *UserDB) Close() error {
	return db.SQLiteDB.Close()
}

func (db *UserDB) URL() string {
	return fmt.Sprintf("%s://%s.%s:%d", db.serverConfig.HostProtocol, db.Subdomain, db.serverConfig.HostName, db.serverConfig.HostPort)
}

func (db *UserDB) GetDeviceSplat(ctx context.Context, name string) (deviceName string, password string, totpSecret string, role string, err error) {
	device, err := db.Queries.GetDevice(ctx, name)
	if err != nil {
		return "", "", "", "", fmt.Errorf("device not found: %w", err)
	}

	return device.Name, device.Password, device.TotpSecret, device.Role, nil
}
