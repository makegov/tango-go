package tango

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
)

// Values the API serves in a budget account's account_category today.
// The field is an open string, so treat any other value as a category this SDK does not know yet rather than an error.
const (
	// BudgetAccountCategoryBudgetary is an ordinary budgetary account.
	BudgetAccountCategoryBudgetary = "budgetary"
	// BudgetAccountCategoryCreditFinancing is a credit financing account, which carries no enacted budget authority (enacted_ba is null).
	BudgetAccountCategoryCreditFinancing = "credit_financing"
)

// Codes the API serves in a source_anomalies element today.
// The code is an open string, so new codes can appear without an SDK release.
const (
	// BudgetAnomalyContractExceedsObligations: reported contract obligations exceed the account's total obligations, and the served contract_obligated is capped.
	BudgetAnomalyContractExceedsObligations = "contract_exceeds_obligations"
	// BudgetAnomalyAssistanceExceedsObligations: reported assistance obligations exceed the account's total obligations; the value is flagged, not changed.
	BudgetAnomalyAssistanceExceedsObligations = "assistance_exceeds_obligations"
	// BudgetAnomalyContractWithoutObligations: contract obligations are reported against an account with no positive total obligations; the value is flagged, not changed.
	BudgetAnomalyContractWithoutObligations = "contract_without_obligations"
)

// BudgetSourceAnomaly is one element of a budget account's source_anomalies list: a problem found in the source data behind that account-year.
//
// The API serves source_anomalies in the default shape, as [] when the row is clean; decode it from a row with BudgetSourceAnomalies.
// Every field is optional, so each is a pointer (or slice) and unknown keys land in Extra.
type BudgetSourceAnomaly struct {
	// Code names the problem, such as BudgetAnomalyContractExceedsObligations. It is an open string.
	Code *string `json:"code,omitempty"`
	// Field is the served field the anomaly is about (for example "contract_obligated").
	Field *string `json:"field,omitempty"`
	// BoundField is the field Field was checked against (for example "obligated_total").
	BoundField *string `json:"bound_field,omitempty"`
	// Action is "capped" when the served value was changed to respect BoundField, or "flagged" when it was left as reported.
	Action *string `json:"action,omitempty"`
	// ReportedValue is the value the source reported; ServedValue is what the API serves in Field.
	ReportedValue *float64 `json:"reported_value,omitempty"`
	ServedValue   *float64 `json:"served_value,omitempty"`
	// LikelyCause is a short machine-readable guess at why the source disagrees with itself.
	LikelyCause *string `json:"likely_cause,omitempty"`
	// AffectedFields lists the served fields whose values the anomaly bears on.
	AffectedFields []string `json:"affected_fields,omitempty"`
	// Message is a human-readable description of the anomaly.
	Message *string `json:"message,omitempty"`
	// Source points at the source data behind the anomaly.
	Source *BudgetSourceAnomalySource `json:"source,omitempty"`

	// Extra captures fields the server adds that aren't in the typed surface yet.
	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *BudgetSourceAnomaly) UnmarshalJSON(data []byte) error {
	type alias BudgetSourceAnomaly
	return unmarshalWithExtra(data, (*alias)(a))
}

// BudgetSourceAnomalySource identifies the dataset behind a BudgetSourceAnomaly and, when available, the rows that produced it.
type BudgetSourceAnomalySource struct {
	Dataset    *string                  `json:"dataset,omitempty"`
	FiscalYear *int                     `json:"fiscal_year,omitempty"`
	Rows       []BudgetSourceAnomalyRow `json:"rows,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (s *BudgetSourceAnomalySource) UnmarshalJSON(data []byte) error {
	type alias BudgetSourceAnomalySource
	return unmarshalWithExtra(data, (*alias)(s))
}

// BudgetSourceAnomalyRow is one source row behind a BudgetSourceAnomaly.
type BudgetSourceAnomalyRow struct {
	FiscalPeriod               *int     `json:"fiscal_period,omitempty"`
	PIID                       *string  `json:"piid,omitempty"`
	ParentPIID                 *string  `json:"parent_piid,omitempty"`
	TAS                        *string  `json:"tas,omitempty"`
	ReportingAgencyID          *string  `json:"reporting_agency_id,omitempty"`
	TransactionObligatedAmount *float64 `json:"transaction_obligated_amount,omitempty"`
	FileCSource                *string  `json:"file_c_source,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (r *BudgetSourceAnomalyRow) UnmarshalJSON(data []byte) error {
	type alias BudgetSourceAnomalyRow
	return unmarshalWithExtra(data, (*alias)(r))
}

// BudgetSourceAnomalies decodes the source_anomalies list from a budget-account row returned by ListBudgetAccounts, IterateBudgetAccounts or GetBudgetAccount.
//
// It returns nil when the row has no source_anomalies key (a shape that did not name it) or the value is null, and an empty slice when the row is clean.
func BudgetSourceAnomalies(row Record) ([]BudgetSourceAnomaly, error) {
	v, ok := row["source_anomalies"]
	if !ok || v == nil {
		return nil, nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	out := []BudgetSourceAnomaly{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BudgetDataThroughPeriod reads a budget account row's data_through_period: the File A fiscal period (1-12) the account-year's figures run through.
// Below 12 the fiscal year is partial.
// ok is false when the row has no File A data (null), the shape did not name the field, or the value is not a whole number.
func BudgetDataThroughPeriod(row Record) (period int, ok bool) {
	switch v := row["data_through_period"].(type) {
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return int(n), true
	case float64:
		if v != float64(int(v)) {
			return 0, false
		}
		return int(v), true
	case int:
		return v, true
	}
	return 0, false
}

// ListBudgetAccountsOptions filters /api/budget/accounts/ — one row per federal account and fiscal year, covering the budget lifecycle from request through outlay.
//
// The record is wide and shape-driven; ShapeBudgetAccountsMinimal is a compact starting point.
// The server's default shape includes data_through_period, account_category and source_anomalies; read them with BudgetDataThroughPeriod and BudgetSourceAnomalies.
// The default ordering is newest fiscal year first, then largest enacted_ba, with null enacted_ba (credit financing accounts) last.
// Every dollar and ratio metric also takes exact, __gte and __lte filters, and the taxonomy fields take __in; reach those through Extra.
type ListBudgetAccountsOptions struct {
	ListOptions

	// FederalAccountSymbol is an exact filter (e.g. "097-0100").
	FederalAccountSymbol string
	// FiscalYear is an exact filter.
	FiscalYear string
	// FiscalYearGte is the inclusive lower bound for fiscal_year.
	FiscalYearGte string
	// FiscalYearLte is the inclusive upper bound for fiscal_year.
	FiscalYearLte string
	// DataThroughPeriod is an exact filter on the File A fiscal period (1-12) the account-year's figures run through.
	DataThroughPeriod string
	// DataThroughPeriodGte is the inclusive lower bound for data_through_period.
	DataThroughPeriodGte string
	// DataThroughPeriodLte is the inclusive upper bound for data_through_period.
	DataThroughPeriodLte string
	// DataThroughPeriodIsNull, when set, keeps only rows with no File A data (true) or only rows with it (false).
	DataThroughPeriodIsNull *bool
	// AgencyCode is the owning agency's code (exact).
	AgencyCode string
	// BEACategory is the Budget Enforcement Act category, such as discretionary or mandatory (exact).
	BEACategory string
	// OnOffBudget is the on/off-budget flag (exact).
	OnOffBudget string
	// BureauName is an exact match on the bureau name.
	BureauName string
	// AccountTitleContains is a case-insensitive substring match on the account title.
	AccountTitleContains string
	// SubfunctionCode is the budget subfunction code (exact).
	SubfunctionCode string
	// AccountCategory is the account category (exact), such as BudgetAccountCategoryBudgetary or BudgetAccountCategoryCreditFinancing.
	AccountCategory string
	// AccountCategoryIn matches any of the listed account categories (account_category__in).
	AccountCategoryIn []string
	// Agency is a Tango organization: a name, abbreviation, code or organization key (for example "EPA"). It matches the whole organization subtree, so a department includes its sub-agencies; join several with "|".
	Agency string
	// Search is a free-text search filter.
	Search string
	// Ordering is the server-side sort spec; prefix "-" for descending.
	Ordering string

	// Extra carries filter keys without a typed field, such as the metric range filters (for example "enacted_ba__gte").
	Extra map[string]any
}

func (o *ListBudgetAccountsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "federal_account_symbol", o.FederalAccountSymbol)
	setIfNotEmpty(q, "fiscal_year", o.FiscalYear)
	setIfNotEmpty(q, "fiscal_year__gte", o.FiscalYearGte)
	setIfNotEmpty(q, "fiscal_year__lte", o.FiscalYearLte)
	setIfNotEmpty(q, "data_through_period", o.DataThroughPeriod)
	setIfNotEmpty(q, "data_through_period__gte", o.DataThroughPeriodGte)
	setIfNotEmpty(q, "data_through_period__lte", o.DataThroughPeriodLte)
	setIfNotNilBool(q, "data_through_period__isnull", o.DataThroughPeriodIsNull)
	setIfNotEmpty(q, "agency_code", o.AgencyCode)
	setIfNotEmpty(q, "bea_category", o.BEACategory)
	setIfNotEmpty(q, "on_off_budget", o.OnOffBudget)
	setIfNotEmpty(q, "bureau_name", o.BureauName)
	setIfNotEmpty(q, "account_title__icontains", o.AccountTitleContains)
	setIfNotEmpty(q, "subfunction_code", o.SubfunctionCode)
	setIfNotEmpty(q, "account_category", o.AccountCategory)
	setIfNotEmpty(q, "account_category__in", strings.Join(o.AccountCategoryIn, ","))
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListBudgetAccounts queries /api/budget/accounts/.
func (c *Client) ListBudgetAccounts(ctx context.Context, opts *ListBudgetAccountsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/budget/accounts/", q)
}

// IterateBudgetAccounts walks every budget-account row matching opts.
func (c *Client) IterateBudgetAccounts(ctx context.Context, opts *ListBudgetAccountsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListBudgetAccountsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListBudgetAccounts(ctx, &next)
		},
	}
}

// GetBudgetAccount fetches one account-year by its numeric id (/api/budget/accounts/{id}/).
func (c *Client) GetBudgetAccount(ctx context.Context, id string, opts *GetEntityOptions) (Record, error) {
	if id == "" {
		return nil, &ValidationError{&APIError{Message: "budget account id is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/budget/accounts/"+pathEscape(id)+"/", opts.toQuery())
}

// BudgetAccountQuartersOptions controls GetBudgetAccountQuarters.
//
// The route takes page, limit and one narrowing filter; its rows are fixed rather than shape-driven.
type BudgetAccountQuartersOptions struct {
	Page  int
	Limit int

	// TAS narrows the rows to a single Treasury Account Symbol.
	// Omit it to get every TAS that rolls up under the federal account.
	TAS string
}

func (o *BudgetAccountQuartersOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	setIfNonZeroInt(q, "page", o.Page)
	setIfNonZeroInt(q, "limit", o.Limit)
	setIfNotEmpty(q, "tas", o.TAS)
	return q
}

// GetBudgetAccountQuarters lists the quarterly TAS-grain obligation and outlay flow for one account-year (/api/budget/accounts/{id}/quarters/).
//
// It returns one row per (TAS, quarter), because a federal account usually rolls up several TAS.
// Coverage starts at FY2021; an earlier account-year returns an empty page.
func (c *Client) GetBudgetAccountQuarters(ctx context.Context, id string, opts *BudgetAccountQuartersOptions) (*PaginatedResponse[Record], error) {
	if id == "" {
		return nil, &ValidationError{&APIError{Message: "budget account id is required"}}
	}
	return listGeneric[Record](ctx, c, "/api/budget/accounts/"+pathEscape(id)+"/quarters/", opts.toQuery())
}

// BudgetAccountRecipientsOptions controls GetBudgetAccountRecipients.
//
// The route takes page, limit and one narrowing filter; its rows are fixed rather than shape-driven.
type BudgetAccountRecipientsOptions struct {
	Page  int
	Limit int

	// FundingOrganizationID narrows the rows to one funding office (an organization UUID, resolvable through GetOrganization).
	FundingOrganizationID string
}

func (o *BudgetAccountRecipientsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	setIfNonZeroInt(q, "page", o.Page)
	setIfNonZeroInt(q, "limit", o.Limit)
	setIfNotEmpty(q, "funding_organization_id", o.FundingOrganizationID)
	return q
}

// GetBudgetAccountRecipients lists the funding-office by recipient contract flows for one account-year (/api/budget/accounts/{id}/recipients/), largest contract obligation first.
//
// Contract flows only.
// Each row carries resolved funding_office and recipient objects and a capped contracts list; a row that hits the cap sets contracts_truncated and carries its full piids array.
func (c *Client) GetBudgetAccountRecipients(ctx context.Context, id string, opts *BudgetAccountRecipientsOptions) (*PaginatedResponse[Record], error) {
	if id == "" {
		return nil, &ValidationError{&APIError{Message: "budget account id is required"}}
	}
	return listGeneric[Record](ctx, c, "/api/budget/accounts/"+pathEscape(id)+"/recipients/", opts.toQuery())
}
