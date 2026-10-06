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

	aihttp "github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/adapters/http"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/adapters/llm"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/adapters/postgres"
	"github.com/Necromemeser/Cringearium-go/services/ai-chat/internal/application"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	databaseURL := os.Getenv("DATABASE_URL")
	llmURL := os.Getenv("LLM_BASE_URL")
	llmModel := os.Getenv("LLM_MODEL")
	if databaseURL == "" || llmURL == "" || llmModel == "" {
		log.Error("configuration error", "missing_database_url", databaseURL == "", "missing_llm_base_url", llmURL == "", "missing_llm_model", llmModel == "")
		os.Exit(1)
	}

	db, err := postgres.New(log, databaseURL)
	if err != nil { log.Error("database connection failed", "error", err); os.Exit(1) }
	defer db.Close()
	if err := db.Migrate(); err != nil { log.Error("migration failed", "error", err); os.Exit(1) }

	repository := postgres.NewRepository(db)
	llmClient := llm.NewClient(log, llmURL, os.Getenv("LLM_API_KEY"), llmModel)
	service := application.NewChatService(repository, llmClient)
	handler := aihttp.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})
	aihttp.RegisterRoutes(mux, handler)

	server := &http.Server{Addr: ":8084", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	serverErrors := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErrors:
		log.Error("server error", "error", err)
	case sig := <-shutdown:
		log.Info("shutdown signal received", "signal", sig.String())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil { log.Error("graceful shutdown failed", "error", err) }
}
