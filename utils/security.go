package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	badrand "math/rand"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
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

func TOTPValidate(token, secret string) bool {
	if secret == config.DEV_ADMIN_TOTP_SECRET {
		return token == "123123"
	}
	return totp.Validate(token, secret)
}

func TOTPGenerate(accountName string, encryptFunc func(string) (string, error)) (totpSecret, totpEncSecret string, err error) {
	totpKey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "NightTroop",
		AccountName: accountName,
	})

	if err != nil {
		return "", "", fmt.Errorf("cannot generate TOTP key: %w", err)
	}

	totpSecret = totpKey.Secret()

	if db.serverConfig.Mode == config.Development {
		// We are going to override this stuff in dev, just to make it easier
		password = config.DEV_ADMIN_PASSWORD
		hashedPassword = utils.HashPassword(password)
		totpSecret = config.DEV_ADMIN_TOTP_SECRET
	}

	encTOTPSecret, err := db.serverConfig.Encrypt(totpSecret)
	if err != nil {
		return fmt.Errorf("cannot encrypt TOTP secret: %w", err)
	}

	return totpSecret, encTOTPSecret, nil
}
