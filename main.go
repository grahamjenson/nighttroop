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
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/security"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run starts the server: from its config, it restores its missing DBs from their replicas in S3, opens the control DB
// and replicates it and every site's, then runs the background jobs and serves HTTP until it is told to stop, when it
// shuts down
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

	// S3 is off without S3_URL; with it, the server only starts once S3 works
	s3Client := s3utils.NewS3Client(serverConfig)
	if err := s3Client.Start(ctx); err != nil {
		return err
	}

	// Run some tests to make sure the server is setup properly
	if err := ValidateServer(); err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	replicas := replica.New(serverConfig, s3Client)
	if !utils.FileExists(controldb.GetPath(serverConfig)) {
		if _, err := replicas.Restore(ctx, "control", controldb.GetPath(serverConfig)); err != nil {
			return fmt.Errorf("cannot restore the control DB: %w", err)
		}
	}
	controlDB, err := openControlDB(ctx, serverConfig, replicas)
	if err != nil {
		return err
	}

	backgroundManager := background.NewBackGroundManager(controlDB, serverConfig)
	backgroundManager.Start()
	defer backgroundManager.Stop()

	fmt.Printf("Starting HTTP server on port %d\n", serverConfig.HostPort)
	httpServer := setupHTTP(controlDB, serverConfig)
	serveErr := serve(httpServer)

	// A server that cannot serve, as when its port is in use, stops as a signal stops it, with a last backup, rather
	// than exiting at once while the background jobs run and before Sentry has its error
	select {
	case <-ctx.Done():
		shutdown(httpServer, backgroundManager, controlDB)
		return nil
	case err := <-serveErr:
		shutdown(httpServer, backgroundManager, controlDB)
		return fmt.Errorf("cannot serve HTTP: %w", err)
	}
}

// openControlDB opens the control DB and every site's, restoring a live site's missing DB from its replica, and
// replicates them, with an admin to log in with in development and staging, and prints the setup code of a server not
// yet set up
func openControlDB(ctx context.Context, serverConfig *config.ServerConfig, replicas *replica.Replicas) (*db.ControlDB, error) {
	controlDB, err := db.NewControlDB(context.Background(), serverConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot create control db: %w", err)
	}
	if err := controlDB.Replicate(ctx, replicas); err != nil {
		return nil, fmt.Errorf("cannot open restored sites: %w", err)
	}
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

// shutdown stops serving HTTP, within 5 seconds, then the background jobs, then sends each DB's last writes to its
// replica as it closes the DBs
func shutdown(httpServer *http.Server, backgroundManager *background.BackGroundManager, controlDB *db.ControlDB) {
	slog.Info("Received shutdown signal, shutting down gracefully...")

	slog.Info("Shutting down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server Shutdown Error", "err", err)
	}
	slog.Info("HTTP server shutdown complete")

	slog.Info("Stopping background jobs...")
	backgroundManager.Stop()
	slog.Info("Stopped background jobs complete")

	slog.Info("Closing Database connections, sending their last writes to S3")
	controlDB.Close()
	slog.Info("Server shutdown complete")
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

func setupHTTP(controlDB *db.ControlDB, serverConfig *config.ServerConfig) *http.Server {
	// The outer most layer is CORS and CSRF protection, so no state changing in GET requests
	httpServer := newHTTPServer(fmt.Sprintf(":%d", serverConfig.HostPort),
		server.WithCrossOriginProtection(server.GetMainHandler(controlDB, serverConfig)))
	// Shutdown does not close websockets, so its start closes every Socket.IO session, and none writes to a site's DB
	// after its last backup
	closing, closeSessions := context.WithCancel(context.Background())
	httpServer.RegisterOnShutdown(closeSessions)
	httpServer.BaseContext = func(net.Listener) context.Context { return engineio.WithClosing(context.Background(), closing) }
	return httpServer
}

// serve serves httpServer in the background, and the channel gets the error that stops it other than its Shutdown,
// such as its port being in use
func serve(httpServer *http.Server) <-chan error {
	failed := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			failed <- err
		}
	}()
	return failed
}

func ValidateServer() error {
	_, _, err := security.SecureCredentials(bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error generating secure credentials: %w", err)
	}

	return nil
}
