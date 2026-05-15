//go:build integration

// Package integration contains live-API smoke tests for tango-go.
// These tests are NOT run in normal CI — they require a real API key and
// hit the production Tango API. They are gated on the //go:build integration
// tag so they are only executed when explicitly requested:
//
//	go test -tags integration ./tests/integration/
//
// All tests read TANGO_API_KEY from the environment and skip when it is
// unset. Tests are expected to run in ≤ 5 seconds each and are idempotent
// (no side effects that can't be re-run).
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	tango "github.com/makegov/tango-go"
)

// newLiveClient returns a Client configured for the live API. It calls
// t.Skip if TANGO_API_KEY is unset.
func newLiveClient(t *testing.T) *tango.Client {
	t.Helper()
	key := os.Getenv("TANGO_API_KEY")
	if key == "" {
		t.Skip("set TANGO_API_KEY for integration tests")
	}
	return tango.NewClient(
		tango.WithAPIKey(key),
		tango.WithRetries(1),
		tango.WithRetryBackoff(250*time.Millisecond),
		tango.WithTimeout(10*time.Second),
	)
}

// ctx returns a context with a sensible deadline for live calls.
func ctx(t *testing.T) context.Context {
	t.Helper()
	c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return c
}
