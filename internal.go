package tango

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// ListOptions is the shared embed for every list endpoint. Per-resource
// options structs embed this so callers can set pagination + shaping on
// any list call.
type ListOptions struct {
	// Page is 1-based and used for offset-paginated endpoints. Mutually
	// exclusive with Cursor; if both are set, Cursor wins.
	Page int

	// Limit is the page size. The server caps this at 100 on most
	// endpoints. 0 means "use the server default".
	Limit int

	// Cursor is the keyset cursor for cursor-paginated endpoints (most
	// list endpoints support it). Pass the Cursor field from the
	// previous response.
	Cursor string

	// Shape is the comma-separated field selector for dynamic response
	// shaping. Use one of the Shape* constants or roll your own. Empty
	// means "use the API default for this endpoint".
	Shape string

	// Flat collapses nested objects into dot-separated keys in the
	// response. Default false.
	Flat bool

	// FlatLists when true flattens list-valued nested fields as well.
	// Only meaningful when Flat is true.
	FlatLists bool
}

// applyTo writes pagination+shape fields into a url.Values, respecting
// "zero means absent". Called by every list method.
func (o *ListOptions) applyTo(q url.Values) {
	if o == nil {
		return
	}
	if o.Cursor != "" {
		q.Set("cursor", o.Cursor)
	} else if o.Page > 0 {
		q.Set("page", strconv.Itoa(o.Page))
	}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Shape != "" {
		q.Set("shape", o.Shape)
	}
	if o.Flat {
		q.Set("flat", "true")
	}
	if o.FlatLists {
		q.Set("flat_lists", "true")
	}
}

// setIfNotEmpty writes key=value only when value isn't empty/zero.
func setIfNotEmpty(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

func setIfNonZeroInt(q url.Values, key string, value int) {
	if value != 0 {
		q.Set(key, strconv.Itoa(value))
	}
}

func setIfNotNilBool(q url.Values, key string, value *bool) {
	if value != nil {
		q.Set(key, strconv.FormatBool(*value))
	}
}

// listGeneric runs a GET against a list endpoint and returns a typed
// PaginatedResponse. Used by every list method on Client.
func listGeneric[T any](ctx context.Context, c *Client, path string, q url.Values) (*PaginatedResponse[T], error) {
	raw, err := c.do(ctx, requestSpec{method: "GET", path: path, query: q})
	if err != nil {
		return nil, err
	}
	return decodePaginated[T](raw)
}

// getGeneric runs a GET against a detail endpoint and decodes into T.
func getGeneric[T any](ctx context.Context, c *Client, path string, q url.Values) (T, error) {
	var zero T
	raw, err := c.do(ctx, requestSpec{method: "GET", path: path, query: q})
	if err != nil {
		return zero, err
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, &APIError{Message: "decode response", Cause: err}
	}
	return out, nil
}

// postGeneric runs a POST and decodes the JSON response into T.
func postGeneric[T any](ctx context.Context, c *Client, path string, body any) (T, error) {
	var zero T
	raw, err := c.do(ctx, requestSpec{method: "POST", path: path, body: body})
	if err != nil {
		return zero, err
	}
	var out T
	if len(raw) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, &APIError{Message: "decode response", Cause: err}
	}
	return out, nil
}

// patchGeneric runs a PATCH and decodes the JSON response into T.
func patchGeneric[T any](ctx context.Context, c *Client, path string, body any) (T, error) {
	var zero T
	raw, err := c.do(ctx, requestSpec{method: "PATCH", path: path, body: body})
	if err != nil {
		return zero, err
	}
	var out T
	if len(raw) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, &APIError{Message: "decode response", Cause: err}
	}
	return out, nil
}

// deleteEndpoint runs a DELETE and discards the response.
func (c *Client) deleteEndpoint(ctx context.Context, path string) error {
	_, err := c.do(ctx, requestSpec{method: "DELETE", path: path})
	return err
}

// pathEscape is url.PathEscape with one extra quirk: the API rejects
// "+" for spaces in path segments, so we use PathEscape (which uses %20)
// rather than QueryEscape. Centralized so callers don't have to think.
func pathEscape(s string) string { return url.PathEscape(s) }
