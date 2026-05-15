package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListOTAsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListOTAsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"awarding_agency", "piid", "search"},
		},
		{
			name: "all filters",
			opts: &ListOTAsOptions{
				Joiner:          ".",
				AwardingAgency:  "9700",
				FundingAgency:   "9800",
				PIID:            "W15P7T19C0001",
				Recipient:       "Acme",
				UEI:             "UEI12345",
				FiscalYear:      "2024",
				FiscalYearGte:   "2023",
				FiscalYearLte:   "2025",
				AwardDate:       "2024-01-01",
				AwardDateGte:    "2024-01-01",
				AwardDateLte:    "2024-12-31",
				ExpiringGte:     "2025-01-01",
				ExpiringLte:     "2026-12-31",
				PopStartDateGte: "2024-01-01",
				PopStartDateLte: "2024-12-31",
				PopEndDateGte:   "2025-01-01",
				PopEndDateLte:   "2025-12-31",
				PSC:             "D302",
				Search:          "AI research",
				Ordering:        "-award_date",
			},
			wantQS: map[string]string{
				"joiner":              ".",
				"awarding_agency":     "9700",
				"funding_agency":      "9800",
				"piid":                "W15P7T19C0001",
				"recipient":           "Acme",
				"uei":                 "UEI12345",
				"fiscal_year":         "2024",
				"fiscal_year_gte":     "2023",
				"fiscal_year_lte":     "2025",
				"award_date":          "2024-01-01",
				"award_date_gte":      "2024-01-01",
				"award_date_lte":      "2024-12-31",
				"expiring_gte":        "2025-01-01",
				"expiring_lte":        "2026-12-31",
				"pop_start_date_gte":  "2024-01-01",
				"pop_start_date_lte":  "2024-12-31",
				"pop_end_date_gte":    "2025-01-01",
				"pop_end_date_lte":    "2025-12-31",
				"psc":                 "D302",
				"search":              "AI research",
				"ordering":            "-award_date",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListOTAsOptions{},
			notInQS: []string{"joiner", "awarding_agency", "piid"},
		},
		{
			name: "extra map",
			opts: &ListOTAsOptions{Extra: map[string]any{"tag": "ota"}},
			wantQS: map[string]string{"tag": "ota"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListOTAs(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetOTARequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetOTA(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetOTABuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetOTA(context.Background(), "ota-key-1234", nil)
	assertPathContains(t, capturedURL, "/api/otas/ota-key-1234/")
}

func TestGetOTAWithOptions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetOTA(context.Background(), "key1", &GetEntityOptions{
		Shape: "otas(minimal)", Flat: true, FlatLists: true,
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"shape":      "otas(minimal)",
		"flat":       "true",
		"flat_lists": "true",
	}, nil)
}

func TestIterateOTAsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateOTAs(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestListOTIDVsFilterMapping(t *testing.T) {
	cases := []struct {
		name   string
		opts   *ListOTIDVsOptions
		wantQS map[string]string
	}{
		{
			name: "key filters",
			opts: &ListOTIDVsOptions{
				AwardingAgency: "9700",
				FiscalYear:     "2024",
				PSC:            "R410",
			},
			wantQS: map[string]string{
				"awarding_agency": "9700",
				"fiscal_year":     "2024",
				"psc":             "R410",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListOTIDVs(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, nil)
		})
	}
}

func TestGetOTIDVRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetOTIDV(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetOTIDVBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetOTIDV(context.Background(), "otidv-key-5678", nil)
	assertPathContains(t, capturedURL, "/api/otidvs/otidv-key-5678/")
}

func TestListOTIDVAwardsRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListOTIDVAwards(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListOTIDVAwardsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListOTIDVAwards(context.Background(), "otidv-key", nil)
	assertPathContains(t, capturedURL, "/api/otidvs/otidv-key/awards/")
}

func TestIterateOTIDVAwardsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateOTIDVAwards(context.Background(), "key1", nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}
