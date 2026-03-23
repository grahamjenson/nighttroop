//go:build ignore
//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"nighttroop.t1d.nz/utils/config"
)

func main() {
	// Create Empty Config and
	configs := config.ServerConfig{}
	// Create
	err := configs.ConfigCreateEncryptionKeys(map[string]string{}, map[string]string{}) // This will create the keysets and store them in the config struct
	if err != nil {
		log.Fatal(fmt.Errorf("cannot create server config: %w", err))
	}

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
