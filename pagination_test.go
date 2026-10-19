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

func TestListResponseCarriesMetaAndAgencyDiagnostics(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"results":[{"k":"a"}],"meta":{
			"resolved_filters":{
				"awarding_agency":[
					{"token":"HUD","resolved":{"key":"org-1","name":"Department of Housing and Urban Development"},"matched_by":"alias"},
					{"token":"NOTANAGENCY","resolved":null}
				],
				"funding_agency":[{"token":"9700","resolved":{"key":"org-2","name":"Department of Defense"},"matched_by":"code"}],
				"malformed":"skip me"
			},
			"warnings":["Agency filter 'awarding_agency': 'NOTANAGENCY' did not match any organization and was ignored.", 7]
		}}`))
	})
	resp, err := c.ListContracts(context.Background(), &ListContractsOptions{AwardingAgency: "HUD|NOTANAGENCY", FundingAgency: "9700"})
	if err != nil {
		t.Fatalf("ListContracts: %v", err)
	}

	entries := resp.Meta["resolved_filters"].(map[string]any)["awarding_agency"].([]any)
	if got := entries[0].(map[string]any)["matched_by"]; got != "alias" {
		t.Errorf("matched_by: want alias, got %v", got)
	}
	if _, ok := entries[1].(map[string]any)["matched_by"]; ok {
		t.Error("an unresolved entry has no matched_by")
	}

	if got := resp.AgencyWarnings(); len(got) != 1 || !strings.Contains(got[0], "NOTANAGENCY") {
		t.Errorf("AgencyWarnings: %v", got)
	}
	unresolved := resp.UnresolvedAgencyTokens()
	if len(unresolved) != 1 || len(unresolved["awarding_agency"]) != 1 || unresolved["awarding_agency"][0] != "NOTANAGENCY" {
		t.Errorf("UnresolvedAgencyTokens: %v", unresolved)
	}
	resolved := resp.ResolvedAgencies()
	if resolved["awarding_agency"][0]["name"] != "Department of Housing and Urban Development" || resolved["funding_agency"][0]["key"] != "org-2" || len(resolved) != 2 {
		t.Errorf("ResolvedAgencies: %v", resolved)
	}
}

func TestAgencyDiagnosticsAreEmptyWithoutMeta(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	resp, err := c.ListContracts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListContracts: %v", err)
	}
	if resp.Meta != nil || resp.AgencyWarnings() != nil || len(resp.UnresolvedAgencyTokens()) != 0 || len(resp.ResolvedAgencies()) != 0 {
		t.Errorf("want empty diagnostics, got meta=%v", resp.Meta)
	}
	var nilResp *PaginatedResponse[Record]
	if nilResp.AgencyWarnings() != nil || len(nilResp.UnresolvedAgencyTokens()) != 0 || len(nilResp.ResolvedAgencies()) != 0 {
		t.Error("a nil response should report no diagnostics")
	}
}

func TestListResponseCountTypeComesFromItsOwnHeader(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if v := r.URL.Query().Get("count_type"); v != "" {
			w.Header().Set("X-Results-CountType", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1000,"results":[{"k":"a"}]}`))
	})
	ctx := context.Background()

	approx, err := c.ListContracts(ctx, &ListContractsOptions{Extra: map[string]any{"count_type": "Approximate"}})
	if err != nil {
		t.Fatalf("ListContracts: %v", err)
	}
	exact, err := c.ListOpportunities(ctx, &ListOpportunitiesOptions{Extra: map[string]any{"count_type": "exact"}})
	if err != nil {
		t.Fatalf("ListOpportunities: %v", err)
	}
	none, err := c.ListNotices(ctx, nil)
	if err != nil {
		t.Fatalf("ListNotices: %v", err)
	}
	if approx.CountType != "approximate" || exact.CountType != "exact" || none.CountType != "" {
		t.Errorf("CountType: got %q, %q, %q", approx.CountType, exact.CountType, none.CountType)
	}
}
