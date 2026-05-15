package tango

import (
	"context"
)

// ListLcatsOptions selects which owner's Labor Categories (LCATs) to
// list. Exactly one of UEI or IDVKey must be set — LCATs live under
// owner resources in the Tango API; there is no top-level `/api/lcats/`
// endpoint. Mirrors the Node SDK's `listLcats` dispatcher.
//
// If both UEI and IDVKey are set, UEI wins.
type ListLcatsOptions struct {
	// UEI dispatches to /api/entities/{uei}/lcats/.
	UEI string

	// IDVKey dispatches to /api/idvs/{key}/lcats/.
	IDVKey string

	// Sub-resource filters: shape / pagination / search / ordering.
	// These are forwarded to the dispatched endpoint as-is.
	EntityLcatsOptions
}

// ListLcats lists Labor Categories (LCATs) owned by either an entity
// (when UEI is set) or an IDV (when IDVKey is set). One of the two
// must be set, or ListLcats returns *ValidationError.
//
// LCATs do not exist at the top level — they're always scoped to a
// parent resource. Use this method as a convenience when you want a
// single entry point; if you already know which owner type you want,
// call ListEntityLcats or ListIDVLcats directly.
func (c *Client) ListLcats(ctx context.Context, opts *ListLcatsOptions) (*PaginatedResponse[Record], error) {
	if opts == nil || (opts.UEI == "" && opts.IDVKey == "") {
		return nil, &ValidationError{&APIError{
			Message: "ListLcats: one of UEI or IDVKey is required",
		}}
	}
	if opts.UEI != "" {
		return c.ListEntityLcats(ctx, opts.UEI, &opts.EntityLcatsOptions)
	}
	return c.ListIDVLcats(ctx, opts.IDVKey, &opts.EntityLcatsOptions)
}

// IterateLcats walks every LCAT for the owner specified by opts.
// One of opts.UEI or opts.IDVKey must be set.
func (c *Client) IterateLcats(ctx context.Context, opts *ListLcatsOptions) *Iterator[Record] {
	if opts == nil {
		opts = &ListLcatsOptions{}
	}
	return &Iterator[Record]{
		ctx: ctx,
		fetch: func(ctx context.Context, page int, cursor string) (*PaginatedResponse[Record], error) {
			next := *opts
			next.Page = page
			next.Cursor = cursor
			return c.ListLcats(ctx, &next)
		},
	}
}
