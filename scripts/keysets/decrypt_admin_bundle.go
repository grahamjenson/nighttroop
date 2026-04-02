//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

func main() {
	// Create Empty Config and
	keysetRaw, ok := os.LookupEnv(consts.KEY_SETS_ENV_KEY)
	if !ok {
		log.Fatal(fmt.Errorf("KEY_SETS not set in environment"))
	}

	configs := security.NewSecurityConfig(keysetRaw, consts.Production, "localhost", false) // This will create the keysets and store them in the config struct

	val, ok := os.LookupEnv("ENCRYPTED_ADMIN_BUNDLE")
	if !ok {
		log.Fatal(fmt.Errorf("ENCRYPTED_ADMIN_BUNDLE not set in environment"))
	}

	plainText, err := configs.Decrypt(val)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot decrypt admin bundle: %w", err))
	}

	fmt.Println("Decrypted admin bundle:", plainText)
}
