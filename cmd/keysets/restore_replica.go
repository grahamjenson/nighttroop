//go:build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"nighttroop.t1d.nz/data/replica"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

// restore_replica restores the control DB from its Litestream replica in S3, with the server's config (its HOST_NAME,
// S3_URL, S3_BUCKET, keys and KEY_SETS, whose MAC key makes the replica's SSE-C key), as the server does at start. A
// site's backup is decrypted with decrypt_tar_lz4_enc.go.
//
//	go run cmd/keysets/restore_replica.go <output path, not there yet>
func main() {
	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	s3 := s3utils.NewS3Client(cfg)
	if err := s3.Start(ctx); err != nil {
		log.Fatal(err)
	}
	found, err := replica.New(cfg, s3).Restore(ctx, flag.Arg(0))
	if err != nil {
		log.Fatal(fmt.Errorf("cannot restore the control DB: %w", err))
	} else if !found {
		log.Fatal("the control DB has no replica")
	}
}
