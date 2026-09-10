package tango

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestListSledOpportunitiesFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListSledOpportunitiesOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"state", "status", "active", "search"},
		},
		{
			// The open-only default is the API's. Synthesizing status=open here
			// would make Active=false unreachable, since it is the complement of
			// open rather than an independent value.
			name:    "no liveness filter is sent when the caller requests none",
			opts:    &ListSledOpportunitiesOptions{State: "TX"},
			wantQS:  map[string]string{"state": "TX"},
			notInQS: []string{"status", "active"},
		},
		{
			name:   "active=false is sent, not dropped",
			opts:   &ListSledOpportunitiesOptions{Active: boolPtr(false)},
			wantQS: map[string]string{"active": "false"},
		},
		{
			name:   "has_documents=false is sent, not dropped",
			opts:   &ListSledOpportunitiesOptions{HasDocuments: boolPtr(false)},
			wantQS: map[string]string{"has_documents": "false"},
		},
		{
			// status is Tango-derived liveness; the portal's frozen word is
			// served but not filterable.
			name:    "source_status is not a filter",
			opts:    &ListSledOpportunitiesOptions{Status: "closed"},
			wantQS:  map[string]string{"status": "closed"},
			notInQS: []string{"source_status"},
		},
		{
			name: "all filters",
			opts: &ListSledOpportunitiesOptions{
				State:                  "TX|OK",
				Jurisdiction:           "local|education",
				Status:                 "open|unknown",
				Active:                 boolPtr(true),
				Agency:                 "Texas Commission",
				SolicitationNumber:     "RFP-2026-001",
				SolicitationType:       "rfp",
				HasDocuments:           boolPtr(true),
				RevisionKind:           "deadline_change",
				Naics:                  "541620",
				Nigp:                   "962-47",
				Unspsc:                 "77101500",
				Category:               "Environmental Services",
				CategoryCode:           "541620",
				PostedAfter:            "2026-01-01",
				PostedBefore:           "2026-12-31",
				ResponseDeadlineAfter:  "2026-09-01",
				ResponseDeadlineBefore: "2026-10-01",
				FirstSeenAfter:         "2026-09-01",
				FirstSeenBefore:        "2026-09-30",
				ChangeSeenAfter:        "2026-09-05",
				ModifiedAfter:          "2026-09-01",
				ModifiedBefore:         "2026-09-30",
				Platform:               "custom",
				NativeID:               "abc-123",
				ExternalID:             "tx:custom:abc-123",
				Search:                 "environmental mitigation",
				Ordering:               "response_deadline",
			},
			wantQS: map[string]string{
				"state":                    "TX|OK",
				"jurisdiction":             "local|education",
				"status":                   "open|unknown",
				"active":                   "true",
				"agency":                   "Texas Commission",
				"solicitation_number":      "RFP-2026-001",
				"solicitation_type":        "rfp",
				"has_documents":            "true",
				"revision_kind":            "deadline_change",
				"naics":                    "541620",
				"nigp":                     "962-47",
				"unspsc":                   "77101500",
				"category":                 "Environmental Services",
				"category_code":            "541620",
				"posted_after":             "2026-01-01",
				"posted_before":            "2026-12-31",
				"response_deadline_after":  "2026-09-01",
				"response_deadline_before": "2026-10-01",
				"first_seen_after":         "2026-09-01",
				"first_seen_before":        "2026-09-30",
				"change_seen_after":        "2026-09-05",
				"modified_after":           "2026-09-01",
				"modified_before":          "2026-09-30",
				"platform":                 "custom",
				"native_id":                "abc-123",
				"external_id":              "tx:custom:abc-123",
				"search":                   "environmental mitigation",
				"ordering":                 "response_deadline",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListSledOpportunitiesOptions{},
			notInQS: []string{"state", "jurisdiction", "status", "naics", "search"},
		},
		{
			name:   "extra map",
			opts:   &ListSledOpportunitiesOptions{Extra: map[string]any{"verbose": "true"}},
			wantQS: map[string]string{"verbose": "true"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListSledOpportunities(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestListSledOpportunitiesPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListSledOpportunities(context.Background(), nil)
	assertPathContains(t, capturedURL, "/api/sled/opportunities/")
}

func TestGetSledOpportunityRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetSledOpportunity(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetSledOpportunityBuildsPathAndShape(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetSledOpportunity(context.Background(), "550e8400-e29b-41d4-a716-446655440000", &GetEntityOptions{
		Shape: ShapeSledOpportunitiesComprehensive, Flat: true,
	})
	assertPathContains(t, capturedURL, "/api/sled/opportunities/550e8400-e29b-41d4-a716-446655440000/")
	assertQueryContains(t, capturedURL, map[string]string{
		"shape": ShapeSledOpportunitiesComprehensive,
		"flat":  "true",
	}, nil)
}

func TestListSledOpportunityRevisions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListSledOpportunityRevisions(context.Background(), "abc", &ListSledOpportunityRevisionsOptions{
		Kind:           "deadline_change",
		SourceDeclared: boolPtr(true),
		ObservedAfter:  "2026-09-01",
	})
	assertPathContains(t, capturedURL, "/api/sled/opportunities/abc/revisions/")
	assertQueryContains(t, capturedURL, map[string]string{
		"kind":            "deadline_change",
		"source_declared": "true",
		"observed_after":  "2026-09-01",
	}, nil)
}

// The revisions(*) expand excludes enrichment rows; this route is how a caller
// reaches them.
func TestListSledOpportunityRevisionsReachesEnrichment(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListSledOpportunityRevisions(context.Background(), "abc", &ListSledOpportunityRevisionsOptions{
		Kind: "enrichment",
	})
	assertQueryContains(t, capturedURL, map[string]string{"kind": "enrichment"}, nil)
}

func TestListSledOpportunityRevisionsRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListSledOpportunityRevisions(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

// `changes` needs a Small plan, so naming it in the suggested revision shape
// would 403 a Free caller on a field they never asked to gate.
func TestSledRevisionsShapeOmitsPlanGatedChanges(t *testing.T) {
	fields := strings.Split(ShapeSledRevisionsMinimal, ",")
	for _, f := range fields {
		if f == "changes" {
			t.Fatalf("ShapeSledRevisionsMinimal must not name the Small-gated `changes` leaf: %q", ShapeSledRevisionsMinimal)
		}
	}
	if !strings.Contains(ShapeSledRevisionsMinimal, "changed_fields") {
		t.Fatalf("ShapeSledRevisionsMinimal should carry `changed_fields`, which every plan can read: %q", ShapeSledRevisionsMinimal)
	}
}

// The API resolves the document body only for a caller who names it, so a
// suggested shape naming it would make every detail fetch pay for it.
func TestSledShapesDoNotNameThePaidDocumentBody(t *testing.T) {
	for name, shape := range map[string]string{
		"ShapeSledOpportunitiesMinimal":       ShapeSledOpportunitiesMinimal,
		"ShapeSledOpportunitiesComprehensive": ShapeSledOpportunitiesComprehensive,
	} {
		if strings.Contains(shape, "extracted_text") {
			t.Errorf("%s must not name the Small-gated extracted_text leaf: %q", name, shape)
		}
	}
}

func TestGetSledCoverage(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"generated_at":"2026-09-10T14:00:00Z","totals":{"opportunities":3,"forecasts":1},"states":[{"state":"TX","total_count":3}]}`))
	})
	got, err := c.GetSledCoverage(context.Background())
	if err != nil {
		t.Fatalf("GetSledCoverage: %v", err)
	}
	assertPathContains(t, capturedURL, "/api/sled/opportunities/coverage/")
	assertQueryContains(t, capturedURL, nil, []string{"shape", "page", "limit"})
	totals, ok := got["totals"].(map[string]any)
	if !ok {
		t.Fatalf("coverage payload has no totals object: %#v", got)
	}
	if totals["opportunities"] != float64(3) {
		t.Errorf("totals.opportunities: want 3, got %#v", totals["opportunities"])
	}
}

func TestListSledForecastsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListSledForecasts(context.Background(), &ListSledForecastsOptions{
		State:               "MD",
		Agency:              "Department of Transportation",
		ProcurementCategory: "Services",
		ProcurementMethod:   "Competitive Sealed Proposals",
		ContractNumber:      "K-2026-001",
		IncumbentName:       "Acme",
		AdvertisementAfter:  "2026-10-01",
		AdvertisementBefore: "2027-03-31",
		FirstSeenAfter:      "2026-09-01",
		FirstSeenBefore:     "2026-09-30",
		ModifiedAfter:       "2026-09-01",
		ModifiedBefore:      "2026-09-30",
		Search:              "data center",
		Ordering:            "estimated_advertisement_date",
	})
	assertPathContains(t, capturedURL, "/api/sled/forecasts/")
	assertQueryContains(t, capturedURL, map[string]string{
		"state":                "MD",
		"agency":               "Department of Transportation",
		"procurement_category": "Services",
		"procurement_method":   "Competitive Sealed Proposals",
		"contract_number":      "K-2026-001",
		"incumbent_name":       "Acme",
		"advertisement_after":  "2026-10-01",
		"advertisement_before": "2027-03-31",
		"first_seen_after":     "2026-09-01",
		"first_seen_before":    "2026-09-30",
		"modified_after":       "2026-09-01",
		"modified_before":      "2026-09-30",
		"search":               "data center",
		"ordering":             "estimated_advertisement_date",
	}, []string{"status", "active"})
}

func TestGetSledForecastRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetSledForecast(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetSledForecastBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetSledForecast(context.Background(), "f1", nil)
	assertPathContains(t, capturedURL, "/api/sled/forecasts/f1/")
}

func TestIterateSledNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	if it := c.IterateSledOpportunities(context.Background(), nil); it == nil {
		t.Fatal("expected non-nil solicitation iterator")
	}
	if it := c.IterateSledForecasts(context.Background(), nil); it == nil {
		t.Fatal("expected non-nil forecast iterator")
	}
}
