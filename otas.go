package tango

import (
	"context"
	"net/url"
)

// ListOTAsOptions filters /api/otas/ (Other Transaction Authority award
// actions). Mirror of the Node `ListOTAsOptions` interface and the Python
// `list_otas` kwargs.
//
// All date / fiscal-year fields are wire-format strings (see the SDK
// design note: dates stay strings to match the API). Empty fields are
// omitted from the request.
type ListOTAsOptions struct {
	ListOptions

	// Joiner is the separator used when Flat is true (e.g. "." or "__").
	// Empty means "use the server default".
	Joiner string

	// Identifiers + agencies
	AwardingAgency string
	FundingAgency  string
	PIID           string
	Recipient      string
	UEI            string

	// Dollar / fiscal-year bounds
	FiscalYear    string
	FiscalYearGte string
	FiscalYearLte string

	// Date bounds
	AwardDate       string
	AwardDateGte    string
	AwardDateLte    string
	ExpiringGte     string
	ExpiringLte     string
	PopStartDateGte string
	PopStartDateLte string
	PopEndDateGte   string
	PopEndDateLte   string

	// Classification
	PSC string

	// Search + ordering
	Search   string
	Ordering string

	// Extra is an escape hatch for filter keys not first-classed here.
	Extra map[string]any
}

func (o *ListOTAsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "joiner", o.Joiner)
	setIfNotEmpty(q, "awarding_agency", o.AwardingAgency)
	setIfNotEmpty(q, "funding_agency", o.FundingAgency)
	setIfNotEmpty(q, "piid", o.PIID)
	setIfNotEmpty(q, "recipient", o.Recipient)
	setIfNotEmpty(q, "uei", o.UEI)
	setIfNotEmpty(q, "fiscal_year", o.FiscalYear)
	setIfNotEmpty(q, "fiscal_year_gte", o.FiscalYearGte)
	setIfNotEmpty(q, "fiscal_year_lte", o.FiscalYearLte)
	setIfNotEmpty(q, "award_date", o.AwardDate)
	setIfNotEmpty(q, "award_date_gte", o.AwardDateGte)
	setIfNotEmpty(q, "award_date_lte", o.AwardDateLte)
	setIfNotEmpty(q, "expiring_gte", o.ExpiringGte)
	setIfNotEmpty(q, "expiring_lte", o.ExpiringLte)
	setIfNotEmpty(q, "pop_start_date_gte", o.PopStartDateGte)
	setIfNotEmpty(q, "pop_start_date_lte", o.PopStartDateLte)
	setIfNotEmpty(q, "pop_end_date_gte", o.PopEndDateGte)
	setIfNotEmpty(q, "pop_end_date_lte", o.PopEndDateLte)
	setIfNotEmpty(q, "psc", o.PSC)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListOTAs queries /api/otas/ — OTA (Other Transaction Authority) award
// actions.
func (c *Client) ListOTAs(ctx context.Context, opts *ListOTAsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/otas/", q)
}

// IterateOTAs walks every OTA matching opts.
func (c *Client) IterateOTAs(ctx context.Context, opts *ListOTAsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListOTAsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListOTAs(ctx, &next)
		},
	}
}

// GetOTA fetches a single OTA by key.
func (c *Client) GetOTA(ctx context.Context, key string, opts *GetEntityOptions) (Record, error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "OTA key is required"}}
	}
	q := url.Values{}
	if opts != nil {
		setIfNotEmpty(q, "shape", opts.Shape)
		if opts.Flat {
			q.Set("flat", "true")
		}
		if opts.FlatLists {
			q.Set("flat_lists", "true")
		}
	}
	return getGeneric[Record](ctx, c, "/api/otas/"+pathEscape(key)+"/", q)
}

// ListOTIDVsOptions filters /api/otidvs/ (Other Transaction IDV parents).
// Same filter set as ListOTAsOptions per the sibling SDKs.
type ListOTIDVsOptions struct {
	ListOptions

	Joiner string

	AwardingAgency string
	FundingAgency  string
	PIID           string
	Recipient      string
	UEI            string

	FiscalYear    string
	FiscalYearGte string
	FiscalYearLte string

	AwardDate       string
	AwardDateGte    string
	AwardDateLte    string
	ExpiringGte     string
	ExpiringLte     string
	PopStartDateGte string
	PopStartDateLte string
	PopEndDateGte   string
	PopEndDateLte   string

	PSC string

	Search   string
	Ordering string

	Extra map[string]any
}

func (o *ListOTIDVsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "joiner", o.Joiner)
	setIfNotEmpty(q, "awarding_agency", o.AwardingAgency)
	setIfNotEmpty(q, "funding_agency", o.FundingAgency)
	setIfNotEmpty(q, "piid", o.PIID)
	setIfNotEmpty(q, "recipient", o.Recipient)
	setIfNotEmpty(q, "uei", o.UEI)
	setIfNotEmpty(q, "fiscal_year", o.FiscalYear)
	setIfNotEmpty(q, "fiscal_year_gte", o.FiscalYearGte)
	setIfNotEmpty(q, "fiscal_year_lte", o.FiscalYearLte)
	setIfNotEmpty(q, "award_date", o.AwardDate)
	setIfNotEmpty(q, "award_date_gte", o.AwardDateGte)
	setIfNotEmpty(q, "award_date_lte", o.AwardDateLte)
	setIfNotEmpty(q, "expiring_gte", o.ExpiringGte)
	setIfNotEmpty(q, "expiring_lte", o.ExpiringLte)
	setIfNotEmpty(q, "pop_start_date_gte", o.PopStartDateGte)
	setIfNotEmpty(q, "pop_start_date_lte", o.PopStartDateLte)
	setIfNotEmpty(q, "pop_end_date_gte", o.PopEndDateGte)
	setIfNotEmpty(q, "pop_end_date_lte", o.PopEndDateLte)
	setIfNotEmpty(q, "psc", o.PSC)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListOTIDVs queries /api/otidvs/ — OTIDV parent records.
func (c *Client) ListOTIDVs(ctx context.Context, opts *ListOTIDVsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/otidvs/", q)
}

// IterateOTIDVs walks every OTIDV matching opts.
func (c *Client) IterateOTIDVs(ctx context.Context, opts *ListOTIDVsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListOTIDVsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListOTIDVs(ctx, &next)
		},
	}
}

// GetOTIDV fetches a single OTIDV by key.
func (c *Client) GetOTIDV(ctx context.Context, key string, opts *GetEntityOptions) (Record, error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "OTIDV key is required"}}
	}
	q := url.Values{}
	if opts != nil {
		setIfNotEmpty(q, "shape", opts.Shape)
		if opts.Flat {
			q.Set("flat", "true")
		}
		if opts.FlatLists {
			q.Set("flat_lists", "true")
		}
	}
	return getGeneric[Record](ctx, c, "/api/otidvs/"+pathEscape(key)+"/", q)
}

// ListOTIDVAwardsOptions filters /api/otidvs/{key}/awards/. Same filter
// shape as ListOTAsOptions per the Node interface.
type ListOTIDVAwardsOptions struct {
	ListOptions

	Joiner string

	AwardingAgency string
	FundingAgency  string
	PIID           string
	Recipient      string
	UEI            string

	FiscalYear    string
	FiscalYearGte string
	FiscalYearLte string

	AwardDate       string
	AwardDateGte    string
	AwardDateLte    string
	ExpiringGte     string
	ExpiringLte     string
	PopStartDateGte string
	PopStartDateLte string
	PopEndDateGte   string
	PopEndDateLte   string

	PSC string

	Search   string
	Ordering string

	Extra map[string]any
}

func (o *ListOTIDVAwardsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "joiner", o.Joiner)
	setIfNotEmpty(q, "awarding_agency", o.AwardingAgency)
	setIfNotEmpty(q, "funding_agency", o.FundingAgency)
	setIfNotEmpty(q, "piid", o.PIID)
	setIfNotEmpty(q, "recipient", o.Recipient)
	setIfNotEmpty(q, "uei", o.UEI)
	setIfNotEmpty(q, "fiscal_year", o.FiscalYear)
	setIfNotEmpty(q, "fiscal_year_gte", o.FiscalYearGte)
	setIfNotEmpty(q, "fiscal_year_lte", o.FiscalYearLte)
	setIfNotEmpty(q, "award_date", o.AwardDate)
	setIfNotEmpty(q, "award_date_gte", o.AwardDateGte)
	setIfNotEmpty(q, "award_date_lte", o.AwardDateLte)
	setIfNotEmpty(q, "expiring_gte", o.ExpiringGte)
	setIfNotEmpty(q, "expiring_lte", o.ExpiringLte)
	setIfNotEmpty(q, "pop_start_date_gte", o.PopStartDateGte)
	setIfNotEmpty(q, "pop_start_date_lte", o.PopStartDateLte)
	setIfNotEmpty(q, "pop_end_date_gte", o.PopEndDateGte)
	setIfNotEmpty(q, "pop_end_date_lte", o.PopEndDateLte)
	setIfNotEmpty(q, "psc", o.PSC)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListOTIDVAwards queries /api/otidvs/{key}/awards/ — the child awards
// under a given OTIDV parent.
func (c *Client) ListOTIDVAwards(ctx context.Context, key string, opts *ListOTIDVAwardsOptions) (*PaginatedResponse[Record], error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "OTIDV key is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/otidvs/"+pathEscape(key)+"/awards/", q)
}

// IterateOTIDVAwards walks every child award under the given OTIDV.
func (c *Client) IterateOTIDVAwards(ctx context.Context, key string, opts *ListOTIDVAwardsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListOTIDVAwardsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListOTIDVAwards(ctx, key, &next)
		},
	}
}
