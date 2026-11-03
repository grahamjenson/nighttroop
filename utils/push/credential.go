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
		return fmt.Errorf("the secret must be 1 to %d characters", MaxSecret)
	}
	if len(c.Apps) == 0 || len(c.Apps) > len(consts.PushApps) {
		return fmt.Errorf("pick the apps it pushes to")
	}
	for i, app := range c.Apps {
		if !slices.Contains(consts.PushApps, consts.PushApp(app)) || slices.Contains(c.Apps[:i], app) {
			return fmt.Errorf("unknown or repeated app: %s", app)
		}
	}
	if len(c.BundleIDs) > 20 {
		return fmt.Errorf("at most 20 bundle IDs")
	}
	for _, id := range c.BundleIDs {
		if !bundleID.MatchString(id) {
			return fmt.Errorf("bad bundle ID: %s", id)
		}
	}
	switch c.Provider {
	case consts.PushProviderAPNs:
		if !appleID.MatchString(c.TeamOrProjectID) || c.KeyID == nil || !appleID.MatchString(*c.KeyID) {
			return fmt.Errorf("an APNs key needs its 10 character team ID and key ID")
		}
		_, err := APNsKey(c.Secret)
		return err
	case consts.PushProviderFCM:
		if !projectID.MatchString(c.TeamOrProjectID) || c.KeyID != nil {
			return fmt.Errorf("a Firebase service account needs its project ID, and no key ID")
		}
		return checkServiceAccount(c.Secret, c.TeamOrProjectID)
	}
	return fmt.Errorf("unknown provider: %s", c.Provider)
}

// APNsKey reads the private key of an APNs .p8 file, which signs the tokens pushes are sent with
func APNsKey(p8 string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(p8))
	if block == nil {
		return nil, fmt.Errorf("the APNs key is not a .p8 file")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if ecKey, ok := key.(*ecdsa.PrivateKey); err == nil && ok && ecKey.Curve == elliptic.P256() {
		return ecKey, nil
	}
	return nil, fmt.Errorf("the APNs key is not a P-256 private key")
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
		return fmt.Errorf("the secret is not a service account's JSON")
	}
	if account.ProjectID != project {
		return fmt.Errorf("the service account is for project %s", account.ProjectID)
	}
	block, _ := pem.Decode([]byte(account.PrivateKey))
	if block == nil {
		return fmt.Errorf("the service account has no private key")
	}
	if _, err := x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
		return fmt.Errorf("the service account's private key cannot be read")
	}
	return nil
}
