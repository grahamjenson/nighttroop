package security

import (
	"bytes"
	"crypto/hkdf"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"

	"strings"

	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/mac"
)

// There are three key sets that we need to manage:
//  1. Encryption key set (AEAD) - for encrypting sensitive data in the database (e.g., user secrets)
//  2. File encryption key set (StreamingAEAD) - unused since the DBs' replicas are encrypted by S3 (SSE-C), kept as
//     KEY_SETS has four
//  3. JWT key set - for signing and verifying JWT tokens
//
// They are base64 encoded and comma separated in KEY_SETS env
func (config *SecurityConfig) ConfigCreateEncryptionKeys(keySetsRaw string) error {
	if keySetsRaw == "" {
		return fmt.Errorf("no key sets provided") // No keys provided, we can generate them later
	}
	rawKeySets := strings.Split(keySetsRaw, ",")
	if len(rawKeySets) != 4 {
		return fmt.Errorf("invalid number of key sets provided: %d", len(rawKeySets))
	}

	encKeyRaw := rawKeySets[0]
	fileEncKeyRaw := rawKeySets[1]
	jwtKeyRaw := rawKeySets[2]
	macKeyRaw := rawKeySets[3]

	encKetSet, err := readKeySetFromBase64(encKeyRaw)
	if err != nil {
		return fmt.Errorf("failed to read encryption key set: %w", err)
	}

	if _, err := readKeySetFromBase64(fileEncKeyRaw); err != nil {
		return fmt.Errorf("failed to read file encryption key set: %w", err)
	}

	jwtKeySet, err := readKeySetFromBase64(jwtKeyRaw)
	if err != nil {
		return fmt.Errorf("failed to read JWT key set: %w", err)

	}

	macKeySet, err := readKeySetFromBase64(macKeyRaw)
	if err != nil {
		return fmt.Errorf("failed to read MAC key set: %w", err)
	}

	slog.Info("Successfully read encryption key sets from config")

	aeadTink, err := aead.New(encKetSet)
	if err != nil {
		return fmt.Errorf("failed to create AEAD primitive: %w", err)
	}

	macEnc, err := mac.New(macKeySet)
	if err != nil {
		return fmt.Errorf("failed to create MAC primitive: %w", err)
	}

	// Store KeySets
	config.EncryptionKeySet = encKetSet
	config.JWTKeySet = jwtKeySet
	config.MACKeySet = macKeySet

	// Store primitives
	config.aeadEnc = aeadTink
	config.mac = macEnc

	return nil

}

// UsesDevKeySets is whether any key in the comma separated key sets is one of DevKeySets', even in a key set of its own
func UsesDevKeySets(keySetsRaw string) bool {
	devKeys := keyValues(DevKeySets)
	for value := range keyValues(keySetsRaw) {
		if devKeys[value] {
			return true
		}
	}
	return false
}

// keyValues are the key data of every key in the comma separated key sets, leaving out any key set that cannot be read
func keyValues(keySetsRaw string) map[string]bool {
	values := map[string]bool{}
	for _, rawKeySet := range strings.Split(keySetsRaw, ",") {
		handle, err := readKeySetFromBase64(rawKeySet)
		if err != nil {
			continue
		}
		for _, key := range insecurecleartextkeyset.KeysetMaterial(handle).GetKey() {
			values[string(key.GetKeyData().GetValue())] = true
		}
	}
	return values
}

func readKeySetFromBase64(rawKeySet string) (*keyset.Handle, error) {
	// 1. Decode the Base64 string to binary
	keySetBytes, err := base64.RawStdEncoding.DecodeString(rawKeySet)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 keyset: %w", err)
	}

	reader := keyset.NewBinaryReader(bytes.NewBuffer(keySetBytes))

	// 2. Read the keyset into a handle
	// Use insecurecleartextkeyset because the keys are currently unencrypted strings.
	handle, err := insecurecleartextkeyset.Read(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read keyset: %w", err)
	}

	return handle, nil
}

// //
// Encryption Methods
// //

// Purpose is what a ciphertext is, bound into it with its site, so it decrypts only as that, for that site: one
// site's Dexcom password cannot be copied in as another's, or as a TOTP secret
type Purpose string

const (
	PurposeConnector       Purpose = "connector" // a connector's URL or username, and its password
	PurposeTOTP            Purpose = "totp"
	PurposePushSecret      Purpose = "push_secret"
	PurposePushToken       Purpose = "push_token"
	PurposeNightscoutToken Purpose = "nightscout_token"
	PurposeWelcomePasskey  Purpose = "welcome_passkey"
	PurposeNightscoutHost  Purpose = "nightscout_host" // a MAC of a Nightscout connector's host
)

func associatedData(purpose Purpose, siteID string) []byte {
	return []byte(string(purpose) + "|" + siteID)
}

// EncryptFor encrypts plaintext for its purpose and site, "" where it has none, as a control admin's TOTP secret
func (config *SecurityConfig) EncryptFor(purpose Purpose, siteID, plaintext string) (string, error) {
	ciphertext, err := config.aeadEnc.Encrypt([]byte(plaintext), associatedData(purpose, siteID))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}
	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

// MACFor is a MAC of the data for its purpose, as hex: the same data gives the same MAC while the MAC key set's primary
// key is the same, so equal values can be found without keeping them
func (config *SecurityConfig) MACFor(purpose Purpose, data string) (string, error) {
	tag, err := config.mac.ComputeMAC(associatedData(purpose, data))
	if err != nil {
		return "", fmt.Errorf("failed to compute MAC: %w", err)
	}
	return hex.EncodeToString(tag), nil
}

// DecryptFor decrypts what EncryptFor encrypted for the purpose and site, and nothing else
func (config *SecurityConfig) DecryptFor(purpose Purpose, siteID, ciphertext string) (string, error) {
	ciphertextBytes, err := base64.RawStdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}
	plaintext, err := config.aeadEnc.Decrypt(ciphertextBytes, associatedData(purpose, siteID))
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plaintext), nil
}

// ReplicaKey is the base64 SSE-C key of the replica of the DB with name, such as "control" or "site/<site ID>": an
// HMAC-SHA256 of its name with the MAC key set, so each DB's replica has its own key and a leaked key exposes one DB.
// It is sent to S3 with each request, so it is never a key that encrypts anything here. A new primary MAC key makes
// new keys, and the replicas made with the old ones cannot be read.
func (config *SecurityConfig) ReplicaKey(name string) (string, error) {
	tag, err := config.mac.ComputeMAC([]byte("nighttroop litestream sse-c v1:" + name))
	if err != nil {
		return "", fmt.Errorf("failed to compute the replica key: %w", err)
	}
	// Tink puts a 5 byte prefix naming the key before a tag, but for a raw key
	if len(tag) == 5+sha256.Size {
		tag = tag[5:]
	}
	if len(tag) != sha256.Size {
		if tag, err = hkdf.Key(sha256.New, tag, nil, "nighttroop litestream sse-c v1", sha256.Size); err != nil {
			return "", err
		}
	}
	return base64.StdEncoding.EncodeToString(tag), nil
}
