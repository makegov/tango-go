package tango

// Client-side CRUD for Tango webhook endpoints + alerts.
//
// Note the package layering: this file lives in the root `tango` package and
// exposes methods on *Client for managing the webhook resource itself. The
// `github.com/makegov/tango-go/webhooks` subpackage handles signing and
// verification — distinct concern, distinct dependency footprint (a receiver
// shouldn't pull the API client). The two are wire-compatible: signatures
// produced by Generate verify here, and vice versa.

import (
	"context"
	"net/url"
)

// ---------------------------------------------------------------------------
// Webhook endpoints
// ---------------------------------------------------------------------------

// ListWebhookEventTypes returns the set of event types the server can emit.
// GET /api/webhooks/event-types/.
func (c *Client) ListWebhookEventTypes(ctx context.Context) (*WebhookEventTypesResponse, error) {
	return getGeneric[*WebhookEventTypesResponse](ctx, c, "/api/webhooks/event-types/", nil)
}

// ListWebhookEndpoints returns the caller's configured webhook endpoints.
// GET /api/webhooks/endpoints/.
func (c *Client) ListWebhookEndpoints(ctx context.Context, opts *ListOptions) (*PaginatedResponse[WebhookEndpoint], error) {
	q := url.Values{}
	if opts != nil {
		opts.applyTo(q)
	}
	return listGeneric[WebhookEndpoint](ctx, c, "/api/webhooks/endpoints/", q)
}

// GetWebhookEndpoint fetches a single endpoint by id.
// GET /api/webhooks/endpoints/{id}/.
func (c *Client) GetWebhookEndpoint(ctx context.Context, id string) (*WebhookEndpoint, error) {
	if id == "" {
		return nil, &ValidationError{&APIError{Message: "GetWebhookEndpoint: id is required"}}
	}
	return getGeneric[*WebhookEndpoint](ctx, c, "/api/webhooks/endpoints/"+pathEscape(id)+"/", nil)
}

// CreateWebhookEndpoint creates a new webhook endpoint.
// POST /api/webhooks/endpoints/.
//
// `Name` and `CallbackURL` are required. The Tango API enforces
// unique(user, name) on endpoints — raising client-side here gives a cleaner
// error than the server's 400 on duplicate. Matches the v1.0.0 behavior in
// `tango-node` and `tango-python` (see tango-node CHANGELOG, 1.0.0).
func (c *Client) CreateWebhookEndpoint(ctx context.Context, input WebhookEndpointCreateInput) (*WebhookEndpoint, error) {
	if input.Name == "" {
		return nil, &ValidationError{&APIError{Message: "CreateWebhookEndpoint: Name is required (the Tango API enforces unique(user, name) on endpoints)"}}
	}
	if input.CallbackURL == "" {
		return nil, &ValidationError{&APIError{Message: "CreateWebhookEndpoint: CallbackURL is required"}}
	}
	return postGeneric[*WebhookEndpoint](ctx, c, "/api/webhooks/endpoints/", input)
}

// UpdateWebhookEndpoint patches an existing endpoint. Only fields set on the
// input struct are sent (nil pointer == omit).
// PATCH /api/webhooks/endpoints/{id}/.
func (c *Client) UpdateWebhookEndpoint(ctx context.Context, id string, input WebhookEndpointUpdateInput) (*WebhookEndpoint, error) {
	if id == "" {
		return nil, &ValidationError{&APIError{Message: "UpdateWebhookEndpoint: id is required"}}
	}
	return patchGeneric[*WebhookEndpoint](ctx, c, "/api/webhooks/endpoints/"+pathEscape(id)+"/", input)
}

// DeleteWebhookEndpoint removes an endpoint by id.
// DELETE /api/webhooks/endpoints/{id}/.
func (c *Client) DeleteWebhookEndpoint(ctx context.Context, id string) error {
	if id == "" {
		return &ValidationError{&APIError{Message: "DeleteWebhookEndpoint: id is required"}}
	}
	return c.deleteEndpoint(ctx, "/api/webhooks/endpoints/"+pathEscape(id)+"/")
}

// TestWebhookEndpoint triggers a test delivery for a specific endpoint.
// POST /api/webhooks/endpoints/test-delivery/ with body {"endpoint": "<id>"}.
//
// The canonical request key is `endpoint` (as of tango#2252); the server
// still accepts the legacy `endpoint_id` alias for backward compatibility,
// but the SDK sends the canonical key.
func (c *Client) TestWebhookEndpoint(ctx context.Context, endpointID string) (*WebhookTestDeliveryResult, error) {
	if endpointID == "" {
		return nil, &ValidationError{&APIError{Message: "TestWebhookEndpoint: endpointID is required"}}
	}
	body := map[string]any{"endpoint": endpointID}
	return postGeneric[*WebhookTestDeliveryResult](ctx, c, "/api/webhooks/endpoints/test-delivery/", body)
}

// GetWebhookSamplePayload fetches a sample webhook delivery body for the
// given event type. If `eventType` is empty, the server returns samples for
// all event types (the "all-types" variant of WebhookSamplePayloadResponse).
// GET /api/webhooks/endpoints/sample-payload/?event_type=<type>.
func (c *Client) GetWebhookSamplePayload(ctx context.Context, eventType string) (*WebhookSamplePayloadResponse, error) {
	q := url.Values{}
	if eventType != "" {
		q.Set("event_type", eventType)
	}
	return getGeneric[*WebhookSamplePayloadResponse](ctx, c, "/api/webhooks/endpoints/sample-payload/", q)
}

// ---------------------------------------------------------------------------
// Webhook alerts (filter-subscription convenience API)
// ---------------------------------------------------------------------------

// ListWebhookAlerts returns the caller's filter-based subscriptions.
// GET /api/webhooks/alerts/.
func (c *Client) ListWebhookAlerts(ctx context.Context, opts *ListOptions) (*PaginatedResponse[WebhookAlert], error) {
	q := url.Values{}
	if opts != nil {
		opts.applyTo(q)
	}
	return listGeneric[WebhookAlert](ctx, c, "/api/webhooks/alerts/", q)
}

// GetWebhookAlert fetches a single alert by id.
// GET /api/webhooks/alerts/{id}/.
func (c *Client) GetWebhookAlert(ctx context.Context, id string) (*WebhookAlert, error) {
	if id == "" {
		return nil, &ValidationError{&APIError{Message: "GetWebhookAlert: id is required"}}
	}
	return getGeneric[*WebhookAlert](ctx, c, "/api/webhooks/alerts/"+pathEscape(id)+"/", nil)
}

// CreateWebhookAlert creates a filter-based subscription.
// POST /api/webhooks/alerts/.
//
// `Name`, `QueryType`, and a non-empty `Filters` map are required.
// `QueryType` is SINGULAR (e.g. `"contract"`, not `"contracts"`).
// For accounts with multiple webhook endpoints, set `Endpoint` to the
// destination endpoint UUID; single-endpoint accounts may omit it.
func (c *Client) CreateWebhookAlert(ctx context.Context, input WebhookAlertCreateInput) (*WebhookAlert, error) {
	if input.Name == "" {
		return nil, &ValidationError{&APIError{Message: "CreateWebhookAlert: Name is required"}}
	}
	if input.QueryType == "" {
		return nil, &ValidationError{&APIError{Message: `CreateWebhookAlert: QueryType is required (singular, e.g. "contract")`}}
	}
	if len(input.Filters) == 0 {
		return nil, &ValidationError{&APIError{Message: "CreateWebhookAlert: Filters must be a non-empty map"}}
	}
	return postGeneric[*WebhookAlert](ctx, c, "/api/webhooks/alerts/", input)
}

// UpdateWebhookAlert patches an existing alert. Only `Name`, `Frequency`,
// `CronExpression`, and `IsActive` are writable server-side.
// PATCH /api/webhooks/alerts/{id}/.
func (c *Client) UpdateWebhookAlert(ctx context.Context, id string, input WebhookAlertUpdateInput) (*WebhookAlert, error) {
	if id == "" {
		return nil, &ValidationError{&APIError{Message: "UpdateWebhookAlert: id is required"}}
	}
	return patchGeneric[*WebhookAlert](ctx, c, "/api/webhooks/alerts/"+pathEscape(id)+"/", input)
}

// DeleteWebhookAlert removes an alert by id.
// DELETE /api/webhooks/alerts/{id}/.
func (c *Client) DeleteWebhookAlert(ctx context.Context, id string) error {
	if id == "" {
		return &ValidationError{&APIError{Message: "DeleteWebhookAlert: id is required"}}
	}
	return c.deleteEndpoint(ctx, "/api/webhooks/alerts/"+pathEscape(id)+"/")
}
