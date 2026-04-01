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

	_ "github.com/mattn/go-sqlite3"
	"nighttroop.t1d.nz/background/backup"
	"nighttroop.t1d.nz/background/sync"
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

	setupLogging()

	// Make all envars and config in one struct
	serverConfig, err := config.NewServerConfig()
	if err != nil {
		return fmt.Errorf("cannot create server config: %w", err)
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
	backupJobs := backup.SetupBackupJobs(serverConfig, controlDB, s3Client)
	backupJobs.Start()

	syncJobs := sync.SetupSyncJobs(serverConfig, controlDB)
	syncJobs.Start()
	//
	// HTTP
	// Run web server.
	httpServer := setupHTTP(controlDB, serverConfig)

	// Wait for signal.
	<-ctx.Done()

	fmt.Println("\nReceived signal, shutting down")

	shutdownCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server Shutdown: %v", err)
	}

	// Stop the loops
	syncJobs.Stop()
	backupJobs.Stop()

	// Force a final backup of all dbs
	backupJobs.BackupToS3() // Force a final backup before exiting

	// Close control and User dbs
	controlDB.Close()
	fmt.Println("shut down")
	return nil
}

func setupHTTP(controlDB *db.ControlDB, serverConfig config.ServerConfig) *http.Server {
	fmt.Printf("listening on 8080\n")

	// The outer most layer is CORS and CSRF protection
	// The big caveat is that GET HEAD and OPTIONS requests will not have CSRF protection
	// SO NO STATE CHANGING IN GET REQUESTS
	corsCSRFMiddleware := http.NewCrossOriginProtection()
	httpServer := &http.Server{Addr: ":8080", Handler: corsCSRFMiddleware.Handler(server.GetMainHandler(controlDB, serverConfig))}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	return httpServer
}

func setupLogging() {
	// 1. Create a JSON handler writing to standard output
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Set minimum log level
	})

	// 2. Create the logger and set it as the global default
	logger := slog.New(handler)
	slog.SetDefault(logger)
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
