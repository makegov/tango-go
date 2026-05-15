package tango

import (
	"context"
	"net/url"
)

// ListAssistanceListings queries /api/assistance_listings/ — the CFDA
// (Catalog of Federal Domestic Assistance) program catalog. Mirrors Node
// `listAssistanceListings` and Python `list_assistance_listings`.
//
// opts may be nil for defaults (server page size).
func (c *Client) ListAssistanceListings(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		opts.applyTo(q)
	}
	return listGeneric[Record](ctx, c, "/api/assistance_listings/", q)
}

// GetAssistanceListing fetches a single Assistance Listing (CFDA
// program) by its CFDA number (e.g. "10.001").
func (c *Client) GetAssistanceListing(ctx context.Context, number string) (Record, error) {
	if number == "" {
		return nil, &ValidationError{&APIError{Message: "assistance listing number is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/assistance_listings/"+pathEscape(number)+"/", nil)
}
