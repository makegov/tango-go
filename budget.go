package tango

import (
	"context"
	"net/url"
)

// ListBudgetAccountsOptions filters /api/budget/accounts/ — one row per federal account and fiscal year, covering the budget lifecycle from request through outlay.
//
// The record is wide and shape-driven; ShapeBudgetAccountsMinimal is a compact starting point.
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
	setIfNotEmpty(q, "agency_code", o.AgencyCode)
	setIfNotEmpty(q, "bea_category", o.BEACategory)
	setIfNotEmpty(q, "on_off_budget", o.OnOffBudget)
	setIfNotEmpty(q, "bureau_name", o.BureauName)
	setIfNotEmpty(q, "account_title__icontains", o.AccountTitleContains)
	setIfNotEmpty(q, "subfunction_code", o.SubfunctionCode)
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
