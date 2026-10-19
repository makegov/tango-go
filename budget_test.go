package tango

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

const budgetAccountWithAnomaly = `{"count":1,"next":null,"previous":null,"results":[{
	"federal_account_symbol":"020-4159","fiscal_year":2024,"account_category":"budgetary",
	"source_anomalies":[{
		"code":"contract_exceeds_obligations","field":"contract_obligated","bound_field":"obligated_total",
		"action":"capped","reported_value":73470455401.8,"served_value":4748670061.77,
		"likely_cause":"single_row_outlier","affected_fields":["contract_obligated","contract_share_of_obligated"],
		"message":"Contract obligations exceed total obligations.","severity":"high",
		"source":{"dataset":"usaspending_file_c","fiscal_year":2024,"rows":[{
			"fiscal_period":12,"piid":"P-1","parent_piid":null,"tas":"020-X-4159-000",
			"reporting_agency_id":"020","transaction_obligated_amount":68721785340.03,"file_c_source":"award"
		}]}
	}]
},{"federal_account_symbol":"012-4391","fiscal_year":2024,"account_category":"credit_financing","enacted_ba":null,"source_anomalies":[]}]}`

func TestListBudgetAccountsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListBudgetAccountsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts produces empty query",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"federal_account_symbol", "fiscal_year", "search"},
		},
		{
			name: "all typed filters",
			opts: &ListBudgetAccountsOptions{
				FederalAccountSymbol: "097-0100",
				FiscalYear:           "2024",
				FiscalYearGte:        "2020",
				FiscalYearLte:        "2025",
				AgencyCode:           "9700",
				BEACategory:          "discretionary",
				OnOffBudget:          "on",
				BureauName:           "Army",
				AccountTitleContains: "operation",
				SubfunctionCode:      "051",
				AccountCategory:      "budgetary",
				Search:               "operations",
				Ordering:             "-enacted_ba",
			},
			wantQS: map[string]string{
				"federal_account_symbol":   "097-0100",
				"fiscal_year":              "2024",
				"fiscal_year__gte":         "2020",
				"fiscal_year__lte":         "2025",
				"agency_code":              "9700",
				"bea_category":             "discretionary",
				"on_off_budget":            "on",
				"bureau_name":              "Army",
				"account_title__icontains": "operation",
				"subfunction_code":         "051",
				"account_category":         "budgetary",
				"search":                   "operations",
				"ordering":                 "-enacted_ba",
			},
		},
		{
			name: "pagination and shape",
			opts: &ListBudgetAccountsOptions{
				ListOptions: ListOptions{Page: 2, Limit: 50, Shape: ShapeBudgetAccountsMinimal},
			},
			wantQS: map[string]string{
				"page":  "2",
				"limit": "50",
				"shape": ShapeBudgetAccountsMinimal,
			},
		},
		{
			name:    "account category in joins with commas",
			opts:    &ListBudgetAccountsOptions{AccountCategoryIn: []string{BudgetAccountCategoryBudgetary, BudgetAccountCategoryCreditFinancing}},
			wantQS:  map[string]string{"account_category__in": "budgetary,credit_financing"},
			notInQS: []string{"account_category"},
		},
		{
			name: "data through period filters",
			opts: &ListBudgetAccountsOptions{
				DataThroughPeriod:       "9",
				DataThroughPeriodGte:    "3",
				DataThroughPeriodLte:    "11",
				DataThroughPeriodIsNull: boolPtr(false),
			},
			wantQS: map[string]string{
				"data_through_period":         "9",
				"data_through_period__gte":    "3",
				"data_through_period__lte":    "11",
				"data_through_period__isnull": "false",
			},
		},
		{
			name:   "extra forwards range filters",
			opts:   &ListBudgetAccountsOptions{Extra: map[string]any{"enacted_ba__gte": "1000000"}},
			wantQS: map[string]string{"enacted_ba__gte": "1000000"},
		},
		{
			name:    "zero values omitted",
			opts:    &ListBudgetAccountsOptions{},
			notInQS: []string{"federal_account_symbol", "fiscal_year", "agency_code", "account_category", "account_category__in", "ordering"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListBudgetAccounts(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestListBudgetAccountsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListBudgetAccounts(context.Background(), nil)
	assertPathContains(t, capturedURL, "/api/budget/accounts/")
}

func TestIterateBudgetAccountsBuildsIterator(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(emptyListHandler))
	it := c.IterateBudgetAccounts(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}

func TestGetBudgetAccountRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetBudgetAccount(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetBudgetAccountBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetBudgetAccount(context.Background(), "097-0100-2024", nil)
	assertPathContains(t, capturedURL, "/api/budget/accounts/097-0100-2024/")
}

func TestGetBudgetAccountQuartersRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetBudgetAccountQuarters(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetBudgetAccountQuartersBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.GetBudgetAccountQuarters(context.Background(), "acct-1", nil)
	assertPathContains(t, capturedURL, "/api/budget/accounts/acct-1/quarters/")
}

func TestGetBudgetAccountRecipientsRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetBudgetAccountRecipients(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetBudgetAccountRecipientsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.GetBudgetAccountRecipients(context.Background(), "acct-1", nil)
	assertPathContains(t, capturedURL, "/api/budget/accounts/acct-1/recipients/")
}

func TestGetBudgetAccountQuartersSendsOnlyItsOwnParams(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, err := c.GetBudgetAccountQuarters(context.Background(), "42", &BudgetAccountQuartersOptions{Page: 2, Limit: 50, TAS: "097-2020/2021-0100"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertPathContains(t, capturedURL, "/api/budget/accounts/42/quarters/")
	assertQueryContains(t, capturedURL, map[string]string{"page": "2", "limit": "50", "tas": "097-2020/2021-0100"}, []string{"shape", "flat", "cursor"})
}

func TestGetBudgetAccountRecipientsSendsOnlyItsOwnParams(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, err := c.GetBudgetAccountRecipients(context.Background(), "42", &BudgetAccountRecipientsOptions{Limit: 10, FundingOrganizationID: "0b6d3c1e-0000-4000-8000-000000000000"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertPathContains(t, capturedURL, "/api/budget/accounts/42/recipients/")
	assertQueryContains(t, capturedURL, map[string]string{"limit": "10", "funding_organization_id": "0b6d3c1e-0000-4000-8000-000000000000"}, []string{"page", "shape", "tas"})
}

func TestGetBudgetAccountForwardsShape(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetBudgetAccount(context.Background(), "42", &GetEntityOptions{Shape: "*,appendix(*)"})
	assertQueryContains(t, capturedURL, map[string]string{"shape": "*,appendix(*)"}, []string{"page", "limit"})
}

func TestBudgetSourceAnomaliesDecodesListRows(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(budgetAccountWithAnomaly))
	}))
	page, err := c.ListBudgetAccounts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListBudgetAccounts: %v", err)
	}
	if got := page.Results[1]["account_category"]; got != BudgetAccountCategoryCreditFinancing {
		t.Errorf("account_category = %v, want credit_financing", got)
	}

	anomalies, err := BudgetSourceAnomalies(page.Results[0])
	if err != nil {
		t.Fatalf("BudgetSourceAnomalies: %v", err)
	}
	if len(anomalies) != 1 {
		t.Fatalf("got %d anomalies, want 1", len(anomalies))
	}
	a := anomalies[0]
	if *a.Code != BudgetAnomalyContractExceedsObligations || *a.Action != "capped" || *a.BoundField != "obligated_total" {
		t.Errorf("code/action/bound_field = %q/%q/%q", *a.Code, *a.Action, *a.BoundField)
	}
	if *a.ReportedValue != 73470455401.8 || *a.ServedValue != 4748670061.77 {
		t.Errorf("reported/served = %v/%v", *a.ReportedValue, *a.ServedValue)
	}
	if len(a.AffectedFields) != 2 || a.AffectedFields[1] != "contract_share_of_obligated" {
		t.Errorf("affected_fields = %v", a.AffectedFields)
	}
	if a.Extra["severity"] != "high" {
		t.Errorf("unknown key not kept in Extra: %v", a.Extra)
	}
	if a.Source == nil || *a.Source.Dataset != "usaspending_file_c" || *a.Source.FiscalYear != 2024 || len(a.Source.Rows) != 1 {
		t.Fatalf("source = %+v", a.Source)
	}
	row := a.Source.Rows[0]
	if *row.FiscalPeriod != 12 || *row.PIID != "P-1" || row.ParentPIID != nil || *row.ReportingAgencyID != "020" || *row.TransactionObligatedAmount != 68721785340.03 {
		t.Errorf("source row = %+v", row)
	}

	clean, err := BudgetSourceAnomalies(page.Results[1])
	if err != nil {
		t.Fatalf("BudgetSourceAnomalies(clean): %v", err)
	}
	if clean == nil || len(clean) != 0 {
		t.Errorf("clean row = %#v, want empty non-nil slice", clean)
	}
}

func TestBudgetDataThroughPeriod(t *testing.T) {
	cases := map[string]struct {
		row    Record
		want   int
		wantOK bool
	}{
		"json number": {Record{"data_through_period": json.Number("9")}, 9, true},
		"float":       {Record{"data_through_period": float64(12)}, 12, true},
		"null":        {Record{"data_through_period": nil}, 0, false},
		"absent":      {Record{"fiscal_year": 2026}, 0, false},
		"fractional":  {Record{"data_through_period": 9.5}, 0, false},
	}
	for name, tc := range cases {
		got, ok := BudgetDataThroughPeriod(tc.row)
		if got != tc.want || ok != tc.wantOK {
			t.Errorf("%s: got (%d, %v), want (%d, %v)", name, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestShapeBudgetAccountsMinimalPutsDataThroughPeriodAfterFiscalYear(t *testing.T) {
	if !strings.Contains(ShapeBudgetAccountsMinimal, "fiscal_year,data_through_period,") {
		t.Errorf("ShapeBudgetAccountsMinimal = %q, want data_through_period right after fiscal_year", ShapeBudgetAccountsMinimal)
	}
}

func TestBudgetSourceAnomaliesAbsentOrNull(t *testing.T) {
	for name, row := range map[string]Record{
		"absent": {"federal_account_symbol": "097-0100"},
		"null":   {"source_anomalies": nil},
	} {
		got, err := BudgetSourceAnomalies(row)
		if err != nil || got != nil {
			t.Errorf("%s: got %#v, %v; want nil, nil", name, got, err)
		}
	}
}

func TestBudgetSourceAnomaliesRejectsNonList(t *testing.T) {
	if _, err := BudgetSourceAnomalies(Record{"source_anomalies": "oops"}); err == nil {
		t.Error("expected an error for a non-list source_anomalies")
	}
}

func TestListBudgetAccountsForwardsAgency(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListBudgetAccounts(context.Background(), &ListBudgetAccountsOptions{Agency: "EPA|DOT"})
	assertPathContains(t, capturedURL, "/api/budget/accounts/")
	assertQueryContains(t, capturedURL, map[string]string{"agency": "EPA|DOT"}, []string{"agency_code"})
}
