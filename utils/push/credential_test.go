package push

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/consts"
)

// pkcs8 is the key as a PEM PKCS #8 private key, as a .p8 file or service account has it
func pkcs8(t *testing.T, key any) string {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func TestCheck(t *testing.T) {
	p256, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	p384, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	p8, keyID := pkcs8(t, p256), new("ABC123DEFG")
	account := func(project string) string {
		b, _ := json.Marshal(map[string]string{"type": "service_account", "project_id": project,
			"client_email": "push@" + project + ".iam.gserviceaccount.com", "private_key": pkcs8(t, rsaKey)})
		return string(b)
	}
	apns := func(change func(*Credential)) Credential {
		c := Credential{Provider: consts.PushProviderAPNs, TeamOrProjectID: "TEAM123456", KeyID: keyID, Secret: p8, Apps: []string{"loop"}}
		change(&c)
		return c
	}
	fcm := func(change func(*Credential)) Credential {
		c := Credential{Provider: consts.PushProviderFCM, TeamOrProjectID: "t1d-carer", Secret: account("t1d-carer"), Apps: []string{"t1dcarer"},
			BundleIDs: []string{"nz.t1d.carer"}}
		change(&c)
		return c
	}

	tests := []struct {
		name       string
		credential Credential
		err        string
	}{
		{"an APNs key", apns(func(c *Credential) {}), ""},
		{"an APNs key for some bundle IDs", apns(func(c *Credential) { c.BundleIDs = []string{"com.TEAM123456.loopkit.Loop"} }), ""},
		{"a Firebase service account", fcm(func(c *Credential) {}), ""},
		{"no secret", apns(func(c *Credential) { c.Secret = "" }), "secret"},
		{"too long a secret", apns(func(c *Credential) { c.Secret = strings.Repeat("a", MaxSecret+1) }), "secret"},
		{"no apps", apns(func(c *Credential) { c.Apps = nil }), "apps"},
		{"an unknown app", apns(func(c *Credential) { c.Apps = []string{"nope"} }), "unknown or repeated app"},
		{"an app twice", apns(func(c *Credential) { c.Apps = []string{"loop", "loop"} }), "unknown or repeated app"},
		{"a bundle ID with a quote", apns(func(c *Credential) { c.BundleIDs = []string{`com.a"b`} }), "bad bundle ID"},
		{"too many bundle IDs", apns(func(c *Credential) { c.BundleIDs = make([]string, 21) }), "at most 20"},
		{"an unknown provider", apns(func(c *Credential) { c.Provider = "sms" }), "unknown provider"},
		{"an APNs key without a key ID", apns(func(c *Credential) { c.KeyID = nil }), "team ID and key ID"},
		{"a short team ID", apns(func(c *Credential) { c.TeamOrProjectID = "TEAM" }), "team ID and key ID"},
		{"an APNs key that is not PEM", apns(func(c *Credential) { c.Secret = "not a key" }), "not a .p8 file"},
		{"an APNs key on another curve", apns(func(c *Credential) { c.Secret = pkcs8(t, p384) }), "not a P-256"},
		{"an APNs key that is RSA", apns(func(c *Credential) { c.Secret = pkcs8(t, rsaKey) }), "not a P-256"},
		{"a service account with a key ID", fcm(func(c *Credential) { c.KeyID = keyID }), "project ID, and no key ID"},
		{"a bad project ID", fcm(func(c *Credential) { c.TeamOrProjectID = "Bad_Project" }), "project ID"},
		{"a service account for another project", fcm(func(c *Credential) { c.Secret = account("other-project") }), "for project other-project"},
		{"a secret that is not a service account", fcm(func(c *Credential) { c.Secret = `{"type":"user"}` }), "not a service account"},
		{"a service account without its key", fcm(func(c *Credential) {
			c.Secret = `{"type":"service_account","project_id":"t1d-carer","client_email":"a@b"}`
		}), "no private key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Check(tt.credential)
			if tt.err == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.err)
		})
	}
}
