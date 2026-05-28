package tango

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestListAgenciesFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListAgenciesOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"page", "limit", "search"},
		},
		{
			name:   "page and limit",
			opts:   &ListAgenciesOptions{Page: 2, Limit: 10},
			wantQS: map[string]string{"page": "2", "limit": "10"},
		},
		{
			name:    "search",
			opts:    &ListAgenciesOptions{Search: "defense"},
			wantQS:  map[string]string{"search": "defense"},
			notInQS: []string{"page"},
		},
		{
			name:   "limit capped at 100",
			opts:   &ListAgenciesOptions{Limit: 9999},
			wantQS: map[string]string{"limit": "100"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListAgencies(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetAgencyRequiresCode(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetAgency(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetAgencyBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetAgency(context.Background(), "9700")
	assertPathContains(t, capturedURL, "/api/agencies/9700/")
}

func TestGetAgencyDecodesTypedRecord(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"agency_id":"DOD001","name":"Department of Defense","abbreviation":"DoD","code":"9700"}`))
	})
	rec, err := c.GetAgency(context.Background(), "9700")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if rec == nil {
		t.Fatal("expected non-nil AgencyRecord")
	}
	assertStrPtr(t, "code", rec.Code, "9700")
	assertStrPtr(t, "name", rec.Name, "Department of Defense")
}

func TestListOrganizationsFilterMapping(t *testing.T) {
	trueBool := true
	cases := []struct {
		name    string
		opts    *ListOrganizationsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"search", "type", "level"},
		},
		{
			name: "all filters",
			opts: &ListOrganizationsOptions{
				Search:          "army",
				Type:            "department",
				Level:           "1",
				CGAC:            "021",
				Parent:          "DOD",
				IncludeInactive: &trueBool,
			},
			wantQS: map[string]string{
				"search":           "army",
				"type":             "department",
				"level":            "1",
				"cgac":             "021",
				"parent":           "DOD",
				"include_inactive": "true",
			},
		},
		{
			name:    "nil bool omitted",
			opts:    &ListOrganizationsOptions{},
			notInQS: []string{"include_inactive"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListOrganizations(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetOrganizationRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetOrganization(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetOrganizationBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetOrganization(context.Background(), "DoD")
	assertPathContains(t, capturedURL, "/api/organizations/DoD/")
}

func TestListNAICSFilterMapping(t *testing.T) {
	cases := []struct {
		name   string
		opts   *ListNAICSOptions
		wantQS map[string]string
	}{
		{
			name: "all filters",
			opts: &ListNAICSOptions{
				Search:           "computer",
				RevenueLimit:     "38.5M",
				EmployeeLimit:    "1500",
				RevenueLimitGte:  "1M",
				RevenueLimitLte:  "100M",
				EmployeeLimitGte: "100",
				EmployeeLimitLte: "5000",
			},
			wantQS: map[string]string{
				"search":             "computer",
				"revenue_limit":      "38.5M",
				"employee_limit":     "1500",
				"revenue_limit_gte":  "1M",
				"revenue_limit_lte":  "100M",
				"employee_limit_gte": "100",
				"employee_limit_lte": "5000",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListNAICS(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, nil)
		})
	}
}

func TestGetNAICSRequiresCode(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetNAICS(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetNAICSBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetNAICS(context.Background(), "541512")
	assertPathContains(t, capturedURL, "/api/naics/541512/")
}

func TestGetPSCRequiresCode(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetPSC(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetPSCBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetPSC(context.Background(), "D302")
	assertPathContains(t, capturedURL, "/api/psc/D302/")
}

func TestListPSCNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	resp, err := c.ListPSC(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestListSubawardsFilterMapping(t *testing.T) {
	cases := []struct {
		name   string
		opts   *ListSubawardsOptions
		wantQS map[string]string
	}{
		{
			name: "all filters",
			opts: &ListSubawardsOptions{
				AwardKey:       "CONT-KEY",
				PrimeUEI:       "PRIME001",
				SubUEI:         "SUB001",
				AwardingAgency: "9700",
				FundingAgency:  "9800",
				FiscalYear:     "2024",
				FiscalYearGte:  "2022",
				FiscalYearLte:  "2025",
				Recipient:      "Acme Sub",
				Ordering:       "last_modified_date",
			},
			wantQS: map[string]string{
				"award_key":       "CONT-KEY",
				"prime_uei":       "PRIME001",
				"sub_uei":         "SUB001",
				"awarding_agency": "9700",
				"funding_agency":  "9800",
				"fiscal_year":     "2024",
				"fiscal_year_gte": "2022",
				"fiscal_year_lte": "2025",
				"recipient":       "Acme Sub",
				"ordering":        "last_modified_date",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListSubawards(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, nil)
		})
	}
}

func TestListPSCWithOpts(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListPSC(context.Background(), &ListOptions{Limit: 10})
	assertQueryContains(t, capturedURL, map[string]string{"limit": "10"}, nil)
}

func TestGetVersionBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetVersion(context.Background())
	assertPathContains(t, capturedURL, "/api/version/")
}

func TestGetSubawardRequiresKey(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetSubaward(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetSubawardBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetSubaward(context.Background(), "SUB-1", nil)
	assertPathContains(t, capturedURL, "/api/subawards/SUB-1/")
}
