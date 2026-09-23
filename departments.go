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
// Deprecated: prefer ListOrganizations with Level=1 for new code.
// The standalone /api/departments/ endpoint is retained for backward compatibility and will be removed in a future API version.
func (c *Client) ListDepartments(ctx context.Context, opts *ListOptions) (*PaginatedResponse[Record], error) {
	q := url.Values{}
	if opts != nil {
		opts.applyTo(q)
	}
	return listGeneric[Record](ctx, c, "/api/departments/", q)
}

// DepartmentRecord is the typed return model for GetDepartment.
//
// Code is an integer on the wire (DoD is 97), not a zero-padded CGAC string, so it decodes into *int.
// Fields are pointers so the zero value means "absent"; unknown wire fields land in Extra.
type DepartmentRecord struct {
	Code         *int    `json:"code,omitempty"`
	Name         *string `json:"name,omitempty"`
	Abbreviation *string `json:"abbreviation,omitempty"`

	// Extra captures any forward-compatible fields the server adds that aren't in the typed surface yet.
	// Not serialized on the way out.
	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (d *DepartmentRecord) UnmarshalJSON(data []byte) error {
	type alias DepartmentRecord
	out := (*alias)(d)
	return unmarshalWithExtra(data, out)
}

// GetDepartment fetches a single department by its integer code (/api/departments/{code}/).
//
// The code is passed as a string for the path; "97" and "097" both resolve to the Department of Defense.
// Returns a typed *DepartmentRecord whose Code is an *int, matching the API.
func (c *Client) GetDepartment(ctx context.Context, code string) (*DepartmentRecord, error) {
	if code == "" {
		return nil, &ValidationError{&APIError{Message: "department code is required"}}
	}
	return getGeneric[*DepartmentRecord](ctx, c, "/api/departments/"+pathEscape(code)+"/", nil)
}
