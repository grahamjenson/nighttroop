// Package push checks the credentials pushes are sent with to apps: an Apple APNs key or a Firebase service account
package push

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"regexp"
	"slices"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/i18n"
)

// MaxSecret is the longest secret, in bytes, a credential can have: a .p8 key is about 250, a service account about 2400
const MaxSecret = 4096

var (
	// appleID is an Apple team or key ID, ten letters and digits
	appleID = regexp.MustCompile(`^[A-Z0-9]{10}$`)
	// projectID is a Firebase (Google Cloud) project ID
	projectID = regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)
	// bundleID is an iOS bundle ID or Android package name, which has no quotes, as a JSON list of them is searched
	bundleID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,254}$`)
)

// Credential is what is checked of a credential before it is kept
type Credential struct {
	Provider        consts.PushProvider
	TeamOrProjectID string
	KeyID           *string
	Secret          string
	Apps            []string
	BundleIDs       []string
}

// Check says what is wrong with the credential, or nil: an APNs one has a team ID, key ID and a .p8 key on the P-256
// curve; an FCM one has a project ID and its service account; and each pushes to apps Nighttroop knows, at most once each
func Check(c Credential) error {
	if len(c.Secret) == 0 || len(c.Secret) > MaxSecret {
		return i18n.NewError("errors.push.secret_length", i18n.Data{"Max": MaxSecret})
	}
	if len(c.Apps) == 0 || len(c.Apps) > len(consts.PushApps) {
		return i18n.NewError("errors.push.pick_apps")
	}
	for i, app := range c.Apps {
		if !slices.Contains(consts.PushApps, consts.PushApp(app)) || slices.Contains(c.Apps[:i], app) {
			return i18n.NewError("errors.push.bad_app", i18n.Data{"App": app})
		}
	}
	if len(c.BundleIDs) > 20 {
		return i18n.NewError("errors.push.too_many_bundle_ids")
	}
	for _, id := range c.BundleIDs {
		if !bundleID.MatchString(id) {
			return i18n.NewError("errors.push.bad_bundle_id", i18n.Data{"ID": id})
		}
	}
	switch c.Provider {
	case consts.PushProviderAPNs:
		if !appleID.MatchString(c.TeamOrProjectID) || c.KeyID == nil || !appleID.MatchString(*c.KeyID) {
			return i18n.NewError("errors.push.apns_ids")
		}
		_, err := APNsKey(c.Secret)
		return err
	case consts.PushProviderFCM:
		if !projectID.MatchString(c.TeamOrProjectID) || c.KeyID != nil {
			return i18n.NewError("errors.push.fcm_ids")
		}
		return checkServiceAccount(c.Secret, c.TeamOrProjectID)
	}
	return fmt.Errorf("unknown provider: %s", c.Provider)
}

// APNsKey reads the private key of an APNs .p8 file, which signs the tokens pushes are sent with
func APNsKey(p8 string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(p8))
	if block == nil {
		return nil, i18n.NewError("errors.push.not_p8")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if ecKey, ok := key.(*ecdsa.PrivateKey); err == nil && ok && ecKey.Curve == elliptic.P256() {
		return ecKey, nil
	}
	return nil, i18n.NewError("errors.push.not_p256")
}

// checkServiceAccount checks a Firebase service account's JSON is for the project, with an RSA private key
func checkServiceAccount(secret string, project string) error {
	var account struct {
		Type        string `json:"type"`
		ProjectID   string `json:"project_id"`
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
	}
	if json.Unmarshal([]byte(secret), &account) != nil || account.Type != "service_account" || account.ClientEmail == "" {
		return i18n.NewError("errors.push.not_service_account")
	}
	if account.ProjectID != project {
		return i18n.NewError("errors.push.other_project", i18n.Data{"Project": account.ProjectID})
	}
	block, _ := pem.Decode([]byte(account.PrivateKey))
	if block == nil {
		return i18n.NewError("errors.push.no_private_key")
	}
	if _, err := x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
		return i18n.NewError("errors.push.bad_private_key")
	}
	return nil
}
