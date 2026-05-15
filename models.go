package tango

// This file holds typed response/input models mirrored from
// `tango-node/src/types.ts` and `tango-node/src/models/Webhooks.ts`. The
// pattern is "permissive + forward-compatible":
//
//   - All fields are pointers (or maps/slices) so the zero value is "absent",
//     distinguishable from a server-supplied zero ("" / 0 / false).
//   - All JSON tags use `omitempty` so round-trips don't synthesize fields the
//     server didn't send.
//   - Every struct includes an `Extra map[string]any` (with `json:"-"`) that
//     UnmarshalJSON populates with any fields not matched by the named fields,
//     so the SDK doesn't lose data when the server adds a column. This mirrors
//     the TypeScript `[key: string]: unknown` index signature.
//
// If you add a new typed struct, mirror the pattern: pointer-optional fields,
// `Extra map[string]any \`json:"-"\``, and a custom UnmarshalJSON that catches
// the leftover keys. Helper `unmarshalWithExtra` does the heavy lifting.

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
)

// unmarshalWithExtra decodes `data` into `dst` (a non-nil pointer-to-struct),
// then collects every unrecognized top-level key into the `Extra` map on the
// struct (if the struct has an exported `Extra map[string]any` field with
// `json:"-"`). This lets every typed model preserve forward-compatible fields
// without forcing callers to drop down to `Record`.
func unmarshalWithExtra(data []byte, dst any) error {
	// First pass: decode into the typed struct.
	if err := json.Unmarshal(data, dst); err != nil {
		return err
	}

	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil
	}
	extraField := v.FieldByName("Extra")
	if !extraField.IsValid() || !extraField.CanSet() {
		return nil
	}

	// Collect named keys from json tags so we know what to skip.
	named := make(map[string]struct{}, v.NumField())
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name := strings.SplitN(tag, ",", 2)[0]
		if name == "" {
			name = f.Name
		}
		named[name] = struct{}{}
	}

	// Second pass: pull every key from the wire body, keep the ones we didn't
	// decode into named fields.
	var raw map[string]json.RawMessage
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		// `dst` may legitimately be non-object JSON (unlikely for these
		// models, but safer than failing the call). Leave Extra empty.
		return nil
	}
	if len(raw) == 0 {
		return nil
	}
	extra := map[string]any{}
	for k, v := range raw {
		if _, ok := named[k]; ok {
			continue
		}
		var anyVal any
		if err := json.Unmarshal(v, &anyVal); err == nil {
			extra[k] = anyVal
		}
	}
	if len(extra) > 0 {
		extraField.Set(reflect.ValueOf(extra))
	}
	return nil
}

// ---------------------------------------------------------------------------
// Agency
// ---------------------------------------------------------------------------

// AgencyRecord is the typed response from GetAgency. Mirrors
// `AgencyRecord` in `tango-node/src/types.ts`. All fields are pointers so the
// zero value distinguishes "absent" from "explicit empty/null".
type AgencyRecord struct {
	AgencyID     *string        `json:"agency_id,omitempty"`
	Name         *string        `json:"name,omitempty"`
	Abbreviation *string        `json:"abbreviation,omitempty"`
	Code         *string        `json:"code,omitempty"`
	Department   map[string]any `json:"department,omitempty"`

	// Extra captures any forward-compatible fields the server adds that
	// aren't in the typed surface yet. Not serialized on the way out.
	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AgencyRecord) UnmarshalJSON(data []byte) error {
	type alias AgencyRecord
	out := (*alias)(a)
	return unmarshalWithExtra(data, out)
}

// ---------------------------------------------------------------------------
// Webhook endpoints
// ---------------------------------------------------------------------------

// WebhookEndpoint is a configured outbound webhook destination. Mirrors the
// `WebhookEndpoint` interface in `tango-node/src/models/Webhooks.ts`.
type WebhookEndpoint struct {
	ID          *string `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	CallbackURL *string `json:"callback_url,omitempty"`
	Secret      *string `json:"secret,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	CreatedAt   *string `json:"created_at,omitempty"`
	UpdatedAt   *string `json:"updated_at,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (w *WebhookEndpoint) UnmarshalJSON(data []byte) error {
	type alias WebhookEndpoint
	out := (*alias)(w)
	return unmarshalWithExtra(data, out)
}

// WebhookEndpointCreateInput is the body for CreateWebhookEndpoint.
//
// `Name` and `CallbackURL` are required server-side (the API enforces
// unique(user, name) on endpoints — see tango#1.0.0 changes). `IsActive`
// defaults to true on the server when omitted; the SDK preserves omit
// semantics so the server gets to apply that default.
type WebhookEndpointCreateInput struct {
	Name        string   `json:"name"`
	CallbackURL string   `json:"callback_url"`
	IsActive    *bool    `json:"is_active,omitempty"`
	EventTypes  []string `json:"event_types,omitempty"`
}

// WebhookEndpointUpdateInput is the patch body for UpdateWebhookEndpoint. All
// fields are optional; only the ones the caller sets are sent to the server.
type WebhookEndpointUpdateInput struct {
	Name        *string  `json:"name,omitempty"`
	CallbackURL *string  `json:"callback_url,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
	EventTypes  []string `json:"event_types,omitempty"`
}

// ---------------------------------------------------------------------------
// Webhook event types + sample payloads + test delivery
// ---------------------------------------------------------------------------

// WebhookEventType describes a single event type the server can emit.
type WebhookEventType struct {
	EventType     *string `json:"event_type,omitempty"`
	Description   *string `json:"description,omitempty"`
	SchemaVersion *int    `json:"schema_version,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (w *WebhookEventType) UnmarshalJSON(data []byte) error {
	type alias WebhookEventType
	out := (*alias)(w)
	return unmarshalWithExtra(data, out)
}

// WebhookEventTypesResponse wraps the response from ListWebhookEventTypes.
type WebhookEventTypesResponse struct {
	EventTypes []WebhookEventType `json:"event_types,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (w *WebhookEventTypesResponse) UnmarshalJSON(data []byte) error {
	type alias WebhookEventTypesResponse
	out := (*alias)(w)
	return unmarshalWithExtra(data, out)
}

// WebhookSampleDelivery is one element of a sample-payload response body —
// a timestamped batch of synthetic events the server will deliver.
type WebhookSampleDelivery struct {
	Timestamp *string          `json:"timestamp,omitempty"`
	Events    []map[string]any `json:"events,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (w *WebhookSampleDelivery) UnmarshalJSON(data []byte) error {
	type alias WebhookSampleDelivery
	out := (*alias)(w)
	return unmarshalWithExtra(data, out)
}

// WebhookSamplePayloadResponse covers both variants returned by
// `/api/webhooks/endpoints/sample-payload/`: the single-event-type variant
// (when `event_type` is passed) and the all-types variant (when it is not).
// Callers can branch on which fields are populated. Mirrors the union
// `WebhookSamplePayloadSingleResponse | WebhookSamplePayloadAllResponse` in
// `tango-node/src/models/Webhooks.ts`.
type WebhookSamplePayloadResponse struct {
	// Single-event-type variant fields:
	EventType       *string                `json:"event_type,omitempty"`
	SampleDelivery  *WebhookSampleDelivery `json:"sample_delivery,omitempty"`
	SignatureHeader *string                `json:"signature_header,omitempty"`
	Note            *string                `json:"note,omitempty"`

	// All-types variant fields:
	Samples map[string]struct {
		SampleDelivery *WebhookSampleDelivery `json:"sample_delivery,omitempty"`
	} `json:"samples,omitempty"`
	Usage *string `json:"usage,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (w *WebhookSamplePayloadResponse) UnmarshalJSON(data []byte) error {
	type alias WebhookSamplePayloadResponse
	out := (*alias)(w)
	return unmarshalWithExtra(data, out)
}

// WebhookTestDeliveryResult is the response from TestWebhookEndpoint.
type WebhookTestDeliveryResult struct {
	Success        *bool          `json:"success,omitempty"`
	StatusCode     *int           `json:"status_code,omitempty"`
	ResponseTimeMs *int           `json:"response_time_ms,omitempty"`
	EndpointURL    *string        `json:"endpoint_url,omitempty"`
	Message        *string        `json:"message,omitempty"`
	Error          *string        `json:"error,omitempty"`
	ResponseBody   *string        `json:"response_body,omitempty"`
	TestPayload    map[string]any `json:"test_payload,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (w *WebhookTestDeliveryResult) UnmarshalJSON(data []byte) error {
	type alias WebhookTestDeliveryResult
	out := (*alias)(w)
	return unmarshalWithExtra(data, out)
}

// ---------------------------------------------------------------------------
// Webhook alerts (filter-based subscription convenience API)
// ---------------------------------------------------------------------------

// WebhookAlert is a filter-based subscription. Mirrors `WebhookAlert` in
// `tango-node/src/models/Webhooks.ts`. Note that the alerts API uses `name`
// + `filters` (whereas the canonical subscriptions API uses
// `subscription_name` + `filter_definition`); see the Node model file for
// the rationale.
type WebhookAlert struct {
	AlertID        *string        `json:"alert_id,omitempty"`
	Name           *string        `json:"name,omitempty"`
	QueryType      *string        `json:"query_type,omitempty"`
	Filters        map[string]any `json:"filters,omitempty"`
	Frequency      *string        `json:"frequency,omitempty"`
	CronExpression *string        `json:"cron_expression,omitempty"`
	Status         *string        `json:"status,omitempty"`
	CreatedAt      *string        `json:"created_at,omitempty"`
	LastCheckedAt  *string        `json:"last_checked_at,omitempty"`
	MatchCount     *int           `json:"match_count,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (w *WebhookAlert) UnmarshalJSON(data []byte) error {
	type alias WebhookAlert
	out := (*alias)(w)
	return unmarshalWithExtra(data, out)
}

// WebhookAlertCreateInput is the body for CreateWebhookAlert.
//
// `Name`, `QueryType`, and `Filters` are required. `QueryType` is SINGULAR
// (e.g. `"contract"`, not `"contracts"`). `Endpoint` is the UUID of the
// destination webhook; required for accounts with multiple endpoints,
// optional otherwise (the server auto-resolves a single endpoint).
type WebhookAlertCreateInput struct {
	Name           string         `json:"name"`
	QueryType      string         `json:"query_type"`
	Filters        map[string]any `json:"filters"`
	Frequency      *string        `json:"frequency,omitempty"`
	CronExpression *string        `json:"cron_expression,omitempty"`
	Endpoint       *string        `json:"endpoint,omitempty"`
}

// WebhookAlertUpdateInput is the patch body for UpdateWebhookAlert. Only
// `name`, `frequency`, `cron_expression`, and `is_active` are writable
// server-side; `query_type` and `filters` are read-only after creation.
type WebhookAlertUpdateInput struct {
	Name           *string `json:"name,omitempty"`
	Frequency      *string `json:"frequency,omitempty"`
	CronExpression *string `json:"cron_expression,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
}
