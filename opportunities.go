package tango

import (
	"context"
	"net/url"
)

// ListOpportunitiesOptions filters /api/opportunities/.
type ListOpportunitiesOptions struct {
	ListOptions

	Active                 *bool
	Agency                 string
	FirstNoticeDateAfter   string
	FirstNoticeDateBefore  string
	LastNoticeDateAfter    string
	LastNoticeDateBefore   string
	NAICS                  string
	NoticeType             string
	Ordering               string
	PlaceOfPerformance     string
	PSC                    string
	ResponseDeadlineAfter  string
	ResponseDeadlineBefore string
	Search                 string
	SetAside               string
	SolicitationNumber     string

	Extra map[string]any
}

func (o *ListOpportunitiesOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotNilBool(q, "active", o.Active)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "first_notice_date_after", o.FirstNoticeDateAfter)
	setIfNotEmpty(q, "first_notice_date_before", o.FirstNoticeDateBefore)
	setIfNotEmpty(q, "last_notice_date_after", o.LastNoticeDateAfter)
	setIfNotEmpty(q, "last_notice_date_before", o.LastNoticeDateBefore)
	setIfNotEmpty(q, "naics", o.NAICS)
	setIfNotEmpty(q, "notice_type", o.NoticeType)
	setIfNotEmpty(q, "ordering", o.Ordering)
	setIfNotEmpty(q, "place_of_performance", o.PlaceOfPerformance)
	setIfNotEmpty(q, "psc", o.PSC)
	setIfNotEmpty(q, "response_deadline_after", o.ResponseDeadlineAfter)
	setIfNotEmpty(q, "response_deadline_before", o.ResponseDeadlineBefore)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "set_aside", o.SetAside)
	setIfNotEmpty(q, "solicitation_number", o.SolicitationNumber)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListOpportunities queries /api/opportunities/.
func (c *Client) ListOpportunities(ctx context.Context, opts *ListOpportunitiesOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/opportunities/", q)
}

// IterateOpportunities walks every opportunity matching opts.
func (c *Client) IterateOpportunities(ctx context.Context, opts *ListOpportunitiesOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListOpportunitiesOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListOpportunities(ctx, &next)
		},
	}
}

// ListNoticesOptions filters /api/notices/. The notices viewset rejects
// every ?ordering= value, so it isn't exposed here (mirrors Python/Node).
type ListNoticesOptions struct {
	ListOptions

	Active                 *bool
	Agency                 string
	NAICS                  string
	NoticeType             string
	PostedDateAfter        string
	PostedDateBefore       string
	PSC                    string
	ResponseDeadlineAfter  string
	ResponseDeadlineBefore string
	Search                 string
	SetAside               string
	SolicitationNumber     string

	Extra map[string]any
}

func (o *ListNoticesOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotNilBool(q, "active", o.Active)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "naics", o.NAICS)
	setIfNotEmpty(q, "notice_type", o.NoticeType)
	setIfNotEmpty(q, "posted_date_after", o.PostedDateAfter)
	setIfNotEmpty(q, "posted_date_before", o.PostedDateBefore)
	setIfNotEmpty(q, "psc", o.PSC)
	setIfNotEmpty(q, "response_deadline_after", o.ResponseDeadlineAfter)
	setIfNotEmpty(q, "response_deadline_before", o.ResponseDeadlineBefore)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "set_aside", o.SetAside)
	setIfNotEmpty(q, "solicitation_number", o.SolicitationNumber)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListNotices queries /api/notices/.
func (c *Client) ListNotices(ctx context.Context, opts *ListNoticesOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/notices/", q)
}

// ListForecastsOptions filters /api/forecasts/.
type ListForecastsOptions struct {
	ListOptions

	Agency          string
	AwardDateAfter  string
	AwardDateBefore string
	FiscalYear      string
	FiscalYearGte   string
	FiscalYearLte   string
	ModifiedAfter   string
	ModifiedBefore  string
	NAICSCode       string
	NAICSStartsWith string
	Ordering        string
	Search          string
	SourceSystem    string
	Status          string

	Extra map[string]any
}

func (o *ListForecastsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "award_date_after", o.AwardDateAfter)
	setIfNotEmpty(q, "award_date_before", o.AwardDateBefore)
	setIfNotEmpty(q, "fiscal_year", o.FiscalYear)
	setIfNotEmpty(q, "fiscal_year_gte", o.FiscalYearGte)
	setIfNotEmpty(q, "fiscal_year_lte", o.FiscalYearLte)
	setIfNotEmpty(q, "modified_after", o.ModifiedAfter)
	setIfNotEmpty(q, "modified_before", o.ModifiedBefore)
	setIfNotEmpty(q, "naics_code", o.NAICSCode)
	setIfNotEmpty(q, "naics_starts_with", o.NAICSStartsWith)
	setIfNotEmpty(q, "ordering", o.Ordering)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "source_system", o.SourceSystem)
	setIfNotEmpty(q, "status", o.Status)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListForecasts queries /api/forecasts/.
func (c *Client) ListForecasts(ctx context.Context, opts *ListForecastsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/forecasts/", q)
}

// ListGrantsOptions filters /api/grants/.
type ListGrantsOptions struct {
	ListOptions

	Agency             string
	ApplicantTypes     string
	CFDANumber         string
	FundingCategories  string
	FundingInstruments string
	OpportunityNumber  string
	Ordering           string
	PostedDateAfter    string
	PostedDateBefore   string
	ResponseDateAfter  string
	ResponseDateBefore string
	Search             string
	Status             string

	Extra map[string]any
}

func (o *ListGrantsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "applicant_types", o.ApplicantTypes)
	setIfNotEmpty(q, "cfda_number", o.CFDANumber)
	setIfNotEmpty(q, "funding_categories", o.FundingCategories)
	setIfNotEmpty(q, "funding_instruments", o.FundingInstruments)
	setIfNotEmpty(q, "opportunity_number", o.OpportunityNumber)
	setIfNotEmpty(q, "ordering", o.Ordering)
	setIfNotEmpty(q, "posted_date_after", o.PostedDateAfter)
	setIfNotEmpty(q, "posted_date_before", o.PostedDateBefore)
	setIfNotEmpty(q, "response_date_after", o.ResponseDateAfter)
	setIfNotEmpty(q, "response_date_before", o.ResponseDateBefore)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "status", o.Status)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListGrants queries /api/grants/.
func (c *Client) ListGrants(ctx context.Context, opts *ListGrantsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/grants/", q)
}

// IterateNotices walks every notice matching opts.
func (c *Client) IterateNotices(ctx context.Context, opts *ListNoticesOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListNoticesOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListNotices(ctx, &next)
		},
	}
}

// IterateForecasts walks every forecast matching opts.
func (c *Client) IterateForecasts(ctx context.Context, opts *ListForecastsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListForecastsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListForecasts(ctx, &next)
		},
	}
}

// IterateGrants walks every grant matching opts.
func (c *Client) IterateGrants(ctx context.Context, opts *ListGrantsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListGrantsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListGrants(ctx, &next)
		},
	}
}
