package tango

import (
	"context"
	"net/url"
)

// ListBudgetAccountsOptions filters /api/budget/accounts/ — federal-account x
// fiscal-year budget rollups. Embed ListOptions for pagination + shape
// control. The BudgetAccount schema is wide (~63 fields) and shape-driven;
// use ShapeBudgetAccountsMinimal for a compact default. The full __gte / __lte
// numeric-range filter set is reachable via Extra.
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
	// AgencyCode is the awarding/funding agency CGAC code (exact).
	AgencyCode string
	// BEACategory is the Bureau of Economic Analysis category (exact).
	BEACategory string
	// OnOffBudget is the on/off-budget flag (exact).
	OnOffBudget string
	// Search is a free-text search filter.
	Search string
	// Ordering is the server-side sort spec; prefix "-" for descending.
	Ordering string

	// Extra is an escape hatch for filter keys not yet first-classed on this
	// struct (e.g. the *_gte / *_lte range filters on the numeric metrics).
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

// IterateBudgetAccounts returns an Iterator that walks every budget-account
// rollup matching opts. The iterator follows ?page= or ?cursor= on the
// server's next URL automatically.
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

// GetBudgetAccount fetches a single budget-account rollup by its id
// (/api/budget/accounts/{id}/).
func (c *Client) GetBudgetAccount(ctx context.Context, id string, opts *ListOptions) (Record, error) {
	if id == "" {
		return nil, &ValidationError{&APIError{Message: "budget account id is required"}}
	}
	q := url.Values{}
	opts.applyTo(q)
	return getGeneric[Record](ctx, c, "/api/budget/accounts/"+pathEscape(id)+"/", q)
}

// GetBudgetAccountQuarters fetches the quarterly lifecycle detail for a single
// account-year (/api/budget/accounts/{id}/quarters/).
func (c *Client) GetBudgetAccountQuarters(ctx context.Context, id string, opts *ListOptions) (*PaginatedResponse[Record], error) {
	if id == "" {
		return nil, &ValidationError{&APIError{Message: "budget account id is required"}}
	}
	q := url.Values{}
	opts.applyTo(q)
	return listGeneric[Record](ctx, c, "/api/budget/accounts/"+pathEscape(id)+"/quarters/", q)
}

// GetBudgetAccountRecipients fetches the funding-office x recipient contract-
// flow detail for a single account-year (/api/budget/accounts/{id}/recipients/).
// The response envelope carries extra keys (federal_account_symbol,
// fiscal_year) alongside the standard pagination fields.
func (c *Client) GetBudgetAccountRecipients(ctx context.Context, id string, opts *ListOptions) (*PaginatedResponse[Record], error) {
	if id == "" {
		return nil, &ValidationError{&APIError{Message: "budget account id is required"}}
	}
	q := url.Values{}
	opts.applyTo(q)
	return listGeneric[Record](ctx, c, "/api/budget/accounts/"+pathEscape(id)+"/recipients/", q)
}
