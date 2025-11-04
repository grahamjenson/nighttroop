package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"

	_ "github.com/mattn/go-sqlite3"
	"sqlitemt.t1d.nz/db"
	mtmiddleware "sqlitemt.t1d.nz/middleware"
	mtsql "sqlitemt.t1d.nz/sql"
	"sqlitemt.t1d.nz/templates"
)

var tenantCreateLimiter = rate.NewLimiter(10, 100)
var tenantCreateMu sync.Mutex

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var err error

	// SETUP
	err = templates.CreateTemplates()
	if err != nil {
		return fmt.Errorf("cannot create templates: %w", err)
	}

	// Create a context that will be cancelled when the program receives a SIGTERM signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	controlDB, err := db.NewControlDB("http://localhost:9000/", "mybucket")
	if err != nil {
		return fmt.Errorf("cannot create control db: %w", err)
	}

	err = controlDB.TryRestoreDBFile()
	if err != nil {
		fmt.Println("Failed to restore Control DB: %w", err)
		// TODO dont fail here if the DB file doesnt exist in S3
		return fmt.Errorf("cannot restore control db: %w", err)
	}

	// Open Databse, hook up litesrtream, and migrate
	err = controlDB.Open(&mtsql.ControlCreate)
	if err != nil {
		return err
	}

	err = controlDB.RestoreAndOpenAllTenants()
	if err != nil {
		return fmt.Errorf("cannot restore tenants: %w", err)
	}

	// Run web server.
	server := setupHTTP(ctx, controlDB)

	controlDB.StartBackupLoops()

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

	controlDB.Close()
	fmt.Println("shut down")
	return nil
}

func setupHTTP(ctx context.Context, controlDB *db.ControlDB) *http.Server {
	fmt.Printf("listening on 8080\n")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(mtmiddleware.GetUserDBMiddbleware(controlDB))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.IndexHTMLTemplate.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/tenant", func(w http.ResponseWriter, r *http.Request) {

		tenantCreateMu.Lock()
		allowed := tenantCreateLimiter.Allow()
		tenantCreateMu.Unlock()
		if !allowed {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		// Check to see if there is 10k tenants and return error if there is
		if controlDB.GetNTenantsDBs() > 10000 {
			http.Error(w, "Full: Too many tenants", http.StatusTooManyRequests)
			return
		}

		// TODO replace middleware with Function, since this is not mostly used
		subdomain := mtmiddleware.GetSubdomain(r.Context())
		if subdomain != "control" {
			// Must use control subdomain
			http.NotFound(w, r)
			return
		}

		tenantID, apiKey, err := controlDB.CreateTenant(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Number of Tenants %d, Number of Routines %d\n", controlDB.GetNTenantsDBs(), runtime.NumGoroutine())

		// buf := make([]byte, 1<<16) // 64 KB buffer
		// n := runtime.Stack(buf, true)
		// fmt.Printf("%s", buf[:n])

		// TODO: Handle JSON and HTMX return
		accept := r.Header.Get("Accept")
		fmt.Println("accept", accept)
		if accept == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("{\"tenantID\": \"%s\", \"apiKey\": \"%s\"}", tenantID, apiKey)))
			return
		}
		if accept == "text/htmx" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			templates.IndexHTMLTemplate.ExecuteTemplate(w, "tenant.html", templates.TenantHTML{Subdomain: tenantID, APIKey: apiKey})
			return
		}

		w.Write([]byte(fmt.Sprintf("Tenant created: %s, apiKey %s \n", tenantID, apiKey)))
	})

	r.Get("/score", func(w http.ResponseWriter, r *http.Request) {
		tenantDB := mtmiddleware.GetTenant(r.Context())

		if tenantDB == nil {
			fmt.Println("Tenant NOT found")
			http.Error(w, "Tenant not found", http.StatusNotFound)
			return
		} else {
			fmt.Println("Tenant found", tenantDB.Subdomain)
		}
		score, err := tenantDB.GetScoresSum()
		if err != nil {
			fmt.Println("Error getting scores sum", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Start a transaction.
		w.Write([]byte(fmt.Sprintf("Tenant found: %s %v\n", tenantDB.Subdomain, score)))
	})

	server := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	return server
}
