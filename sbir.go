package tango

import (
	"context"
	"net/url"
)

// SBIR/STTR — /api/sbir/topics/ and /api/sbir/solicitations/.
//
// A topic is one research area a program office will fund; a solicitation is the program cycle (for example a DoD BAA) the topic is released under.

// ListSbirTopicsOptions filters /api/sbir/topics/.
type ListSbirTopicsOptions struct {
	ListOptions

	// Activity is "open", "closed" or "unknown" (no close date and no parent solicitation to inherit one from).
	Activity string

	// Agency is a partial, case-insensitive match on the raw agency text (for example "DOD" or "NIH"); it is not organization-resolved.
	Agency             string
	TopicNumber        string
	SolicitationNumber string
	Year               int
	DocSource          string

	// Date bounds (wire format: YYYY-MM-DD).
	CloseDateAfter    string
	CloseDateBefore   string
	OpenDateAfter     string
	OpenDateBefore    string
	ReleaseDateAfter  string
	ReleaseDateBefore string

	Search string
	// Ordering is one of activity, close_date, modified, open_date, release_date or year; prefix "-" for descending.
	Ordering string

	Extra map[string]any
}

func (o *ListSbirTopicsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "activity", o.Activity)
	setIfNotEmpty(q, "agency", o.Agency)
	setIfNotEmpty(q, "topic_number", o.TopicNumber)
	setIfNotEmpty(q, "solicitation_number", o.SolicitationNumber)
	setIfNonZeroInt(q, "year", o.Year)
	setIfNotEmpty(q, "doc_source", o.DocSource)
	setIfNotEmpty(q, "close_date_after", o.CloseDateAfter)
	setIfNotEmpty(q, "close_date_before", o.CloseDateBefore)
	setIfNotEmpty(q, "open_date_after", o.OpenDateAfter)
	setIfNotEmpty(q, "open_date_before", o.OpenDateBefore)
	setIfNotEmpty(q, "release_date_after", o.ReleaseDateAfter)
	setIfNotEmpty(q, "release_date_before", o.ReleaseDateBefore)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListSbirTopics queries /api/sbir/topics/.
func (c *Client) ListSbirTopics(ctx context.Context, opts *ListSbirTopicsOptions) (*PaginatedResponse[Record], error) {
	return listGeneric[Record](ctx, c, "/api/sbir/topics/", opts.toQuery())
}

// IterateSbirTopics walks every topic matching opts.
func (c *Client) IterateSbirTopics(ctx context.Context, opts *ListSbirTopicsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListSbirTopicsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListSbirTopics(ctx, &next)
		},
	}
}

// GetSbirTopic fetches one topic by its topic id (/api/sbir/topics/{topic_id}/).
func (c *Client) GetSbirTopic(ctx context.Context, topicID string, opts *GetEntityOptions) (Record, error) {
	if topicID == "" {
		return nil, &ValidationError{&APIError{Message: "SBIR topic id is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/sbir/topics/"+pathEscape(topicID)+"/", opts.toQuery())
}

// ListSbirSolicitationsOptions filters /api/sbir/solicitations/.
type ListSbirSolicitationsOptions struct {
	ListOptions

	// Activity is "open" or "closed", derived at query time from end_date.
	Activity string

	// Program is "SBIR" or "STTR".
	Program            string
	SolicitationNumber string
	CycleName          string
	// SolicitationStatus matches the raw feed status.
	SolicitationStatus string
	// OutOfCycle is a pointer so that false is a real filter value rather than an absent one.
	OutOfCycle *bool
	Year       int

	// Date bounds (wire format: YYYY-MM-DD).
	StartDateAfter  string
	StartDateBefore string
	EndDateAfter    string
	EndDateBefore   string

	Search string
	// Ordering is one of activity, end_date, modified, start_date or year; prefix "-" for descending.
	Ordering string

	Extra map[string]any
}

func (o *ListSbirSolicitationsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "activity", o.Activity)
	setIfNotEmpty(q, "program", o.Program)
	setIfNotEmpty(q, "solicitation_number", o.SolicitationNumber)
	setIfNotEmpty(q, "cycle_name", o.CycleName)
	setIfNotEmpty(q, "solicitation_status", o.SolicitationStatus)
	setIfNotNilBool(q, "out_of_cycle", o.OutOfCycle)
	setIfNonZeroInt(q, "year", o.Year)
	setIfNotEmpty(q, "start_date_after", o.StartDateAfter)
	setIfNotEmpty(q, "start_date_before", o.StartDateBefore)
	setIfNotEmpty(q, "end_date_after", o.EndDateAfter)
	setIfNotEmpty(q, "end_date_before", o.EndDateBefore)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListSbirSolicitations queries /api/sbir/solicitations/.
func (c *Client) ListSbirSolicitations(ctx context.Context, opts *ListSbirSolicitationsOptions) (*PaginatedResponse[Record], error) {
	return listGeneric[Record](ctx, c, "/api/sbir/solicitations/", opts.toQuery())
}

// IterateSbirSolicitations walks every solicitation matching opts.
func (c *Client) IterateSbirSolicitations(ctx context.Context, opts *ListSbirSolicitationsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListSbirSolicitationsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListSbirSolicitations(ctx, &next)
		},
	}
}

// GetSbirSolicitation fetches one solicitation by its solicitation id (/api/sbir/solicitations/{solicitation_id}/).
func (c *Client) GetSbirSolicitation(ctx context.Context, solicitationID string, opts *GetEntityOptions) (Record, error) {
	if solicitationID == "" {
		return nil, &ValidationError{&APIError{Message: "SBIR solicitation id is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/sbir/solicitations/"+pathEscape(solicitationID)+"/", opts.toQuery())
}
