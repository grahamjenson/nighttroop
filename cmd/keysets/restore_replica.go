//go:build ignore

package main

import (
	"context"
	"flag"
	"log"

	"nighttroop.t1d.nz/data/replica"
	"nighttroop.t1d.nz/utils/config"
)

// restore_replica restores the control DB from its Litestream replica in S3, with the server's config (its HOST_NAME,
// S3_URL, S3_BUCKET, keys and KEY_SETS, whose MAC key makes the replica's SSE-C key), as the server does at start. It
// only reads the bucket, so it works while the server runs. A site's backup is decrypted with decrypt_tar_lz4_enc.go.
//
//	go run cmd/keysets/restore_replica.go <output path, not there yet>
func main() {
	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err := replica.RestoreByHand(context.Background(), cfg, flag.Arg(0)); err != nil {
		log.Fatal(err)
	}
}
