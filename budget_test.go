package tango

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestListBudgetAccountsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListBudgetAccountsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts produces empty query",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"federal_account_symbol", "fiscal_year", "search"},
		},
		{
			name: "all typed filters",
			opts: &ListBudgetAccountsOptions{
				FederalAccountSymbol: "097-0100",
				FiscalYear:           "2024",
				FiscalYearGte:        "2020",
				FiscalYearLte:        "2025",
				AgencyCode:           "9700",
				BEACategory:          "discretionary",
				OnOffBudget:          "on",
				Search:               "operations",
				Ordering:             "-enacted_ba",
			},
			wantQS: map[string]string{
				"federal_account_symbol": "097-0100",
				"fiscal_year":            "2024",
				"fiscal_year__gte":       "2020",
				"fiscal_year__lte":       "2025",
				"agency_code":            "9700",
				"bea_category":           "discretionary",
				"on_off_budget":          "on",
				"search":                 "operations",
				"ordering":               "-enacted_ba",
			},
		},
		{
			name: "pagination and shape",
			opts: &ListBudgetAccountsOptions{
				ListOptions: ListOptions{Page: 2, Limit: 50, Shape: ShapeBudgetAccountsMinimal},
			},
			wantQS: map[string]string{
				"page":  "2",
				"limit": "50",
				"shape": ShapeBudgetAccountsMinimal,
			},
		},
		{
			name:   "extra forwards range filters",
			opts:   &ListBudgetAccountsOptions{Extra: map[string]any{"enacted_ba__gte": "1000000"}},
			wantQS: map[string]string{"enacted_ba__gte": "1000000"},
		},
		{
			name:    "zero values omitted",
			opts:    &ListBudgetAccountsOptions{},
			notInQS: []string{"federal_account_symbol", "fiscal_year", "agency_code", "ordering"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListBudgetAccounts(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestListBudgetAccountsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListBudgetAccounts(context.Background(), nil)
	assertPathContains(t, capturedURL, "/api/budget/accounts/")
}

func TestIterateBudgetAccountsBuildsIterator(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(emptyListHandler))
	it := c.IterateBudgetAccounts(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestGetBudgetAccountRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetBudgetAccount(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetBudgetAccountBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetBudgetAccount(context.Background(), "097-0100-2024", nil)
	assertPathContains(t, capturedURL, "/api/budget/accounts/097-0100-2024/")
}

func TestGetBudgetAccountQuartersRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetBudgetAccountQuarters(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetBudgetAccountQuartersBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.GetBudgetAccountQuarters(context.Background(), "acct-1", nil)
	assertPathContains(t, capturedURL, "/api/budget/accounts/acct-1/quarters/")
}

func TestGetBudgetAccountRecipientsRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetBudgetAccountRecipients(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetBudgetAccountRecipientsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.GetBudgetAccountRecipients(context.Background(), "acct-1", nil)
	assertPathContains(t, capturedURL, "/api/budget/accounts/acct-1/recipients/")
}
