package tango

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestListAssistanceAwardsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListAssistanceAwardsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"fain", "cfda_number", "page", "ordering"},
		},
		{
			name: "every filter",
			opts: &ListAssistanceAwardsOptions{
				ListOptions:               ListOptions{Limit: 25, Cursor: "abc", Shape: "key,title", Flat: true, FlatLists: true},
				Joiner:                    "__",
				Key:                       "ASST_NON_1_075|ASST_NON_2_075",
				FAIN:                      "2303OHAIPP",
				URI:                       "AGG-1",
				FundingOpportunityNumber:  "USDA-FS-2024-CWDG-CWSF",
				UEI:                       "CBD4TX1Y41D5",
				Recipient:                 "ohio department",
				AwardingAgency:            "HHS",
				FundingAgency:             "7590",
				CFDANumber:                "93.778|93.767",
				AssistanceType:            "F001",
				RecordType:                "2",
				BusinessFundsIndicator:    "NON",
				DisasterEmergencyFundCode: "L",
				AwardDate:                 "2024-09-11",
				AwardDateGte:              "2024-01-01",
				AwardDateLte:              "2024-12-31",
				PopStartDateGte:           "2022-10-01",
				PopStartDateLte:           "2023-09-30",
				PopEndDateGte:             "2026-01-01",
				PopEndDateLte:             "2026-12-31",
				FiscalYear:                "2024",
				FiscalYearGte:             "2023",
				FiscalYearLte:             "2025",
				PopState:                  "OH|VA",
				PopCountry:                "USA",
				PopZip:                    "43215",
				ObligatedGte:              "1000",
				ObligatedLte:              "500000",
				Search:                    "wildfire resilience",
				Ordering:                  "-pragmatic_obligations",
			},
			wantQS: map[string]string{
				"limit":                        "25",
				"cursor":                       "abc",
				"shape":                        "key,title",
				"flat":                         "true",
				"flat_lists":                   "true",
				"joiner":                       "__",
				"key":                          "ASST_NON_1_075|ASST_NON_2_075",
				"fain":                         "2303OHAIPP",
				"uri":                          "AGG-1",
				"funding_opportunity_number":   "USDA-FS-2024-CWDG-CWSF",
				"uei":                          "CBD4TX1Y41D5",
				"recipient":                    "ohio department",
				"awarding_agency":              "HHS",
				"funding_agency":               "7590",
				"cfda_number":                  "93.778|93.767",
				"assistance_type":              "F001",
				"record_type":                  "2",
				"business_funds_indicator":     "NON",
				"disaster_emergency_fund_code": "L",
				"award_date":                   "2024-09-11",
				"award_date_gte":               "2024-01-01",
				"award_date_lte":               "2024-12-31",
				"pop_start_date_gte":           "2022-10-01",
				"pop_start_date_lte":           "2023-09-30",
				"pop_end_date_gte":             "2026-01-01",
				"pop_end_date_lte":             "2026-12-31",
				"fiscal_year":                  "2024",
				"fiscal_year_gte":              "2023",
				"fiscal_year_lte":              "2025",
				"pop_state":                    "OH|VA",
				"pop_country":                  "USA",
				"pop_zip":                      "43215",
				"obligated_gte":                "1000",
				"obligated_lte":                "500000",
				"search":                       "wildfire resilience",
				"ordering":                     "-pragmatic_obligations",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListAssistanceAwardsOptions{},
			notInQS: []string{"key", "fain", "uri", "uei", "recipient", "record_type", "obligated_gte", "joiner", "search"},
		},
		{
			name:   "extra map",
			opts:   &ListAssistanceAwardsOptions{Extra: map[string]any{"exact": true}},
			wantQS: map[string]string{"exact": "true"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListAssistanceGrants(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestAssistanceAwardRoutes(t *testing.T) {
	ctx := context.Background()
	const key = "ASST NON/1"
	const escaped = "ASST%20NON%2F1"
	cases := []struct {
		name     string
		wantPath string
		call     func(c *Client) error
	}{
		{"ListAssistanceGrants", "/api/assistance/grants/", func(c *Client) error { _, err := c.ListAssistanceGrants(ctx, nil); return err }},
		{"ListAssistanceLoans", "/api/assistance/loans/", func(c *Client) error { _, err := c.ListAssistanceLoans(ctx, nil); return err }},
		{"ListAssistanceDirectPayments", "/api/assistance/direct_payments/", func(c *Client) error { _, err := c.ListAssistanceDirectPayments(ctx, nil); return err }},
		{"ListAssistanceOther", "/api/assistance/other/", func(c *Client) error { _, err := c.ListAssistanceOther(ctx, nil); return err }},
		{"GetAssistanceGrant", "/api/assistance/grants/" + escaped + "/", func(c *Client) error { _, err := c.GetAssistanceGrant(ctx, key, nil); return err }},
		{"GetAssistanceLoan", "/api/assistance/loans/" + escaped + "/", func(c *Client) error { _, err := c.GetAssistanceLoan(ctx, key, nil); return err }},
		{"GetAssistanceDirectPayment", "/api/assistance/direct_payments/" + escaped + "/", func(c *Client) error { _, err := c.GetAssistanceDirectPayment(ctx, key, nil); return err }},
		{"GetAssistanceOther", "/api/assistance/other/" + escaped + "/", func(c *Client) error { _, err := c.GetAssistanceOther(ctx, key, nil); return err }},
		{"ListAssistanceGrantTransactions", "/api/assistance/grants/" + escaped + "/transactions/", func(c *Client) error { _, err := c.ListAssistanceGrantTransactions(ctx, key, nil); return err }},
		{"ListAssistanceLoanTransactions", "/api/assistance/loans/" + escaped + "/transactions/", func(c *Client) error { _, err := c.ListAssistanceLoanTransactions(ctx, key, nil); return err }},
		{"ListAssistanceDirectPaymentTransactions", "/api/assistance/direct_payments/" + escaped + "/transactions/", func(c *Client) error {
			_, err := c.ListAssistanceDirectPaymentTransactions(ctx, key, nil)
			return err
		}},
		{"ListAssistanceOtherTransactions", "/api/assistance/other/" + escaped + "/transactions/", func(c *Client) error { _, err := c.ListAssistanceOtherTransactions(ctx, key, nil); return err }},
		{"ListAssistanceGrantListings", "/api/assistance/grants/" + escaped + "/listings/", func(c *Client) error { _, err := c.ListAssistanceGrantListings(ctx, key, nil); return err }},
		{"ListAssistanceLoanListings", "/api/assistance/loans/" + escaped + "/listings/", func(c *Client) error { _, err := c.ListAssistanceLoanListings(ctx, key, nil); return err }},
		{"ListAssistanceDirectPaymentListings", "/api/assistance/direct_payments/" + escaped + "/listings/", func(c *Client) error { _, err := c.ListAssistanceDirectPaymentListings(ctx, key, nil); return err }},
		{"ListAssistanceOtherListings", "/api/assistance/other/" + escaped + "/listings/", func(c *Client) error { _, err := c.ListAssistanceOtherListings(ctx, key, nil); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath string
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.EscapedPath()
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"count":0,"results":[]}`))
			})
			if err := tc.call(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if gotPath != tc.wantPath {
				t.Errorf("path: want %q, got %q", tc.wantPath, gotPath)
			}
		})
	}
}

func TestAssistanceAwardKeyRequired(t *testing.T) {
	ctx := context.Background()
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	calls := map[string]func() error{
		"GetAssistanceGrant":                      func() error { _, err := c.GetAssistanceGrant(ctx, "", nil); return err },
		"GetAssistanceLoan":                       func() error { _, err := c.GetAssistanceLoan(ctx, "", nil); return err },
		"GetAssistanceDirectPayment":              func() error { _, err := c.GetAssistanceDirectPayment(ctx, "", nil); return err },
		"GetAssistanceOther":                      func() error { _, err := c.GetAssistanceOther(ctx, "", nil); return err },
		"ListAssistanceGrantTransactions":         func() error { _, err := c.ListAssistanceGrantTransactions(ctx, "", nil); return err },
		"ListAssistanceLoanTransactions":          func() error { _, err := c.ListAssistanceLoanTransactions(ctx, "", nil); return err },
		"ListAssistanceDirectPaymentTransactions": func() error { _, err := c.ListAssistanceDirectPaymentTransactions(ctx, "", nil); return err },
		"ListAssistanceOtherTransactions":         func() error { _, err := c.ListAssistanceOtherTransactions(ctx, "", nil); return err },
		"ListAssistanceGrantListings":             func() error { _, err := c.ListAssistanceGrantListings(ctx, "", nil); return err },
		"ListAssistanceLoanListings":              func() error { _, err := c.ListAssistanceLoanListings(ctx, "", nil); return err },
		"ListAssistanceDirectPaymentListings":     func() error { _, err := c.ListAssistanceDirectPaymentListings(ctx, "", nil); return err },
		"ListAssistanceOtherListings":             func() error { _, err := c.ListAssistanceOtherListings(ctx, "", nil); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			var ve *ValidationError
			if err := call(); !errors.As(err, &ve) {
				t.Fatalf("expected *ValidationError, got %T: %v", err, err)
			}
		})
	}
}

const assistanceLoanDefaultShapeJSON = `{
	"assistance_listing":{"number":"59.012","title":"7(a) Loan Guarantees"},
	"assistance_type":{"code":"F004","description":"Loan Guarantee"},
	"award_date":"2026-10-15",
	"awarding_office":{"organization_id":"7d23b0cc-0292-51fc-ac50-6ba0f26d0f2d","office_code":"737010","office_name":"OFC OF CAPITAL ACCESS","agency_code":"7300","agency_name":"SMALL BUSINESS ADMINISTRATION","department_code":"073","department_name":"SMALL BUSINESS ADMINISTRATION"},
	"cfda_number":"59.012",
	"face_value_of_loan":425000.0,
	"fain":"8606059100",
	"fiscal_year":2027,
	"funding_office":{"organization_id":"7d23b0cc-0292-51fc-ac50-6ba0f26d0f2d","office_code":"737010","office_name":"OFC OF CAPITAL ACCESS","agency_code":"7300","agency_name":"SMALL BUSINESS ADMINISTRATION","department_code":"073","department_name":"SMALL BUSINESS ADMINISTRATION"},
	"key":"ASST_NON_8606059100_073",
	"loan_subsidy_cost":0.0,
	"obligated":0.0,
	"period_of_performance":{"start_date":null,"current_end_date":null},
	"place_of_performance":{"city_name":"","code":"GA**067","country":"USA","county_fips":"13067","county_name":"COBB","foreign_location":null,"scope":"COUNTY-WIDE","state_code":"GA","state_fips":"13","state_name":"GEORGIA","zip_code":"","zip_plus_four":""},
	"pragmatic_obligations":0.0,
	"recipient":{"uuid":"536c50dc-3456-5ff2-979d-0f898b8e85c9","uei":null,"duns":null,"legal_business_name":"SLEDGE HAMMER CONSULTING CORP.","legal_business_name_raw":"SLEDGE HAMMER CONSULTING CORP.","county":"COBB","congressional_district":"GA-11","congressional_district_original":"GA-11","highest_owner":{"uei":"PARENT12345","duns":null,"legal_business_name":"HAMMER HOLDINGS","legal_business_name_raw":"Hammer Holdings"},"business_types":[{"code":"R","description":"Small Business"}]},
	"record_type":"2",
	"title":"SBA 7(a) Program Loans",
	"uri":null
}`

func TestGetAssistanceLoanDecodesDefaultShape(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(assistanceLoanDefaultShapeJSON))
	})
	got, err := c.GetAssistanceLoan(context.Background(), "ASST_NON_8606059100_073", nil)
	if err != nil {
		t.Fatalf("GetAssistanceLoan: %v", err)
	}
	assertStrPtr(t, "key", got.Key, "ASST_NON_8606059100_073")
	assertStrPtr(t, "fain", got.FAIN, "8606059100")
	if got.URI != nil {
		t.Errorf("uri: want nil for a JSON null, got %q", *got.URI)
	}
	assertStrPtr(t, "award_date", got.AwardDate, "2026-10-15")
	if got.FiscalYear == nil || *got.FiscalYear != 2027 {
		t.Errorf("fiscal_year: want 2027, got %v", got.FiscalYear)
	}
	assertStrPtr(t, "assistance_type.code", got.AssistanceType.Code, "F004")
	assertStrPtr(t, "assistance_type.description", got.AssistanceType.Description, "Loan Guarantee")
	assertStrPtr(t, "record_type", got.RecordType, "2")
	assertStrPtr(t, "cfda_number", got.CFDANumber, "59.012")
	assertStrPtr(t, "assistance_listing.title", got.AssistanceListing.Title, "7(a) Loan Guarantees")
	assertStrPtr(t, "title", got.Title, "SBA 7(a) Program Loans")
	if got.FaceValueOfLoan == nil || *got.FaceValueOfLoan != 425000 {
		t.Errorf("face_value_of_loan: want 425000, got %v", got.FaceValueOfLoan)
	}
	if got.Obligated == nil || *got.Obligated != 0 {
		t.Errorf("obligated: want a present 0, got %v", got.Obligated)
	}
	if got.PragmaticObligations == nil || got.LoanSubsidyCost == nil {
		t.Errorf("pragmatic_obligations / loan_subsidy_cost: want present, got %v / %v", got.PragmaticObligations, got.LoanSubsidyCost)
	}
	assertStrPtr(t, "awarding_office.office_code", got.AwardingOffice.OfficeCode, "737010")
	assertStrPtr(t, "funding_office.department_code", got.FundingOffice.DepartmentCode, "073")
	assertStrPtr(t, "recipient.legal_business_name", got.Recipient.LegalBusinessName, "SLEDGE HAMMER CONSULTING CORP.")
	if got.Recipient.UEI != nil {
		t.Errorf("recipient.uei: want nil, got %q", *got.Recipient.UEI)
	}
	assertStrPtr(t, "recipient.highest_owner.uei", got.Recipient.HighestOwner.UEI, "PARENT12345")
	if len(got.Recipient.BusinessTypes) != 1 {
		t.Fatalf("recipient.business_types: want 1, got %d", len(got.Recipient.BusinessTypes))
	}
	assertStrPtr(t, "recipient.business_types[0].description", got.Recipient.BusinessTypes[0].Description, "Small Business")
	assertStrPtr(t, "place_of_performance.code", got.PlaceOfPerformance.Code, "GA**067")
	assertStrPtr(t, "place_of_performance.scope", got.PlaceOfPerformance.Scope, "COUNTY-WIDE")
	if got.PeriodOfPerformance == nil || got.PeriodOfPerformance.StartDate != nil {
		t.Errorf("period_of_performance: want present with a nil start_date, got %+v", got.PeriodOfPerformance)
	}
	if len(got.Extra) != 0 {
		t.Errorf("default shape should leave Extra empty, got %v", got.Extra)
	}
}

func TestGetAssistanceGrantPassesShapeAndKeepsExtra(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"key":"ASST_NON_2303OHAIPP_075","assistance_type":"F001","summary":"Adoption incentives","listings":[{"cfda_number":"93.603","is_primary":true}]}`))
	})
	got, err := c.GetAssistanceGrant(context.Background(), "ASST_NON_2303OHAIPP_075", &GetEntityOptions{
		Shape: "key,assistance_type,summary,listings(cfda_number,is_primary)",
		Flat:  true,
	})
	if err != nil {
		t.Fatalf("GetAssistanceGrant: %v", err)
	}
	assertQueryContains(t, capturedURL, map[string]string{
		"shape": "key,assistance_type,summary,listings(cfda_number,is_primary)",
		"flat":  "true",
	}, nil)
	assertStrPtr(t, "assistance_type.code", got.AssistanceType.Code, "F001")
	if got.AssistanceType.Description != nil {
		t.Errorf("assistance_type.description: want nil from a bare code, got %q", *got.AssistanceType.Description)
	}
	if got.Extra["summary"] != "Adoption incentives" {
		t.Errorf("Extra[summary]: got %#v", got.Extra["summary"])
	}
	if _, ok := got.Extra["listings"]; !ok {
		t.Errorf("Extra[listings]: missing from %v", got.Extra)
	}
}

func TestAssistanceAwardRecordObligationsByFiscalYearFromExtra(t *testing.T) {
	var rec AssistanceAwardRecord
	raw := `{"key":"ASST_NON_2303OHAIPP_075","obligations_by_fiscal_year":[{"fiscal_year":2023,"obligated":1500.25},{"fiscal_year":2024,"obligated":40000.0}]}`
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	years, err := ObligationsByFiscalYear(rec.Extra)
	if err != nil {
		t.Fatalf("ObligationsByFiscalYear: %v", err)
	}
	if len(years) != 2 || years[1].FiscalYear == nil || *years[1].FiscalYear != 2024 || years[0].Obligated == nil || *years[0].Obligated != 1500.25 {
		t.Errorf("years: want 2023=1500.25 and 2024, got %+v", years)
	}
}

func TestListAssistanceGrantTransactionsDecodesTypedRows(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":1,"next":null,"previous":null,"results":[{
			"key":"7590_2303OHAIPP_2303OHAIPP-73532728_93.603_00","modification_number":"00","transaction_date":"2024-09-11",
			"action_type":"A1","assistance_type_code":"03","correction_delete_indicator":null,"cfda_number":"93.603",
			"obligated":40000.0,"pragmatic_obligations":40000.0,"non_federal_funding_amount":0.0,"indirect_cost_federal_share_amount":0.0,
			"face_value_of_loan":0.0,"loan_subsidy_cost":0.0,"award_amount_basis":null,"award_recipient_basis":null,
			"action_date_fiscal_year":2024,"last_modified_date":"2024-09-20T20:44:06Z","initial_report_date":"2024-09-20T20:44:06Z"
		}]}`))
	})
	page, err := c.ListAssistanceGrantTransactions(context.Background(), "ASST_NON_2303OHAIPP_075", &AssistanceSubresourceOptions{Page: 2, Limit: 50})
	if err != nil {
		t.Fatalf("ListAssistanceGrantTransactions: %v", err)
	}
	u, _ := url.Parse(capturedURL)
	if got := u.Query(); len(got) != 2 || got.Get("page") != "2" || got.Get("limit") != "50" {
		t.Errorf("query: want only page=2&limit=50, got %q", u.RawQuery)
	}
	if page.Count != 1 || len(page.Results) != 1 {
		t.Fatalf("want 1 result, got count=%d len=%d", page.Count, len(page.Results))
	}
	tx := page.Results[0]
	assertStrPtr(t, "action_type", tx.ActionType, "A1")
	assertStrPtr(t, "cfda_number", tx.CFDANumber, "93.603")
	if tx.Obligated == nil || *tx.Obligated != 40000 {
		t.Errorf("obligated: want 40000, got %v", tx.Obligated)
	}
	if tx.ActionDateFiscalYear == nil || *tx.ActionDateFiscalYear != 2024 {
		t.Errorf("action_date_fiscal_year: want 2024, got %v", tx.ActionDateFiscalYear)
	}
	if tx.CorrectionDeleteIndicator != nil {
		t.Errorf("correction_delete_indicator: want nil, got %q", *tx.CorrectionDeleteIndicator)
	}
}

func TestListAssistanceLoanListingsDecodesTypedRows(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"count":2,"next":null,"previous":null,"results":[
			{"cfda_number":"59.012","is_primary":true,"obligated":0.0,"pragmatic_obligations":1200.5,"face_value_of_loan":425000.0,"loan_subsidy_cost":1200.5,"transactions_count":3,"first_action_date":"2026-01-02","last_action_date":"2026-10-15"},
			{"cfda_number":"59.008","is_primary":false,"transactions_count":1}
		]}`))
	})
	page, err := c.ListAssistanceLoanListings(context.Background(), "ASST_NON_8606059100_073", nil)
	if err != nil {
		t.Fatalf("ListAssistanceLoanListings: %v", err)
	}
	u, _ := url.Parse(capturedURL)
	if u.RawQuery != "" {
		t.Errorf("nil opts should send no query, got %q", u.RawQuery)
	}
	if len(page.Results) != 2 {
		t.Fatalf("want 2 results, got %d", len(page.Results))
	}
	primary, other := page.Results[0], page.Results[1]
	if primary.IsPrimary == nil || !*primary.IsPrimary {
		t.Errorf("is_primary: want true, got %v", primary.IsPrimary)
	}
	if primary.LoanSubsidyCost == nil || *primary.LoanSubsidyCost != 1200.5 {
		t.Errorf("loan_subsidy_cost: want 1200.5, got %v", primary.LoanSubsidyCost)
	}
	assertStrPtr(t, "last_action_date", primary.LastActionDate, "2026-10-15")
	if other.IsPrimary == nil || *other.IsPrimary {
		t.Errorf("is_primary: want a present false, got %v", other.IsPrimary)
	}
	if other.Obligated != nil {
		t.Errorf("obligated: want nil when absent, got %v", *other.Obligated)
	}
}

func TestIterateAssistanceDirectPaymentsFollowsCursor(t *testing.T) {
	var queries []url.Values
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("cursor") == "" {
			w.Write([]byte(`{"count":2,"next":"` + srv.URL + `/api/assistance/direct_payments/?cursor=NEXT&record_type=2","results":[{"key":"A"}]}`))
			return
		}
		w.Write([]byte(`{"count":2,"next":null,"results":[{"key":"B"}]}`))
	}))
	t.Cleanup(srv.Close)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))

	var keys []string
	for row, err := range c.IterateAssistanceDirectPayments(context.Background(), &ListAssistanceAwardsOptions{RecordType: "2"}).Seq() {
		if err != nil {
			t.Fatalf("iterate: %v", err)
		}
		keys = append(keys, row["key"].(string))
	}
	if len(keys) != 2 || keys[0] != "A" || keys[1] != "B" {
		t.Fatalf("keys: want [A B], got %v", keys)
	}
	if len(queries) != 2 {
		t.Fatalf("requests: want 2, got %d", len(queries))
	}
	second := queries[1]
	if second.Get("cursor") != "NEXT" || second.Get("record_type") != "2" || second.Has("page") {
		t.Errorf("second request: want cursor=NEXT&record_type=2 and no page, got %v", second)
	}
}

func TestIterateAssistanceAwardsNilOpts(t *testing.T) {
	ctx := context.Background()
	srv, calls := makeSinglePageServer(t)
	c := NewClient(WithAPIKey("k"), WithBaseURL(srv.URL), WithRetries(0))
	iterators := []*Iterator[Record]{
		c.IterateAssistanceGrants(ctx, nil),
		c.IterateAssistanceLoans(ctx, nil),
		c.IterateAssistanceDirectPayments(ctx, nil),
		c.IterateAssistanceOther(ctx, nil),
	}
	for _, it := range iterators {
		for it.Next() {
			_ = it.Item()
		}
		if err := it.Err(); err != nil {
			t.Fatalf("iterate: %v", err)
		}
	}
	if *calls != len(iterators) {
		t.Errorf("server calls: want %d, got %d", len(iterators), *calls)
	}
}

func TestAssistanceTypeRefUnmarshal(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		wantCode string
		wantDesc string
	}{
		{"expand object", `{"code":"F002","description":"Cooperative Agreement"}`, "F002", "Cooperative Agreement"},
		{"bare leaf code", `"F006"`, "F006", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ref AssistanceTypeRef
			if err := json.Unmarshal([]byte(tc.raw), &ref); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			assertStrPtr(t, "code", ref.Code, tc.wantCode)
			if tc.wantDesc == "" {
				if ref.Description != nil {
					t.Errorf("description: want nil, got %q", *ref.Description)
				}
				return
			}
			assertStrPtr(t, "description", ref.Description, tc.wantDesc)
		})
	}
}
