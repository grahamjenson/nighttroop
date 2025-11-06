package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"

	_ "github.com/mattn/go-sqlite3"
	"nighttroop.t1d.nz/db"
	"nighttroop.t1d.nz/router"
	"nighttroop.t1d.nz/sql"
	"nighttroop.t1d.nz/templates"
)

var userCreateLimiter = rate.NewLimiter(10, 100)
var userCreateMu sync.Mutex

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ServerConfig struct {
	s3Bucket string
	s3URL    string
}

func run() error {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// BOOTSTRAP
	err = templates.CreateTemplates()
	if err != nil {
		return fmt.Errorf("cannot create templates: %w", err)
	}
	var serverConfig = ServerConfig{
		s3Bucket: os.Getenv("S3_BUCKET"),
		s3URL:    os.Getenv("S3_URL"),
	}

	// Create a context that will be cancelled when the program receives a SIGTERM signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	controlDB, err := db.NewControlDB(serverConfig.s3URL, serverConfig.s3Bucket)
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
	err = controlDB.Open(&sql.ControlCreate)
	if err != nil {
		return err
	}

	err = controlDB.RestoreAndOpenAllUsers()
	if err != nil {
		return fmt.Errorf("cannot restore users: %w", err)
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
	r.Use(router.GetUserMiddleware(controlDB))

	r.Post("/guess", func(w http.ResponseWriter, r *http.Request) {
		userDB := router.GetUser(r.Context())
		if userDB == nil {
			fmt.Println("User NOT found")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		choice := r.FormValue("choice")
		if choice != "even" && choice != "odd" {
			http.Error(w, "Invalid choice", http.StatusBadRequest)
			return
		}

		correctGuess := false
		// Simple Game of Guessing if the number is even or odd
		randomBool := rand.NextBool()
		if randomBool == true && choice == "even" {
			// Correct guess
			correctGuess = true
		}
		if randomBool == false && choice == "odd" {
			correctGuess = true
		}

		accept := r.Header.Get("Accept")

		if accept == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("{\"userID\": \"%s\", \"apiKey\": \"%s\"}", userID, apiKey)))
			return
		}

		if accept == "text/htmx" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			templates.IndexHTMLTemplate.ExecuteTemplate(w, "htmx_guess.html", templates.GuessHTML{Choise: choice, Correct: correctGuess})
			return
		}

		w.Write([]byte(fmt.Sprintf("User created: %s, apiKey %s \n", userID, apiKey)))

	})

	r.Get("/score", func(w http.ResponseWriter, r *http.Request) {
		userDB := router.GetUser(r.Context())

		if userDB == nil {
			fmt.Println("User NOT found")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		score, err := userDB.GetScoresSum()
		if err != nil {
			fmt.Println("Error getting scores sum", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Start a transaction.
		w.Write([]byte(fmt.Sprintf("User found: %s %v\n", userDB.Subdomain, score)))
	})

	server := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	return server
}

func controlRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(AdminOnly)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ControlHTMLTemplate.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/user", func(w http.ResponseWriter, r *http.Request) {
		userCreateMu.Lock()
		allowed := userCreateLimiter.Allow()
		userCreateMu.Unlock()
		if !allowed {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		// Check to see if there is 10k users and return error if there is
		if controlDB.GetNUsersDBs() > 10000 {
			http.Error(w, "Full: Too many users", http.StatusTooManyRequests)
			return
		}

		// TODO replace middleware with Function, since this is not mostly used
		subdomain := router.GetSubdomain(r.Context())
		if subdomain != "control" {
			// Must use control subdomain
			http.NotFound(w, r)
			return
		}

		userID, apiKey, err := controlDB.CreateUser(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Number of Users %d, Number of Routines %d\n", controlDB.GetNUsersDBs(), runtime.NumGoroutine())

		// buf := make([]byte, 1<<16) // 64 KB buffer
		// n := runtime.Stack(buf, true)
		// fmt.Printf("%s", buf[:n])

		// TODO: Handle JSON and HTMX return
		accept := r.Header.Get("Accept")

		if accept == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("{\"userID\": \"%s\", \"apiKey\": \"%s\"}", userID, apiKey)))
			return
		}
		if accept == "text/htmx" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			templates.IndexHTMLTemplate.ExecuteTemplate(w, "htmx_user.html", templates.UserHTML{Subdomain: userID, APIKey: apiKey})
			return
		}

		w.Write([]byte(fmt.Sprintf("User created: %s, apiKey %s \n", userID, apiKey)))
	})

	return r
}

func userRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(router.GetUserDBMiddbleware(controlDB))
	r.Get("/", userIndex)

	return r
}
