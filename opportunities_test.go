package tango

import (
	"context"
	"testing"
)

func TestListOpportunitiesFilterMapping(t *testing.T) {
	trueBool := true
	cases := []struct {
		name    string
		opts    *ListOpportunitiesOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"active", "agency", "search"},
		},
		{
			name: "all filters",
			opts: &ListOpportunitiesOptions{
				Active:                 &trueBool,
				Agency:                 "9700",
				FirstNoticeDateAfter:   "2024-01-01",
				FirstNoticeDateBefore:  "2024-06-30",
				LastNoticeDateAfter:    "2024-07-01",
				LastNoticeDateBefore:   "2024-12-31",
				NAICS:                  "541512",
				NoticeType:             "PRESOL",
				Ordering:               "-response_deadline",
				PlaceOfPerformance:     "VA",
				PSC:                    "D302",
				ResponseDeadlineAfter:  "2024-01-15",
				ResponseDeadlineBefore: "2024-03-31",
				Search:                 "cloud computing",
				SetAside:               "8A",
				SolicitationNumber:     "W15P7T24R0001",
			},
			wantQS: map[string]string{
				"active":                   "true",
				"agency":                   "9700",
				"first_notice_date_after":  "2024-01-01",
				"first_notice_date_before": "2024-06-30",
				"last_notice_date_after":   "2024-07-01",
				"last_notice_date_before":  "2024-12-31",
				"naics":                    "541512",
				"notice_type":              "PRESOL",
				"ordering":                 "-response_deadline",
				"place_of_performance":     "VA",
				"psc":                      "D302",
				"response_deadline_after":  "2024-01-15",
				"response_deadline_before": "2024-03-31",
				"search":                   "cloud computing",
				"set_aside":                "8A",
				"solicitation_number":      "W15P7T24R0001",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListOpportunitiesOptions{},
			notInQS: []string{"active", "agency", "naics", "ordering"},
		},
		{
			name:   "extra map",
			opts:   &ListOpportunitiesOptions{Extra: map[string]any{"x": "y"}},
			wantQS: map[string]string{"x": "y"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListOpportunities(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestListNoticesFilterMapping(t *testing.T) {
	trueBool := true
	cases := []struct {
		name    string
		opts    *ListNoticesOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"active", "agency"},
		},
		{
			name: "all filters",
			opts: &ListNoticesOptions{
				Active:                 &trueBool,
				Agency:                 "9700",
				NAICS:                  "541512",
				NoticeType:             "AWARD",
				PostedDateAfter:        "2024-01-01",
				PostedDateBefore:       "2024-12-31",
				PSC:                    "D302",
				ResponseDeadlineAfter:  "2024-02-01",
				ResponseDeadlineBefore: "2024-11-30",
				Search:                 "cybersecurity",
				SetAside:               "HZC",
				SolicitationNumber:     "PROJ-2024",
			},
			wantQS: map[string]string{
				"active":                   "true",
				"agency":                   "9700",
				"naics":                    "541512",
				"notice_type":              "AWARD",
				"posted_date_after":        "2024-01-01",
				"posted_date_before":       "2024-12-31",
				"psc":                      "D302",
				"response_deadline_after":  "2024-02-01",
				"response_deadline_before": "2024-11-30",
				"search":                   "cybersecurity",
				"set_aside":                "HZC",
				"solicitation_number":      "PROJ-2024",
			},
		},
		{
			name:   "extra map",
			opts:   &ListNoticesOptions{Extra: map[string]any{"tag": "v2"}},
			wantQS: map[string]string{"tag": "v2"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListNotices(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestListForecastsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListForecastsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"agency", "fiscal_year"},
		},
		{
			name: "all filters",
			opts: &ListForecastsOptions{
				Agency:          "9700",
				AwardDateAfter:  "2024-01-01",
				AwardDateBefore: "2024-12-31",
				FiscalYear:      "2024",
				FiscalYearGte:   "2023",
				FiscalYearLte:   "2025",
				ModifiedAfter:   "2024-06-01",
				ModifiedBefore:  "2024-06-30",
				NAICSCode:       "541512",
				NAICSStartsWith: "5415",
				Ordering:        "award_date",
				Search:          "AI",
				SourceSystem:    "SAM",
				Status:          "active",
			},
			wantQS: map[string]string{
				"agency":            "9700",
				"award_date_after":  "2024-01-01",
				"award_date_before": "2024-12-31",
				"fiscal_year":       "2024",
				"fiscal_year_gte":   "2023",
				"fiscal_year_lte":   "2025",
				"modified_after":    "2024-06-01",
				"modified_before":   "2024-06-30",
				"naics_code":        "541512",
				"naics_starts_with": "5415",
				"ordering":          "award_date",
				"search":            "AI",
				"source_system":     "SAM",
				"status":            "active",
			},
		},
		{
			name:   "extra",
			opts:   &ListForecastsOptions{Extra: map[string]any{"x": "1"}},
			wantQS: map[string]string{"x": "1"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListForecasts(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestListGrantsFilterMapping(t *testing.T) {
	cases := []struct {
		name   string
		opts   *ListGrantsOptions
		wantQS map[string]string
	}{
		{
			name: "all filters",
			opts: &ListGrantsOptions{
				Agency:             "9700",
				ApplicantTypes:     "11",
				CFDANumber:         "10.001",
				FundingCategories:  "AR",
				FundingInstruments: "G",
				OpportunityNumber:  "OPP-001",
				Ordering:           "response_date",
				PostedDateAfter:    "2024-01-01",
				PostedDateBefore:   "2024-12-31",
				ResponseDateAfter:  "2024-02-01",
				ResponseDateBefore: "2024-11-30",
				Search:             "environment",
				Status:             "posted",
			},
			wantQS: map[string]string{
				"agency":               "9700",
				"applicant_types":      "11",
				"cfda_number":          "10.001",
				"funding_categories":   "AR",
				"funding_instruments":  "G",
				"opportunity_number":   "OPP-001",
				"ordering":             "response_date",
				"posted_date_after":    "2024-01-01",
				"posted_date_before":   "2024-12-31",
				"response_date_after":  "2024-02-01",
				"response_date_before": "2024-11-30",
				"search":               "environment",
				"status":               "posted",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListGrants(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, nil)
		})
	}
}

func TestIterateOpportunitiesNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateOpportunities(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestIterateNoticesNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateNotices(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestIterateForecastsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateForecasts(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestIterateGrantsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateGrants(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}
