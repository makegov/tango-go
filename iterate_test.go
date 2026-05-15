package tango

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// This file exercises the Iterate* methods' closure logic (the fetch function
// that gets called internally by the iterator). Since these are all
// structurally identical thin wrappers, one representative test per resource
// is sufficient to cover the closure path.

func makeSinglePageServer(t *testing.T) (*httptest.Server, *int) {
	t.Helper()
	calls := new(int)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":1,"next":null,"results":[{"id":"x"}]}`))
	}))
	t.Cleanup(srv.Close)
	return srv, calls
}

func TestIterateEntitiesRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateEntities(context.Background(), &ListEntitiesOptions{UEI: "TEST"})
	for it.Next() {
		_ = it.Item()
	}
	if err := it.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateIDVsRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateIDVs(context.Background(), &ListIDVsOptions{AwardingAgency: "9700"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateVehiclesRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateVehicles(context.Background(), &ListVehiclesOptions{Search: "oasis"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateOTAsRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateOTAs(context.Background(), &ListOTAsOptions{AwardingAgency: "9700"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateOTIDVsRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateOTIDVs(context.Background(), &ListOTIDVsOptions{FiscalYear: "2024"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateOTIDVAwardsRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateOTIDVAwards(context.Background(), "key1", &ListOTIDVAwardsOptions{FiscalYear: "2024"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateOpportunitiesRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateOpportunities(context.Background(), &ListOpportunitiesOptions{Agency: "9700"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateNoticesRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateNotices(context.Background(), &ListNoticesOptions{Agency: "9700"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateForecastsRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateForecasts(context.Background(), &ListForecastsOptions{Agency: "9700"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateGrantsRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateGrants(context.Background(), &ListGrantsOptions{Agency: "9700"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateProtestsRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateProtests(context.Background(), &ListProtestsOptions{Agency: "9700"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateItDashboardRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateItDashboard(context.Background(), &ListItDashboardOptions{AgencyCode: "9700"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateGsaElibraryContractsRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateGsaElibraryContracts(context.Background(), &ListGsaElibraryContractsOptions{Schedule: "MAS"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

func TestIterateLcatsRunsClosure(t *testing.T) {
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	it := c.IterateLcats(context.Background(), &ListLcatsOptions{UEI: "UEI12345"})
	for it.Next() {
		_ = it.Item()
	}
	if *calls == 0 {
		t.Error("expected at least 1 server call")
	}
}

// TestIteratorSeqErrorPath tests that Seq yields an error when the fetch fails.
func TestIteratorSeqErrorPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	var gotErr error
	for _, err := range c.IterateContracts(context.Background(), nil).Seq() {
		if err != nil {
			gotErr = err
		}
	}
	if gotErr == nil {
		t.Error("expected error from Seq() when server returns 500")
	}
}
