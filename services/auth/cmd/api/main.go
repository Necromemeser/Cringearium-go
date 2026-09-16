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

	"github.com/Necromemeser/Cringearium-go/services/auth/internal/adapters/bcrypt"
	httpadapter "github.com/Necromemeser/Cringearium-go/services/auth/internal/adapters/http"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/adapters/jwt"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/adapters/postgres"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/application"
	"github.com/Necromemeser/Cringearium-go/services/auth/internal/config"
)

const bcryptCost = 12

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	db, err := postgres.New(logger, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	userRepository := postgres.NewUserRepository(db)
	passwordHasher := bcrypt.NewHasher(bcryptCost)
	tokenService := jwt.NewTokenService(
		cfg.JWTSecret,
		cfg.JWTTTL,
	)

	auth := application.NewAuth(
		userRepository,
		passwordHasher,
		tokenService,
	)

	handler := httpadapter.NewHandler(auth)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /api/auth/register", handler.Register)
	mux.HandleFunc("POST /api/auth/login", handler.Login)
	mux.HandleFunc("GET /api/auth/users/{id}", handler.GetByID)
	mux.HandleFunc("GET /api/auth/users", handler.GetUser)

	server := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Cringearium Auth started on %s", cfg.ServerAddr)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(
		shutdown,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case err := <-serverErrors:
		log.Fatalf("server error: %v", err)

	case sig := <-shutdown:
		log.Printf("received signal: %v", sig)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	log.Println("Cringearium Auth stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, `{"status":"ok"}`)
}
