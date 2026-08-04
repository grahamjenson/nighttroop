package utils

import (
	"fmt"
	"regexp"

	"nighttroop.t1d.nz/utils/security"
)

// CONSTS

const SITE_CHAR_SET = "abcdefghijklmnopqrstuvwxyz"
const SITE_BATCH_LENGTH = 26
const SITE_ID_LENGTH = 8

const ID_CHAR_SET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567891234567890"
const ID_LENGTH = 8

const DB_FOLDER = ".dbs/db"
const CONTROL_DB_PATH = ".dbs/db/control.sqlite"
const CONTROL_DB_BACKUP_PATH = ".dbs/backup/control.sqlite"

const SITE_DB_FOLDER = ".dbs/db/site"
const SITE_DB_BACKUP_FOLDER = ".dbs/backup/site"

func SiteDBPaths(subdomain string) (string, string) {
	if len(subdomain) == 0 {
		return "", ""
	}
	batchPath, batchBackupPath := SiteBatchPaths(string(subdomain[0]))

	dbPath := fmt.Sprintf("%s/%s.sqlite", batchPath, subdomain)
	backupPath := fmt.Sprintf("%s/%s.sqlite", batchBackupPath, subdomain)

	return dbPath, backupPath
}

func SiteBatchS3Path(hostname string, char string) string {
	return fmt.Sprintf("%s/backup/site/%s.lz4", hostname, char)
}

func SiteBatchPaths(char string) (string, string) {
	dbPath := fmt.Sprintf("%s/%s", SITE_DB_FOLDER, char)
	backupPath := fmt.Sprintf("%s/%s", SITE_DB_BACKUP_FOLDER, char)
	return dbPath, backupPath
}

// Length of strings must be SITE_ID_LENGTH and chars must be in SITE_CHAR_SET
var firstSubdomains = []string{
	"alicezzz",
	"bobzzzzz",
	"charliez",
	"davezzzz",
	"evezzzzz",
	"frankzzz",
}

func NextSiteSubdomain(usePrecomputedList bool) string {
	if usePrecomputedList {
		if !(len(firstSubdomains) == 0) {
			next := firstSubdomains[0]
			firstSubdomains = firstSubdomains[1:]
			return next
		}
	}
	return security.SecureRandomString(SITE_ID_LENGTH, SITE_CHAR_SET)
}

func SyntacticallyValidateSiteID(siteID string) bool {
	if len(siteID) != SITE_ID_LENGTH {
		return false
	}
	// validate with regex that all chars are lowercase a-z
	re := regexp.MustCompile("^[a-z]+$")
	return re.MatchString(siteID)
}

func NextID() string {
	return security.InsecureRandomString(ID_LENGTH, ID_CHAR_SET)
}
