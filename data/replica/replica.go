// Package replica replicates the control DB and each site's DB to S3 with Litestream, as each one is written, and
// restores one from its replica
package replica

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/benbjohnson/litestream"
	"golang.org/x/sync/errgroup"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
)

// Each site's writes are sent within siteSync, as a site's apps send theirs every 5 minutes or so, so a site sends
// about one upload a cycle: about 300 a day while it is used and none while it is not. The control DB, which every
// sign in, site and identity change writes, is sent within controlSync, as often as Litestream does. Compaction keeps
// what a restore downloads to a daily snapshot, the hours since, and for the control DB the minutes since.
const (
	siteSync    = time.Minute
	controlSync = time.Second
	snapshots   = 24 * time.Hour
)

var (
	siteLevels    = litestream.CompactionLevels{{Level: 0}, {Level: 1, Interval: time.Hour}}
	controlLevels = litestream.CompactionLevels{{Level: 0}, {Level: 1, Interval: time.Minute}, {Level: 2, Interval: time.Hour}}
)

// Replicas are the replicas of the server's DBs. Nil, as when S3 is off, replicates and restores nothing.
type Replicas struct {
	sites, control *litestream.Store
	// client is the replica client of the DB with name, "control" or "site/<site ID>"
	client func(name string) litestream.ReplicaClient
}

// New are the replicas of the server's DBs in its S3 bucket, or nil when S3 is off
func New(cfg *config.ServerConfig, s3 *s3utils.S3Client) *Replicas {
	if !cfg.Status.S3Working {
		return nil
	}
	return NewWithClients(func(name string) litestream.ReplicaClient {
		return &client{s3: s3, security: cfg.Security, prefix: s3.ReplicaPrefix() + name}
	})
}

// NewWithClients are replicas kept with client, such as a test's in a folder
func NewWithClients(client func(name string) litestream.ReplicaClient) *Replicas {
	return &Replicas{client: client}
}

// SiteName is the name of the site's replica
func SiteName(siteID string) string {
	return "site/" + siteID
}

// Start replicates the control DB at controlPath and the DBs of sites, by path to name
func (r *Replicas) Start(ctx context.Context, controlPath string, sites map[string]string) error {
	if r == nil {
		return nil
	}
	dbs := []*litestream.DB{}
	for path, name := range sites {
		dbs = append(dbs, r.db(path, name, siteSync))
	}
	r.sites = store(dbs, siteLevels, time.Hour)
	r.control = store([]*litestream.DB{r.db(controlPath, "control", controlSync)}, controlLevels, time.Minute)
	return errors.Join(r.control.Open(ctx), r.sites.Open(ctx))
}

// store keeps dbs, compacted at levels, with snapshots kept a day and the replica's L0 files removed once compacted,
// checked each l0Check
func store(dbs []*litestream.DB, levels litestream.CompactionLevels, l0Check time.Duration) *litestream.Store {
	s := litestream.NewStore(dbs, levels)
	s.SnapshotInterval, s.SnapshotRetention, s.L0RetentionCheckInterval = snapshots, snapshots, l0Check
	s.Logger = slog.New(warnings{slog.Default().Handler()})
	for _, db := range dbs {
		db.SetLogger(s.Logger.With(litestream.LogKeyDB, filepath.Base(db.Path())))
	}
	return s
}

func (r *Replicas) db(path, name string, sync time.Duration) *litestream.DB {
	db := litestream.NewDB(path)
	db.MonitorInterval = sync
	db.Replica = litestream.NewReplicaWithClient(db, r.client(name))
	db.Replica.SyncInterval = sync
	return db
}

// Add replicates the site DB at path to name, sending it at once
func (r *Replicas) Add(ctx context.Context, path, name string) error {
	if r == nil {
		return nil
	}
	if err := r.sites.RegisterDB(r.db(path, name, siteSync)); err != nil {
		return err
	}
	return r.Sync(ctx, path)
}

// Sync sends the DB at path's writes to its replica now, such as before an archive or after a change to the control
// DB that a restore must have
func (r *Replicas) Sync(ctx context.Context, path string) error {
	if r == nil {
		return nil
	}
	store := r.sites
	if r.control.FindDB(path) != nil {
		store = r.control
	}
	_, err := store.SyncDB(ctx, path, true)
	return err
}

// Remove sends the site DB at path's last writes and stops replicating it, removing what Litestream keeps beside it,
// as for a site archived or deleted; deleted also deletes its replica, name
func (r *Replicas) Remove(ctx context.Context, path, name string, deleted bool) error {
	if r == nil {
		return nil
	}
	db := r.sites.FindDB(path)
	if db == nil {
		return fmt.Errorf("the DB is not replicated")
	}
	// An archived one is synced first, as Close does
	if !deleted {
		if _, err := r.sites.SyncDB(ctx, path, true); err != nil {
			return err
		}
	}
	if err := r.sites.UnregisterDB(ctx, path); err != nil {
		return err
	}
	if err := os.RemoveAll(db.MetaPath()); err != nil {
		return err
	}
	if deleted {
		return db.Replica.Client.DeleteAll(ctx)
	}
	return nil
}

// Checkpoint moves the site DB at path's write-ahead log into it and truncates it, as Vacuum does without
// replication, once its writes are sent: a checkpoint of its own cannot, while Litestream reads the DB
func (r *Replicas) Checkpoint(ctx context.Context, path string) error {
	db := r.sites.FindDB(path)
	if db == nil {
		return fmt.Errorf("the DB is not replicated")
	}
	return db.Checkpoint(ctx, litestream.CheckpointModeTruncate)
}

// Replicated is when the site DB at path was last in step with its replica, or nil if it has not been since start
func (r *Replicas) Replicated(path string) *int64 {
	if r == nil {
		return nil
	}
	db := r.sites.FindDB(path)
	if db == nil || db.LastSuccessfulSyncAt().IsZero() {
		return nil
	}
	ms := db.LastSuccessfulSyncAt().UnixMilli()
	return &ms
}

// Close sends each DB's last writes and stops replicating them. Each is synced first, a few at once: Litestream only
// sends a DB's last writes as it closes it once it has read the DB, which it first does a sync interval after start.
func (r *Replicas) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	g := errgroup.Group{}
	g.SetLimit(closeParallelism)
	for _, store := range []*litestream.Store{r.sites, r.control} {
		for _, db := range store.DBs() {
			g.Go(func() error {
				_, err := store.SyncDB(ctx, db.Path(), true)
				return err
			})
		}
	}
	return errors.Join(g.Wait(), r.sites.Close(ctx), r.control.Close(ctx))
}

// closeParallelism is how many DBs Close syncs at once
const closeParallelism = 16

// Restore restores the DB with name from its replica to path, which must not be there, and is false when it has no
// replica
func (r *Replicas) Restore(ctx context.Context, name, path string) (bool, error) {
	if r == nil {
		return false, nil
	}
	opt := litestream.NewRestoreOptions()
	opt.OutputPath = path
	err := litestream.NewReplicaWithClient(nil, r.client(name)).Restore(ctx, opt)
	if errors.Is(err, litestream.ErrTxNotAvailable) || errors.Is(err, litestream.ErrNoSnapshots) {
		return false, nil
	}
	return err == nil, err
}

// warnings logs Litestream's warnings and errors, each with its DB, but not its routine syncs and compactions, which
// for thousands of sites would be most of the logs
type warnings struct{ slog.Handler }

func (w warnings) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= slog.LevelWarn && w.Handler.Enabled(ctx, level)
}
func (w warnings) WithAttrs(attrs []slog.Attr) slog.Handler {
	return warnings{w.Handler.WithAttrs(attrs)}
}
func (w warnings) WithGroup(name string) slog.Handler { return warnings{w.Handler.WithGroup(name)} }
