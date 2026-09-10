package tango

import (
	"context"
	"net/url"
)

// State, local and education (SLED) procurement — /api/sled/opportunities/ and
// /api/sled/forecasts/. Solicitations that never appear on SAM.gov because they
// were never federal.
//
// Beta: coverage is partial and grows one jurisdiction at a time. There is no
// national SLED feed, so a thin per-state result is at least as likely to be a
// portal Tango does not read as a quiet market — GetSledCoverage answers that.
//
// This data does not join to the federal data. There is no UEI, no PIID, no
// agency-hierarchy key and no NAICS/PSC crosswalk; the organization expand here
// is three strings, not the federal 7-key office payload.

// ListSledOpportunitiesOptions filters /api/sled/opportunities/. Mirrors the
// Python `list_sled_opportunities` kwargs and the Node
// `ListSledOpportunitiesOptions` interface.
//
// Two things behave unlike the federal endpoints.
//
// Leaving both Status and Active unset returns open solicitations only — the API
// defaults the list to status=open, because only about a fifth of the corpus is
// open and portals drop a closed solicitation rather than restating it. Set
// Status explicitly to page the whole corpus. GetSledOpportunity returns a
// solicitation whatever its status.
//
// Status is Tango-derived liveness, refreshed every fifteen minutes. The
// portal's own word is served as source_status, is frozen at last capture, and
// is not filterable — most of what it calls open already has a passed deadline.
type ListSledOpportunitiesOptions struct {
	ListOptions

	// State is a two-letter state or territory code. Multi-value: "TX|OK".
	State string

	// Jurisdiction is the level of government: "state", "local", "education",
	// or "unknown" for the aggregator rows that cannot tell state from local.
	Jurisdiction string

	// Status is "open", "closed", "awarded", "cancelled" or "unknown". With
	// Active also unset the API returns open only; "unknown" is hidden by that
	// default and reachable with "open|unknown".
	Status string

	// Active is sugar for federal-shaped callers: true is status=open, false is
	// its complement (so it includes "unknown"). A pointer so that false is a
	// real filter value rather than an absent one.
	Active *bool

	// Agency is a substring match on the buyer's published text (min 2
	// characters). No code resolution behind it — state agencies have no entry
	// in the federal organization tree.
	Agency string

	// SolicitationNumber is the number a human would quote. Null on roughly a
	// third of the corpus, where the portal publishes none.
	SolicitationNumber string

	// SolicitationType is "rfp", "ifb", "rfq", "rfi", "itb", "sole_source",
	// "grant" or "other". "null" (the portal states no type) is a distinct
	// answer from "other" (a type the vocabulary does not recognize).
	SolicitationType string

	// HasDocuments filters on whether the solicitation advertises at least one
	// document. A pointer so false is distinguishable from unset.
	HasDocuments *bool

	// RevisionKind matches the kind of the most recent substantive revision:
	// "deadline_change", "status_change", "documents_added",
	// "documents_removed", "documents_replaced", "title_change" or
	// "content_change".
	RevisionKind string

	// Naics matches within the "naics" category scheme only, and is thin on
	// purpose: scheme tagging is mid-migration, so only a small share of
	// entries are tagged NAICS. Prefer CategoryCode unless you need scheme
	// precision.
	Naics string

	// Nigp matches within the "nigp" scheme, the most widely tagged of the four.
	Nigp string

	// Unspsc matches within the "unspsc" scheme.
	Unspsc string

	// Category matches within the "text" scheme, where the code is the portal's
	// own human label.
	Category string

	// CategoryCode matches a code under ANY scheme, including the untagged
	// pre-migration strings. The escape hatch when a scheme-specific filter
	// returns less than you expected.
	CategoryCode string

	// Date bounds (wire format: YYYY-MM-DD).
	PostedAfter            string
	PostedBefore           string
	ResponseDeadlineAfter  string
	ResponseDeadlineBefore string

	// FirstSeenAfter / FirstSeenBefore bound when Tango FIRST OBSERVED the
	// solicitation. FirstSeenAfter is the polling primitive: everything new
	// since your last call.
	FirstSeenAfter  string
	FirstSeenBefore string

	// ChangeSeenAfter bounds when Tango OBSERVED the last substantive change.
	// A scrape date, not an amendment date: a state portal restates one page in
	// place and publishes no amendment date, so the resolution is that state's
	// crawl cadence. There is no ChangeSeenBefore — the API exposes only the
	// lower bound.
	ChangeSeenAfter string

	ModifiedAfter  string
	ModifiedBefore string

	// Platform, NativeID and ExternalID are support filters for reproducing a
	// specific record with us. Platform identifies the source portal's platform
	// family. None appears in a response shape and none is a stable value.
	// ExternalID accepts at most 500 values.
	Platform   string
	NativeID   string
	ExternalID string

	// Search is a ranked full-text query over title, agency, identifiers,
	// category labels and description, widened by the solicitations whose
	// ATTACHMENT text matched (min 2 characters). A row that matched on its
	// description gains a "snippet" field; a title-or-agency match carries none.
	Search string

	// Ordering is one of rank, response_deadline, posted_date, first_seen_at,
	// last_seen_at, last_change_seen_at, modified. "rank" requires a non-empty
	// Search.
	Ordering string

	Extra map[string]any
}

func (o *ListSledOpportunitiesOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "state", o.State)
	setIfNotEmpty(q, "jurisdiction", o.Jurisdiction)
	setIfNotEmpty(q, "status", o.Status)
	setIfNotNilBool(q, "active", o.Active)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "solicitation_number", o.SolicitationNumber)
	setIfNotEmpty(q, "solicitation_type", o.SolicitationType)
	setIfNotNilBool(q, "has_documents", o.HasDocuments)
	setIfNotEmpty(q, "revision_kind", o.RevisionKind)
	setIfNotEmpty(q, "naics", o.Naics)
	setIfNotEmpty(q, "nigp", o.Nigp)
	setIfNotEmpty(q, "unspsc", o.Unspsc)
	setIfNotEmpty(q, "category", o.Category)
	setIfNotEmpty(q, "category_code", o.CategoryCode)
	setIfNotEmpty(q, "posted_after", o.PostedAfter)
	setIfNotEmpty(q, "posted_before", o.PostedBefore)
	setIfNotEmpty(q, "response_deadline_after", o.ResponseDeadlineAfter)
	setIfNotEmpty(q, "response_deadline_before", o.ResponseDeadlineBefore)
	setIfNotEmpty(q, "first_seen_after", o.FirstSeenAfter)
	setIfNotEmpty(q, "first_seen_before", o.FirstSeenBefore)
	setIfNotEmpty(q, "change_seen_after", o.ChangeSeenAfter)
	setIfNotEmpty(q, "modified_after", o.ModifiedAfter)
	setIfNotEmpty(q, "modified_before", o.ModifiedBefore)
	setIfNotEmpty(q, "platform", o.Platform)
	setIfNotEmpty(q, "native_id", o.NativeID)
	setIfNotEmpty(q, "external_id", o.ExternalID)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListSledOpportunitiesRevisionOptions filters the nested
// /api/sled/opportunities/{opportunity_id}/revisions/ route.
type ListSledOpportunityRevisionsOptions struct {
	ListOptions

	// Kind is a revision kind plus "enrichment" — which the revisions(*) expand
	// excludes and this route serves. Multi-value: use "|".
	Kind string

	// SourceDeclared filters on whether a portal's own amendment marker moved at
	// this emission. True on about 5% of revisions; everything else is Tango
	// inferring the change from the diff.
	SourceDeclared *bool

	ObservedAfter  string
	ObservedBefore string

	Extra map[string]any
}

func (o *ListSledOpportunityRevisionsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "kind", o.Kind)
	setIfNotNilBool(q, "source_declared", o.SourceDeclared)
	setIfNotEmpty(q, "observed_after", o.ObservedAfter)
	setIfNotEmpty(q, "observed_before", o.ObservedBefore)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListSledForecastsOptions filters /api/sled/forecasts/.
//
// Forecasts carry no liveness at all — there is no deadline to have passed, so
// there is no Status field, no Active field, and no open-only default. Currency
// is the caller's call from estimated_advertisement_date, which is the START of
// the published quarter rather than a posting date.
type ListSledForecastsOptions struct {
	ListOptions

	// State is a two-letter state code. Multi-value: use "|".
	State string

	// Agency is a substring match on the buyer's published text (min 2
	// characters).
	Agency string

	// ProcurementCategory and ProcurementMethod are the portal's own words,
	// passed through verbatim.
	ProcurementCategory string
	ProcurementMethod   string

	// ContractNumber is the number the state expects to award under, where it
	// publishes one in advance.
	ContractNumber string

	// IncumbentName is a substring match on the incumbent vendor's name as
	// published. NOT resolved to a Tango entity.
	IncumbentName string

	// AdvertisementAfter / AdvertisementBefore bound the estimated advertisement
	// date (YYYY-MM-DD). Remember the date is a QUARTER START, not a posting
	// date.
	AdvertisementAfter  string
	AdvertisementBefore string

	FirstSeenAfter  string
	FirstSeenBefore string
	ModifiedAfter   string
	ModifiedBefore  string

	// Search is a ranked full-text query over title, agency and description
	// (min 2 characters).
	Search string

	// Ordering is one of rank, estimated_advertisement_date, first_seen_at,
	// last_seen_at, modified. "rank" requires a non-empty Search.
	Ordering string

	Extra map[string]any
}

func (o *ListSledForecastsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "state", o.State)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "procurement_category", o.ProcurementCategory)
	setIfNotEmpty(q, "procurement_method", o.ProcurementMethod)
	setIfNotEmpty(q, "contract_number", o.ContractNumber)
	setIfNotEmpty(q, "incumbent_name", o.IncumbentName)
	setIfNotEmpty(q, "advertisement_after", o.AdvertisementAfter)
	setIfNotEmpty(q, "advertisement_before", o.AdvertisementBefore)
	setIfNotEmpty(q, "first_seen_after", o.FirstSeenAfter)
	setIfNotEmpty(q, "first_seen_before", o.FirstSeenBefore)
	setIfNotEmpty(q, "modified_after", o.ModifiedAfter)
	setIfNotEmpty(q, "modified_before", o.ModifiedBefore)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// sledShapeQuery builds the shape/flat query for a SLED detail fetch.
func sledShapeQuery(opts *GetEntityOptions) url.Values {
	q := url.Values{}
	if opts == nil {
		return q
	}
	setIfNotEmpty(q, "shape", opts.Shape)
	if opts.Flat {
		q.Set("flat", "true")
	}
	if opts.FlatLists {
		q.Set("flat_lists", "true")
	}
	return q
}

// ListSledOpportunities queries /api/sled/opportunities/ — state, local and
// education solicitations.
//
// Leaving both Status and Active unset returns open solicitations only. That
// default is the API's, and this method deliberately does not synthesize one:
// sending status=open here would make Active=false unreachable.
func (c *Client) ListSledOpportunities(ctx context.Context, opts *ListSledOpportunitiesOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/sled/opportunities/", q)
}

// IterateSledOpportunities walks every solicitation matching opts.
func (c *Client) IterateSledOpportunities(ctx context.Context, opts *ListSledOpportunitiesOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListSledOpportunitiesOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListSledOpportunities(ctx, &next)
		},
	}
}

// GetSledOpportunity fetches one solicitation by opportunity_id, whatever its
// status. The open-only default applies to the list endpoint, not here.
//
// A document's body is available as the attachments(extracted_text) leaf on a
// Small plan or above, from API version 4.25.1. It must be NAMED — no suggested
// shape includes it and attachments(*) does not carry it — and its key is ABSENT
// rather than null whenever the text is not being served: below Small (where it
// is withheld and named in meta.upgrade_hints), on a contested document, or
// where it could not be resolved. A contested document never returns text at any
// plan, because its stored bytes disagree with what the record advertised.
//
// Searching document text and reading it are separate. Search matches inside
// attachment text on every plan and returns no fragment of it; the body is a
// per-record read.
func (c *Client) GetSledOpportunity(ctx context.Context, opportunityID string, opts *GetEntityOptions) (Record, error) {
	if opportunityID == "" {
		return nil, &ValidationError{&APIError{Message: "sled opportunity_id is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/sled/opportunities/"+pathEscape(opportunityID)+"/", sledShapeQuery(opts))
}

// ListSledOpportunityRevisions queries one solicitation's observed revision
// history.
//
// observed_at is the scrape that saw the change, not the date the agency made
// it: no state portal emits amendment notices, so kind is Tango's inference from
// the diff on about 95% of revisions, resolution is that state's crawl cadence,
// and history starts when Tango began reading the jurisdiction rather than when
// the solicitation was posted.
//
// Unlike the revisions(*) expand, this route serves "enrichment" rows — Tango's
// own detail fetch filling in coverage rather than an agency amendment. Pass
// Kind: "enrichment" for only those. The per-field before/after ("changes")
// requires a Small plan; "changed_fields" names what moved at every plan.
func (c *Client) ListSledOpportunityRevisions(ctx context.Context, opportunityID string, opts *ListSledOpportunityRevisionsOptions) (*PaginatedResponse[Record], error) {
	if opportunityID == "" {
		return nil, &ValidationError{&APIError{Message: "sled opportunity_id is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/sled/opportunities/"+pathEscape(opportunityID)+"/revisions/", q)
}

// GetSledCoverage fetches the per-state coverage rollup.
//
// Returns corpus totals plus one row per jurisdiction — the total, the count in
// each of the five statuses, the jurisdiction levels present, and when a
// solicitation there last changed. It answers one question: whether a thin
// result for a state is a thin market or a portal Tango does not read.
//
// Every state row carries all five status buckets whether or not they have rows,
// so a total and two buckets never invite subtraction. Takes no parameters and
// is neither shaped nor paginated.
func (c *Client) GetSledCoverage(ctx context.Context) (Record, error) {
	return getGeneric[Record](ctx, c, "/api/sled/opportunities/coverage/", url.Values{})
}

// ListSledForecasts queries /api/sled/forecasts/ — planned state procurements.
func (c *Client) ListSledForecasts(ctx context.Context, opts *ListSledForecastsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/sled/forecasts/", q)
}

// IterateSledForecasts walks every forecast matching opts.
func (c *Client) IterateSledForecasts(ctx context.Context, opts *ListSledForecastsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListSledForecastsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListSledForecasts(ctx, &next)
		},
	}
}

// GetSledForecast fetches one forecast by forecast_id.
func (c *Client) GetSledForecast(ctx context.Context, forecastID string, opts *GetEntityOptions) (Record, error) {
	if forecastID == "" {
		return nil, &ValidationError{&APIError{Message: "sled forecast_id is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/sled/forecasts/"+pathEscape(forecastID)+"/", sledShapeQuery(opts))
}
