package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResearchSourceFailureIsNotAnEmptySuccessfulDataset(t *testing.T) {
	for _, tc := range []struct {
		status  int
		body    string
		failure bool
	}{{200, `{"events":[],"truncated":false}`, false}, {200, `{"events":[],"truncated":true}`, true}, {503, `unavailable`, true}} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("courses") != "1,2" {
				t.Error(r.URL)
			}
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(tc.body))
		}))
		_, err := NewResearchClient(server.URL).Events(context.Background(), []int64{1, 2})
		server.Close()
		if (err != nil) != tc.failure {
			t.Fatalf("status %d: err=%v", tc.status, err)
		}
	}
}
