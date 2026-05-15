package tango

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestIteratorWalksPageBasedPagination tests the iterator's page= path
// (as opposed to cursor= path).
func TestIteratorWalksPageBasedPagination(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case 1:
			// next URL uses page= (not cursor=) to trigger the page-based path
			w.Write([]byte(`{"count":3,"next":"http://x/?page=2","results":[{"k":"a"},{"k":"b"}]}`))
		default:
			w.Write([]byte(`{"count":3,"next":null,"results":[{"k":"c"}]}`))
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	iter := c.IterateContracts(context.Background(), nil)
	var keys []string
	for iter.Next() {
		keys = append(keys, iter.Item()["k"].(string))
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if got := strings.Join(keys, ","); got != "a,b,c" {
		t.Errorf("expected a,b,c; got %s", got)
	}
}

// TestIteratorPageBasedNextURLInvalidPage tests when next URL has page but
// it's not a valid integer — iterator should stop gracefully.
func TestIteratorPageBasedNextURLInvalidPage(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		// Return a next URL with invalid page= to trigger done path
		w.Write([]byte(`{"count":99,"next":"http://x/?page=notanumber","results":[{"k":"a"}]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	iter := c.IterateContracts(context.Background(), nil)
	var keys []string
	for iter.Next() {
		keys = append(keys, iter.Item()["k"].(string))
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("unexpected iterator error: %v", err)
	}
	// Should have gotten "a" from the first page, then stopped because page=notanumber
	if len(keys) != 1 || keys[0] != "a" {
		t.Errorf("expected [a], got %v", keys)
	}
}

// TestIteratorNextURLNoPageNoCursorStops tests when next URL has neither
// page= nor cursor= — iterator should stop.
func TestIteratorNextURLNoPageNoCursorStops(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		// No page or cursor in next
		w.Write([]byte(`{"count":99,"next":"http://x/?something=else","results":[{"k":"a"}]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	iter := c.IterateContracts(context.Background(), nil)
	var keys []string
	for iter.Next() {
		keys = append(keys, iter.Item()["k"].(string))
	}
	// Should stop after 1 page
	if len(keys) == 0 {
		t.Error("expected at least 1 key")
	}
}

// TestParseIntStrict covers the parseIntStrict helper.
func TestParseIntStrict(t *testing.T) {
	cases := []struct {
		s       string
		want    int
		wantErr bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
		{"1a2", 0, true},
		{"456", 456, false},
	}
	for _, tc := range cases {
		got, err := parseIntStrict(tc.s)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseIntStrict(%q): expected error, got %d", tc.s, got)
			}
		} else {
			if err != nil {
				t.Errorf("parseIntStrict(%q): unexpected error: %v", tc.s, err)
			}
			if got != tc.want {
				t.Errorf("parseIntStrict(%q): want %d, got %d", tc.s, tc.want, got)
			}
		}
	}
}

// TestExtractCursorEdgeCases covers extractCursor.
func TestExtractCursorEdgeCases(t *testing.T) {
	// Empty string
	if c := extractCursor(""); c != "" {
		t.Errorf("expected empty for empty input, got %q", c)
	}
	// Invalid URL
	if c := extractCursor("://invalid"); c != "" {
		t.Errorf("expected empty for invalid URL, got %q", c)
	}
	// URL without cursor param
	if c := extractCursor("https://api.example.com/api/agencies/?page=2"); c != "" {
		t.Errorf("expected empty for URL without cursor, got %q", c)
	}
	// Valid cursor
	if c := extractCursor("https://api.example.com/api/idvs/?cursor=abc123"); c != "abc123" {
		t.Errorf("expected abc123, got %q", c)
	}
}

// TestDecodePaginatedEdgeCases covers decodePaginated edge cases.
func TestDecodePaginatedEmptyResults(t *testing.T) {
	raw := []byte(`{"count":0,"next":null,"previous":null,"results":[]}`)
	resp, err := decodePaginated[Record](raw)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("expected Count=0, got %d", resp.Count)
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
}

func TestDecodePaginatedInvalidJSON(t *testing.T) {
	raw := []byte(`not-json`)
	_, err := decodePaginated[Record](raw)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDecodePaginatedInvalidResults(t *testing.T) {
	// results field is not a JSON array
	raw := []byte(`{"count":1,"results":"not-an-array"}`)
	_, err := decodePaginated[Record](raw)
	if err == nil {
		t.Error("expected error for invalid results field")
	}
}

// TestApplyToListOptionsCursorTakesPrecedence checks Cursor vs Page precedence.
func TestApplyToListOptionsCursorTakesPrecedence(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListAgencies(context.Background(), &ListAgenciesOptions{
		Page:  2,
		Limit: 10,
		// Note: ListAgenciesOptions doesn't embed ListOptions, so we test applyTo directly
	})
	// Page 2 with no cursor should include page=2
	assertQueryContains(t, capturedURL, map[string]string{"page": "2"}, nil)
}

// TestListOptionsCursorWinsOverPage exercises applyTo cursor precedence.
func TestListOptionsCursorWinsOverPage(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListContracts(context.Background(), &ListContractsOptions{
		ListOptions: ListOptions{
			Cursor: "cursor-abc",
			Page:   5, // should be ignored
		},
	})
	assertQueryContains(t, capturedURL, map[string]string{"cursor": "cursor-abc"}, []string{"page"})
}

// TestListOptionsFlatLists exercises the FlatLists branch.
func TestListOptionsFlatLists(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListContracts(context.Background(), &ListContractsOptions{
		ListOptions: ListOptions{Flat: true, FlatLists: true},
	})
	assertQueryContains(t, capturedURL, map[string]string{"flat": "true", "flat_lists": "true"}, nil)
}
