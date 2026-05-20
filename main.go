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

	"github.com/getsentry/sentry-go"
	_ "github.com/mattn/go-sqlite3"
	slogmulti "github.com/samber/slog-multi"
	slogsentry "github.com/samber/slog-sentry/v2"
	"nighttroop.t1d.nz/background"
	"nighttroop.t1d.nz/background/backup"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/config"
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

	logger := setupLogging()

	// Make all envars and config in one struct
	serverConfig, err := config.NewServerConfig(logger)
	if err != nil {
		return fmt.Errorf("cannot create server config: %w", err)
	}

	if serverConfig.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:           serverConfig.SentryDSN,
			EnableTracing: false,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	// Create a context that will be cancelled when the program receives a SIGTERM signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	//
	// Try and restore database files from S3 (we should )
	//
	s3Client := s3utils.NewS3Client(serverConfig)

	// Run some tests to make sure the server is setup properly
	err = ValidateServer(s3Client)
	if err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	// This will download all files from S3 if they exist
	err = backup.TryRestoreAllFiles(s3Client)
	if err != nil {
		return fmt.Errorf("Failed to restore backups: %w", err)
	}

	//
	// Setup Control DB
	//
	controlDB, err := db.NewControlDB(serverConfig)
	if err != nil {
		return fmt.Errorf("cannot create control db: %w", err)
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server Shutdown Error", "err", err)
	}

	slog.Info("Stopping background jobs...")
	backgroundManager.Stop() // Stop background jobs

	backgroundManager.BackupJobManager.BackupToS3() // Force a final backup before exiting

	// Close control and User dbs
	slog.Info("Closing Database connections")
	controlDB.Close()
	slog.Info("Server shutdown complete")
	return nil
}

func setupHTTP(controlDB *db.ControlDB, serverConfig config.ServerConfig) *http.Server {
	// The outer most layer is CORS and CSRF protection
	// The big caveat is that GET HEAD and OPTIONS requests will not have CSRF protection
	// SO NO STATE CHANGING IN GET REQUESTS
	corsCSRFMiddleware := http.NewCrossOriginProtection()
	httpServer := &http.Server{
		Addr:         ":8080",
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

func setupLogging() *slog.Logger {
	// 1. Create a JSON handler writing to standard output
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Set minimum log level
	})

	sentryHandler := slogsentry.Option{
		Level: slog.LevelError, // Only send errors to Sentry
	}.NewSentryHandler()

	logger := slog.New(slogmulti.Fanout(jsonHandler, sentryHandler))

	slog.SetDefault(logger)
	return logger
}

func ValidateServer(c *s3utils.S3Client) error {
	_, _, err := security.SecureCredentials()
	if err != nil {
		return fmt.Errorf("error generating secure credentials: %w", err)
	}

	// Test s3 client is working wither permissions otherwise fail
	err = c.TestS3Client()
	if err != nil {
		return fmt.Errorf("error testing S3 client: %w", err)
	}

	return nil
}
