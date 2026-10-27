package main

import (
	"context"
	"fmt"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration load failed", "error", err)
		os.Exit(1)
	}

	db, err := postgres.New(logger, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	userRepository := postgres.NewUserRepository(db)
	passwordHasher := bcrypt.NewHasher(bcryptCost)
	tokenService := jwt.NewTokenService(cfg.JWTSecret, cfg.JWTTTL)
	auth := application.NewAuth(userRepository, passwordHasher, tokenService)

	handler := httpadapter.NewHandler(auth)
	middleware := httpadapter.NewMiddleware(tokenService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /api/auth/register", handler.Register)
	mux.HandleFunc("POST /api/auth/login", handler.Login)
	mux.Handle("GET /api/auth/users/{id}", middleware.Admin(http.HandlerFunc(handler.GetByID)))
	mux.Handle("GET /api/auth/users", middleware.Admin(http.HandlerFunc(handler.GetUser)))
	mux.Handle("GET /api/auth/me", middleware.Auth(http.HandlerFunc(handler.Me)))

	server := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           requestLogger(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("auth service started", "address", cfg.ServerAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("http server failed", "error", err)
	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("auth service stopped")
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		}
		if status >= http.StatusInternalServerError {
			logger.Error("http request completed", attrs...)
			return
		}
		logger.Info("http request completed", attrs...)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status":"ok"}`)
}
