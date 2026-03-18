package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	badrand "math/rand"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// CONSTS

const USER_CHAR_SET = "abcdefghijklmnopqrstuvwxyz"
const USER_BATCH_LENGTH = 26
const USER_ID_LENGTH = 8

const CONTROL_DB_PATH = "data/db/control.sqlite"
const CONTROL_DB_BACKUP_PATH = "data/backup/control.sqlite"

const API_KEY_CHAR_SET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const API_KEY_CHAR_SET_LENGTH = len(API_KEY_CHAR_SET)

const USER_DB_FOLDER = "data/db/user"
const USER_DB_BACKUP_FOLDER = "data/backup/user"

func ControlS3Path(hostname string) string {
	return fmt.Sprintf("%s/backup/control.sqlite", hostname)
}

func UserDBPaths(subdomain string) (string, string) {
	if len(subdomain) == 0 {
		return "", ""
	}
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
	return SecureRandomString(USER_ID_LENGTH, USER_CHAR_SET)
}

func ValidateUserID(userID string) bool {
	if len(userID) != USER_ID_LENGTH {
		return false
	}
	// validate with regex that all chars are lowercase a-z
	re := regexp.MustCompile("^[a-z]+$")
	return re.MatchString(userID)
}

func InsecureRandomString(length int, charset string) string {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		num := badrand.Intn(len(charset))
		bytes[i] = charset[num]
	}

	return string(bytes)
}

func SecureRandomString(length int, charset string) string {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			fmt.Println("Bad Secure Random Error: ", err)
			return InsecureRandomString(length, charset)
		}
		bytes[i] = charset[num.Int64()]
	}

	return string(bytes)
}

func SecureCredentials() (string, string, error) {
	// 1. Generate a secure random string for the password
	password := SecureRandomString(12, API_KEY_CHAR_SET)

	// 2. Hash the password
	// Bcrypt automatically generates a unique salt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	return password, string(hashedBytes), nil
}

func VerifyPassword(plainPassword, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword))
	return err == nil
}
