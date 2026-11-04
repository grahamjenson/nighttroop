package push

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// MaxRegistrations is the most apps on devices an identity can have pushes sent to, far more than one person has
const MaxRegistrations = 10

// ErrTooManyRegistrations is a registration for an identity that already has MaxRegistrations
var ErrTooManyRegistrations = fmt.Errorf("an identity can have pushes sent to at most %d apps", MaxRegistrations)

// maxToken is the longest token kept, which its encrypted column holds
const maxToken = 1024

var (
	// apnsToken is an APNs device token in hex, 32 bytes today, which Apple says may grow
	apnsToken = regexp.MustCompile(`^[0-9a-fA-F]{64,200}$`)
	// fcmToken is a Firebase registration token, about 160 characters, at most maxToken
	fcmToken = regexp.MustCompile(`^[A-Za-z0-9_:-]{32,}$`)
)

// Registration is an app on a device registering for pushes. Production is whether an APNs token is from a TestFlight
// or App Store build, rather than a development one.
type Registration struct {
	Provider   consts.PushProvider
	App        string
	BundleID   string
	Token      string
	Production bool
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

// Register checks the registration and keeps it for the site's identity, its token encrypted, moving the token there if
// another of the site's identities had it. It is refused with ErrTooManyRegistrations when the identity has
// MaxRegistrations other than this token, counted in q's transaction, if it has one.
func Register(ctx context.Context, q *controldb.Queries, security *security.SecurityConfig, siteID string, identityID int64, r Registration) (*controldb.PushRegistration, error) {
	if err := CheckRegistration(r); err != nil {
		return nil, err
	}
	if r.Provider == consts.PushProviderAPNs {
		// An APNs token is hex, so in upper or lower case it is the same token
		r.Token = strings.ToLower(r.Token)
	}
	tokenHash := TokenHash(r.Token)
	count, err := q.CountPushRegistrations(ctx, controldb.CountPushRegistrationsParams{SiteID: siteID, IdentityID: identityID, TokenHash: tokenHash})
	if err != nil {
		return nil, err
	}
	if count >= MaxRegistrations {
		return nil, ErrTooManyRegistrations
	}
	tokenEnc, err := security.Encrypt(r.Token)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot encrypt a push token"), err)
	}
	return q.UpsertPushRegistration(ctx, controldb.UpsertPushRegistrationParams{SiteID: siteID, IdentityID: identityID,
		App: consts.PushApp(r.App), Provider: r.Provider, BundleID: r.BundleID, Production: r.Production, TokenEnc: tokenEnc,
		TokenHash: tokenHash})
}
