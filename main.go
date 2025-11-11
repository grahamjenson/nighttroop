package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"syscall"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
	"nighttroop.t1d.nz/db"
	"nighttroop.t1d.nz/router"
	"nighttroop.t1d.nz/sql"
	"nighttroop.t1d.nz/templates"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ServerConfig struct {
	isDev bool

	s3Bucket string
	s3URL    string

	// User Rate Limits
	UserReadRateLimit  int
	UserReadRateBurst  int
	UserWriteRateLimit int
	UserWriteRateBurst int

	// User Limits
	MaxUsers            int
	MaxUserDBSizeMB     int
	UserCreateRateLimit int
	UserCreateRateBurst int
}

func getServerConfig() ServerConfig {
	sc := ServerConfig{
		isDev: false,
	}
	flag.BoolVar(&sc.isDev, "dev", false, "Run in development mode")
	flag.Parse()

	envFiles := []string{}
	if sc.isDev {
		envFiles = append(envFiles, ".env.dev")
	}
	err := godotenv.Load(envFiles...)

	if err != nil {
		// Non Fatal error
		fmt.Fprintln(os.Stderr, "Error loading .env file")
	}

	s3Bucket, s3BucketPresent := os.LookupEnv("S3_BUCKET")
	if !s3BucketPresent || s3Bucket == "" {
		log.Fatal("S3_BUCKET environment variable not set")
	}
	sc.s3Bucket = s3Bucket

	s3URL, s3URLPresent := os.LookupEnv("S3_URL")
	if !s3URLPresent || s3URL == "" {
		log.Fatal("S3_URL environment variable not set")
	}
	sc.s3URL = s3URL

	return sc
}

func loadAndProcessStaticContent() error {
	// Load static content
	err := templates.CreateTemplates()
	if err != nil {
		return fmt.Errorf("cannot load templates: %w", err)
	}
	return nil
}

func setUpDBs(serverConfig ServerConfig) (*db.ControlDB, error) {
	controlDB, err := db.NewControlDB(serverConfig.s3URL, serverConfig.s3Bucket)
	if err != nil {
		return nil, fmt.Errorf("cannot create control db: %w", err)
	}

	err = controlDB.TryRestoreDBFile()
	if err != nil {
		fmt.Println("Failed to restore Control DB: %w", err)
		// TODO dont fail here if the DB file doesnt exist in S3
		return nil, fmt.Errorf("cannot restore control db: %w", err)
	}

	// Open Databse, hook up litesrtream, and migrate
	err = controlDB.Open(&sql.ControlCreate)
	if err != nil {
		return nil, err
	}

	err = controlDB.RestoreAndOpenAllUsers()
	if err != nil {
		return nil, fmt.Errorf("cannot restore users: %w", err)
	}
	return controlDB, nil
}

func run() error {
	var err error

	// Make all envars and config in one struct
	serverConfig := getServerConfig()

	// BOOTSTRAP
	err = loadAndProcessStaticContent()
	if err != nil {
		return fmt.Errorf("cannot create templates: %w", err)
	}

	// Create a context that will be cancelled when the program receives a SIGTERM signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	controlDB, err := setUpDBs(serverConfig)
	if err != nil {
		return fmt.Errorf("cannot set up dbs: %w", err)
	}
	// Run web server.
	server := setupHTTP(ctx, controlDB)

	// Setup other background jobs
	setupBackgroundJobs(controlDB)

	// Wait for signal.
	<-ctx.Done()

	fmt.Println("\nreceived signal, shutting down")

	shutdownCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server Shutdown: %v", err)
	}

	stopBackgroundJobs(controlDB)

	// Force a final backup before exiting
	controlDB.ForceBackupSync()

	controlDB.Close()
	fmt.Println("shut down")
	return nil
}

func setupHTTP(ctx context.Context, controlDB *db.ControlDB) *http.Server {
	fmt.Printf("listening on 8080\n")

	server := &http.Server{Addr: ":8080", Handler: router.GetMainHandler(controlDB)}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	return server
}

func setupBackgroundJobs(controlDB *db.ControlDB) {
	// Start backup Jobs
	controlDB.StartBackupLoops()
	//loop that infinitely goes over a series of background jobs and runs them
}

func stopBackgroundJobs(controlDB *db.ControlDB) error {
	// Stop the backup loops
	controlDB.StopBackupLoop()

	return nil
}
