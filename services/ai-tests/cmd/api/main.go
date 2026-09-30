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

	aihttp "github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/adapters/http"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/adapters/llm"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/adapters/postgres"
	"github.com/Necromemeser/Cringearium-go/services/ai-tests/internal/application"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	coursesURL := envOrDefault("COURSES_URL", "http://courses:8082")
	llmURL := os.Getenv("LLM_BASE_URL")
	llmModel := os.Getenv("LLM_MODEL")
	if llmURL == "" || llmModel == "" {
		log.Fatal("LLM_BASE_URL and LLM_MODEL are required")
	}

	db, err := postgres.New(logger, databaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	repository := postgres.NewRepository(db)
	courses := aihttp.NewCoursesClient(coursesURL)
	llmClient := llm.NewClient(llmURL, os.Getenv("LLM_API_KEY"), llmModel)
	service := application.NewAdaptiveTestService(repository, courses, llmClient)
	handler := aihttp.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	aihttp.RegisterRoutes(mux, handler)

	server := &http.Server{
		Addr:              ":8085",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Println("Cringearium AI-tests started on :8085")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("server error: %v", err)
	case sig := <-shutdown:
		log.Printf("received signal: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("Cringearium AI-tests stopped")
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
