package db

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	migrations "nighttroop.t1d.nz/sql/migrations"
	usersql "nighttroop.t1d.nz/sql/user"
	"nighttroop.t1d.nz/utils"
)

// //
// TenentDB
// //

type UserDB struct {
	*SQLDB
	Subdomain string
	char      byte

	// Prepared statements
	DevicesGet *usersql.DevicesGetSQL
}

func NewUserDB(ctx context.Context, subdomain string) *UserDB {
	char := subdomain[0]
	topPath := fmt.Sprintf("/user/%s/%s.sqlite", string(char), subdomain)
	userDB := &UserDB{
		SQLDB: &SQLDB{
			Path:       "data/db" + topPath,
			BackupPath: "data/backup" + topPath,
			CacheSize:  500,
		},
		Subdomain: subdomain,
		char:      char,
	}

	migrations.MigrateUser(ctx, userDB.sqldb)

	return userDB
}

func (db *UserDB) ArchivePath() string {
	return fmt.Sprintf("archive/%s/%s.tar.lz4", string(db.char), db.Subdomain)
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

	// Store the hashed key in the database
	_, err = db.Exec("INSERT INTO api_tokens (token) VALUES (?)", hashedKey)
	if err != nil {
		return "", fmt.Errorf("cannot store API token: %w", err)
	}

	// Return the plain API key to the user (hashed key is stored in the DB)
	return apiKey, nil
}

func (db *UserDB) ValidateAPIToken(token string) error {
	var err error
	// Create an HMAC hash of the provided token
	h := hmac.New(sha256.New, hmacSecretKey)
	h.Write([]byte(token))
	hashedToken := hex.EncodeToString(h.Sum(nil))

	// Check if the hashed token exists in the database
	rows, err := db.Query("SELECT COUNT(*) FROM api_tokens WHERE token = ?", hashedToken)
	if err != nil {
		return fmt.Errorf("cannot validate API token: %w", err)
	}

	if rows == nil {
		return fmt.Errorf("no rows returned in validate API token")
	}

	defer rows.Close()
	err = rows.Err()
	if err != nil {
		return fmt.Errorf("rows error in validate API token: %w", err)
	}

	rows.Next()

	var count int
	err = rows.Scan(&count)

	if err != nil {
		return fmt.Errorf("cannot validate API token: %w", err)
	}

	if count == 0 {
		return errors.New("invalid API token")
	}

	// Update last_seen timestamp for the token
	_, err = db.Exec("UPDATE api_tokens SET last_seen = CURRENT_TIMESTAMP WHERE token = ?", hashedToken)
	if err != nil {
		return fmt.Errorf("cannot update last_seen for API token: %w", err)
	}

	return nil
}
