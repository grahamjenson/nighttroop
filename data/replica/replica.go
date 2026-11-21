// Package replica replicates the control DB to S3 with Litestream, as it is written, and restores it from its replica
package replica

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	// location is where the replica is in S3, which the control DB it is of keeps beside it, see Owns
	location string
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
	client := Client(cfg, bucket, encrypted)
	return &Replica{client: client, location: client.Endpoint + "/" + client.Bucket + "/" + client.Path}
}

// ownerFile is the file beside the control DB at path that names the replica it was restored from or replicates to
func ownerFile(path string) string {
	return path + ".replica"
}

// Owns is whether the control DB at path is known to be the one this replica is of: it was restored from it, or has
// replicated to it. A DB that replicated before it kept this file, known by Litestream's folder beside it, is too.
func (r *Replica) Owns(path string) bool {
	owner, err := os.ReadFile(ownerFile(path))
	if errors.Is(err, fs.ErrNotExist) {
		_, err = os.Stat(litestream.NewDB(path).MetaPath())
		return err == nil
	}
	return err == nil && string(owner) == r.location
}

func (r *Replica) own(path string) error {
	return os.WriteFile(ownerFile(path), []byte(r.location), 0o600)
}

// Adopt settles a control DB at path that this replica is not known to be of, as one made while S3 was off, or was
// another bucket or host, before it is replicated over the replica. It is whether to replicate it:
//   - with no replica in S3, it is the replica's first DB;
//   - one never set up holds nothing, so it is moved aside and the replica restored in its place;
//   - one set up is left as it is and not replicated, as only a person knows which of the two to keep.
func (r *Replica) Adopt(ctx context.Context, path string, setUp func() (bool, error)) (bool, error) {
	if r == nil || r.Owns(path) {
		return true, nil
	}
	staged := path + ".restoring"
	_ = os.Remove(staged)
	found, err := r.Restore(ctx, staged)
	_ = os.Remove(ownerFile(staged))
	if err != nil {
		return false, err
	}
	if !found {
		return true, r.own(path)
	}
	if keep, err := setUp(); err != nil || keep {
		return false, errors.Join(err, os.Remove(staged))
	}
	if err := moveDB(path, fmt.Sprintf("%s.replaced-%d", path, time.Now().UnixMilli())); err != nil {
		return false, err
	}
	if err := os.Rename(staged, path); err != nil {
		return false, err
	}
	return true, r.own(path)
}

// moveDB renames the DB at from, with its write-ahead log and shared memory, to to, leaving out any not there
func moveDB(from, to string) error {
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.Rename(from+suffix, to+suffix); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

// RestoreOver restores the control DB from its replica in place of the one at path, which does not open, as a file a
// disk damaged. The one that does not open is kept beside it, with its log, as path.unopened-<time>, and put back if
// there is no replica or the restore fails. It is done once: while a DB is kept from an earlier time, it does nothing,
// so a server that is started again and again does not fill its disk with copies. It is whether it restored.
func (r *Replica) RestoreOver(ctx context.Context, path string) (bool, error) {
	if r == nil {
		return false, nil
	}
	if kept, _ := filepath.Glob(path + ".unopened-*"); len(kept) > 0 {
		return false, nil
	}
	aside := fmt.Sprintf("%s.unopened-%d", path, time.Now().UnixMilli())
	if err := moveDB(path, aside); err != nil {
		return false, err
	}
	if found, err := r.Restore(ctx, path); err != nil || !found {
		return false, errors.Join(err, moveDB(aside, path))
	}
	// Litestream's own files beside the DB are of the one that did not open
	return true, os.RemoveAll(litestream.NewDB(path).MetaPath())
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
	if err := r.own(path); err != nil {
		return err
	}
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

// restoreTimeout is the longest a restore of the control DB takes
var restoreTimeout = 15 * time.Minute

// Restore restores the control DB from its replica to path, which must not be there, and is false when it has no
// replica. A write-ahead log left at path, by a DB moved away without it, is kept beside it as path.replaced-<time>-wal.
func (r *Replica) Restore(ctx context.Context, path string) (bool, error) {
	if r == nil {
		return false, nil
	}
	// The log of a DB that was at path, and Litestream's files of it, are not the restored one's: SQLite would write
	// the log over the restored DB when it opens
	aside := fmt.Sprintf("%s.replaced-%d", path, time.Now().UnixMilli())
	for _, suffix := range []string{"-wal", "-shm"} {
		if err := os.Rename(path+suffix, aside+suffix); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return false, err
		}
	}
	if err := os.RemoveAll(litestream.NewDB(path).MetaPath()); err != nil {
		return false, err
	}
	// A replica that S3 stops sending would otherwise hold the start for ever
	ctx, cancel := context.WithTimeout(ctx, restoreTimeout)
	defer cancel()
	opt := litestream.NewRestoreOptions()
	opt.OutputPath = path
	err := litestream.NewReplicaWithClient(nil, r.client).Restore(ctx, opt)
	if errors.Is(err, litestream.ErrTxNotAvailable) || errors.Is(err, litestream.ErrNoSnapshots) {
		return false, nil
	}
	if err == nil {
		err = r.own(path)
	}
	return err == nil, err
}

// RestoreByHand restores the control DB from its replica to path, which must not be there, for a person with the
// server's config. It only reads the bucket: it does not take the S3 lock, which a running server holds and a tool
// would wait for, or leave behind for the next start to wait out.
func RestoreByHand(ctx context.Context, cfg *config.ServerConfig, path string) error {
	if cfg.S3URL == "" || path == "" {
		return errors.New("set S3_URL, and give the path to restore the control DB to")
	}
	cfg.Status.S3Working = true
	found, err := New(cfg, s3utils.NewS3Client(cfg)).Restore(ctx, path)
	if err != nil {
		return fmt.Errorf("cannot restore the control DB: %w", err)
	} else if !found {
		return errors.New("the control DB has no replica")
	}
	return nil
}
