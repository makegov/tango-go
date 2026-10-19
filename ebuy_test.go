package tango

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestListEbuyRequestsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, err := c.ListEbuyRequests(context.Background(), &ListEbuyRequestsOptions{
		ListOptions:     ListOptions{Page: 2, Limit: 50, Shape: ShapeEbuyRequestsMinimal},
		Search:          "cybersecurity",
		RfqID:           "RFQ1835158",
		ReferenceNumber: "W912DY-26-Q-0012",
		RequestType:     "RFQ|RFI",
		Status:          "Open",
		Sin:             "54151S",
		Schedule:        "MAS",
		BuyerAgency:     "Department of the Army",
		Agency:          "DOD",
		ContractNumber:  "CONTRACT-1",
		IssueDateAfter:  "2026-01-01",
		IssueDateBefore: "2026-01-31",
		CloseDateAfter:  "2026-02-01",
		CloseDateBefore: "2026-02-28",
		Ordering:        "-close_date",
		Extra:           map[string]any{"flat": true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertPathContains(t, capturedURL, "/api/ebuy/requests/")
	assertQueryContains(t, capturedURL, map[string]string{
		"page":              "2",
		"limit":             "50",
		"shape":             ShapeEbuyRequestsMinimal,
		"search":            "cybersecurity",
		"rfq_id":            "RFQ1835158",
		"reference_number":  "W912DY-26-Q-0012",
		"request_type":      "RFQ|RFI",
		"status":            "Open",
		"sin":               "54151S",
		"schedule":          "MAS",
		"buyer_agency":      "Department of the Army",
		"agency":            "DOD",
		"contract_number":   "CONTRACT-1",
		"issue_date_after":  "2026-01-01",
		"issue_date_before": "2026-01-31",
		"close_date_after":  "2026-02-01",
		"close_date_before": "2026-02-28",
		"ordering":          "-close_date",
		"flat":              "true",
	}, nil)
}

func TestListEbuyRequestsNilOptionsSendsNoFilters(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	if _, err := c.ListEbuyRequests(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(capturedURL, "?") {
		t.Fatalf("expected no query string, got %q", capturedURL)
	}
}

func TestIterateEbuyRequestsWalksPages(t *testing.T) {
	var pages []string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		pages = append(pages, r.URL.Query().Get("page"))
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`{"count":2,"next":null,"results":[{"rfq_id":"RFQ2"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"count":2,"next":"http://x/api/ebuy/requests/?page=2","results":[{"rfq_id":"RFQ1"}]}`))
	})
	var ids []string
	for rec, err := range c.IterateEbuyRequests(context.Background(), &ListEbuyRequestsOptions{Status: "Open"}).Seq() {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ids = append(ids, rec["rfq_id"].(string))
	}
	if strings.Join(ids, ",") != "RFQ1,RFQ2" {
		t.Fatalf("expected RFQ1,RFQ2, got %v (pages requested %v)", ids, pages)
	}
}

func TestGetEbuyRequest(t *testing.T) {
	offline := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	var ve *ValidationError
	if _, err := offline.GetEbuyRequest(context.Background(), "", nil); !errors.As(err, &ve) {
		t.Fatalf("empty id: expected *ValidationError, got %T: %v", err, err)
	}

	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	rec, err := c.GetEbuyRequest(context.Background(), "RFQ1835158", &GetEntityOptions{Shape: "rfq_id,attachments(*)"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec["id"] != "x" {
		t.Fatalf("unexpected record: %v", rec)
	}
	assertPathContains(t, capturedURL, "/api/ebuy/requests/RFQ1835158/")
	assertQueryContains(t, capturedURL, map[string]string{"shape": "rfq_id,attachments(*)"}, nil)
}

func TestGetEbuyRequestNamedExtractedTextLeaf(t *testing.T) {
	const shape = "rfq_id,attachments(doc_seq_num,doc_name,extracted_text)"
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rfq_id":"RFQ1835158","attachments":[` +
			`{"doc_seq_num":1,"doc_name":"sow.pdf","extracted_text":"Statement of work"},` +
			`{"doc_seq_num":2,"doc_name":"portal link"}]}`))
	})
	rec, err := c.GetEbuyRequest(context.Background(), "RFQ1835158", &GetEntityOptions{Shape: shape})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertQueryContains(t, capturedURL, map[string]string{"shape": shape}, nil)

	attachments, ok := rec["attachments"].([]any)
	if !ok || len(attachments) != 2 {
		t.Fatalf("expected two attachments, got %v", rec["attachments"])
	}
	if got := attachments[0].(map[string]any)["extracted_text"]; got != "Statement of work" {
		t.Fatalf("expected the extracted text, got %v", got)
	}
	if _, present := attachments[1].(map[string]any)["extracted_text"]; present {
		t.Fatalf("expected no extracted_text key on an attachment without text, got %v", attachments[1])
	}
}

func TestGetEbuyRequestOutOfScopeIsNotFound(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"Not found."}`))
	})
	var nf *NotFoundError
	if _, err := c.GetEbuyRequest(context.Background(), "RFQ0", nil); !errors.As(err, &nf) {
		t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
	}
}

func TestGetEbuyAttachmentURLReturnsRedirectTargetWithoutFollowingIt(t *testing.T) {
	const target = "https://documents.example.com/rfq/doc.pdf?X-Amz-Signature=abc"
	var requested []string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requested = append(requested, r.URL.Path)
		if r.Header.Get("X-API-KEY") != "test-key" {
			t.Errorf("expected API key header on the download request")
		}
		http.Redirect(w, r, target, http.StatusFound)
	})
	got, err := c.GetEbuyAttachmentURL(context.Background(), "RFQ1835158", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != target {
		t.Fatalf("expected %q, got %q", target, got)
	}
	if len(requested) != 1 || requested[0] != "/api/ebuy/requests/RFQ1835158/attachments/3/download/" {
		t.Fatalf("expected exactly one request to the download path, got %v", requested)
	}
}

func TestGetEbuyAttachmentURLDoesNotChangeRedirectHandlingForOtherCalls(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/moved/" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"enabled":true,"reason":null,"contracts":[]}`))
			return
		}
		http.Redirect(w, r, "/moved/", http.StatusFound)
	})
	if _, err := c.GetEbuyAttachmentURL(context.Background(), "RFQ1", 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	access, err := c.GetEbuyAccess(context.Background())
	if err != nil {
		t.Fatalf("expected the shared client to still follow redirects, got %v", err)
	}
	if !access.Enabled {
		t.Fatalf("expected the followed response, got %+v", access)
	}
}

func TestGetEbuyAttachmentURLLinkEntrySurfacesURL(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"This entry is an external link, not a stored document.","url":"https://agency.example.gov/sow"}`))
	})
	_, err := c.GetEbuyAttachmentURL(context.Background(), "RFQ1", 2)
	var linkErr *EbuyLinkAttachmentError
	if !errors.As(err, &linkErr) {
		t.Fatalf("expected *EbuyLinkAttachmentError, got %T: %v", err, err)
	}
	if linkErr.URL != "https://agency.example.gov/sow" {
		t.Fatalf("unexpected URL %q", linkErr.URL)
	}
	var ve *ValidationError
	if !errors.As(err, &ve) || ve.StatusCode != 400 {
		t.Fatalf("expected the error to unwrap to a 400 *ValidationError, got %v", err)
	}
}

func TestGetEbuyAttachmentURLErrors(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/ebuy/requests/RFQ1/attachments/1/download/":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"The document for this attachment has not been captured yet."}`))
		case "/api/ebuy/requests/RFQ1/attachments/2/download/":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"detail":"bad"}`))
		case "/api/ebuy/requests/RFQ1/attachments/3/download/":
			w.WriteHeader(http.StatusFound)
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}
	})
	ctx := context.Background()

	var nf *NotFoundError
	if _, err := c.GetEbuyAttachmentURL(ctx, "RFQ1", 1); !errors.As(err, &nf) {
		t.Fatalf("uncaptured document: expected *NotFoundError, got %T: %v", err, err)
	}

	var linkErr *EbuyLinkAttachmentError
	var ve *ValidationError
	if _, err := c.GetEbuyAttachmentURL(ctx, "RFQ1", 2); errors.As(err, &linkErr) || !errors.As(err, &ve) {
		t.Fatalf("400 without url: expected a plain *ValidationError, got %T: %v", err, err)
	}

	if _, err := c.GetEbuyAttachmentURL(ctx, "RFQ1", 3); err == nil || !strings.Contains(err.Error(), "no Location") {
		t.Fatalf("redirect without Location: expected an error, got %v", err)
	}

	if _, err := c.GetEbuyAttachmentURL(ctx, "RFQ1", 4); err == nil || !strings.Contains(err.Error(), "expected a redirect") {
		t.Fatalf("2xx: expected an error, got %v", err)
	}

	if _, err := c.GetEbuyAttachmentURL(ctx, "", 1); !errors.As(err, &ve) {
		t.Fatalf("empty id: expected *ValidationError, got %T: %v", err, err)
	}
	if _, err := c.GetEbuyAttachmentURL(ctx, "RFQ1", -1); !errors.As(err, &ve) {
		t.Fatalf("negative doc_seq_num: expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetEbuyAccess(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		enabled    bool
		reason     string
		contracts  []string
		extraField string
	}{
		{"enabled", `{"enabled":true,"reason":null,"contracts":["A-1","B-2"]}`, true, "", []string{"A-1", "B-2"}, ""},
		{"tier required", `{"enabled":false,"reason":"tier_required","contracts":[]}`, false, "tier_required", []string{}, ""},
		{"no grant", `{"enabled":false,"reason":"no_contract_grant","contracts":[],"note":"x"}`, false, "no_contract_grant", []string{}, "note"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				capturedURL = r.URL.RequestURI()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			})
			got, err := c.GetEbuyAccess(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if capturedURL != "/api/ebuy/access/" {
				t.Fatalf("unexpected path %q", capturedURL)
			}
			if got.Enabled != tc.enabled {
				t.Fatalf("Enabled: want %v, got %v", tc.enabled, got.Enabled)
			}
			gotReason := ""
			if got.Reason != nil {
				gotReason = *got.Reason
			}
			if gotReason != tc.reason {
				t.Fatalf("Reason: want %q, got %q", tc.reason, gotReason)
			}
			if strings.Join(got.Contracts, ",") != strings.Join(tc.contracts, ",") {
				t.Fatalf("Contracts: want %v, got %v", tc.contracts, got.Contracts)
			}
			if tc.extraField != "" {
				if _, ok := got.Extra[tc.extraField]; !ok {
					t.Fatalf("expected %q in Extra, got %v", tc.extraField, got.Extra)
				}
			}
		})
	}
}
