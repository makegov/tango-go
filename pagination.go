package tango

import (
	"context"
	"encoding/json"
	"iter"
	"net/url"
)

// Record is the default result type for list endpoints. It mirrors the
// permissive map[string]any shape used by the API — fields vary based on
// the requested shape, so callers either work with the map directly or
// json.Unmarshal it into their own struct.
type Record = map[string]any

// PaginatedResponse is the standard envelope for list endpoints.
//
// For keyset/cursor endpoints (e.g. /api/idvs/), the next cursor is
// extracted from Next for convenience and exposed on Cursor.
type PaginatedResponse[T any] struct {
	// Count is the total number of items across all pages, when the
	// server reports it. For some endpoints this matches len(Results).
	Count int `json:"count"`

	// Next is the URL of the next page, or "" when on the last page.
	Next string `json:"next"`

	// Previous is the URL of the previous page, or "" when on the first.
	Previous string `json:"previous"`

	// PageMetadata carries opaque page metadata for keyset endpoints.
	PageMetadata map[string]any `json:"page_metadata,omitempty"`

	// Cursor is the cursor extracted from Next on keyset-paginated
	// endpoints. Pass it back via the Cursor option to fetch the next
	// page. Empty when the endpoint is page-based or no cursor is set.
	Cursor string `json:"-"`

	// Results is the actual data for this page.
	Results []T `json:"results"`
}

// rawListResponse is what the wire actually returns; we decode into this
// and then build a typed PaginatedResponse[T].
type rawListResponse struct {
	Count        json.Number     `json:"count"`
	Next         *string         `json:"next"`
	Previous     *string         `json:"previous"`
	PageMetadata map[string]any  `json:"page_metadata"`
	Results      json.RawMessage `json:"results"`
}

func decodePaginated[T any](raw []byte) (*PaginatedResponse[T], error) {
	var envelope rawListResponse
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, &APIError{Message: "decode list response: " + err.Error()}
	}
	out := &PaginatedResponse[T]{}
	if envelope.Count != "" {
		if n, err := envelope.Count.Int64(); err == nil {
			out.Count = int(n)
		}
	}
	if envelope.Next != nil {
		out.Next = *envelope.Next
		out.Cursor = extractCursor(out.Next)
	}
	if envelope.Previous != nil {
		out.Previous = *envelope.Previous
	}
	out.PageMetadata = envelope.PageMetadata

	if len(envelope.Results) > 0 {
		if err := json.Unmarshal(envelope.Results, &out.Results); err != nil {
			return nil, &APIError{Message: "decode list results: " + err.Error()}
		}
	}
	if out.Count == 0 {
		out.Count = len(out.Results)
	}
	return out, nil
}

func extractCursor(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Query().Get("cursor")
}

// Iterator walks every result of a paginated list endpoint, following
// either page= or cursor= on the server's Next URL.
//
// Usage:
//
//	iter := client.IterateContracts(ctx, &tango.ListContractsOptions{
//	    AwardingAgency: "9700",
//	})
//	for iter.Next() {
//	    contract := iter.Item()
//	    // ...
//	}
//	if err := iter.Err(); err != nil { ... }
type Iterator[T any] struct {
	ctx     context.Context
	fetch   func(ctx context.Context, page int, cursor string) (*PaginatedResponse[T], error)
	page    int    // 0 means "haven't set"
	cursor  string // empty means "haven't set"
	buf     []T
	idx     int
	current T
	done    bool
	err     error
}

// Next advances the iterator and returns true if Item() is now valid.
// Returns false when the iteration is finished or an error occurred (in
// which case Err() returns the error).
func (it *Iterator[T]) Next() bool {
	if it.err != nil {
		return false
	}
	// Drain the buffer first — even after the previous fetch flagged
	// "done" (last page reached), buf may still hold un-yielded items.
	// Yielding them before checking done is what makes single-page and
	// multi-item-last-page iteration correct.
	if it.idx < len(it.buf) {
		it.current = it.buf[it.idx]
		it.idx++
		return true
	}
	if it.done {
		return false
	}
	// Need a new page.
	resp, err := it.fetch(it.ctx, it.page, it.cursor)
	if err != nil {
		it.err = err
		return false
	}
	if len(resp.Results) == 0 {
		it.done = true
		return false
	}
	it.buf = resp.Results
	it.idx = 0

	// Decide how to advance.
	if resp.Next == "" {
		it.done = true // last page; we'll still drain buf
		it.page = 0
		it.cursor = ""
	} else if resp.Cursor != "" {
		it.cursor = resp.Cursor
		it.page = 0
	} else if u, err := url.Parse(resp.Next); err == nil {
		if p := u.Query().Get("page"); p != "" {
			if n, err := parseIntStrict(p); err == nil {
				it.page = n
				it.cursor = ""
			} else {
				it.done = true
			}
		} else {
			it.done = true
		}
	} else {
		it.done = true
	}

	it.current = it.buf[it.idx]
	it.idx++
	return true
}

// Item returns the current item. Only valid after a true return from Next.
func (it *Iterator[T]) Item() T { return it.current }

// Err returns the error that ended iteration, if any.
func (it *Iterator[T]) Err() error { return it.err }

// Seq adapts the iterator to a Go 1.23+ range-over-func sequence.
// Yields (item, nil) for each result and (zero, err) once if iteration
// fails. The Next/Item/Err surface still works — Seq is additive sugar.
//
// Usage:
//
//	for record, err := range client.IterateContracts(ctx, opts).Seq() {
//	    if err != nil {
//	        return err
//	    }
//	    // use record
//	}
//
// Break out of the loop normally; the underlying iterator stops fetching
// pages as soon as the range body returns false.
func (it *Iterator[T]) Seq() iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for it.Next() {
			if !yield(it.Item(), nil) {
				return
			}
		}
		if err := it.Err(); err != nil {
			var zero T
			yield(zero, err)
		}
	}
}

func parseIntStrict(s string) (int, error) {
	n := 0
	if s == "" {
		return 0, &APIError{Message: "empty page"}
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, &APIError{Message: "invalid page: " + s}
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}
