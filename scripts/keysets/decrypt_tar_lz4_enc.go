//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"nighttroop.t1d.nz/config"
	"nighttroop.t1d.nz/utils"
)

// something like cat file.tar.lz4.enc | go run decrypt_tar_lz4_enc.go
func main() {
	// Create Empty Config and
	configs := config.ServerConfig{}
	s3Client := &utils.S3Client{ServerConfig: &configs}

	// Create
	err := configs.ConfigCreateEncryptionKeys(map[string]string{}, map[string]string{}) // This will create the keysets and store them in the config struct
	if err != nil {
		log.Fatal(fmt.Errorf("cannot create server config: %w", err))
	}

	// This takes in stdin and outputs decrypted and decompressed data to folders
	s3Client.ExtractTarLZ4(os.Stdin, ".")
}
