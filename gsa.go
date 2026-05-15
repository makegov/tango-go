package tango

import (
	"context"
	"net/url"
)

// ListGsaElibraryContractsOptions filters /api/gsa_elibrary_contracts/.
// Mirrors the Node `ListGsaElibraryContractsOptions` interface and the
// Python `list_gsa_elibrary_contracts` kwargs.
type ListGsaElibraryContractsOptions struct {
	ListOptions

	Schedule       string
	ContractNumber string
	Key            string
	PIID           string
	UEI            string
	SIN            string
	Search         string
	Ordering       string

	Extra map[string]any
}

func (o *ListGsaElibraryContractsOptions) toQuery() url.Values {
	q := url.Values{}
	if o == nil {
		return q
	}
	o.ListOptions.applyTo(q)
	setIfNotEmpty(q, "schedule", o.Schedule)
	setIfNotEmpty(q, "contract_number", o.ContractNumber)
	setIfNotEmpty(q, "key", o.Key)
	setIfNotEmpty(q, "piid", o.PIID)
	setIfNotEmpty(q, "uei", o.UEI)
	setIfNotEmpty(q, "sin", o.SIN)
	setIfNotEmpty(q, "search", o.Search)
	setIfNotEmpty(q, "ordering", o.Ordering)
	for k, v := range o.Extra {
		q.Set(k, valueToString(v))
	}
	return q
}

// ListGsaElibraryContracts queries /api/gsa_elibrary_contracts/ — GSA
// eLibrary contract listings.
func (c *Client) ListGsaElibraryContracts(ctx context.Context, opts *ListGsaElibraryContractsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		q = opts.toQuery()
	}
	return listGeneric[Record](ctx, c, "/api/gsa_elibrary_contracts/", q)
}

// IterateGsaElibraryContracts walks every GSA eLibrary contract matching
// opts.
func (c *Client) IterateGsaElibraryContracts(ctx context.Context, opts *ListGsaElibraryContractsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListGsaElibraryContractsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListGsaElibraryContracts(ctx, &next)
		},
	}
}

// GetGsaElibraryContract fetches a single GSA eLibrary contract by UUID.
//
// Note: this is a Python-only method (the Node SDK has no equivalent
// single-record getter). Included here for parity with tango-python.
func (c *Client) GetGsaElibraryContract(ctx context.Context, uuid string, opts *GetEntityOptions) (Record, error) {
	if uuid == "" {
		return nil, &ValidationError{&APIError{Message: "GSA eLibrary contract UUID is required"}}
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
	return getGeneric[Record](ctx, c, "/api/gsa_elibrary_contracts/"+pathEscape(uuid)+"/", q)
}
