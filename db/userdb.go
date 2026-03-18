package db

import (
	"context"
	"database/sql"
	"fmt"

	migrations "nighttroop.t1d.nz/sql/migrations"
	usersql "nighttroop.t1d.nz/sql/user"
	"nighttroop.t1d.nz/utils"
)

// //
// TenentDB
// //

type UserDB struct {
	SQLiteDB  *sql.DB
	Subdomain string

	// Prepared statements
	DevicesGet *usersql.DevicesCreateStmt
}

func NewUserDB(ctx context.Context, subdomain string, mustExist bool) (*UserDB, error) {
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
		SQLiteDB:  sqldb,
		Subdomain: subdomain,
	}

	err = migrations.MigrateUser(ctx, sqldb)
	if err != nil {
		sqldb.Close()
		return userDB, fmt.Errorf("Migration Error Occured user db: %s %w", subdomain, err)
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
