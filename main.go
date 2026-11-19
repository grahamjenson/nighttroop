package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"

	"syscall"
	"time"

	// Every IANA timezone, for the timezone setting, wherever the server runs
	_ "time/tzdata"

	"github.com/getsentry/sentry-go"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
	"nighttroop.t1d.nz/background"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/replica"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/migrateup"
	"nighttroop.t1d.nz/data/sql/sqltiming"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/server/components/errortoast"
	"nighttroop.t1d.nz/socketio/engineio"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/i18n"
	"nighttroop.t1d.nz/utils/page"
	"nighttroop.t1d.nz/utils/restore"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/usage"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run starts the server: from its config, it restores the control DB from its replica in S3, opens and replicates it,
// then runs the background jobs and serves HTTP until it is told to stop, when it shuts down
func run() (err error) {
	// Make all envars and config in one struct
	serverConfig, err := config.NewServerConfig()
	if err != nil {
		return fmt.Errorf("cannot create server config: %w", err)
	}

	setupSentry(serverConfig)
	// Events are sent in the background, so wait for them on the way out: otherwise the errors of a failed start or of
	// the final backup are lost
	defer sentry.Flush(2 * time.Second)
	// A server that does not start, or stops with an error, has no sites up and restarts in a loop, so the admin is
	// paged: before Sentry is flushed, as a page is logged as an error too
	admin := page.New(serverConfig.PagerDutyRoutingKey, serverConfig.PagerDutyEventsURL, serverConfig.HostName).Admin
	started := false
	defer func() { pageFailure(admin, started, err) }()
	sqltiming.SetSlowThreshold(time.Duration(serverConfig.SlowQueryMS) * time.Millisecond)
	sqltiming.SetSamplePercent(serverConfig.QuerySamplePercent)

	// Create a context that will be cancelled when the program receives a SIGTERM signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Listens before the slow steps of a start, so a port in use fails at once and each request is answered "starting"
	starting := &startingHandler{}
	httpServer := setupHTTP(serverConfig, starting)
	serveErr, err := serve(httpServer)
	if err != nil {
		return fmt.Errorf("cannot serve HTTP: %w", err)
	}
	defer httpServer.Close()

	// S3 is off without S3_URL; with it, the server only starts once S3 works
	s3Client := s3utils.NewS3Client(serverConfig)
	restart, err := startS3(ctx, s3Client.Start, controldb.GetPath(serverConfig), s3Client.LockWhenWorking, admin)
	if err != nil {
		return err
	}
	serverConfig.Status.S3Down = restart != nil
	// Released on the way out, after the last backup, so only then may another server back up to the bucket
	defer s3Client.Unlock()

	// Before anything is restored or opened, as none of it works on a full disk
	if err := makeRoom(serverConfig, admin, func() (float64, bool) {
		used, total, ok := usage.Disk(serverConfig.DataDir)
		return total - used, ok
	}); err != nil {
		return err
	}

	// Run some tests to make sure the server is setup properly
	if err := ValidateServer(); err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	control := replica.New(serverConfig, s3Client)
	if err := restoreFromS3(ctx, serverConfig, control, admin); err != nil {
		return err
	}
	if !serverConfig.Status.S3Working {
		control = nil // as restoreFromS3 turned S3 off
	}
	if err := restoreUnopened(ctx, serverConfig, control, admin); err != nil {
		return err
	}
	controlDB, err := openControlDB(ctx, serverConfig, s3Client, control, admin)
	if err != nil {
		return err
	}

	// A server whose lock another took stops replicating at once, as the replica is that server's now
	go func() {
		<-s3Client.Lost()
		control.Abandon()
	}()

	backgroundManager := background.NewBackGroundManager(context.Background(), controlDB, serverConfig, s3Client)
	backgroundManager.Start()
	defer backgroundManager.Stop()

	fmt.Printf("Serving HTTP on port %d\n", serverConfig.HostPort)
	starting.ready(server.WithCrossOriginProtection(server.GetMainHandler(controlDB, serverConfig)))
	started = true

	// A server that cannot serve stops as a signal stops it, with a last backup, rather than exiting at once while the
	// background jobs run and before Sentry has its error
	select {
	case <-ctx.Done():
		shutdown(httpServer, backgroundManager, controlDB)
		return nil
	case err := <-serveErr:
		shutdown(httpServer, backgroundManager, controlDB)
		return fmt.Errorf("cannot serve HTTP: %w", err)
	case <-s3Client.Lost():
		control.Abandon()
		shutdown(httpServer, backgroundManager, controlDB)
		return errors.New("another server took this server's S3 lock, so the server restarts and waits for the lock")
	case <-restart:
		shutdown(httpServer, backgroundManager, controlDB)
		return errRestartWithS3
	}
}

// errRestartWithS3 stops a server that started without S3, for it to start again with S3: not a failure
var errRestartWithS3 = errors.New("S3 works again, so the server restarts to restore, replicate and back up with it")

// pageFailure pages the admin with a call when the server did not start, or stopped with an error after it had: its
// sites are down. Each has one message, so a server that fails each time it is started again is one incident. A server
// told to stop, or restarting with S3, is not a failure.
func pageFailure(admin page.Func, started bool, err error) {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, errRestartWithS3) {
		return
	}
	message := "The server did not start, so its sites are down"
	if started {
		message = "The server stopped with an error, so its sites are down"
	}
	admin(context.Background(), page.Call, message, "error", err)
}

// startS3 starts S3. If S3 is not working and the control DB is here, the server starts without S3, so an S3 outage
// does not take the sites down: nothing is restored, replicated or backed up, and the channel closes once S3 works again
// and this server holds its lock, for the server to restart with S3. Without its control DB the server does not
// start, as it would start new, and later replicate that over its replica. A server that starts without S3 pages the
// admin.
func startS3(ctx context.Context, start func(context.Context) error, controlPath string, lockWhenWorking func(context.Context) bool, admin page.Func) (<-chan struct{}, error) {
	err := start(ctx)
	if err == nil || ctx.Err() != nil || !utils.FileExists(controlPath) {
		return nil, err
	}
	// The start has tried S3 again for about half a minute, so the admin is paged: the sites are up, with no backups
	admin(ctx, page.SMS, "S3 is not working, so the server started from its DBs here without backups, and restarts when S3 works", "error", err)
	restart := make(chan struct{})
	go func() {
		if lockWhenWorking(ctx) {
			close(restart)
		}
	}()
	return restart, nil
}

// minFreeDisk is the least space the DBs' disk must have for the server to start: room for the DBs to migrate and for
// their write-ahead logs
const minFreeDisk = 100 << 20

// makeRoom frees space on the DBs' disk when a start finds less than minFreeDisk free, as a server on a full disk
// cannot open its DBs. It removes only what the server makes again, the lesser loss first: the local copies of the
// backups, which are made again from the DBs, then the logs DB, and pages the admin. With too little free after that
// the server does not start, and says so.
func makeRoom(serverConfig *config.ServerConfig, admin page.Func, free func() (float64, bool)) error {
	before, ok := free()
	if !ok || before >= minFreeDisk {
		return nil
	}
	_ = filepath.WalkDir(serverConfig.BackupDir(), func(path string, entry fs.DirEntry, err error) error {
		if err == nil && entry.Type().IsRegular() {
			_ = os.Remove(path)
		}
		return nil
	})
	if after, _ := free(); after < minFreeDisk {
		for _, suffix := range []string{"", "-wal", "-shm"} {
			_ = os.Remove(logsdb.GetPath(serverConfig) + suffix)
		}
	}
	after, _ := free()
	admin(context.Background(), page.SMS, "The DBs' disk was nearly full at start, so the server removed its local backup copies and, if still short, its logs DB",
		"free_mb_before", int(before)>>20, "free_mb_after", int(after)>>20)
	if after < minFreeDisk {
		return fmt.Errorf("the DBs' disk has %d MB free, and the server needs %d MB to start: free some space on it", int(after)>>20, minFreeDisk>>20)
	}
	return nil
}

// startingHandler answers every request 503 "starting" until the server is ready, then with its handler. A start can
// take minutes, to restore from S3 and open every site, and a client or a deploy must tell that from a server that is
// down: the process of a start that fails exits.
type startingHandler struct{ handler atomic.Pointer[http.Handler] }

func (s *startingHandler) ready(handler http.Handler) { s.handler.Store(&handler) }

func (s *startingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := s.handler.Load(); handler != nil {
		(*handler).ServeHTTP(w, r)
		return
	}
	w.Header().Set("Retry-After", "5")
	// An htmx request puts the reply in a toast, so it gets one to read rather than the JSON
	if httputils.GetAcceptType(r) == httputils.AcceptTypeHTMX {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = errortoast.Toast("", i18n.T(r.Context(), "errors.server_starting")).Render(r.Context(), w)
		return
	}
	apiutils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "starting"})
}

// restoreFromS3 restores a missing control DB from its replica in S3, where there is one; without one the server
// starts with a new control DB. A site whose DB is missing or does not open is restored from its own backup once the
// control DB is open (KeepRestoringUnavailableSites).
func restoreFromS3(ctx context.Context, serverConfig *config.ServerConfig, control *replica.Replica, admin page.Func) error {
	// Before any restore runs, so none of the staging folders it removes is in use
	restore.RemoveStaging(serverConfig)
	path := controldb.GetPath(serverConfig)
	switch {
	case utils.FileExists(path):
		// One made while S3 was off, or was another bucket or host, is not replicated over the replica
		replicate, err := control.Adopt(ctx, path, func() (bool, error) { return controlIsSetUp(ctx, serverConfig) })
		if err != nil {
			// The control DB is here, so the sites are served from it, without S3, as it is not known to be its replica's
			turnS3Off(serverConfig, admin, config.S3OffCheck, "The control DB could not be checked against its replica in S3, so S3 is off: nothing is replicated or backed up", err)
			return nil
		}
		if !replicate {
			turnS3Off(serverConfig, admin, config.S3OffConflict, "The control DB here is set up, and is not the one replicated in S3, so S3 is off: nothing is replicated or backed up. Move it away to restore the replica, or delete the replica to replicate this one", nil)
			return nil
		}
		slog.Info("Control DB is here, so it is not restored", "path", path)
		return nil
	case control == nil:
		slog.Warn("Control DB is missing and S3 is off, so the server starts with a new control DB", "path", path)
		return nil
	}
	slog.Info("Control DB is missing, so it is restored from its replica in S3", "path", path)
	found, err := control.Restore(ctx, path)
	if err != nil {
		return fmt.Errorf("cannot restore the control DB from its replica: %w", err)
	}
	if !found {
		slog.Warn("Control DB has no replica in S3, so the server starts with a new control DB", "path", path)
		return nil
	}
	slog.Info("Control DB restored from its replica in S3", "path", path)
	return nil
}

// restoreUnopened restores the control DB from its replica when the one here does not open, as a file a disk damaged,
// and pages the admin: the one here is kept beside it. A control DB whose migration fails is left, as that is the
// release's fault, which update.sh rolls back, and its replica has the same DB.
func restoreUnopened(ctx context.Context, serverConfig *config.ServerConfig, control *replica.Replica, admin page.Func) error {
	path := controldb.GetPath(serverConfig)
	if control == nil || !utils.FileExists(path) {
		return nil
	}
	opened, err := controldb.NewDB(ctx, serverConfig)
	if err == nil {
		return opened.Close()
	}
	if errors.As(err, new(*migrateup.FailedError)) {
		return nil
	}
	restored, rerr := control.RestoreOver(ctx, path)
	if rerr != nil {
		return fmt.Errorf("the control DB does not open (%w), and cannot be restored from its replica: %w", err, rerr)
	}
	if restored {
		admin(ctx, page.SMS, "The control DB did not open, so it was restored from its replica in S3. The one that did not open is kept beside it", "error", err, "path", path)
	}
	return nil
}

// turnS3Off turns S3 off for this run of a server whose S3 works, as its control DB must not be replicated or cannot
// be: its sites are served, nothing is replicated or backed up, the control site's banner says why, and the admin is
// paged, as only a person can fix it
func turnS3Off(serverConfig *config.ServerConfig, admin page.Func, reason, message string, err error) {
	serverConfig.Status.S3Working, serverConfig.Status.S3Off = false, reason
	admin(context.Background(), page.SMS, message, "error", err)
}

// controlIsSetUp is whether the control DB here has a control admin, so holds a server someone set up
func controlIsSetUp(ctx context.Context, serverConfig *config.ServerConfig) (bool, error) {
	control, err := controldb.NewDB(ctx, serverConfig)
	if err != nil {
		return false, err
	}
	defer control.Close()
	admins, err := control.Queries.CountControlAdmins(ctx)
	return admins > 0, err
}

// openControlDB opens the control DB and every site's, and restores, in the background, each site whose DB is missing
// or does not open from its backup, with an admin to log in with in development and staging, and prints the setup code
// of a server not yet set up
func openControlDB(ctx context.Context, serverConfig *config.ServerConfig, s3Client *s3utils.S3Client, control *replica.Replica, admin page.Func) (*db.ControlDB, error) {
	controlDB, err := db.NewControlDB(context.Background(), serverConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot create control db: %w", err)
	}
	if err := controlDB.Replicate(ctx, control); err != nil {
		// The control DB is open, so the sites are served from it, without S3
		turnS3Off(serverConfig, admin, config.S3OffReplication, "The control DB's replication could not start, so S3 is off: nothing is replicated or backed up", err)
	}
	// A site whose DB is missing or does not open is unavailable, so the server starts without it and serves the rest
	// while it is restored from its backup, tried again for about an hour, then the admin is paged
	go controlDB.KeepRestoringUnavailableSites(ctx, s3Client.DownloadFilesFromS3, db.RestoreRetries, setupPagerDuty(serverConfig).Admin)
	// Development has an admin to log in with, and staging its configured one; anywhere else a new server is set up
	// with the code it prints
	if serverConfig.IsDev() {
		if err := controlDB.SeedDevAdmin(ctx); err != nil {
			return nil, fmt.Errorf("cannot create the development admin: %w", err)
		}
	}
	if serverConfig.IsStaging() {
		if err := controlDB.SeedStagingAdmin(ctx); err != nil {
			return nil, fmt.Errorf("cannot create the staging admin: %w", err)
		}
	}
	if code := controlDB.SetupCode(); code != "" {
		fmt.Printf("######## SET UP NIGHTTROOP ########\nOpen %s and enter the setup code: %s\n###################################\n", serverConfig.SetupURL(), code)
	}
	return controlDB, nil
}

// shutdown stops serving HTTP, within 5 seconds, then the background jobs, backs the sites up to S3 a last time, and
// closes the DBs, sending the control DB's last writes to its replica
func shutdown(httpServer *http.Server, backgroundManager *background.BackGroundManager, controlDB *db.ControlDB) {
	slog.Info("Received shutdown signal, shutting down gracefully...")

	slog.Info("Shutting down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server Shutdown Error", "err", err)
	}
	slog.Info("HTTP server shutdown complete")
	aslog.RefusedRequests.Flush(context.Background())

	slog.Info("Stopping background jobs...")
	backgroundManager.Stop()
	slog.Info("Stopped background jobs complete")

	slog.Info("Performing final backup to S3...")
	finalBackup := aslog.NewJob(context.Background(), "FinalBackup")
	backgroundManager.BackupSiteJobManager.BackupToS3(finalBackup)
	slog.Info("Final backup complete")

	slog.Info("Closing Database connections")
	controlDB.Close()
	slog.Info("Server shutdown complete")
}

// setupPagerDuty is the pager of the admin, through PagerDuty when PAGERDUTY_ROUTING_KEY is set, and records whether
// it is in serverConfig.Status. PagerDuty is optional: without it each page is still logged as an error.
func setupPagerDuty(serverConfig *config.ServerConfig) *page.Pager {
	if serverConfig.PagerDutyRoutingKey == "" {
		slog.Warn("PagerDuty is not set up, set PAGERDUTY_ROUTING_KEY to page the admin")
	}
	serverConfig.Status.PagerDutyWorking = serverConfig.PagerDutyRoutingKey != ""
	return page.New(serverConfig.PagerDutyRoutingKey, serverConfig.PagerDutyEventsURL, serverConfig.HostName)
}

// setupSentry starts Sentry error reporting and records whether it is working in serverConfig.Status.
// Sentry is optional: without it errors are still logged, just not reported.
func setupSentry(serverConfig *config.ServerConfig) {
	if serverConfig.SentryDSN == "" {
		slog.Warn("Sentry is not set up, set SENTRY_DSN to report errors")
		return
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:           serverConfig.SentryDSN,
		Environment:   serverConfig.Deployment,
		EnableTracing: false,
		// Each event is searchable by what is logged with it, such as its site and identity, then cleared of tokens
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			return httputils.RedactSentryEvent(aslog.SentryEvent(event), hint)
		},
	})
	if err != nil {
		slog.Warn("Sentry is not working, running without error reporting", "error", err)
		return
	}
	serverConfig.Status.SentryWorking = true
}

// maxHeaderBytes is the most a request's headers may hold, far above a browser's cookies or an uploader's token, so a
// client cannot make each connection hold Go's default of 1 MB. Larger headers are refused with 431.
const maxHeaderBytes = 64 << 10

// newHTTPServer is the server for handler on addr, with the limits every connection to it has
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    5 * time.Second,  // Max time to read the entire request
		WriteTimeout:   10 * time.Second, // Max time to write the response
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}
}

func setupHTTP(serverConfig *config.ServerConfig, handler http.Handler) *http.Server {
	// The outer most layer is CORS and CSRF protection, so no state changing in GET requests
	httpServer := newHTTPServer(fmt.Sprintf(":%d", serverConfig.HostPort), handler)
	// Shutdown does not close websockets, so its start closes every Socket.IO session, and none writes to a site's DB
	// after its last backup
	closing, closeSessions := context.WithCancel(context.Background())
	httpServer.RegisterOnShutdown(closeSessions)
	httpServer.BaseContext = func(net.Listener) context.Context { return engineio.WithClosing(context.Background(), closing) }
	return httpServer
}

// serve listens on httpServer's address, or gives why it cannot, such as its port being in use, then serves in the
// background, and the channel gets the error that stops it other than its Shutdown or Close
func serve(httpServer *http.Server) (<-chan error, error) {
	listener, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		return nil, err
	}
	failed := make(chan error, 1)
	go func() {
		if err := httpServer.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			failed <- err
		}
	}()
	return failed, nil
}

func ValidateServer() error {
	_, _, err := security.SecureCredentials(bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error generating secure credentials: %w", err)
	}

	return nil
}
