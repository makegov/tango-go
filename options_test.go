package tango

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWithUserAgent(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(
		WithAPIKey("k"),
		WithBaseURL(srv.URL),
		WithUserAgent("my-app/1.0"),
		WithRetries(0),
		WithTimeout(2*time.Second),
	)
	_, _ = c.ListAgencies(context.Background(), nil)
	if captured != "my-app/1.0" {
		t.Errorf("expected User-Agent=my-app/1.0, got %q", captured)
	}
}

func TestWithRetriesAndBackoff(t *testing.T) {
	// Ensure the options are accepted without panic.
	c := NewClient(
		WithAPIKey("k"),
		WithRetries(3),
		WithRetryBackoff(100*time.Millisecond),
	)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := NewClient(
		WithAPIKey("k"),
		WithHTTPClient(custom),
	)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}
