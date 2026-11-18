// Package replica replicates the control DB to S3 with Litestream, as it is written, and restores it from its replica
package replica

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/benbjohnson/litestream"
	"github.com/benbjohnson/litestream/s3"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

// Name is the name of the control DB's replica, under the bucket's replica prefix
const Name = "control"

// Replica is the control DB's replica. Nil, as when S3 is off, replicates and restores nothing.
type Replica struct {
	client litestream.ReplicaClient
	// mu guards store, which Abandon ends while the server runs
	mu    sync.Mutex
	store *litestream.Store
	path  string
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

// Start replicates the control DB at path with Litestream's defaults: its writes are sent within a second
func (r *Replica) Start(ctx context.Context, path string) error {
	if r == nil {
		return nil
	}
	db := litestream.NewDB(path)
	db.Replica = litestream.NewReplicaWithClient(db, r.client)
	r.path, r.store = path, litestream.NewStore([]*litestream.DB{db}, litestream.DefaultCompactionLevels)
	// Litestream logs each second's sync, and each compaction, so only its warnings and errors are kept
	r.store.Logger = slog.New(warnings{slog.Default().Handler()}).With(litestream.LogKeySystem, litestream.LogSystemStore)
	db.SetLogger(r.store.Logger.With(litestream.LogKeyDB, filepath.Base(path)))
	return r.store.Open(ctx)
}

// warnings passes on only the logs of handler that are warnings or errors
type warnings struct{ slog.Handler }

func (w warnings) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= slog.LevelWarn && w.Handler.Enabled(ctx, level)
}

func (w warnings) WithAttrs(attrs []slog.Attr) slog.Handler {
	return warnings{w.Handler.WithAttrs(attrs)}
}

func (w warnings) WithGroup(name string) slog.Handler {
	return warnings{w.Handler.WithGroup(name)}
}

// Close sends the control DB's last writes and stops replicating it. It is synced first: Litestream only sends a DB's
// last writes as it closes it once it has read the DB, which it first does a second after start.
func (r *Replica) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.store == nil {
		return nil
	}
	_, err := r.store.SyncDB(ctx, r.path, true)
	return errors.Join(err, r.store.Close(ctx))
}

// Abandon stops replicating at once and sends nothing more, as when another server took this server's S3 lock, and
// the replica is that server's. Close then does nothing.
func (r *Replica) Abandon() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.store != nil {
		// Litestream sends a DB's last writes as it closes it, but not with a context that has ended
		ended, end := context.WithCancel(context.Background())
		end()
		_ = r.store.Close(ended)
		r.store = nil
	}
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
