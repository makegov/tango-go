package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListIDVsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListIDVsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"award_date", "awarding_agency", "naics"},
		},
		{
			name: "all filters",
			opts: &ListIDVsOptions{
				AwardDate:              "2024-01-01",
				AwardDateGte:           "2024-01-01",
				AwardDateLte:           "2024-12-31",
				AwardingAgency:         "9700",
				FundingAgency:          "9800",
				ExpiringGte:            "2025-01-01",
				ExpiringLte:            "2026-12-31",
				FiscalYear:             "2024",
				FiscalYearGte:          "2023",
				FiscalYearLte:          "2025",
				IDVType:                "A",
				LastDateToOrderGte:     "2026-01-01",
				LastDateToOrderLte:     "2027-01-01",
				NAICS:                  "541512",
				Ordering:               "-award_date",
				PIID:                   "W15P7T19D0001",
				PopStartDateGte:        "2024-01-01",
				PopStartDateLte:        "2024-12-31",
				PSC:                    "D302",
				Recipient:              "Acme",
				Search:                 "keyword",
				SetAside:               "8A",
				SolicitationIdentifier: "SOL001",
				UEI:                    "UEI12345",
			},
			wantQS: map[string]string{
				"award_date":              "2024-01-01",
				"award_date_gte":          "2024-01-01",
				"award_date_lte":          "2024-12-31",
				"awarding_agency":         "9700",
				"funding_agency":          "9800",
				"expiring_gte":            "2025-01-01",
				"expiring_lte":            "2026-12-31",
				"fiscal_year":             "2024",
				"fiscal_year_gte":         "2023",
				"fiscal_year_lte":         "2025",
				"idv_type":                "A",
				"last_date_to_order_gte":  "2026-01-01",
				"last_date_to_order_lte":  "2027-01-01",
				"naics":                   "541512",
				"ordering":                "-award_date",
				"piid":                    "W15P7T19D0001",
				"pop_start_date_gte":      "2024-01-01",
				"pop_start_date_lte":      "2024-12-31",
				"psc":                     "D302",
				"recipient":               "Acme",
				"search":                  "keyword",
				"set_aside":               "8A",
				"solicitation_identifier": "SOL001",
				"uei":                     "UEI12345",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListIDVsOptions{},
			notInQS: []string{"award_date", "awarding_agency", "naics", "ordering"},
		},
		{
			name:   "extra map",
			opts:   &ListIDVsOptions{Extra: map[string]any{"custom_x": "xval"}},
			wantQS: map[string]string{"custom_x": "xval"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListIDVs(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetIDVRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetIDV(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetIDVBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetIDV(context.Background(), "PIID-XYZ", nil)
	assertPathContains(t, capturedURL, "/api/idvs/PIID-XYZ/")
}

func TestGetIDVWithOptions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetIDV(context.Background(), "KEY1", &GetEntityOptions{
		Shape: "idvs(minimal)", Flat: true, FlatLists: true,
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"shape":      "idvs(minimal)",
		"flat":       "true",
		"flat_lists": "true",
	}, nil)
}

func TestIterateIDVsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateIDVs(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}
