package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"strings"

	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/jwt"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/streamingaead"
)

// There are three key sets that we need to manage:
// 1. Encryption key set (AEAD) - for encrypting sensitive data in the database (e.g., user secrets)
// 2. File encryption key set (StreamingAEAD) - for encrypting large files (e.g., backups)
// 3. JWT key set - for signing and verifying JWT tokens
// They are base64 encoded and comma separated in KEY_SETS env
func (config *ServerConfig) ConfigCreateEncryptionKeys(dotenvdev map[string]string, dotenv map[string]string) error {
	keySetsRaw := configStringSelect(dotenvdev, dotenv, "KEY_SETS", "")
	if keySetsRaw == "" {
		return fmt.Errorf("no key sets provided") // No keys provided, we can generate them later
	}
	rawKeySets := strings.Split(keySetsRaw, ",")
	if len(rawKeySets) != 3 {
		return fmt.Errorf("invalid number of key sets provided: %d", len(rawKeySets))
	}

	encKeyRaw := rawKeySets[0]
	fileEncKeyRaw := rawKeySets[1]
	jwtKeyRaw := rawKeySets[2]

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

	fmt.Println("Successfully read encryption key sets from config")

	aeadTink, err := aead.New(encKetSet)
	if err != nil {
		return fmt.Errorf("failed to create AEAD primitive: %w", err)
	}

	streamEnc, err := streamingaead.New(fileEncKeySet)
	if err != nil {
		return fmt.Errorf("failed to create StreamingAEAD primitive: %w", err)
	}

	mac, err := jwt.NewMAC(jwtKeySet)
	if err != nil {
		return fmt.Errorf("failed to create JWT MAC primitive: %w", err)
	}

	validator, err := jwt.NewValidator(&jwt.ValidatorOpts{
		ExpectIssuedInThePast: true,
	})

	if err != nil {
		return fmt.Errorf("failed to create JWT validator: %w", err)
	}
	// Store KeySets
	config.EncryptionKeySet = encKetSet
	config.StreamEncryptionKeySet = fileEncKeySet // for files
	config.JWTKeySet = jwtKeySet

	// Store primitives
	config.aeadEnc = aeadTink
	config.streamEnc = streamEnc
	config.jwtMac = mac
	config.jwwtValidator = validator

	return nil

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

func (config *ServerConfig) Encrypt(plaintext string) (string, error) {

	ciphertext, err := config.aeadEnc.Encrypt([]byte(plaintext), nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func (config *ServerConfig) Decrypt(ciphertext string) (string, error) {
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

func (config *ServerConfig) EncryptingWriter(writer io.Writer) (io.WriteCloser, error) {
	return config.streamEnc.NewEncryptingWriter(writer, nil)
}

func (config *ServerConfig) DecryptingReader(reader io.Reader) (io.Reader, error) {
	return config.streamEnc.NewDecryptingReader(reader, nil)
}

func (config *ServerConfig) GenerateUserJWTToken(userID string, role UserRole) (string, error) {

	now := time.Now()
	expiresAt := now.Add(time.Hour * 24)

	rawJWT, err := jwt.NewRawJWT(&jwt.RawJWTOptions{
		Subject:   &userID,
		ExpiresAt: &expiresAt,
		IssuedAt:  &now,
		CustomClaims: map[string]any{
			"role": role.String(),
		},
	})
	if err != nil {
		return "", err
	}

	// 3. Sign and return
	return config.jwtMac.ComputeMACAndEncode(rawJWT)
}

func (config *ServerConfig) GenerateControlJWTToken(userID string, role ControlRole) (string, error) {

	now := time.Now()
	expiresAt := now.Add(time.Hour * 24)

	rawJWT, err := jwt.NewRawJWT(&jwt.RawJWTOptions{
		Subject:   &userID,
		ExpiresAt: &expiresAt,
		IssuedAt:  &now,
		CustomClaims: map[string]any{
			"role": role.String(),
		},
	})
	if err != nil {
		return "", err
	}

	// 3. Sign and return
	return config.jwtMac.ComputeMACAndEncode(rawJWT)
}

func (config *ServerConfig) VerifyJWTToken(handle *keyset.Handle, tokenStr string) (*jwt.VerifiedJWT, error) {
	primitive, _ := jwt.NewMAC(handle)

	// Define expected validator (e.g., audience, issuer)
	validator, err := jwt.NewValidator(&jwt.ValidatorOpts{
		ExpectIssuedInThePast: true,
	})
	if err != nil {
		return nil, err
	}

	// Verify and parse in one step
	return primitive.VerifyMACAndDecode(tokenStr, validator)
}
