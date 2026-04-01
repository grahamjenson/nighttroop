package security

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	badrand "math/rand"

	"golang.org/x/crypto/bcrypt"
)

const (
	ChunkSize = 64 * 1024 // 64KB (tweakable)
	KeyIDSize = 6
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

func SecureCredentials() (password string, hashedPassword string, err error) {
	return SecureCredentialsWithPrefix("")
}

func SecureCredentialsWithPrefix(prefix string) (password string, hashedPassword string, err error) {
	// 1. Generate a secure random string for the password
	password = SecureRandomString(12, API_KEY_CHAR_SET)
	password = prefix + password

	// 2. Hash the password
	// Bcrypt automatically generates a unique salt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	return password, string(hashedBytes), nil
}

// This is not used normally so we should be fine
func HashPassword(password string) string {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error hashing password: ", err)
	}
	return string(hashedBytes)
}

func VerifyPassword(plainPassword, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword))
	return err == nil
}
