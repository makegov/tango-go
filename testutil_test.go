package tango

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// emptyListHandler returns a minimal valid paginated response with zero results.
func emptyListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"count":0,"results":[]}`))
}

// assertQueryContains checks that gotURL contains all wantKV pairs and
// that none of notIn keys appear.
func assertQueryContains(t *testing.T, rawURL string, wantKV map[string]string, notIn []string) {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("assertQueryContains: could not parse URL %q: %v", rawURL, err)
	}
	q := u.Query()
	for k, want := range wantKV {
		got := q.Get(k)
		if got != want {
			t.Errorf("query param %q: want %q, got %q (full query: %s)", k, want, got, u.RawQuery)
		}
	}
	for _, k := range notIn {
		if v := q.Get(k); v != "" {
			t.Errorf("query param %q should not be present, but got %q", k, v)
		}
	}
}

// captureURL returns a handler that records the request URL and serves an
// empty list, then calls check with the captured URL.
func captureURLHandler(captured *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*captured = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":0,"results":[]}`))
	}
}

// captureURLRecordHandler records URL and serves a bare JSON object.
func captureURLRecordHandler(captured *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*captured = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"x"}`))
	}
}

// assertPathContains checks that the captured URL path contains substr.
func assertPathContains(t *testing.T, rawURI, substr string) {
	t.Helper()
	if !strings.Contains(rawURI, substr) {
		t.Errorf("expected URL %q to contain %q", rawURI, substr)
	}
}
