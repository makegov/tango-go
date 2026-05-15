package tango

import (
	"context"
	"net/url"
)

// ListDepartments queries /api/departments/ — the legacy standalone
// departments endpoint. Mirrors Node `listDepartments` (marked
// @deprecated in the Node SDK — see comment below) and Python
// `list_departments`.
//
// Deprecated: prefer ListOrganizations with Level=1 for new code. The
// standalone /api/departments/ endpoint is retained for backward
// compatibility and will be removed in a future API version (tango#1461,
// legacy agency tables retirement).
func (c *Client) ListDepartments(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		opts.applyTo(q)
	}
	return listGeneric[Record](ctx, c, "/api/departments/", q)
}

// GetDepartment fetches a single department by its code (typically the
// CGAC department code, e.g. "097" for DoD).
func (c *Client) GetDepartment(ctx context.Context, code string) (Record, error) {
	if code == "" {
		return nil, &ValidationError{&APIError{Message: "department code is required"}}
	}
	return getGeneric[Record](ctx, c, "/api/departments/"+pathEscape(code)+"/", nil)
}
