package tango

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestListFederalRegisterDocumentsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListFederalRegisterDocuments(context.Background(), &ListFederalRegisterDocumentsOptions{
		ListOptions:           ListOptions{Limit: 10, Shape: ShapeFederalRegisterMinimal},
		Search:                `"clean air"`,
		DocumentNumber:        "2016-31922|2017-00001",
		Type:                  "Rule|Proposed Rule",
		Agency:                "EPA",
		FRAgency:              "environmental-protection-agency",
		PublicationDateAfter:  "2024-01-01",
		PublicationDateBefore: "2024-12-31",
		EffectiveOnAfter:      "2024-02-01",
		EffectiveOnBefore:     "2024-11-30",
		CommentsCloseOnAfter:  "2024-03-01",
		CommentsCloseOnBefore: "2024-10-31",
		CommentsOpen:          boolPtr(false),
		CFRTitle:              "40",
		CFRPart:               "52",
		Significant:           boolPtr(false),
		RIN:                   "2060-AV16",
		ExecutiveOrderNumber:  "14028",
		Ordering:              "rank",
	})
	assertPathContains(t, capturedURL, "/api/federal_register/")
	assertQueryContains(t, capturedURL, map[string]string{
		"limit":                    "10",
		"shape":                    ShapeFederalRegisterMinimal,
		"search":                   `"clean air"`,
		"document_number":          "2016-31922|2017-00001",
		"type":                     "Rule|Proposed Rule",
		"agency":                   "EPA",
		"fr_agency":                "environmental-protection-agency",
		"publication_date_after":   "2024-01-01",
		"publication_date_before":  "2024-12-31",
		"effective_on_after":       "2024-02-01",
		"effective_on_before":      "2024-11-30",
		"comments_close_on_after":  "2024-03-01",
		"comments_close_on_before": "2024-10-31",
		"comments_open":            "false",
		"cfr_title":                "40",
		"cfr_part":                 "52",
		"significant":              "false",
		"rin":                      "2060-AV16",
		"executive_order_number":   "14028",
		"ordering":                 "rank",
	}, nil)
}

func TestListFederalRegisterDocumentsOmitsUnsetFilters(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListFederalRegisterDocuments(context.Background(), nil)
	assertQueryContains(t, capturedURL, nil, []string{"comments_open", "significant", "search", "agency", "shape"})
}

func TestGetFederalRegisterDocument(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	rec, err := c.GetFederalRegisterDocument(context.Background(), "0b6c8f8e-1f4a-4c55-9d6f-2c2a3c2f1a10", &GetEntityOptions{Shape: "uuid,full_text"})
	if err != nil {
		t.Fatalf("GetFederalRegisterDocument: %v", err)
	}
	if rec["id"] != "x" {
		t.Errorf("unexpected record %v", rec)
	}
	assertPathContains(t, capturedURL, "/api/federal_register/0b6c8f8e-1f4a-4c55-9d6f-2c2a3c2f1a10/")
	assertQueryContains(t, capturedURL, map[string]string{"shape": "uuid,full_text"}, nil)

	var ve *ValidationError
	if _, err := c.GetFederalRegisterDocument(context.Background(), "", nil); !errors.As(err, &ve) {
		t.Errorf("empty uuid: want *ValidationError, got %v", err)
	}
}

func TestIterateFederalRegisterDocumentsForwardsFilters(t *testing.T) {
	var seen []string
	c, _ := newTestClient(t, twoPageHandler(&seen))
	it := c.IterateFederalRegisterDocuments(context.Background(), &ListFederalRegisterDocumentsOptions{Agency: "EPA"})
	n := 0
	for it.Next() {
		n++
	}
	if err := it.Err(); err != nil {
		t.Fatalf("iterate: %v", err)
	}
	if n != 2 || len(seen) != 2 {
		t.Fatalf("want 2 rows over 2 requests, got %d rows over %v", n, seen)
	}
	assertQueryContains(t, seen[1], map[string]string{"agency": "EPA", "page": "2"}, nil)

	if c.IterateFederalRegisterDocuments(context.Background(), nil) == nil {
		t.Fatal("nil opts should still build an iterator")
	}
}

func TestFederalRegisterShapesDoNotNameFullText(t *testing.T) {
	for name, shape := range map[string]string{
		"ShapeFederalRegisterMinimal":       ShapeFederalRegisterMinimal,
		"ShapeFederalRegisterComprehensive": ShapeFederalRegisterComprehensive,
	} {
		for _, f := range strings.Split(shape, ",") {
			if f == "full_text" {
				t.Errorf("%s names full_text, which can run to several MB", name)
			}
		}
		if !strings.HasPrefix(shape, "uuid,document_number,") {
			t.Errorf("%s should lead with the uuid the detail route takes: %q", name, shape)
		}
	}
}
