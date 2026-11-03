package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
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
	"nighttroop.t1d.nz/server"
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

func run() error {
	var err error

	// Make all envars and config in one struct
	serverConfig, err := config.NewServerConfig()
	if err != nil {
		return fmt.Errorf("cannot create server config: %w", err)
	}

	setupSentry(serverConfig)

	// Create a context that will be cancelled when the program receives a SIGTERM signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	//
	// Try and restore database files from S3 (we should )
	//
	s3Client := s3utils.NewS3Client(serverConfig)

	// S3 is optional: without it there is no restore on start or backup upload
	if err := s3Client.TestS3Client(ctx); err != nil {
		slog.Warn("S3 is not working, running without S3 restore or backups", "error", err)
	} else {
		serverConfig.Status.S3Working = true
		// Without versioning each backup replaces the last, which is worth an error but not stopping backups
		if err := s3Client.ConfigureVersioning(ctx); err != nil {
			slog.Error("S3 versioning is not set up, so each backup replaces the last", "error", err)
		}
	}

	// Run some tests to make sure the server is setup properly
	err = ValidateServer()
	if err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	// Before any restore runs, so none of the staging folders it removes is in use
	restore.RemoveStaging(serverConfig)

	// This will download all files from S3 if they exist
	err = restore.TryRestoreControlDB(s3Client)
	if err != nil {
		return fmt.Errorf("Failed to restore backups: %w", err)
	}

	// This will download all files from S3 if they exist
	err = restore.TryRestoreAllFiles(s3Client)
	if err != nil {
		return fmt.Errorf("Failed to restore backups: %w", err)
	}

	//
	// Setup Control DB
	//
	controlDB, err := db.NewControlDB(context.Background(), serverConfig)
	if err != nil {
		return fmt.Errorf("cannot create control db: %w", err)
	}
	// A live site whose DB file is missing, though its batch's folder is there, is restored from its batch's backup
	controlDB.RestoreMissingSites(ctx, s3Client.DownloadFilesFromS3)
	// Development has an admin to log in with, and staging its configured one; anywhere else a new server is set up
	// with the code it prints
	if serverConfig.IsDev() {
		if err := controlDB.SeedDevAdmin(ctx); err != nil {
			return fmt.Errorf("cannot create the development admin: %w", err)
		}
	}
	if serverConfig.IsStaging() {
		if err := controlDB.SeedStagingAdmin(ctx); err != nil {
			return fmt.Errorf("cannot create the staging admin: %w", err)
		}
	}
	if code := controlDB.SetupCode(); code != "" {
		fmt.Printf("######## SET UP NIGHTTROOP ########\nOpen %s and enter the setup code: %s\n###################################\n", serverConfig.SetupURL(), code)
	}

	//
	// Background jobs
	//
	backgroundManager := background.NewBackGroundManager(context.Background(), controlDB, serverConfig, s3Client)
	backgroundManager.Start()
	defer backgroundManager.Stop()
	// backupJobs := backup.SetupBackupJobs(serverConfig, controlDB, s3Client)
	// backupJobs.Start()

	// syncJobs := sync.SetupSyncJobs(serverConfig, controlDB)
	// syncJobs.Start()
	//
	// HTTP
	// Run web server.
	fmt.Printf("Starting HTTP server on port %d\n", serverConfig.HostPort)
	httpServer := setupHTTP(controlDB, serverConfig)

	// Wait for signal.
	<-ctx.Done()

	slog.Info("Received shutdown signal, shutting down gracefully...")

	slog.Info("Shutting down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server Shutdown Error", "err", err)
	}
	slog.Info("HTTP server shutdown complete")

	slog.Info("Stopping background jobs...")
	backgroundManager.Stop() // Stop background jobs

	slog.Info("Stopped background jobs complete")

	slog.Info("Performing final backup to S3...")
	// Force a final backup before exiting; a failure here can lose the latest control data, so log it
	finalBackup := aslog.NewJob(context.Background(), "FinalBackup")
	if err := backgroundManager.BackupControlJobManager.BackupToS3(finalBackup); err != nil {
		slog.ErrorContext(finalBackup, "Final control DB backup failed", "error", err)
	}
	backgroundManager.BackupSiteJobManager.BackupToS3(finalBackup) // Force a final backup before exiting
	slog.Info("Final backup complete")

	// Close control and User dbs
	slog.Info("Closing Database connections")
	controlDB.Close()
	slog.Info("Server shutdown complete")
	return nil
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
		EnableTracing: false,
		BeforeSend:    httputils.RedactSentryEvent,
	})
	if err != nil {
		slog.Warn("Sentry is not working, running without error reporting", "error", err)
		return
	}
	serverConfig.Status.SentryWorking = true
}

func setupHTTP(controlDB *db.ControlDB, serverConfig *config.ServerConfig) *http.Server {
	// The outer most layer is CORS and CSRF protection
	// The big caveat is that GET HEAD and OPTIONS requests will not have CSRF protection
	// SO NO STATE CHANGING IN GET REQUESTS
	corsCSRFMiddleware := http.NewCrossOriginProtection()
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", serverConfig.HostPort),
		Handler:      corsCSRFMiddleware.Handler(server.GetMainHandler(controlDB, serverConfig)),
		ReadTimeout:  5 * time.Second,  // Max time to read the entire request
		WriteTimeout: 10 * time.Second, // Max time to write the response
		IdleTimeout:  120 * time.Second,
	}

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
