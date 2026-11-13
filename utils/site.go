package utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"nighttroop.t1d.nz/utils/security"
)

// CONSTS

const SITE_CHAR_SET = "abcdefghijklmnopqrstuvwxyz"
const SITE_BATCH_LENGTH = 26
const SITE_ID_LENGTH = 8

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

// NextID is a new record's ID, shaped as a MongoDB ObjectId, as Nightscout's are: 24 lower case hex characters, the
// unix seconds it was made in then 8 random bytes. xDrip+ only deletes a treatment whose _id is that long, so one made
// here, or sent by another app without an _id, could not otherwise be deleted from xDrip+, and came back.
func NextID() string {
	var id [12]byte
	binary.BigEndian.PutUint32(id[:4], uint32(time.Now().Unix()))
	_, _ = rand.Read(id[4:])
	return hex.EncodeToString(id[:])
}
