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

	aihttp "github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/adapters/http"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/adapters/llm"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/adapters/postgres"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/application"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Error("configuration error", "missing", "DATABASE_URL")
		return
	}

	coursesURL := envOrDefault("COURSES_URL", "http://courses:8082")
	llmURL := os.Getenv("LLM_BASE_URL")
	llmModel := os.Getenv("LLM_MODEL")
	if llmURL == "" || llmModel == "" {
		logger.Error("configuration error", "missing", "LLM_BASE_URL and LLM_MODEL")
		return
	}

	db, err := postgres.New(logger, databaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		return
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		logger.Error("migration failed", "error", err)
		return
	}

	repository := postgres.NewRepository(db)
	courses := aihttp.NewCoursesClient(logger, coursesURL)
	llmClient := llm.NewClient(logger, llmURL, os.Getenv("LLM_API_KEY"), llmModel)
	service := application.NewAdaptiveTestService(repository, courses, llmClient)
	handler := aihttp.NewHandler(logger, service)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	aihttp.RegisterRoutes(mux, handler)

	server := &http.Server{
		Addr:              ":8085",
		Handler:           requestLogger(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("ai-tests service started", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("server error", "error", err)
	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		return
	}
	logger.Info("ai-tests service stopped")
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status":"ok"}`)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(body)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		status := recorder.status
		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		}

		if status >= http.StatusInternalServerError {
			logger.Error("http request", attrs...)
		} else {
			logger.Info("http request", attrs...)
		}
	})
}
