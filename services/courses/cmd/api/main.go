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

	courseshttp "github.com/Necromemeser/Cringearium-go/services/courses/internal/adapters/http"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/adapters/postgres"
	"github.com/Necromemeser/Cringearium-go/services/courses/internal/application"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	databaseURL := os.Getenv("DATABASE_URL"); if databaseURL == "" { log.Fatal("DATABASE_URL is not set") }
	db, err := postgres.New(logger, databaseURL); if err != nil { log.Fatalf("database connection failed: %v", err) }; defer db.Close()
	if err := db.Migrate(); err != nil { log.Fatalf("migration failed: %v", err) }
	courseService := application.NewCourses(db); handler := courseshttp.NewHandler(courseService)
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
	server := &http.Server{Addr: ":8082", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	serverErrors := make(chan error, 1)
	go func() { log.Println("Cringearium Courses started on :8082"); if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed { serverErrors <- err } }()
	shutdown := make(chan os.Signal, 1); signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	select { case err := <-serverErrors: log.Fatalf("server error: %v", err); case sig := <-shutdown: log.Printf("received signal: %v", sig) }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second); defer cancel(); if err := server.Shutdown(ctx); err != nil { log.Fatalf("graceful shutdown failed: %v", err) }; log.Println("Cringearium Courses stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(http.StatusOK); fmt.Fprintln(w, `{"status":"ok"}`) }
