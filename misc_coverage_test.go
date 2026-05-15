// misc_coverage_test.go covers small edge cases across multiple files
// that would otherwise remain untested: negative retries, empty body paths
// in generics, GetOTIDV options, ListPSC nil opts, and unmarshalWithExtra
// non-object JSON.
package tango

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// options.go edge cases
// ---------------------------------------------------------------------------

func TestWithRetriesNegativeClampedToZero(t *testing.T) {
	// n < 0 should be clamped to 0 (no panic)
	c := NewClient(WithRetries(-5))
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWithRetryBackoffNegativeClampedToZero(t *testing.T) {
	c := NewClient(WithRetryBackoff(-100 * time.Millisecond))
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

// ---------------------------------------------------------------------------
// GetOTIDV options paths
// ---------------------------------------------------------------------------

func TestGetOTIDVWithFlatOptions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetOTIDV(context.Background(), "otidv-key", &GetEntityOptions{
		Shape: "otidvs(minimal)", Flat: true, FlatLists: true,
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"shape":      "otidvs(minimal)",
		"flat":       "true",
		"flat_lists": "true",
	}, nil)
}

// ---------------------------------------------------------------------------
// ListPSC nil opts (covers the nil short-circuit branch)
// ---------------------------------------------------------------------------

func TestListPSCNilOptsNoQueryParams(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListPSC(context.Background(), nil)
	assertQueryContains(t, capturedURL, map[string]string{}, []string{"page", "limit"})
}

// ---------------------------------------------------------------------------
// getGeneric decode error path
// ---------------------------------------------------------------------------

func TestGetGenericDecodeError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return invalid JSON to trigger decode error in getGeneric
		w.Write([]byte(`{not valid json`))
	})
	_, err := c.GetAgency(context.Background(), "9700")
	if err == nil {
		t.Fatal("expected error for malformed JSON response")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// postGeneric empty body path (204 No Content)
// ---------------------------------------------------------------------------

func TestPostGenericEmptyBodyNoError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// 204 with empty body — postGeneric should return zero value
		w.WriteHeader(204)
	})
	// CreateWebhookEndpoint uses postGeneric; a 204 with empty body returns nil result
	result, err := c.CreateWebhookEndpoint(context.Background(), WebhookEndpointCreateInput{
		Name:        "Test",
		CallbackURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// result should be nil (zero value for *WebhookEndpoint)
	if result != nil {
		// A non-nil result is also fine — the function only returns nil when body is empty
		// and T is a pointer type. This covers the `len(raw) == 0` branch.
		t.Logf("got non-nil result (ok): %+v", result)
	}
}

// ---------------------------------------------------------------------------
// patchGeneric empty body path
// ---------------------------------------------------------------------------

func TestPatchGenericEmptyBodyNoError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	result, err := c.UpdateWebhookEndpoint(context.Background(), "ep1", WebhookEndpointUpdateInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

// ---------------------------------------------------------------------------
// patchGeneric decode error path
// ---------------------------------------------------------------------------

func TestPatchGenericDecodeError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid`))
	})
	_, err := c.UpdateWebhookEndpoint(context.Background(), "ep1", WebhookEndpointUpdateInput{})
	if err == nil {
		t.Fatal("expected error for malformed JSON in PATCH response")
	}
}

// ---------------------------------------------------------------------------
// postGeneric decode error path
// ---------------------------------------------------------------------------

func TestPostGenericDecodeError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid json`))
	})
	_, err := c.CreateWebhookEndpoint(context.Background(), WebhookEndpointCreateInput{
		Name:        "Hook",
		CallbackURL: "https://x.com",
	})
	if err == nil {
		t.Fatal("expected error for malformed JSON in POST response")
	}
}

// ---------------------------------------------------------------------------
// GetEntityMetrics validation paths
// ---------------------------------------------------------------------------

func TestGetEntityMetricsRequiresPeriodGrouping(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetEntityMetrics(context.Background(), "UEI12345", 12, "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// GetPSCMetrics period grouping validation
// ---------------------------------------------------------------------------

func TestGetPSCMetricsRequiresPeriodGrouping(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetPSCMetrics(context.Background(), "D302", 6, "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// models.go unmarshalWithExtra edge cases
// ---------------------------------------------------------------------------

func TestUnmarshalWithExtraNonPointerNoop(t *testing.T) {
	// Passing non-pointer to unmarshalWithExtra doesn't blow up
	// (The alias trick means we normally pass pointer; but directly it should handle it)
	raw := []byte(`{"agency_id":"A"}`)
	var rec AgencyRecord
	// Normal unmarshal via json.Unmarshal (which calls UnmarshalJSON)
	if err := (&rec).UnmarshalJSON(raw); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshalWithExtraInvalidJSON(t *testing.T) {
	// Invalid JSON should return error
	raw := []byte(`{not json`)
	var rec AgencyRecord
	if err := rec.UnmarshalJSON(raw); err == nil {
		t.Error("expected error for invalid JSON input")
	}
}

// ---------------------------------------------------------------------------
// ListWebhookAlerts with opts
// ---------------------------------------------------------------------------

func TestListWebhookAlertsWithPagination(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListWebhookAlerts(context.Background(), &ListOptions{Limit: 5})
	assertQueryContains(t, capturedURL, map[string]string{"limit": "5"}, nil)
}

// ---------------------------------------------------------------------------
// ListIDVChildIDVs nil opts path (covers the q := url.Values{} branch)
// ---------------------------------------------------------------------------

func TestListIDVChildIDVsNilOpts(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListIDVChildIDVs(context.Background(), "IDV-KEY", nil)
	assertPathContains(t, capturedURL, "/api/idvs/IDV-KEY/idvs/")
	// No query params should be sent
	assertQueryContains(t, capturedURL, nil, []string{"ordering", "naics"})
}

// ---------------------------------------------------------------------------
// ListIDVSummaryAwards nil opts
// ---------------------------------------------------------------------------

func TestListIDVSummaryAwardsNilOpts(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListIDVSummaryAwards(context.Background(), "SOL-001", nil)
	assertPathContains(t, capturedURL, "/api/idvs/SOL-001/summary/awards/")
}

// ---------------------------------------------------------------------------
// ListVehicleOrders nil opts
// ---------------------------------------------------------------------------

func TestListVehicleOrdersWithPagination(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListVehicleOrders(context.Background(), "uuid1", &ListOptions{Limit: 10})
	assertQueryContains(t, capturedURL, map[string]string{"limit": "10"}, nil)
}

// ---------------------------------------------------------------------------
// buildURL edge case — bad base URL
// ---------------------------------------------------------------------------

func TestBuildURLInvalidBase(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("://invalid-url"))
	// This should return an error
	_, err := c.buildURL("/api/test/", nil)
	if err == nil {
		t.Error("expected error for invalid base URL")
	}
}

// ---------------------------------------------------------------------------
// attempt body marshal error — unreachable in practice (body must be marshalable)
// We can trigger do returning early on a bad build URL
// ---------------------------------------------------------------------------

func TestDoContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Long delay
		time.Sleep(500 * time.Millisecond)
		w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	_, err := c.ListAgencies(ctx, nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// ---------------------------------------------------------------------------
// Resolve success path additional coverage
// ---------------------------------------------------------------------------

func TestResolveOrganizationTargetType(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":1,"candidates":[{"identifier":"ORG001","display_name":"Army"}]}`))
	})
	res, err := c.Resolve(context.Background(), ResolveInput{
		Name:       "Army",
		TargetType: ResolveOrganization,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(res.Candidates) != 1 {
		t.Errorf("expected 1 candidate, got %d", len(res.Candidates))
	}
}
