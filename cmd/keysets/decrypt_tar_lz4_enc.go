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

// something like cat file.tar.lz4.enc | go run decrypt_tar_lz4_enc.go <its object key>, as host/backup/a.tar.lz4.enc,
// which it was encrypted for; without one, only a backup from before backups were bound decrypts
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
	objectKey := ""
	if len(os.Args) > 1 {
		objectKey = os.Args[1]
	}
	if err := s3Client.ExtractTarLZ4(os.Stdin, ".", objectKey); err != nil {
		log.Fatal(fmt.Errorf("failed to extract backup: %w", err))
	}
}
