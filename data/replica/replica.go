// Package replica replicates the control DB to S3 with Litestream, as it is written, and restores it from its replica
package replica

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/benbjohnson/litestream"
	"github.com/benbjohnson/litestream/s3"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

// The control DB, which every sign in, site and identity change writes, is sent within a second, as often as
// Litestream does. Compaction keeps what a restore downloads to Litestream's daily snapshot, the hours since, and the
// minutes since.
var levels = litestream.CompactionLevels{{Level: 0}, {Level: 1, Interval: time.Minute}, {Level: 2, Interval: time.Hour}}

// Name is the name of the control DB's replica, under the bucket's replica prefix
const Name = "control"

// Replica is the control DB's replica. Nil, as when S3 is off, replicates and restores nothing.
type Replica struct {
	client litestream.ReplicaClient
	store  *litestream.Store
	path   string
}

// New is the replica of the control DB in the server's S3 bucket, or nil when S3 is off. Over HTTPS S3 encrypts it
// with the server's key (SSE-C); over HTTP, as the development MinIO is, S3 refuses SSE-C, so it is not.
func New(cfg *config.ServerConfig, bucket *s3utils.S3Client) *Replica {
	if !cfg.Status.S3Working {
		return nil
	}
	encrypted := strings.HasPrefix(bucket.Endpoint, "https://")
	if !encrypted {
		slog.Warn("S3_URL is not HTTPS, so the control DB's replica is not encrypted")
	}
	return &Replica{client: Client(cfg, bucket, encrypted)}
}

// Client is the Litestream S3 client of the control DB's replica, encrypted with its key when encrypted
func Client(cfg *config.ServerConfig, bucket *s3utils.S3Client, encrypted bool) *s3.ReplicaClient {
	client := s3.NewReplicaClient()
	client.AccessKeyID, client.SecretAccessKey = bucket.AccessKeyID, bucket.SecretAccessKey
	client.Bucket, client.Region, client.Endpoint, client.ForcePathStyle = bucket.Bucket, bucket.Region, bucket.Endpoint, bucket.ForcePathStyle
	client.Path = bucket.ReplicaPrefix() + Name
	if encrypted {
		key, err := cfg.Security.ReplicaKey(Name)
		if err != nil {
			// Not a key, so the client refuses to start rather than replicate unencrypted
			slog.Error("Cannot make the replica's key", "error", err)
			key = "none"
		}
		client.SSECustomerAlgorithm, client.SSECustomerKey = "AES256", key
	}
	return client
}

// Start replicates the control DB at path
func (r *Replica) Start(ctx context.Context, path string) error {
	if r == nil {
		return nil
	}
	db := litestream.NewDB(path)
	db.MonitorInterval = time.Second
	db.Replica = litestream.NewReplicaWithClient(db, r.client)
	db.Replica.SyncInterval = time.Second
	r.path, r.store = path, litestream.NewStore([]*litestream.DB{db}, levels)
	// Litestream's 15 seconds lists two levels of the replica each time
	r.store.L0RetentionCheckInterval = time.Minute
	return r.store.Open(ctx)
}

// Close sends the control DB's last writes and stops replicating it. It is synced first: Litestream only sends a DB's
// last writes as it closes it once it has read the DB, which it first does a second after start.
func (r *Replica) Close(ctx context.Context) error {
	if r == nil || r.store == nil {
		return nil
	}
	_, err := r.store.SyncDB(ctx, r.path, true)
	return errors.Join(err, r.store.Close(ctx))
}

// Restore restores the control DB from its replica to path, which must not be there, and is false when it has no
// replica
func (r *Replica) Restore(ctx context.Context, path string) (bool, error) {
	if r == nil {
		return false, nil
	}
	opt := litestream.NewRestoreOptions()
	opt.OutputPath = path
	err := litestream.NewReplicaWithClient(nil, r.client).Restore(ctx, opt)
	if errors.Is(err, litestream.ErrTxNotAvailable) || errors.Is(err, litestream.ErrNoSnapshots) {
		return false, nil
	}
	return err == nil, err
}
