package tango

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestListContractsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListContractsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts produces empty query",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"award_date", "awarding_agency", "search"},
		},
		{
			name:    "award dates",
			opts:    &ListContractsOptions{AwardDate: "2024-01-01", AwardDateGte: "2024-01-01", AwardDateLte: "2024-12-31"},
			wantQS:  map[string]string{"award_date": "2024-01-01", "award_date_gte": "2024-01-01", "award_date_lte": "2024-12-31"},
			notInQS: nil,
		},
		{
			name: "fiscal year bounds",
			opts: &ListContractsOptions{FiscalYear: "2024", FiscalYearGte: "2023", FiscalYearLte: "2025"},
			wantQS: map[string]string{
				"fiscal_year":     "2024",
				"fiscal_year_gte": "2023",
				"fiscal_year_lte": "2025",
			},
		},
		{
			name: "obligated bounds",
			opts: &ListContractsOptions{ObligatedGte: "1000", ObligatedLte: "999999"},
			wantQS: map[string]string{
				"obligated_gte": "1000",
				"obligated_lte": "999999",
			},
		},
		{
			name: "pop dates",
			opts: &ListContractsOptions{
				PopStartDateGte: "2024-01-01",
				PopStartDateLte: "2024-06-30",
				PopEndDateGte:   "2025-01-01",
				PopEndDateLte:   "2025-12-31",
			},
			wantQS: map[string]string{
				"pop_start_date_gte": "2024-01-01",
				"pop_start_date_lte": "2024-06-30",
				"pop_end_date_gte":   "2025-01-01",
				"pop_end_date_lte":   "2025-12-31",
			},
		},
		{
			name: "expiring bounds",
			opts: &ListContractsOptions{ExpiringGte: "2024-01-01", ExpiringLte: "2024-12-31"},
			wantQS: map[string]string{
				"expiring_gte": "2024-01-01",
				"expiring_lte": "2024-12-31",
			},
		},
		{
			name: "agencies and identifiers",
			opts: &ListContractsOptions{
				AwardingAgency:         "9700",
				FundingAgency:          "9800",
				PIID:                   "W15P7T19C0001",
				SolicitationIdentifier: "SOL001",
			},
			wantQS: map[string]string{
				"awarding_agency":         "9700",
				"funding_agency":          "9800",
				"piid":                    "W15P7T19C0001",
				"solicitation_identifier": "SOL001",
			},
		},
		{
			name: "naics sdk alias wins over raw",
			opts: &ListContractsOptions{NAICS: "541330", NAICSCode: "541512"},
			wantQS: map[string]string{
				"naics": "541512", // alias wins
			},
		},
		{
			name: "naics raw when alias empty",
			opts: &ListContractsOptions{NAICS: "541330"},
			wantQS: map[string]string{
				"naics": "541330",
			},
		},
		{
			name:   "psc alias wins",
			opts:   &ListContractsOptions{PSC: "R410", PSCCode: "D302"},
			wantQS: map[string]string{"psc": "D302"},
		},
		{
			name:   "recipient alias wins",
			opts:   &ListContractsOptions{Recipient: "OldCo", RecipientName: "NewCo"},
			wantQS: map[string]string{"recipient": "NewCo"},
		},
		{
			name:   "uei alias wins",
			opts:   &ListContractsOptions{UEI: "OLD123", RecipientUEI: "NEW456"},
			wantQS: map[string]string{"uei": "NEW456"},
		},
		{
			name:   "set_aside alias wins",
			opts:   &ListContractsOptions{SetAside: "8A", SetAsideType: "HZC"},
			wantQS: map[string]string{"set_aside": "HZC"},
		},
		{
			name:   "keyword alias wins over search",
			opts:   &ListContractsOptions{Search: "test", Keyword: "better"},
			wantQS: map[string]string{"search": "better"},
		},
		{
			name:   "sort asc",
			opts:   &ListContractsOptions{Sort: "award_date", Order: "asc"},
			wantQS: map[string]string{"ordering": "award_date"},
		},
		{
			name:   "sort desc",
			opts:   &ListContractsOptions{Sort: "award_date", Order: "desc"},
			wantQS: map[string]string{"ordering": "-award_date"},
		},
		{
			name:   "ordering direct (no sort)",
			opts:   &ListContractsOptions{Ordering: "piid"},
			wantQS: map[string]string{"ordering": "piid"},
		},
		{
			name:   "extra map string",
			opts:   &ListContractsOptions{Extra: map[string]any{"custom_filter": "val"}},
			wantQS: map[string]string{"custom_filter": "val"},
		},
		{
			name:   "extra map int",
			opts:   &ListContractsOptions{Extra: map[string]any{"year": 2024}},
			wantQS: map[string]string{"year": "2024"},
		},
		{
			name:   "extra map bool",
			opts:   &ListContractsOptions{Extra: map[string]any{"active": true}},
			wantQS: map[string]string{"active": "true"},
		},
		{
			name:    "zero values omitted",
			opts:    &ListContractsOptions{},
			notInQS: []string{"award_date", "awarding_agency", "naics", "psc", "uei", "ordering"},
		},
		{
			name: "list options applied",
			opts: &ListContractsOptions{
				ListOptions: ListOptions{Limit: 10, Shape: ShapeContractsMinimal},
			},
			wantQS: map[string]string{"limit": "10", "shape": ShapeContractsMinimal},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))

			_, _ = c.ListContracts(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestListContractsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(emptyListHandler))
	resp, err := c.ListContracts(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIterateContractsBuildsIterator(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(emptyListHandler))
	it := c.IterateContracts(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestValueToString(t *testing.T) {
	cases := []struct {
		v    any
		want string
	}{
		{"hello", "hello"},
		{42, "42"},
		{int64(100), "100"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
		{struct{}{}, ""}, // unsupported falls back to ""
	}
	for _, tc := range cases {
		got := valueToString(tc.v)
		if got != tc.want {
			t.Errorf("valueToString(%v): want %q, got %q", tc.v, tc.want, got)
		}
	}
}

func TestListContractsExtraEmptyStringOmitted(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListContracts(context.Background(), &ListContractsOptions{
		Extra: map[string]any{"empty_str": ""},
	})
	assertQueryContains(t, capturedURL, nil, []string{"empty_str"})
}

func TestListContractsExtraFloat64(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListContracts(context.Background(), &ListContractsOptions{
		Extra: map[string]any{"amount": float64(1234.5)},
	})
	assertQueryContains(t, capturedURL, map[string]string{"amount": "1234.5"}, nil)
}

func TestListContractsAwardType(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListContracts(context.Background(), &ListContractsOptions{
		AwardType: "A",
	})
	assertQueryContains(t, capturedURL, map[string]string{"award_type": "A"}, nil)
}

func TestGetContractRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetContract(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetContractBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetContract(context.Background(), "KEY-1", nil)
	assertPathContains(t, capturedURL, "/api/contracts/KEY-1/")
}

func TestListContractSubawardsRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListContractSubawards(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListContractSubawardsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListContractSubawards(context.Background(), "KEY-1", nil)
	assertPathContains(t, capturedURL, "/api/contracts/KEY-1/subawards/")
}

func TestListContractTransactionsRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListContractTransactions(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListContractTransactionsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListContractTransactions(context.Background(), "KEY-1", nil)
	assertPathContains(t, capturedURL, "/api/contracts/KEY-1/transactions/")
}

func TestListContractsServerError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	_, err := c.ListContracts(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
}
