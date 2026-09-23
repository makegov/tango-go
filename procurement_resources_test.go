package tango

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

// twoPageHandler serves two one-row pages, recording every request URI so a test can check the iterator forwards filters to the second page.
func twoPageHandler(seen *[]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*seen = append(*seen, r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, `{"count":2,"next":null,"results":[{"n":2}]}`)
			return
		}
		fmt.Fprintf(w, `{"count":2,"next":"http://%s%s?page=2","results":[{"n":1}]}`, r.Host, r.URL.Path)
	}
}

func TestListExclusionsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListExclusions(context.Background(), &ListExclusionsOptions{
		ListOptions:           ListOptions{Limit: 10},
		Active:                boolPtr(false),
		Delisted:              boolPtr(true),
		ClassificationType:    "Firm",
		ExclusionType:         "Ineligible (Proceedings Completed)",
		ExclusionProgram:      "Reciprocal",
		ExcludingAgencyCode:   "HHS",
		ExcludingAgencyName:   "Department of Health and Human Services",
		UEI:                   "ABC123DEF456",
		CageCode:              "1ABC2",
		NPI:                   "1234567890",
		EntityUEI:             "ZZZ999YYY888",
		ActivateDateAfter:     "2024-01-01",
		ActivateDateBefore:    "2024-12-31",
		TerminationDateAfter:  "2025-01-01",
		TerminationDateBefore: "2025-12-31",
		UpdateDateAfter:       "2026-01-01",
		UpdateDateBefore:      "2026-06-30",
		Search:                "acme",
		Ordering:              "-activate_date",
	})
	assertPathContains(t, capturedURL, "/api/exclusions/")
	assertQueryContains(t, capturedURL, map[string]string{
		"limit":                   "10",
		"active":                  "false",
		"delisted":                "true",
		"classification_type":     "Firm",
		"exclusion_type":          "Ineligible (Proceedings Completed)",
		"exclusion_program":       "Reciprocal",
		"excluding_agency_code":   "HHS",
		"excluding_agency_name":   "Department of Health and Human Services",
		"uei":                     "ABC123DEF456",
		"cage_code":               "1ABC2",
		"npi":                     "1234567890",
		"entity_uei":              "ZZZ999YYY888",
		"activate_date_after":     "2024-01-01",
		"activate_date_before":    "2024-12-31",
		"termination_date_after":  "2025-01-01",
		"termination_date_before": "2025-12-31",
		"update_date_after":       "2026-01-01",
		"update_date_before":      "2026-06-30",
		"search":                  "acme",
		"ordering":                "-activate_date",
	}, nil)
}

func TestListDibbsRfqsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListDibbsRfqs(context.Background(), &ListDibbsRfqsOptions{
		Open:               boolPtr(false),
		NSN:                "5305-01-123-4567|5305-01-765-4321",
		PartNumber:         "MS24693",
		Solicitation:       "SPE1C126Q0337",
		PurchaseRequest:    "7012345678",
		SetAside:           "Y",
		StatusCode:         "O",
		Organization:       "100000000",
		QuantityMin:        5,
		QuantityMax:        500,
		ReturnByDateAfter:  "2026-10-01",
		ReturnByDateBefore: "2026-10-31",
		IssueDateAfter:     "2026-09-01",
		IssueDateBefore:    "2026-09-30",
		Search:             "screw",
		Ordering:           "return_by_date",
	})
	assertPathContains(t, capturedURL, "/api/dibbs/rfqs/")
	assertQueryContains(t, capturedURL, map[string]string{
		"open":                  "false",
		"nsn":                   "5305-01-123-4567|5305-01-765-4321",
		"part_number":           "MS24693",
		"solicitation":          "SPE1C126Q0337",
		"purchase_request":      "7012345678",
		"set_aside":             "Y",
		"status_code":           "O",
		"organization":          "100000000",
		"quantity_min":          "5",
		"quantity_max":          "500",
		"return_by_date_after":  "2026-10-01",
		"return_by_date_before": "2026-10-31",
		"issue_date_after":      "2026-09-01",
		"issue_date_before":     "2026-09-30",
		"search":                "screw",
		"ordering":              "return_by_date",
	}, nil)
}

func TestListDibbsRfpsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListDibbsRfps(context.Background(), &ListDibbsRfpsOptions{
		Open:             boolPtr(true),
		NSN:              "2530-01-111-2222",
		PartNumber:       "A1",
		Solicitation:     "SPE7M126R0001",
		BuyerCode:        "PMCA",
		Organization:     "100000000",
		IssuedDateAfter:  "2026-08-01",
		IssuedDateBefore: "2026-08-31",
		ClosesDateAfter:  "2026-10-01",
		ClosesDateBefore: "2026-11-01",
		Search:           "brake",
		Ordering:         "-closes_date",
	})
	assertPathContains(t, capturedURL, "/api/dibbs/rfps/")
	assertQueryContains(t, capturedURL, map[string]string{
		"open":               "true",
		"nsn":                "2530-01-111-2222",
		"part_number":        "A1",
		"solicitation":       "SPE7M126R0001",
		"buyer_code":         "PMCA",
		"organization":       "100000000",
		"issued_date_after":  "2026-08-01",
		"issued_date_before": "2026-08-31",
		"closes_date_after":  "2026-10-01",
		"closes_date_before": "2026-11-01",
		"search":             "brake",
		"ordering":           "-closes_date",
	}, nil)
}

func TestListDibbsAwardsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListDibbsAwards(context.Background(), &ListDibbsAwardsOptions{
		NSN:                   "5305-01-123-4567",
		PartNumber:            "MS24693",
		Solicitation:          "SPE1C126Q0337",
		AwardNumber:           "SPE1C126P1234",
		DeliveryOrderNumber:   "0001",
		PurchaseRequest:       "7012345678",
		AwardeeCage:           "1ABC2",
		Entity:                "ABC123DEF456",
		Organization:          "100000000",
		AwardDateAfter:        "2026-01-01",
		AwardDateBefore:       "2026-06-30",
		PostedDateAfter:       "2026-01-02",
		PostedDateBefore:      "2026-07-01",
		TotalContractPriceMin: "1000",
		TotalContractPriceMax: "250000.50",
		Search:                "screw",
		Ordering:              "-award_date",
	})
	assertPathContains(t, capturedURL, "/api/dibbs/awards/")
	assertQueryContains(t, capturedURL, map[string]string{
		"nsn":                      "5305-01-123-4567",
		"part_number":              "MS24693",
		"solicitation":             "SPE1C126Q0337",
		"award_number":             "SPE1C126P1234",
		"delivery_order_number":    "0001",
		"purchase_request":         "7012345678",
		"awardee_cage":             "1ABC2",
		"entity":                   "ABC123DEF456",
		"organization":             "100000000",
		"award_date_after":         "2026-01-01",
		"award_date_before":        "2026-06-30",
		"posted_date_after":        "2026-01-02",
		"posted_date_before":       "2026-07-01",
		"total_contract_price_min": "1000",
		"total_contract_price_max": "250000.50",
		"search":                   "screw",
		"ordering":                 "-award_date",
	}, nil)
}

func TestListSbirTopicsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListSbirTopics(context.Background(), &ListSbirTopicsOptions{
		Activity:           "open",
		Agency:             "DOD",
		TopicNumber:        "A26-001",
		SolicitationNumber: "DOD_SBIR_2026_P1_C1",
		Year:               2026,
		DocSource:          "dsip",
		CloseDateAfter:     "2026-10-01",
		CloseDateBefore:    "2026-10-31",
		OpenDateAfter:      "2026-09-01",
		OpenDateBefore:     "2026-09-30",
		ReleaseDateAfter:   "2026-08-01",
		ReleaseDateBefore:  "2026-08-31",
		Search:             "autonomy",
		Ordering:           "close_date",
	})
	assertPathContains(t, capturedURL, "/api/sbir/topics/")
	assertQueryContains(t, capturedURL, map[string]string{
		"activity":            "open",
		"agency":              "DOD",
		"topic_number":        "A26-001",
		"solicitation_number": "DOD_SBIR_2026_P1_C1",
		"year":                "2026",
		"doc_source":          "dsip",
		"close_date_after":    "2026-10-01",
		"close_date_before":   "2026-10-31",
		"open_date_after":     "2026-09-01",
		"open_date_before":    "2026-09-30",
		"release_date_after":  "2026-08-01",
		"release_date_before": "2026-08-31",
		"search":              "autonomy",
		"ordering":            "close_date",
	}, nil)
}

func TestListSbirSolicitationsFilterMapping(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListSbirSolicitations(context.Background(), &ListSbirSolicitationsOptions{
		Activity:           "closed",
		Program:            "STTR",
		SolicitationNumber: "DOD_STTR_2026_P1_C1",
		CycleName:          "2026.1",
		SolicitationStatus: "Closed",
		OutOfCycle:         boolPtr(false),
		Year:               2026,
		StartDateAfter:     "2026-01-01",
		StartDateBefore:    "2026-01-31",
		EndDateAfter:       "2026-03-01",
		EndDateBefore:      "2026-03-31",
		Search:             "quantum",
		Ordering:           "-end_date",
	})
	assertPathContains(t, capturedURL, "/api/sbir/solicitations/")
	assertQueryContains(t, capturedURL, map[string]string{
		"activity":            "closed",
		"program":             "STTR",
		"solicitation_number": "DOD_STTR_2026_P1_C1",
		"cycle_name":          "2026.1",
		"solicitation_status": "Closed",
		"out_of_cycle":        "false",
		"year":                "2026",
		"start_date_after":    "2026-01-01",
		"start_date_before":   "2026-01-31",
		"end_date_after":      "2026-03-01",
		"end_date_before":     "2026-03-31",
		"search":              "quantum",
		"ordering":            "-end_date",
	}, nil)
}

func TestProcurementResourceListsOmitUnsetFilters(t *testing.T) {
	cases := map[string]func(c *Client) error{
		"exclusions": func(c *Client) error { _, err := c.ListExclusions(context.Background(), nil); return err },
		"dibbs rfqs": func(c *Client) error {
			_, err := c.ListDibbsRfqs(context.Background(), &ListDibbsRfqsOptions{})
			return err
		},
		"dibbs rfps":   func(c *Client) error { _, err := c.ListDibbsRfps(context.Background(), nil); return err },
		"dibbs awards": func(c *Client) error { _, err := c.ListDibbsAwards(context.Background(), nil); return err },
		"sbir topics": func(c *Client) error {
			_, err := c.ListSbirTopics(context.Background(), &ListSbirTopicsOptions{})
			return err
		},
		"sbir solicitations": func(c *Client) error { _, err := c.ListSbirSolicitations(context.Background(), nil); return err },
	}
	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			if err := call(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertQueryContains(t, capturedURL, nil, []string{"active", "open", "out_of_cycle", "quantity_min", "year", "ordering", "search"})
		})
	}
}

func TestProcurementResourceGetters(t *testing.T) {
	cases := []struct {
		name string
		path string
		get  func(c *Client, ctx context.Context, id string, opts *GetEntityOptions) (Record, error)
	}{
		{"exclusion", "/api/exclusions/", (*Client).GetExclusion},
		{"dibbs rfq", "/api/dibbs/rfqs/", (*Client).GetDibbsRfq},
		{"dibbs rfp", "/api/dibbs/rfps/", (*Client).GetDibbsRfp},
		{"dibbs award", "/api/dibbs/awards/", (*Client).GetDibbsAward},
		{"sbir topic", "/api/sbir/topics/", (*Client).GetSbirTopic},
		{"sbir solicitation", "/api/sbir/solicitations/", (*Client).GetSbirSolicitation},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			offline := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
			var ve *ValidationError
			if _, err := tc.get(offline, context.Background(), "", nil); !errors.As(err, &ve) {
				t.Fatalf("empty id: expected *ValidationError, got %T: %v", err, err)
			}

			var capturedURL string
			c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
			rec, err := tc.get(c, context.Background(), "a b/c", &GetEntityOptions{Shape: "*", Flat: true})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rec["id"] != "x" {
				t.Errorf("decoded record: want id x, got %#v", rec)
			}
			assertPathContains(t, capturedURL, tc.path+"a%20b%2Fc/")
			assertQueryContains(t, capturedURL, map[string]string{"shape": "*", "flat": "true"}, []string{"page", "limit"})
		})
	}
}

func TestProcurementResourceIteratorsWalkEveryPage(t *testing.T) {
	cases := map[string]func(c *Client) *Iterator[Record]{
		"exclusions": func(c *Client) *Iterator[Record] {
			return c.IterateExclusions(context.Background(), &ListExclusionsOptions{Search: "keep"})
		},
		"dibbs rfqs": func(c *Client) *Iterator[Record] {
			return c.IterateDibbsRfqs(context.Background(), &ListDibbsRfqsOptions{Search: "keep"})
		},
		"dibbs rfps": func(c *Client) *Iterator[Record] {
			return c.IterateDibbsRfps(context.Background(), &ListDibbsRfpsOptions{Search: "keep"})
		},
		"dibbs awards": func(c *Client) *Iterator[Record] {
			return c.IterateDibbsAwards(context.Background(), &ListDibbsAwardsOptions{Search: "keep"})
		},
		"sbir topics": func(c *Client) *Iterator[Record] {
			return c.IterateSbirTopics(context.Background(), &ListSbirTopicsOptions{Search: "keep"})
		},
		"sbir solicitations": func(c *Client) *Iterator[Record] {
			return c.IterateSbirSolicitations(context.Background(), &ListSbirSolicitationsOptions{Search: "keep"})
		},
	}
	for name, start := range cases {
		t.Run(name, func(t *testing.T) {
			var seen []string
			c, _ := newTestClient(t, twoPageHandler(&seen))
			it := start(c)
			var got []any
			for it.Next() {
				got = append(got, it.Item()["n"])
			}
			if err := it.Err(); err != nil {
				t.Fatalf("iterator error: %v", err)
			}
			if len(got) != 2 || got[0] != float64(1) || got[1] != float64(2) {
				t.Fatalf("want rows 1 and 2, got %#v", got)
			}
			if len(seen) != 2 {
				t.Fatalf("want 2 requests, got %d: %v", len(seen), seen)
			}
			assertQueryContains(t, seen[1], map[string]string{"page": "2", "search": "keep"}, nil)
		})
	}
}

func TestIterateProcurementResourcesNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	iters := []*Iterator[Record]{
		c.IterateExclusions(context.Background(), nil),
		c.IterateDibbsRfqs(context.Background(), nil),
		c.IterateDibbsRfps(context.Background(), nil),
		c.IterateDibbsAwards(context.Background(), nil),
		c.IterateSbirTopics(context.Background(), nil),
		c.IterateSbirSolicitations(context.Background(), nil),
	}
	for i, it := range iters {
		if it.Next() {
			t.Errorf("iterator %d: empty list should yield nothing", i)
		}
		if err := it.Err(); err != nil {
			t.Errorf("iterator %d: unexpected error %v", i, err)
		}
	}
}
