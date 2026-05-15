package tango

import (
	"context"
	"net/url"
)

// ListMasSinsOptions filters /api/mas_sins/. Embeds ListOptions for
// pagination + shape control; adds a Search knob.
type ListMasSinsOptions struct {
	ListOptions

	// Search is a free-text query against the SIN number, title, and
	// description.
	Search string
}

// ListMasSins queries /api/mas_sins/ — the canonical list of GSA
// Multiple Award Schedule (MAS) Special Item Numbers. Mirrors Node
// `listMasSins` and Python `list_mas_sins`.
func (c *Client) ListMasSins(ctx context.Context, opts *ListMasSinsOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		opts.ListOptions.applyTo(q)
		setIfNotEmpty(q, "search", opts.Search)
	}
	return listGeneric[Record](ctx, c, "/api/mas_sins/", q)
}

// GetMasSin fetches a single MAS SIN by its identifier (e.g. "54151S").
func (c *Client) GetMasSin(ctx context.Context, sin string) (Record, error) {
	if sin == "" {
		return nil, &ValidationError{&APIError{Message: "MAS SIN is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/mas_sins/"+pathEscape(sin)+"/", nil)
}
