package tango

import "context"

// ListAPIKeys returns the authenticated user's API keys.
// Endpoint: GET /api/api-keys/.
//
// Unlike the other list endpoints, this one does not paginate — it
// returns a structured object with the caller's keys and metadata.
// Mirrors Node `listAPIKeys` and Python `list_api_keys`.
func (c *Client) ListAPIKeys(ctx context.Context) (Record, error) {
	return getGeneric[Record](ctx, c, "/api/api-keys/", nil)
}
