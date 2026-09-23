package tango

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestListProtestsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListProtestsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"source_system", "outcome", "agency"},
		},
		{
			name: "all filters",
			opts: &ListProtestsOptions{
				SourceSystem:       "GAO",
				Outcome:            "SUSTAINED",
				CaseType:           "PROTEST",
				Agency:             "9700",
				CaseNumber:         "2024-001",
				SolicitationNumber: "SOL-001",
				NaicsCode:          "541512",
				Protester:          "Acme Corp",
				Search:             "evaluation",
				FiledDateAfter:     "2024-01-01",
				FiledDateBefore:    "2024-12-31",
				DecisionDateAfter:  "2024-06-01",
				DecisionDateBefore: "2024-12-31",
			},
			wantQS: map[string]string{
				"source_system":        "GAO",
				"outcome":              "SUSTAINED",
				"case_type":            "PROTEST",
				"agency":               "9700",
				"case_number":          "2024-001",
				"solicitation_number":  "SOL-001",
				"naics_code":           "541512",
				"protester":            "Acme Corp",
				"search":               "evaluation",
				"filed_date_after":     "2024-01-01",
				"filed_date_before":    "2024-12-31",
				"decision_date_after":  "2024-06-01",
				"decision_date_before": "2024-12-31",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListProtestsOptions{},
			notInQS: []string{"source_system", "outcome", "case_type"},
		},
		{
			name:   "extra map",
			opts:   &ListProtestsOptions{Extra: map[string]any{"tier": "pro"}},
			wantQS: map[string]string{"tier": "pro"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListProtests(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetProtestRequiresCaseID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetProtest(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetProtestBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"case_number":"2024-001","outcome":"SUSTAINED"}`))
	})
	_, _ = c.GetProtest(context.Background(), "89026562-0000-4000-8000-000000000000", nil)
	assertPathContains(t, capturedURL, "/api/protests/89026562-0000-4000-8000-000000000000/")
}

func TestGetProtestWithOptions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})
	_, _ = c.GetProtest(context.Background(), "2024-001", &GetEntityOptions{
		Shape: "protests(with_docket)", Flat: true,
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"shape": "protests(with_docket)",
		"flat":  "true",
	}, nil)
}

func TestIterateProtestsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateProtests(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestGetProtestDecodesDetailPayload(t *testing.T) {
	const body = `{
		"case_id": "89026562-0000-4000-8000-000000000000",
		"source_system": "cofc",
		"case_number": "26-292",
		"title": "EXAMPLE CORP v. USA",
		"protester": "Example Corp",
		"agency": "N/A",
		"solicitation_number": null,
		"case_type": "Bid Protest",
		"outcome": "Sustained",
		"filed_date": "2026-03-02",
		"posted_date": null,
		"decision_date": "2026-06-15",
		"due_date": null,
		"docket_url": "",
		"decision_url": "https://ecf.cofc.uscourts.gov/cgi-bin/show_public_doc?2026cv0292-71-0",
		"organization": null,
		"dockets": [{"docket_number": "1:26-cv-00292", "filed_date": "2026-03-02"}],
		"decisions": [{"title": "Opinion", "decision_date": "2026-06-15", "outcome": "Sustained", "judges": "e_roumel", "document_url": "https://example.test/doc", "courtlistener_url": null, "document_type": null}],
		"resolved_agency": {"key": "9700", "name": "Department of Defense", "match_confidence": 0.9, "rationale": "exact"}
	}`
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	})
	got, err := c.GetProtest(context.Background(), "89026562-0000-4000-8000-000000000000", nil)
	if err != nil {
		t.Fatalf("GetProtest decode failed: %v", err)
	}
	if got.Protester != "Example Corp" || got.Agency != "N/A" {
		t.Fatalf("agency/protester = %q/%q", got.Agency, got.Protester)
	}
	if len(got.Dockets) != 1 || got.Dockets[0]["docket_number"] != "1:26-cv-00292" {
		t.Fatalf("dockets = %#v", got.Dockets)
	}
	if len(got.Decisions) != 1 || got.Decisions[0]["outcome"] != "Sustained" {
		t.Fatalf("decisions = %#v", got.Decisions)
	}
	if got.DecisionURL == "" || got.Title != "EXAMPLE CORP v. USA" || got.SolicitationNumber != "" {
		t.Fatalf("scalar fields = %+v", got)
	}
	if got.ResolvedAgency["name"] != "Department of Defense" {
		t.Fatalf("resolved_agency = %#v", got.ResolvedAgency)
	}
}
