package utils

import (
	"fmt"
	"regexp"

	"nighttroop.t1d.nz/utils/security"
)

// CONSTS

const USER_CHAR_SET = "abcdefghijklmnopqrstuvwxyz"
const USER_BATCH_LENGTH = 26
const USER_ID_LENGTH = 8

const DB_FOLDER = ".dbs/db"
const CONTROL_DB_PATH = ".dbs/db/control.sqlite"
const CONTROL_DB_BACKUP_PATH = ".dbs/backup/control.sqlite"

const USER_DB_FOLDER = ".dbs/db/user"
const USER_DB_BACKUP_FOLDER = ".dbs/backup/user"

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

// Length of strings must be USER_ID_LENGTH and chars must be in USER_CHAR_SET
var firstUserSubdomains = []string{
	"alicezzz",
	"bobzzzzz",
	"charliez",
	"davezzzz",
	"evezzzzz",
	"frankzzz",
}

func NextUserID(usePrecomputedList bool) string {
	if usePrecomputedList {
		if !(len(firstUserSubdomains) == 0) {
			next := firstUserSubdomains[0]
			firstUserSubdomains = firstUserSubdomains[1:]
			return next
		}
	}
	return security.SecureRandomString(USER_ID_LENGTH, USER_CHAR_SET)
}

func SyntacticallyValidateUserID(userID string) bool {
	if len(userID) != USER_ID_LENGTH {
		return false
	}
	// validate with regex that all chars are lowercase a-z
	re := regexp.MustCompile("^[a-z]+$")
	return re.MatchString(userID)
}
