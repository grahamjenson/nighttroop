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

// restore_replica restores a DB from its replica in S3, decrypted with KEY_SETS, with the server's config (its
// HOST_NAME, S3_URL, S3_BUCKET and keys), as the server does at start:
//
//	go run cmd/keysets/restore_replica.go <control, or site/<site ID>> <output path, not there yet>
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
	found, err := replica.New(cfg, s3).Restore(ctx, flag.Arg(0), flag.Arg(1))
	if err != nil {
		log.Fatal(fmt.Errorf("cannot restore %s: %w", flag.Arg(0), err))
	} else if !found {
		log.Fatalf("%s has no replica", flag.Arg(0))
	}
}
