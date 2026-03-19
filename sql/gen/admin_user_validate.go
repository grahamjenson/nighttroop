//go:build ignore

// This file generates the SQL creation for an admin user
// it will print out
package main

import (
	"flag"
	"fmt"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// TODO: store run migrations so they dont run again and again
	var code string
	var password string
	var passwordHash string
	var secret string
	flag.StringVar(&password, "pw", "", "Password for the admin user")
	flag.StringVar(&passwordHash, "pwh", "", "Hashed password for the admin user")
	flag.StringVar(&secret, "secret", "", "TOTP secret for the admin user")
	flag.StringVar(&code, "code", "", "TOTP code for the admin user (will generate secret and hash if not provided)")

	flag.Parse()

	if password == "" {
		panic("Password is required")
	}

	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		panic(fmt.Sprintf("Password does not match hash: %v", err))
	}

	valid := totp.Validate(code, secret)
	if !valid {
		panic("Invalid TOTP code")
	}

	fmt.Println("All Good")

}
