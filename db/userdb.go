package db

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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

// TODO HMAC value to encrypt tokens
var hmacSecretKey = []byte("your-secure-secret-key")

func (db *UserDB) CreateAPIToken() (string, error) {
	// Create a new API token for the user
	apiKey, err := utils.GenerateSecureAPIKey(12)
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Create an HMAC hash of the API key
	h := hmac.New(sha256.New, hmacSecretKey)
	h.Write([]byte(apiKey))
	hashedKey := hex.EncodeToString(h.Sum(nil))
	fmt.Printf("%s\n", hashedKey)
	// Store the hashed key in the database
	// _, err = db.Exec("INSERT INTO api_tokens (token) VALUES (?)", hashedKey)
	// if err != nil {
	// 	return "", fmt.Errorf("cannot store API token: %w", err)
	// }

	// Return the plain API key to the user (hashed key is stored in the DB)
	return apiKey, nil
}

func (db *UserDB) ValidateAPIToken(token string) error {

	return nil
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
