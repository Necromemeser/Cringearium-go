package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type userResponse struct {
	ID int64 `json:"id"`
}

type Gateway struct {
	authURL    *url.URL
	coursesURL *url.URL
	client     *http.Client
	log        *slog.Logger
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	authURL, err := serviceURL("AUTH_URL", "http://auth:8081")
	if err != nil {
		log.Fatal(err)
	}

	coursesURL, err := serviceURL("COURSES_URL", "http://courses:8082")
	if err != nil {
		log.Fatal(err)
	}

	gateway := &Gateway{
		authURL:    authURL,
		coursesURL: coursesURL,
		client:     &http.Client{Timeout: 5 * time.Second},
		log:        logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", gateway.health)
	mux.Handle("POST /api/auth/register", gateway.authProxy())
	mux.Handle("POST /api/auth/login", gateway.authProxy())
	mux.Handle("GET /api/auth/me", gateway.authProxy())
	mux.Handle("GET /api/courses", gateway.coursesProxy(false))
	mux.Handle("GET /api/courses/{id}", gateway.coursesProxy(false))
	mux.Handle("POST /api/courses/{id}/enroll", gateway.coursesProxy(true))
	mux.Handle("GET /api/courses/{id}/access", gateway.coursesProxy(true))

	server := &http.Server{
		Addr:              envOrDefault("SERVER_ADDR", ":8080"),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Cringearium Gateway started on %s", server.Addr)
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
}

func (g *Gateway) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (g *Gateway) authProxy() http.Handler {
	return g.reverseProxy(g.authURL, false)
}

func (g *Gateway) coursesProxy(protected bool) http.Handler {
	proxy := g.reverseProxy(g.coursesURL, true)
	if !protected {
		return proxy
	}

	return g.requireAuth(proxy)
}

func (g *Gateway) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, g.authURL.String()+"/api/auth/me", nil)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", "Bearer "+parts[1])

		resp, err := g.client.Do(req)
		if err != nil {
			g.log.Error("auth service request failed", "error", err)
			http.Error(w, "authentication service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if resp.StatusCode != http.StatusOK {
			g.log.Error("auth service returned unexpected status", "status", resp.StatusCode)
			http.Error(w, "authentication service error", http.StatusBadGateway)
			return
		}

		var user userResponse
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil || user.ID <= 0 {
			http.Error(w, "authentication service error", http.StatusBadGateway)
			return
		}

		r.Header.Del("X-User-ID")
		r.Header.Set("X-User-ID", strconv.FormatInt(user.ID, 10))
		next.ServeHTTP(w, r)
	})
}

func (g *Gateway) reverseProxy(target *url.URL, stripAuth bool) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)
		if stripAuth {
			r.Header.Del("Authorization")
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		g.log.Error("upstream request failed", "path", r.URL.Path, "error", err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}
	return proxy
}

func serviceURL(name, fallback string) (*url.URL, error) {
	value := envOrDefault(name, fallback)
	return url.Parse(value)
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
