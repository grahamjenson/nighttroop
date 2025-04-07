package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/mattn/go-sqlite3"
	"sqlitemt.t1d.nz/db"
	"sqlitemt.t1d.nz/html"
	mtmiddleware "sqlitemt.t1d.nz/middleware"
	mtsql "sqlitemt.t1d.nz/sql"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var err error

	// SETUP
	err = html.CreateTemplates()
	if err != nil {
		return fmt.Errorf("cannot create templates: %w", err)
	}

	// Create a context that will be cancelled when the program receives a SIGTERM signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	controlDB, err := db.NewControlDB("s3://mybucket.localhost:9000/")
	if err != nil {
		return fmt.Errorf("cannot create control db: %w", err)
	}

	err = controlDB.TryRestoreDBFile()
	if err != nil {
		fmt.Println("Failed to restore Control DB: %w", err)

	}
	// Open Databse, hook up litesrtream, and migrate
	err = controlDB.Open(mtsql.ControlCreate)
	if err != nil {
		return err
	}
	defer controlDB.Close()

	err = controlDB.CheckTenantDBs()
	if err != nil {
		return fmt.Errorf("cannot check tenant dbs: %w", err)
	}
	// Run web server.
	fmt.Printf("listening on 8080\n")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(mtmiddleware.GetUserDBMiddbleware(controlDB))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		html.IndexHTMLTemplate.Execute(w, nil)
	})

	r.Post("/tenant", func(w http.ResponseWriter, r *http.Request) {
		subdomain := mtmiddleware.GetSubdomain(r.Context())
		if subdomain != "control" {
			// Must use control subdomain
			http.NotFound(w, r)
			return
		}

		tenantID, err := controlDB.CreateTenant(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Number of Tenants %d, Number of Routines %d\n", controlDB.GetNTenantsDBs(), runtime.NumGoroutine())

		// buf := make([]byte, 1<<16) // 64 KB buffer
		// n := runtime.Stack(buf, true)
		// fmt.Printf("%s", buf[:n])

		w.Write([]byte(fmt.Sprintf("Tenant created: %s\n", tenantID)))
	})

	r.Get("/score", func(w http.ResponseWriter, r *http.Request) {
		tenantDB := mtmiddleware.GetTenant(r.Context())
		if tenantDB == nil {
			http.Error(w, "Tenant not found", http.StatusNotFound)
			return
		}

		// Start a transaction.
		w.Write([]byte(fmt.Sprintf("Tenant found: %s\n", tenantDB.Subdomain)))
	})

	controlDB.StartBackupLoop()

	server := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	// Wait for signal.
	<-ctx.Done()

	fmt.Println("\nreceived signal, shutting down")

	shutdownCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server Shutdown: %v", err)
	}

	// Stop the backup loops
	controlDB.StopBackupLoop()

	// Force a final backup before exiting
	controlDB.ForceBackupSync()

	// Close the database connections
	controlDB.Close()

	fmt.Println("shut down")
	return nil
}
