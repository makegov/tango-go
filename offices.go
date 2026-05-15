package tango

import (
	"context"
	"net/url"
)

// ListOfficesOptions filters /api/offices/. Embeds ListOptions for
// pagination + shape control; adds a Search knob to full-text-search
// office names and codes.
type ListOfficesOptions struct {
	ListOptions

	// Search is a free-text query matched against office name + code.
	Search string
}

// ListOffices queries /api/offices/ — the canonical list of federal
// contracting offices (FPDS-NG hierarchy). Mirrors Node `listOffices`
// and Python `list_offices`.
func (c *Client) ListOffices(ctx context.Context, opts *ListOfficesOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		opts.ListOptions.applyTo(q)
		setIfNotEmpty(q, "search", opts.Search)
	}
	return listGeneric[Record](ctx, c, "/api/offices/", q)
}

// GetOffice fetches a single office by its FPDS-NG office code.
func (c *Client) GetOffice(ctx context.Context, code string) (Record, error) {
	if code == "" {
		return nil, &ValidationError{&APIError{Message: "office code is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/offices/"+pathEscape(code)+"/", nil)
}
