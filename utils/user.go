package utils

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/mattn/go-sqlite3"
)

// CONSTS

const USER_CHAR_SET = "abcdefghijklmnopqrstuvwxyz"
const USER_POOL_LENGTH = 26
const USER_ID_LENGTH = 8

const CONTROL_DB_PATH = "data/db/control.sqlite"
const CONTROL_DB_BACKUP_PATH = "data/backup/control.sqlite"

const API_KEY_CHAR_SET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func UserDBPaths(subdomain string) (string, string) {
	char := subdomain[0]
	dbPath := fmt.Sprintf("data/db/user/%s/%s.sqlite", string(char), subdomain)
	backupPath := fmt.Sprintf("data/backup/user/%s/%s.sqlite", string(char), subdomain)

	return dbPath, backupPath
}

func UserBatchS3Path(hostname string, char byte) string {
	return fmt.Sprintf("%s/backup/user/%s.lz4", hostname, string(char))
}

func ControlS3Path(hostname string) string {
	return fmt.Sprintf("%s/backup/control.sqlite", hostname)
}

func NextUserID() string {
	result := make([]byte, USER_ID_LENGTH)
	rand.Read(result)
	for i, b := range result {
		result[i] = USER_CHAR_SET[b%byte(len(USER_CHAR_SET))]
	}
	return string(result)
}

func ValidateUserID(userID string) bool {
	if len(userID) != USER_ID_LENGTH {
		return false
	}
	// validate with regex that all chars are lowercase a-z
	re := regexp.MustCompile("^[a-z]+$")
	return re.MatchString(userID)
}

func GenerateSecureAPIKey(length int) (string, error) {
	// Create a byte slice to hold the random bytes
	bytes := make([]byte, length)

	// Generate random bytes
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Map random bytes to characters in the character set
	for i, b := range bytes {
		bytes[i] = API_KEY_CHAR_SET[b%byte(len(API_KEY_CHAR_SET))]
	}

	return string(bytes), nil
}

func SqliteBackup(ctx context.Context, db *sql.DB, path string, backupPath string) error {
	defer TrackTime(time.Now(), "SqliteBackup "+path)
	fmt.Println("SqliteBackup", path, backupPath)

	backupDB, err := sql.Open("sqlite3", backupPath)
	if err != nil {
		return fmt.Errorf("cannot open backup db: %w", err)
	}
	defer backupDB.Close()

	backupCon, err := backupDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("cannot get backup connection: %w", err)
	}
	defer backupCon.Close()

	srcConn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("cannot get source connection: %w", err)
	}
	defer srcConn.Close()

	err = backupCon.Raw(func(backupConRaw any) error {
		backupConSQLite, ok := backupConRaw.(*sqlite3.SQLiteConn)
		if !ok {
			fmt.Println("SqliteBackup err backupConSQLite", path, err)
			return fmt.Errorf("failed to get SQLiteConn backupConSQLite")
		}
		srcConn.Raw(func(srcConnRaw any) error {

			srcConnSQLite, ok := srcConnRaw.(*sqlite3.SQLiteConn)
			if !ok {
				fmt.Println("SqliteBackup err srcConnSQLite", path, err)
				return fmt.Errorf("failed to get SQLiteConn srcConnSQLite")
			}

			backup, err := backupConSQLite.Backup("main", srcConnSQLite, "main")
			if err != nil {
				fmt.Println("SqliteBackup err Backup", path, err)
				return fmt.Errorf("cannot create backup no create: %w", err)
			}

			defer backup.Finish()
			_, err = backup.Step(-1)

			if err != nil {
				fmt.Println("SqliteBackup err 3", path, err)
				return fmt.Errorf("cannot create backup no step 1: %w", err)
			}

			return nil
		})
		return nil
	})
	if err != nil {
		fmt.Println("SqliteBackup Error on Raw", path, err)
		return fmt.Errorf("cannot backup db: %w", err)
	}

	return nil
}
