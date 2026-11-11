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

func (config *SecurityConfig) Encrypt(plaintext string) (string, error) {

	ciphertext, err := config.aeadEnc.Encrypt([]byte(plaintext), nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func (config *SecurityConfig) Decrypt(ciphertext string) (string, error) {
	ciphertextBytes, err := base64.RawStdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}

	plaintext, err := config.aeadEnc.Decrypt(ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

func (config *SecurityConfig) EncryptingWriter(writer io.Writer) (io.WriteCloser, error) {
	return config.streamEnc.NewEncryptingWriter(writer, nil)
}

func (config *SecurityConfig) DecryptingReader(reader io.Reader) (io.Reader, error) {
	return config.streamEnc.NewDecryptingReader(reader, nil)
}
