package tango

import (
	"context"
	"net/url"
)

// AgencyContractsOptions controls filtering and shaping for the agency
// contract sub-resource endpoints:
// /api/agencies/{code}/contracts/awarding/ and
// /api/agencies/{code}/contracts/funding/.
//
// Mirrors the Node SDK AgencyContractsOptions interface.
type AgencyContractsOptions struct {
	ListOptions

	// Joiner is the separator used when flat=true to join nested keys.
	// Default on the server side is ".". Only meaningful when Flat is set.
	Joiner string

	// Ordering is the server-side sort spec. Endpoint-specific allowlist;
	// prefix with "-" for descending.
	Ordering string

	// Search is a free-text search filter.
	Search string

	// Extra forwards arbitrary query params that don't have a typed field.
	Extra map[string]any
}

func (o *AgencyContractsOptions) toQuery() url.Values {
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

// listAgencyContracts implements the shared plumbing for the
// /api/agencies/{code}/contracts/{which}/ endpoint family. "which" is
// "awarding" or "funding".
func (c *Client) listAgencyContracts(ctx context.Context, code, which string, opts *AgencyContractsOptions) (*PaginatedResponse[Record], error) {
	if code == "" {
		return nil, &ValidationError{&APIError{Message: "agency code is required"}}
	}
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/agencies/"+pathEscape(code)+"/contracts/"+which+"/", q)
}

// ListAgencyAwardingContracts lists contracts where the given agency is
// the awarding agency (/api/agencies/{code}/contracts/awarding/).
func (c *Client) ListAgencyAwardingContracts(ctx context.Context, code string, opts *AgencyContractsOptions) (*PaginatedResponse[Record], error) {
	return c.listAgencyContracts(ctx, code, "awarding", opts)
}

// ListAgencyFundingContracts lists contracts where the given agency is
// the funding agency (/api/agencies/{code}/contracts/funding/).
func (c *Client) ListAgencyFundingContracts(ctx context.Context, code string, opts *AgencyContractsOptions) (*PaginatedResponse[Record], error) {
	return c.listAgencyContracts(ctx, code, "funding", opts)
}
