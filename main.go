package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
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
	"nighttroop.t1d.nz/data/sql/sqltiming"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/socketio/engineio"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/page"
	"nighttroop.t1d.nz/utils/restore"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/security"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run starts the server: from its config, it restores the control DB from its replica in S3, opens and replicates it,
// then runs the background jobs and serves HTTP until it is told to stop, when it shuts down
func run() error {
	// Make all envars and config in one struct
	serverConfig, err := config.NewServerConfig()
	if err != nil {
		return fmt.Errorf("cannot create server config: %w", err)
	}

	setupSentry(serverConfig)
	// Events are sent in the background, so wait for them on the way out: otherwise the errors of a failed start or of
	// the final backup are lost
	defer sentry.Flush(2 * time.Second)
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
	if err := s3Client.Start(ctx); err != nil {
		return err
	}

	// Run some tests to make sure the server is setup properly
	if err := ValidateServer(); err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	control := replica.New(serverConfig, s3Client)
	if err := restoreFromS3(ctx, serverConfig, control); err != nil {
		return err
	}
	controlDB, err := openControlDB(ctx, serverConfig, s3Client, control)
	if err != nil {
		return err
	}

	backgroundManager := background.NewBackGroundManager(context.Background(), controlDB, serverConfig, s3Client)
	backgroundManager.Start()
	defer backgroundManager.Stop()

	fmt.Printf("Serving HTTP on port %d\n", serverConfig.HostPort)
	starting.ready(server.WithCrossOriginProtection(server.GetMainHandler(controlDB, serverConfig)))

	// A server that cannot serve stops as a signal stops it, with a last backup, rather than exiting at once while the
	// background jobs run and before Sentry has its error
	select {
	case <-ctx.Done():
		shutdown(httpServer, backgroundManager, controlDB)
		return nil
	case err := <-serveErr:
		shutdown(httpServer, backgroundManager, controlDB)
		return fmt.Errorf("cannot serve HTTP: %w", err)
	}
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
	apiutils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "starting"})
}

// restoreFromS3 restores a missing control DB from its replica in S3, where there is one; without one the server
// starts with a new control DB. A site whose DB is missing or does not open is restored from its own backup once the
// control DB is open (KeepRestoringUnavailableSites).
func restoreFromS3(ctx context.Context, serverConfig *config.ServerConfig, control *replica.Replica) error {
	// Before any restore runs, so none of the staging folders it removes is in use
	restore.RemoveStaging(serverConfig)
	path := controldb.GetPath(serverConfig)
	switch {
	case utils.FileExists(path):
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

// openControlDB opens the control DB and every site's, and restores, in the background, each site whose DB is missing
// or does not open from its backup, with an admin to log in with in development and staging, and prints the setup code
// of a server not yet set up
func openControlDB(ctx context.Context, serverConfig *config.ServerConfig, s3Client *s3utils.S3Client, control *replica.Replica) (*db.ControlDB, error) {
	controlDB, err := db.NewControlDB(context.Background(), serverConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot create control db: %w", err)
	}
	if err := controlDB.Replicate(ctx, control); err != nil {
		return nil, fmt.Errorf("cannot replicate the control db: %w", err)
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
