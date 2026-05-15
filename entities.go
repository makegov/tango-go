package tango

import (
	"context"
	"net/url"
)

// ListEntitiesOptions filters /api/entities/ (federal vendors / recipients).
type ListEntitiesOptions struct {
	ListOptions

	Search                       string
	CageCode                     string
	NAICS                        string
	Name                         string
	PSC                          string
	PurposeOfRegistrationCode    string
	Socioeconomic                string
	State                        string
	TotalAwardsObligatedGte      string
	TotalAwardsObligatedLte      string
	UEI                          string
	ZipCode                      string

	Extra map[string]any
}

func (o *ListEntitiesOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "cage_code", o.CageCode)
	setIfNotEmpty(q, "naics", o.NAICS)
	setIfNotEmpty(q, "name", o.Name)
	setIfNotEmpty(q, "psc", o.PSC)
	setIfNotEmpty(q, "purpose_of_registration_code", o.PurposeOfRegistrationCode)
	setIfNotEmpty(q, "socioeconomic", o.Socioeconomic)
	setIfNotEmpty(q, "state", o.State)
	setIfNotEmpty(q, "total_awards_obligated_gte", o.TotalAwardsObligatedGte)
	setIfNotEmpty(q, "total_awards_obligated_lte", o.TotalAwardsObligatedLte)
	setIfNotEmpty(q, "uei", o.UEI)
	setIfNotEmpty(q, "zip_code", o.ZipCode)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListEntities queries /api/entities/.
func (c *Client) ListEntities(ctx context.Context, opts *ListEntitiesOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/entities/", q)
}

// GetEntity fetches a single entity by UEI or CAGE code.
//
// When shape is empty, the API returns its comprehensive default. Pass
// ShapeEntitiesMinimal (or your own) for a slimmer payload.
func (c *Client) GetEntity(ctx context.Context, key string, opts *GetEntityOptions) (Record, error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "entity key is required"}}
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
	return getGeneric[Record](ctx, c, "/api/entities/"+pathEscape(key)+"/", q)
}

// GetEntityOptions controls shaping for GetEntity.
type GetEntityOptions struct {
	Shape     string
	Flat      bool
	FlatLists bool
}

// IterateEntities walks every entity matching opts.
func (c *Client) IterateEntities(ctx context.Context, opts *ListEntitiesOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListEntitiesOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListEntities(ctx, &next)
		},
	}
}

