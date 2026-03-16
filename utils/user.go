package utils

import (
	"crypto/rand"
	"fmt"
	"regexp"
)

// CONSTS

const USER_CHAR_SET = "abcdefghijklmnopqrstuvwxyz"
const USER_BATCH_LENGTH = 26
const USER_ID_LENGTH = 8

const CONTROL_DB_PATH = "data/db/control.sqlite"
const CONTROL_DB_BACKUP_PATH = "data/backup/control.sqlite"

const API_KEY_CHAR_SET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const USER_DB_FOLDER = "data/db/user"
const USER_DB_BACKUP_FOLDER = "data/backup/user"

func ControlS3Path(hostname string) string {
	return fmt.Sprintf("%s/backup/control.sqlite", hostname)
}

func UserDBPaths(subdomain string) (string, string) {
	batchPath, batchBackupPath := UserBatchPaths(string(subdomain[0]))

	dbPath := fmt.Sprintf("%s/%s.sqlite", batchPath, subdomain)
	backupPath := fmt.Sprintf("%s/%s.sqlite", batchBackupPath, subdomain)

	return dbPath, backupPath
}

func UserBatchS3Path(hostname string, char string) string {
	return fmt.Sprintf("%s/backup/user/%s.lz4", hostname, char)
}

func UserBatchPaths(char string) (string, string) {
	dbPath := fmt.Sprintf("%s/%s", USER_DB_FOLDER, char)
	backupPath := fmt.Sprintf("%s/%s", USER_DB_BACKUP_FOLDER, char)
	return dbPath, backupPath
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
