package tango

import (
	"context"
	"net/http"
	"testing"
)

func TestListAPIKeysBuildsCorrectPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.ListAPIKeys(context.Background())
	assertPathContains(t, capturedURL, "/api/api-keys/")
}

func TestListAPIKeysReturnsResult(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"keys":[{"id":"key1","name":"My Key"}],"count":1}`))
	})
	result, err := c.ListAPIKeys(context.Background())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
