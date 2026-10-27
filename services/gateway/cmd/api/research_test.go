package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestResearchReadWriteAndExportRequireAdmin(t *testing.T) {
	hits := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if !strings.HasPrefix(r.URL.Path, "/internal/admin/research/") {
			t.Error(r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := "student"
		if r.Header.Get("Authorization") == "Bearer admin-test" {
			role = "admin"
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7,"role":"` + role + `"}`))
	}))
	defer auth.Close()
	coursesURL, _ := url.Parse(upstream.URL)
	authURL, _ := url.Parse(auth.URL)
	gateway := &Gateway{coursesURL: coursesURL, authURL: authURL, client: http.DefaultClient, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	for _, method := range []string{"GET", "POST"} {
		for _, test := range []struct {
			token  string
			status int
		}{{"", 401}, {"student-test", 403}, {"admin-test", 200}} {
			req := httptest.NewRequest(method, "/api/admin/research/studies/1/export?dataset=participants", nil)
			req.Header.Set("X-User-ID", "1")
			if test.token != "" {
				req.Header.Set("Authorization", "Bearer "+test.token)
			}
			response := httptest.NewRecorder()
			gateway.researchProxy().ServeHTTP(response, req)
			if response.Code != test.status {
				t.Fatalf("%s %s: got %d", method, test.token, response.Code)
			}
		}
	}
	if hits != 2 {
		t.Fatalf("non-admin reached research service: %d", hits)
	}
}
