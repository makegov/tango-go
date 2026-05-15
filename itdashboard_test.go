package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListItDashboardFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListItDashboardOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"search", "agency_code", "cio_rating"},
		},
		{
			name: "all filters",
			opts: &ListItDashboardOptions{
				Search:            "cloud",
				AgencyCode:        "9700",
				AgencyName:        "Department of Defense",
				TypeOfInvestment:  "Major IT Investment",
				UpdatedTimeAfter:  "2024-01-01T00:00:00Z",
				UpdatedTimeBefore: "2024-12-31T23:59:59Z",
				CIORating:         "3",
				CIORatingMax:      "5",
				PerformanceRisk:   "medium",
			},
			wantQS: map[string]string{
				"search":              "cloud",
				"agency_code":         "9700",
				"agency_name":         "Department of Defense",
				"type_of_investment":  "Major IT Investment",
				"updated_time_after":  "2024-01-01T00:00:00Z",
				"updated_time_before": "2024-12-31T23:59:59Z",
				"cio_rating":          "3",
				"cio_rating_max":      "5",
				"performance_risk":    "medium",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListItDashboardOptions{},
			notInQS: []string{"search", "agency_code", "cio_rating"},
		},
		{
			name: "extra map",
			opts: &ListItDashboardOptions{Extra: map[string]any{"tier": "business"}},
			wantQS: map[string]string{"tier": "business"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListItDashboard(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetItDashboardRequiresUII(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetItDashboard(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetItDashboardBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetItDashboard(context.Background(), "024-000001001", nil)
	assertPathContains(t, capturedURL, "/api/itdashboard/024-000001001/")
}

func TestGetItDashboardWithOptions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetItDashboard(context.Background(), "uii-1234", &GetEntityOptions{
		Flat: true,
	})
	assertQueryContains(t, capturedURL, map[string]string{"flat": "true"}, nil)
}

func TestIterateItDashboardNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateItDashboard(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}
