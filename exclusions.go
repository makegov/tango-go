package tango

import (
	"context"
	"net/url"
)

// ListExclusionsOptions filters /api/exclusions/ — SAM.gov exclusions (debarments, suspensions and other ineligibility actions).
type ListExclusionsOptions struct {
	ListOptions

	// Active filters on whether the exclusion is in force today: not delisted, already activated, and not yet terminated.
	// It is derived at query time, so an exclusion reaching its termination date changes this without any write.
	// A pointer so that false is a real filter value rather than an absent one.
	Active *bool

	// Delisted filters on whether the exclusion has dropped out of the SAM catalog, which is distinct from reaching its termination date.
	Delisted *bool

	ClassificationType  string
	ExclusionType       string
	ExclusionProgram    string
	ExcludingAgencyCode string
	ExcludingAgencyName string

	// UEI matches the exclusion's own UEI; most exclusions name individuals and carry none.
	UEI      string
	CageCode string
	NPI      string

	// EntityUEI matches the linked Tango entity, set only when the exclusion's UEI matches a registered entity.
	EntityUEI string

	// Date bounds (wire format: YYYY-MM-DD).
	ActivateDateAfter     string
	ActivateDateBefore    string
	TerminationDateAfter  string
	TerminationDateBefore string
	UpdateDateAfter       string
	UpdateDateBefore      string

	Search string

	// Ordering is one of activate_date, create_date, modified, termination_date or update_date; prefix "-" for descending.
	Ordering string

	Extra map[string]any
}

func (o *ListExclusionsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotNilBool(q, "active", o.Active)
	setIfNotNilBool(q, "delisted", o.Delisted)
	setIfNotEmpty(q, "classification_type", o.ClassificationType)
	setIfNotEmpty(q, "exclusion_type", o.ExclusionType)
	setIfNotEmpty(q, "exclusion_program", o.ExclusionProgram)
	setIfNotEmpty(q, "excluding_agency_code", o.ExcludingAgencyCode)
	setIfNotEmpty(q, "excluding_agency_name", o.ExcludingAgencyName)
	setIfNotEmpty(q, "uei", o.UEI)
	setIfNotEmpty(q, "cage_code", o.CageCode)
	setIfNotEmpty(q, "npi", o.NPI)
	setIfNotEmpty(q, "entity_uei", o.EntityUEI)
	setIfNotEmpty(q, "activate_date_after", o.ActivateDateAfter)
	setIfNotEmpty(q, "activate_date_before", o.ActivateDateBefore)
	setIfNotEmpty(q, "termination_date_after", o.TerminationDateAfter)
	setIfNotEmpty(q, "termination_date_before", o.TerminationDateBefore)
	setIfNotEmpty(q, "update_date_after", o.UpdateDateAfter)
	setIfNotEmpty(q, "update_date_before", o.UpdateDateBefore)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListExclusions queries /api/exclusions/.
func (c *Client) ListExclusions(ctx context.Context, opts *ListExclusionsOptions) (*PaginatedResponse[Record], error) {
	return listGeneric[Record](ctx, c, "/api/exclusions/", opts.toQuery())
}

// IterateExclusions walks every exclusion matching opts.
func (c *Client) IterateExclusions(ctx context.Context, opts *ListExclusionsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListExclusionsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListExclusions(ctx, &next)
		},
	}
}

// GetExclusion fetches one exclusion by its exclusion key (/api/exclusions/{exclusion_key}/).
func (c *Client) GetExclusion(ctx context.Context, exclusionKey string, opts *GetEntityOptions) (Record, error) {
	if exclusionKey == "" {
		return nil, &ValidationError{&APIError{Message: "exclusion key is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/exclusions/"+pathEscape(exclusionKey)+"/", opts.toQuery())
}
