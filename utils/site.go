package utils

import (
	"strings"
	"sync"

	"nighttroop.t1d.nz/utils/security"
)

// CONSTS

const SITE_CHAR_SET = "abcdefghijklmnopqrstuvwxyz"
const SITE_BATCH_LENGTH = 26
const SITE_ID_LENGTH = 8

const ID_CHAR_SET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567891234567890"
const ID_LENGTH = 8

// Length of strings must be SITE_ID_LENGTH and chars must be in SITE_CHAR_SET
var firstSubdomains = []string{
	"alicezzz",
	"bobzzzzz",
	"charliez",
	"davezzzz",
	"evezzzzz",
	"frankzzz",
}

var firstSubdomainsLock sync.Mutex

func NextSiteSubdomain(usePrecomputedList bool) string {
	if usePrecomputedList {
		firstSubdomainsLock.Lock()
		defer firstSubdomainsLock.Unlock()
		if !(len(firstSubdomains) == 0) {
			next := firstSubdomains[0]
			firstSubdomains = firstSubdomains[1:]
			return next
		}
	}
	return security.SecureRandomString(SITE_ID_LENGTH, SITE_CHAR_SET)
}

// SyntacticallyValidateSiteID is whether siteID could be a site's: SITE_ID_LENGTH letters from SITE_CHAR_SET. Every
// request to a subdomain asks, so it allocates nothing, where compiling a regex each time did.
func SyntacticallyValidateSiteID(siteID string) bool {
	return len(siteID) == SITE_ID_LENGTH && strings.Trim(siteID, SITE_CHAR_SET) == ""
}

func NextID() string {
	return security.InsecureRandomString(ID_LENGTH, ID_CHAR_SET)
}
