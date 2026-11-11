package security

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"

	"strings"

	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/mac"
	"github.com/tink-crypto/tink-go/v2/streamingaead"
)

// There are three key sets that we need to manage:
// 1. Encryption key set (AEAD) - for encrypting sensitive data in the database (e.g., user secrets)
// 2. File encryption key set (StreamingAEAD) - for encrypting large files (e.g., backups)
// 3. JWT key set - for signing and verifying JWT tokens
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

	fileEncKeySet, err := readKeySetFromBase64(fileEncKeyRaw)
	if err != nil {
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

	streamEnc, err := streamingaead.New(fileEncKeySet)
	if err != nil {
		return fmt.Errorf("failed to create StreamingAEAD primitive: %w", err)
	}

	macEnc, err := mac.New(macKeySet)
	if err != nil {
		return fmt.Errorf("failed to create MAC primitive: %w", err)
	}

	// Store KeySets
	config.EncryptionKeySet = encKetSet
	config.StreamEncryptionKeySet = fileEncKeySet // for files
	config.JWTKeySet = jwtKeySet
	config.MACKeySet = macKeySet

	// Store primitives
	config.aeadEnc = aeadTink
	config.streamEnc = streamEnc
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

// DecryptFor decrypts what EncryptFor encrypted for the purpose and site, or what was encrypted before ciphertexts were
// bound, with no associated data, until those are encrypted again (TODOs.md)
func (config *SecurityConfig) DecryptFor(purpose Purpose, siteID, ciphertext string) (string, error) {
	ciphertextBytes, err := base64.RawStdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}
	plaintext, err := config.aeadEnc.Decrypt(ciphertextBytes, associatedData(purpose, siteID))
	if err != nil {
		var unboundErr error
		if plaintext, unboundErr = config.aeadEnc.Decrypt(ciphertextBytes, nil); unboundErr != nil {
			return "", fmt.Errorf("failed to decrypt: %w", err)
		}
	}
	return string(plaintext), nil
}

// backupAssociatedData binds a backup to its object key without its host name, so a bucket copied to another host
// still restores, but one site batch's archive cannot be put at another's key
func backupAssociatedData(objectKey string) []byte {
	_, path, _ := strings.Cut(objectKey, "/")
	return []byte("backup|" + path)
}

// EncryptingWriterFor encrypts a backup to be kept at the object key
func (config *SecurityConfig) EncryptingWriterFor(writer io.Writer, objectKey string) (io.WriteCloser, error) {
	return config.streamEnc.NewEncryptingWriter(writer, backupAssociatedData(objectKey))
}

// backupProbeBytes are read ahead to tell an older, unbound backup from a bound one, by whether its first segment, of
// 4KB or at most 1MB, decrypts; more than one segment, so the first is not taken as the last
const backupProbeBytes = 2 << 20

// DecryptingReaderFor decrypts a backup kept at the object key, or one encrypted before backups were bound, with no
// associated data, until retention has removed those (TODOs.md)
func (config *SecurityConfig) DecryptingReaderFor(reader io.Reader, objectKey string) (io.Reader, error) {
	first := make([]byte, backupProbeBytes)
	n, err := io.ReadFull(reader, first)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}
	first = first[:n]
	associated := backupAssociatedData(objectKey)
	probe, err := config.streamEnc.NewDecryptingReader(bytes.NewReader(first), associated)
	if err == nil {
		_, err = io.ReadFull(probe, make([]byte, 1))
	}
	if err != nil && err != io.EOF {
		associated = nil
	}
	return config.streamEnc.NewDecryptingReader(io.MultiReader(bytes.NewReader(first), reader), associated)
}
