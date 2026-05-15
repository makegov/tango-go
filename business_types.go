package tango

import (
	"context"
	"net/url"
)

// ListBusinessTypes queries /api/business_types/ — the reference list of
// SBA / SAM.gov socioeconomic and structural designations (8(a),
// woman-owned, veteran-owned, non-profit, etc.). Mirrors the Node
// `listBusinessTypes` and Python `list_business_types` methods.
//
// opts may be nil for defaults (server page size).
func (c *Client) ListBusinessTypes(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		opts.applyTo(q)
	}
	return listGeneric[Record](ctx, c, "/api/business_types/", q)
}

// GetBusinessType fetches a single business-type reference record by its
// short code. Returns *NotFoundError when the code is unknown.
func (c *Client) GetBusinessType(ctx context.Context, code string) (Record, error) {
	if code == "" {
		return nil, &ValidationError{&APIError{Message: "business type code is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/business_types/"+pathEscape(code)+"/", nil)
}
