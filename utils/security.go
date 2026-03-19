package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	badrand "math/rand"

	"golang.org/x/crypto/bcrypt"
)

// //
// Generating And checking API tokens
// //
const API_KEY_CHAR_SET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const API_KEY_CHAR_SET_LENGTH = len(API_KEY_CHAR_SET)

func InsecureRandomString(length int, charset string) string {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		num := badrand.Intn(len(charset))
		bytes[i] = charset[num]
	}

	return string(bytes)
}

func SecureRandomString(length int, charset string) string {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			fmt.Println("Bad Secure Random Error: ", err)
			return InsecureRandomString(length, charset)
		}
		bytes[i] = charset[num.Int64()]
	}

	return string(bytes)
}

func SecureCredentials() (string, string, error) {
	// 1. Generate a secure random string for the password
	password := SecureRandomString(12, API_KEY_CHAR_SET)

	// 2. Hash the password
	// Bcrypt automatically generates a unique salt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	return password, string(hashedBytes), nil
}

func VerifyPassword(plainPassword, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword))
	return err == nil
}

// //
// Encrypting and Decrypting with key rotation
// //

// ALL KEY IDS ARE 6 BYTES LONG
func EncryptWithRotation(keyID string, keys map[string][]byte, plaintext []byte) ([]byte, error) {
	if len([]byte(keyID)) != 6 {
		return []byte{}, errors.New("invalid key ID length")
	}
	key, ok := keys[keyID]
	if !ok {
		return []byte{}, errors.New("unknown key version " + keyID)
	}

	// 1. Perform standard AES-GCM encryption
	ciphertext, err := Encrypt(plaintext, []byte(key))
	if err != nil {
		return []byte{}, err
	}

	// 2. Prepend the Key ID so we know which key to use later
	return append([]byte(keyID), ciphertext...), nil
}

func DecryptWithRotation(keys map[string][]byte, encodedData []byte) ([]byte, error) {
	if len(encodedData) < 6 {
		return []byte{}, errors.New("encoded data too short to contain key ID")
	}

	keyID, ciphertext := encodedData[0:6], encodedData[6:]
	if len([]byte(keyID)) != 6 {
		return []byte{}, errors.New("invalid key ID length")
	}

	// 2. Fetch the correct key version from the keyring
	key, ok := keys[string(keyID)]
	if !ok {
		return []byte{}, errors.New("unknown key version " + string(keyID))
	}

	return Decrypt(ciphertext, key)
}

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal appends the ciphertext to the nonce and returns it
	// No need for Base64 here; handle that at the very end of your pipeline if needed
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

func Decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
