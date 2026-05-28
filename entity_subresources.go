package tango

import (
	"context"
	"net/url"
)

// EntitySubresourceOptions controls filtering and shaping for the
// entity sub-resource list endpoints that share a common parameter set:
// /api/entities/{uei}/contracts/, /idvs/, /otas/, /otidvs/.
//
// Mirrors the Node SDK EntitySubresourceOptions interface.
type EntitySubresourceOptions struct {
	ListOptions

	// Joiner is the separator used when flat=true to join nested keys.
	// Default on the server side is ".". Only meaningful when Flat is set.
	Joiner string

	// Ordering is the server-side sort spec. Endpoint-specific allowlist;
	// prefix with "-" for descending.
	Ordering string

	// Search is a free-text search filter where supported by the endpoint.
	Search string

	// Extra forwards arbitrary query params that don't have a typed field.
	Extra map[string]any
}

func (o *EntitySubresourceOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	if o.Flat && o.Joiner != "" {
		q.Set("joiner", o.Joiner)
	}
	setIfNotEmpty(q, "ordering", o.Ordering)
	setIfNotEmpty(q, "search", o.Search)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// EntitySubawardsOptions filters /api/entities/{uei}/subawards/.
// The subawards endpoint enforces a strict server-side ordering allowlist
// (see tango#2254); pass values through as-is and let the server validate.
type EntitySubawardsOptions struct {
	ListOptions

	// Ordering must match the subawards endpoint's allowlist (e.g.
	// "last_modified_date" / "-last_modified_date").
	Ordering string

	// Extra forwards arbitrary query params that don't have a typed field.
	Extra map[string]any
}

func (o *EntitySubawardsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// EntityLcatsOptions filters /api/entities/{uei}/lcats/ and
// /api/idvs/{key}/lcats/. Minimal: pagination + shape + ordering + search.
type EntityLcatsOptions struct {
	ListOptions

	// Ordering is the server-side sort spec.
	Ordering string

	// Search is a free-text search filter.
	Search string

	// Extra forwards arbitrary query params that don't have a typed field.
	Extra map[string]any
}

func (o *EntityLcatsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "ordering", o.Ordering)
	setIfNotEmpty(q, "search", o.Search)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// listEntitySubresource implements the shared path/params plumbing for
// the family of /api/entities/{uei}/<segment>/ endpoints that take
// EntitySubresourceOptions.
func (c *Client) listEntitySubresource(ctx context.Context, uei, segment string, opts *EntitySubresourceOptions) (*PaginatedResponse[Record], error) {
	if uei == "" {
		return nil, &ValidationError{&APIError{Message: "UEI is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/entities/"+pathEscape(uei)+"/"+segment+"/", q)
}

// ListEntityContracts lists contracts awarded to an entity
// (/api/entities/{uei}/contracts/).
func (c *Client) ListEntityContracts(ctx context.Context, uei string, opts *EntitySubresourceOptions) (*PaginatedResponse[Record], error) {
	return c.listEntitySubresource(ctx, uei, "contracts", opts)
}

// ListEntityIDVs lists IDVs held by an entity
// (/api/entities/{uei}/idvs/).
func (c *Client) ListEntityIDVs(ctx context.Context, uei string, opts *EntitySubresourceOptions) (*PaginatedResponse[Record], error) {
	return c.listEntitySubresource(ctx, uei, "idvs", opts)
}

// ListEntityOTAs lists Other Transaction Awards (OTAs) held by an entity
// (/api/entities/{uei}/otas/).
func (c *Client) ListEntityOTAs(ctx context.Context, uei string, opts *EntitySubresourceOptions) (*PaginatedResponse[Record], error) {
	return c.listEntitySubresource(ctx, uei, "otas", opts)
}

// ListEntityOTIDVs lists Other Transaction IDVs (OTIDVs) held by an entity
// (/api/entities/{uei}/otidvs/).
func (c *Client) ListEntityOTIDVs(ctx context.Context, uei string, opts *EntitySubresourceOptions) (*PaginatedResponse[Record], error) {
	return c.listEntitySubresource(ctx, uei, "otidvs", opts)
}

// ListEntitySubawards lists subawards reported against contracts held by
// the given entity (/api/entities/{uei}/subawards/).
func (c *Client) ListEntitySubawards(ctx context.Context, uei string, opts *EntitySubawardsOptions) (*PaginatedResponse[Record], error) {
	if uei == "" {
		return nil, &ValidationError{&APIError{Message: "UEI is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/entities/"+pathEscape(uei)+"/subawards/", q)
}

// ListEntityLcats lists Labor Categories (LCATs) advertised by an entity
// (/api/entities/{uei}/lcats/).
func (c *Client) ListEntityLcats(ctx context.Context, uei string, opts *EntityLcatsOptions) (*PaginatedResponse[Record], error) {
	if uei == "" {
		return nil, &ValidationError{&APIError{Message: "UEI is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/entities/"+pathEscape(uei)+"/lcats/", q)
}

// GetEntityBudgetFlows lists funding-account budget flows attributed to an
// entity (/api/entities/{uei}/budget-flows/). Returns a paginated list of
// funding-account rows.
func (c *Client) GetEntityBudgetFlows(ctx context.Context, uei string, opts *EntitySubresourceOptions) (*PaginatedResponse[Record], error) {
	return c.listEntitySubresource(ctx, uei, "budget-flows", opts)
}
