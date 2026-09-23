package tango

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestListContractAppealsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListContractAppealsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"board", "docket", "appellant", "listed"},
		},
		{
			name: "all filters",
			opts: &ListContractAppealsOptions{
				ListOptions:        ListOptions{Limit: 25, Shape: ShapeContractAppealsMinimal},
				Board:              "cbca",
				Docket:             "CBCA 1234",
				Appellant:          "Acme Corp",
				Judge:              "Somers",
				DecisionType:       "denied",
				DecisionDateAfter:  "2024-01-01",
				DecisionDateBefore: "2024-12-31",
				Listed:             boolPtr(true),
				DocumentID:         "doc-9",
				Search:             "differing site conditions",
				Ordering:           "-decision_date",
			},
			wantQS: map[string]string{
				"board":                "cbca",
				"docket":               "CBCA 1234",
				"appellant":            "Acme Corp",
				"judge":                "Somers",
				"decision_type":        "denied",
				"decision_date_after":  "2024-01-01",
				"decision_date_before": "2024-12-31",
				"listed":               "true",
				"document_id":          "doc-9",
				"search":               "differing site conditions",
				"ordering":             "-decision_date",
				"limit":                "25",
				"shape":                ShapeContractAppealsMinimal,
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListContractAppealsOptions{},
			notInQS: []string{"board", "docket", "decision_type", "listed", "ordering"},
		},
		{
			name:   "listed false is a filter value, not an absent one",
			opts:   &ListContractAppealsOptions{Listed: boolPtr(false)},
			wantQS: map[string]string{"listed": "false"},
		},
		{
			name:   "extra map",
			opts:   &ListContractAppealsOptions{Extra: map[string]any{"verbose": true}},
			wantQS: map[string]string{"verbose": "true"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListContractAppeals(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestListContractAppealsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListContractAppeals(context.Background(), nil)
	assertPathContains(t, capturedURL, "/api/contract_appeals/")
}

func TestListContractAppealsDecodesResults(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":1,"results":[{"uuid":"3f2b","board":"asbca","appellant":"Acme Corp"}]}`))
	})
	page, err := c.ListContractAppeals(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListContractAppeals: %v", err)
	}
	if len(page.Results) != 1 {
		t.Fatalf("want 1 result, got %d", len(page.Results))
	}
	if page.Results[0]["board"] != "asbca" {
		t.Errorf("board: want asbca, got %#v", page.Results[0]["board"])
	}
}

func TestGetContractAppealRequiresUUID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetContractAppeal(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetContractAppealBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"uuid":"9c1e-4a","board":"cbca"}`))
	})
	got, err := c.GetContractAppeal(context.Background(), "9c1e-4a", nil)
	if err != nil {
		t.Fatalf("GetContractAppeal: %v", err)
	}
	assertPathContains(t, capturedURL, "/api/contract_appeals/9c1e-4a/")
	assertStrPtr(t, "board", got.Board, "cbca")
}

func TestGetContractAppealWithOptions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})
	_, _ = c.GetContractAppeal(context.Background(), "9c1e-4a", &GetEntityOptions{
		Shape: ShapeContractAppealsComprehensive, Flat: true,
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"shape": ShapeContractAppealsComprehensive,
		"flat":  "true",
	}, nil)
}

func TestIterateContractAppealsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateContractAppeals(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestContractAppealRecordUnmarshalFullRow(t *testing.T) {
	raw := []byte(`{
		"uuid":"3f2b","board":"cbca","docket_numbers":["CBCA 1234","CBCA 1235"],
		"docket_source":"listing","docket_raw":"CBCA 1234, 1235",
		"decision_date":"2024-06-04","decision_date_raw":"June 4, 2024","decision_date_repaired":true,
		"appellant":"Acme Corp","judge":"Somers","decision_type":"denied","decision_type_raw":"DENIED",
		"url":"https://example.invalid/d.pdf","document_id":"doc-9",
		"listing_url":"https://example.invalid/2024","listing_year":2024,
		"first_listed_at":"2024-06-05T12:00:00Z","listed":true,
		"text_status":"extracted","text_char_count":48213,
		"decision_text":"OPINION BY ADMINISTRATIVE JUDGE"
	}`)
	var rec ContractAppealRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	assertStrPtr(t, "uuid", rec.UUID, "3f2b")
	assertStrPtr(t, "board", rec.Board, "cbca")
	assertStrPtr(t, "appellant", rec.Appellant, "Acme Corp")
	assertStrPtr(t, "decision_date", rec.DecisionDate, "2024-06-04")
	assertStrPtr(t, "decision_text", rec.DecisionText, "OPINION BY ADMINISTRATIVE JUDGE")
	if len(rec.DocketNumbers) != 2 || rec.DocketNumbers[1] != "CBCA 1235" {
		t.Errorf("docket_numbers: want both dockets, got %#v", rec.DocketNumbers)
	}
	if rec.ListingYear == nil || *rec.ListingYear != 2024 {
		t.Errorf("listing_year: want 2024, got %v", rec.ListingYear)
	}
	if rec.TextCharCount == nil || *rec.TextCharCount != 48213 {
		t.Errorf("text_char_count: want 48213, got %v", rec.TextCharCount)
	}
	if rec.Listed == nil || !*rec.Listed {
		t.Errorf("listed: want true, got %v", rec.Listed)
	}
	if rec.DecisionDateRepaired == nil || !*rec.DecisionDateRepaired {
		t.Errorf("decision_date_repaired: want true, got %v", rec.DecisionDateRepaired)
	}
}

// Below the Enterprise plan the key is absent rather than null, so nil must survive as "not served to you" rather than collapsing into an empty string.
func TestContractAppealRecordWithoutDecisionText(t *testing.T) {
	raw := []byte(`{"uuid":"3f2b","board":"asbca","decision_date":"2024-06-04","text_status":"extracted","text_char_count":48213}`)
	var rec ContractAppealRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rec.DecisionText != nil {
		t.Errorf("decision_text: want nil when the server withholds it, got %q", *rec.DecisionText)
	}
	if rec.Listed != nil {
		t.Errorf("listed: want nil when absent, got %v", *rec.Listed)
	}
	if rec.TextCharCount == nil || *rec.TextCharCount != 48213 {
		t.Errorf("text_char_count: want 48213, got %v", rec.TextCharCount)
	}
}

func TestContractAppealRecordUnmarshalExtra(t *testing.T) {
	raw := []byte(`{"uuid":"3f2b","unknown_future_field":"future_val"}`)
	var rec ContractAppealRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rec.Extra["unknown_future_field"] != "future_val" {
		t.Errorf("Extra[unknown_future_field]: want %q, got %v", "future_val", rec.Extra["unknown_future_field"])
	}
	if _, ok := rec.Extra["uuid"]; ok {
		t.Error("uuid is a named field and should not appear in Extra")
	}
}

// The API serves the decision body only to an Enterprise caller, so a suggested shape naming it would ask every detail fetch for something most callers cannot read.
func TestContractAppealShapesDoNotNameTheEnterpriseDecisionText(t *testing.T) {
	for name, shape := range map[string]string{
		"ShapeContractAppealsMinimal":       ShapeContractAppealsMinimal,
		"ShapeContractAppealsComprehensive": ShapeContractAppealsComprehensive,
	} {
		if strings.Contains(shape, "decision_text") {
			t.Errorf("%s must not name the Enterprise-gated decision_text leaf: %q", name, shape)
		}
	}
}
