package push

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"slices"

	"nighttroop.t1d.nz/utils/consts"
)

// maxToken is the longest token kept, which its encrypted column holds
const maxToken = 1024

var (
	// apnsToken is an APNs device token in hex, 32 bytes today, which Apple says may grow
	apnsToken = regexp.MustCompile(`^[0-9a-fA-F]{64,200}$`)
	// fcmToken is a Firebase registration token, about 160 characters, at most maxToken
	fcmToken = regexp.MustCompile(`^[A-Za-z0-9_:-]{32,}$`)
)

// Registration is what is checked of an app on a device registering for pushes
type Registration struct {
	Provider consts.PushProvider
	App      string
	BundleID string
	Token    string
}

// CheckRegistration says what is wrong with the registration, or nil: an app Nighttroop knows, its bundle ID, and a
// token as its provider makes them
func CheckRegistration(r Registration) error {
	if !slices.Contains(consts.PushApps, consts.PushApp(r.App)) {
		return fmt.Errorf("unknown app: %s", r.App)
	}
	if !bundleID.MatchString(r.BundleID) {
		return fmt.Errorf("bad bundle ID: %s", r.BundleID)
	}
	switch r.Provider {
	case consts.PushProviderAPNs:
		if !apnsToken.MatchString(r.Token) {
			return fmt.Errorf("an APNs token is its device token in hex")
		}
	case consts.PushProviderFCM:
		if len(r.Token) > maxToken || !fcmToken.MatchString(r.Token) {
			return fmt.Errorf("not a Firebase registration token")
		}
	default:
		return fmt.Errorf("unknown provider: %s", r.Provider)
	}
	return nil
}

// TokenHash is the SHA-256 of a token in hex, which finds it without decrypting it
func TokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
