package main

import (
	"context"
	"fmt"
	"log"
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
	"nighttroop.t1d.nz/data/sql/sqltiming"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/socketio/engineio"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/httputils"
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

// run starts the server: from its config, it restores its DBs from S3, opens the control DB, then runs the background
// jobs and serves HTTP until it is told to stop, when it shuts down
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

	if err := restoreFromS3(serverConfig, s3Client); err != nil {
		return err
	}
	controlDB, err := openControlDB(ctx, serverConfig, s3Client)
	if err != nil {
		return err
	}

	backgroundManager := background.NewBackGroundManager(context.Background(), controlDB, serverConfig, s3Client)
	backgroundManager.Start()
	defer backgroundManager.Stop()

	fmt.Printf("Starting HTTP server on port %d\n", serverConfig.HostPort)
	httpServer := setupHTTP(controlDB, serverConfig)

	<-ctx.Done()
	shutdown(httpServer, backgroundManager, controlDB)
	return nil
}

// restoreFromS3 downloads the control DB and every site's files from their backups in S3, where there are any
func restoreFromS3(serverConfig *config.ServerConfig, s3Client *s3utils.S3Client) error {
	// Before any restore runs, so none of the staging folders it removes is in use
	restore.RemoveStaging(serverConfig)
	if err := restore.TryRestoreControlDB(s3Client); err != nil {
		return fmt.Errorf("Failed to restore backups: %w", err)
	}
	if err := restore.TryRestoreAllFiles(s3Client); err != nil {
		return fmt.Errorf("Failed to restore backups: %w", err)
	}
	return nil
}

// openControlDB opens the control DB and every site's, restoring a live site's missing DB from its batch's backup,
// with an admin to log in with in development and staging, and prints the setup code of a server not yet set up
func openControlDB(ctx context.Context, serverConfig *config.ServerConfig, s3Client *s3utils.S3Client) (*db.ControlDB, error) {
	controlDB, err := db.NewControlDB(context.Background(), serverConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot create control db: %w", err)
	}
	// A live site whose DB file is missing, though its batch's folder is there, is restored from its batch's backup
	if err := controlDB.RestoreMissingSites(ctx, s3Client.DownloadFilesFromS3); err != nil {
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

// shutdown stops serving HTTP, within 5 seconds, then the background jobs, backs the control DB and the sites up to
// S3 a last time, and closes the DBs
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

	slog.Info("Performing final backup to S3...")
	// Force a final backup before exiting; a failure here can lose the latest control data, so log it
	finalBackup := aslog.NewJob(context.Background(), "FinalBackup")
	if err := backgroundManager.BackupControlJobManager.BackupToS3(finalBackup); err != nil {
		slog.ErrorContext(finalBackup, "Final control DB backup failed", "error", err)
	}
	backgroundManager.BackupSiteJobManager.BackupToS3(finalBackup)
	slog.Info("Final backup complete")

	slog.Info("Closing Database connections")
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

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	return httpServer
}

func ValidateServer() error {
	_, _, err := security.SecureCredentials(bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error generating secure credentials: %w", err)
	}

	return nil
}
