package tango

import (
	"context"
	"net/url"
)

// ListIDVAwards lists task-order awards under a parent IDV
// (/api/idvs/{key}/awards/). Filtering mirrors ListIDVs; pagination is
// cursor-based on the server side.
func (c *Client) ListIDVAwards(ctx context.Context, key string, opts *ListIDVsOptions) (*PaginatedResponse[Record], error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "IDV key is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/idvs/"+pathEscape(key)+"/awards/", q)
}

// ListIDVChildIDVs lists child IDVs nested under a parent IDV
// (/api/idvs/{key}/idvs/). Filtering mirrors ListIDVs.
func (c *Client) ListIDVChildIDVs(ctx context.Context, key string, opts *ListIDVsOptions) (*PaginatedResponse[Record], error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "IDV key is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/idvs/"+pathEscape(key)+"/idvs/", q)
}

// ListIDVTransactions lists the raw transaction history backing an IDV
// (/api/idvs/{key}/transactions/). The endpoint only accepts pagination
// params; pass a *ListOptions with Limit/Cursor (or Page).
func (c *Client) ListIDVTransactions(ctx context.Context, key string, opts *ListOptions) (*PaginatedResponse[Record], error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "IDV key is required"}}
	}
	q := url.Values{}
	if opts != nil {
		opts.applyTo(q)
	}
	return listGeneric[Record](ctx, c, "/api/idvs/"+pathEscape(key)+"/transactions/", q)
}

// ListIDVLcats lists Labor Categories (LCATs) under an IDV
// (/api/idvs/{key}/lcats/). Re-uses EntityLcatsOptions because the server
// accepts the same parameter shape on both /entities/{uei}/lcats/ and
// /idvs/{key}/lcats/.
func (c *Client) ListIDVLcats(ctx context.Context, key string, opts *EntityLcatsOptions) (*PaginatedResponse[Record], error) {
	if key == "" {
		return nil, &ValidationError{&APIError{Message: "IDV key is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/idvs/"+pathEscape(key)+"/lcats/", q)
}
