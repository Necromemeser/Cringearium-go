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

	courseshttp "github.com/Necromemeser/Cringearium-go/services/courses/internal/adapters/http"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/adapters/postgres"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/application"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Error("configuration load failed", "error", "DATABASE_URL is not set")
		os.Exit(1)
	}

	db, err := postgres.New(logger, databaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	courseService := application.NewCourses(db)
	handler := courseshttp.NewHandler(courseService)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("GET /api/courses", handler.GetAll)
	mux.HandleFunc("GET /api/courses/{id}", handler.GetByID)
	mux.HandleFunc("POST /api/courses/{id}/enroll", handler.Enroll)
	mux.HandleFunc("GET /api/courses/{id}/access", handler.GetAccess)
	mux.HandleFunc("GET /api/users/me/courses", handler.GetEnrolled)
	mux.HandleFunc("GET /api/courses/{id}/progress", handler.GetProgress)
	mux.HandleFunc("POST /api/pages/{pageId}/complete", handler.CompletePage)
	mux.HandleFunc("GET /api/pages/{pageId}/test", handler.GetTest)
	mux.HandleFunc("POST /api/tests/{testId}/attempts", handler.SubmitTest)
	mux.HandleFunc("GET /internal/courses/{id}/ai-context", handler.GetAIContext)

	server := &http.Server{
		Addr:              ":8082",
		Handler:           requestLogger(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("courses service started", "address", server.Addr)
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

	logger.Info("courses service stopped")
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
