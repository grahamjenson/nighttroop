package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"nighttroop.t1d.nz/backup"
	"nighttroop.t1d.nz/config"
	"nighttroop.t1d.nz/db"
	"nighttroop.t1d.nz/router"
	"nighttroop.t1d.nz/sync"
	"nighttroop.t1d.nz/utils"
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

	// Create a context that will be cancelled when the program receives a SIGTERM signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	//
	// Try and restore database files from S3 (we should )
	//
	s3Client, err := utils.NewS3Client(serverConfig.S3URL, serverConfig.S3Bucket, serverConfig.HostName)
	if err != nil {
		return fmt.Errorf("cannot create s3 client: %w", err)
	}

	// This will download all files from S3 if they exist
	err = backup.TryRestoreAllFiles(s3Client)
	if err != nil {
		return fmt.Errorf("Failed to restore backups: %w", err)
	}

	//
	// Setup Control DB
	//
	controlDB, err := db.NewControlDB()
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
	server := setupHTTP(controlDB, serverConfig)

	// Wait for signal.
	<-ctx.Done()

	fmt.Println("\nReceived signal, shutting down")

	shutdownCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(shutdownCtx); err != nil {
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

	server := &http.Server{Addr: ":8080", Handler: router.GetMainHandler(controlDB, serverConfig)}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	return server
}
