package tango

import (
	"context"
	"net/url"
)

// ListItDashboardOptions filters /api/itdashboard/ — federal IT
// investments from the IT Dashboard. Mirrors the Node
// `ListItDashboardOptions` interface and the Python
// `list_itdashboard_investments` kwargs.
//
// Filters are tier-gated by the API:
//   - Free:      Search
//   - Pro:       AgencyCode, TypeOfInvestment, UpdatedTimeAfter,
//     UpdatedTimeBefore
//   - Business+: AgencyName, CIORating, CIORatingMax, PerformanceRisk
//
// Hitting a gated filter on a lower tier returns a 403.
//
// CIO ratings: 1=High Risk, 2=Moderately High, 3=Medium, 4=Moderately
// Low, 5=Low.
type ListItDashboardOptions struct {
	ListOptions

	Search           string
	AgencyCode       string // accept string to allow numeric or text codes
	AgencyName       string
	TypeOfInvestment string

	// Date / timestamp bounds (wire: ISO 8601).
	UpdatedTimeAfter  string
	UpdatedTimeBefore string

	// CIO ratings + performance risk. Accept strings for flexibility
	// (the API accepts both "3" and integer-typed values; keeping
	// strings avoids ambiguity around "unset" for numeric zero).
	CIORating       string
	CIORatingMax    string
	PerformanceRisk string

	Extra map[string]any
}

func (o *ListItDashboardOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "agency_code", o.AgencyCode)
	setIfNotEmpty(q, "agency_name", o.AgencyName)
	setIfNotEmpty(q, "type_of_investment", o.TypeOfInvestment)
	setIfNotEmpty(q, "updated_time_after", o.UpdatedTimeAfter)
	setIfNotEmpty(q, "updated_time_before", o.UpdatedTimeBefore)
	setIfNotEmpty(q, "cio_rating", o.CIORating)
	setIfNotEmpty(q, "cio_rating_max", o.CIORatingMax)
	setIfNotEmpty(q, "performance_risk", o.PerformanceRisk)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListItDashboard queries /api/itdashboard/ — federal IT investments.
func (c *Client) ListItDashboard(ctx context.Context, opts *ListItDashboardOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/itdashboard/", q)
}

// IterateItDashboard walks every IT Dashboard investment matching opts.
func (c *Client) IterateItDashboard(ctx context.Context, opts *ListItDashboardOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListItDashboardOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListItDashboard(ctx, &next)
		},
	}
}

// GetItDashboard fetches a single IT Dashboard investment by UII (Unique
// Investment Identifier).
func (c *Client) GetItDashboard(ctx context.Context, uii string, opts *GetEntityOptions) (Record, error) {
	if uii == "" {
		return nil, &ValidationError{&APIError{Message: "IT Dashboard UII is required"}}
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
	return getGeneric[Record](ctx, c, "/api/itdashboard/"+pathEscape(uii)+"/", q)
}
