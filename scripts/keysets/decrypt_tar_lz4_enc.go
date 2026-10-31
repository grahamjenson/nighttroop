//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/security"
)

// something like cat file.tar.lz4.enc | go run decrypt_tar_lz4_enc.go
func main() {
	// Create Empty Config and
	configs := config.ServerConfig{}
	keysetRaw, ok := os.LookupEnv(consts.KEY_SETS_ENV_KEY)
	if !ok {
		log.Fatal(fmt.Errorf("KEY_SETS not set in environment"))
	}

	sc := security.NewSecurityConfig(keysetRaw, consts.Production, "localhost", false) // This will create the keysets and store them in the config struct
	configs.Security = sc

	s3Client := &s3utils.S3Client{ServerConfig: &configs}

	// This takes in stdin and outputs decrypted and decompressed data to folders
	s3Client.ExtractTarLZ4(os.Stdin, ".")
}
