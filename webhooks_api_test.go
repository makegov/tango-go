package tango

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// ListWebhookEventTypes
// ---------------------------------------------------------------------------

func TestListWebhookEventTypesBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"event_types":[]}`))
	})
	_, _ = c.ListWebhookEventTypes(context.Background())
	assertPathContains(t, capturedURL, "/api/webhooks/event-types/")
}

func TestListWebhookEventTypesDecodesResponse(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"event_types":[{"event_type":"contract.awarded","schema_version":1}]}`))
	})
	resp, err := c.ListWebhookEventTypes(context.Background())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(resp.EventTypes) != 1 {
		t.Fatalf("expected 1 event type, got %d", len(resp.EventTypes))
	}
	if resp.EventTypes[0].EventType == nil || *resp.EventTypes[0].EventType != "contract.awarded" {
		t.Errorf("unexpected event type: %v", resp.EventTypes[0].EventType)
	}
}

// ---------------------------------------------------------------------------
// ListWebhookEndpoints
// ---------------------------------------------------------------------------

func TestListWebhookEndpointsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListWebhookEndpoints(context.Background(), nil)
	assertPathContains(t, capturedURL, "/api/webhooks/endpoints/")
}

func TestListWebhookEndpointsWithPagination(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListWebhookEndpoints(context.Background(), &ListOptions{Limit: 10})
	assertQueryContains(t, capturedURL, map[string]string{"limit": "10"}, nil)
}

// ---------------------------------------------------------------------------
// GetWebhookEndpoint
// ---------------------------------------------------------------------------

func TestGetWebhookEndpointRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetWebhookEndpoint(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetWebhookEndpointBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"ep1","name":"Test"}`))
	})
	_, _ = c.GetWebhookEndpoint(context.Background(), "ep-uuid-1234")
	assertPathContains(t, capturedURL, "/api/webhooks/endpoints/ep-uuid-1234/")
}

func TestGetWebhookEndpointPathEscape(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"ep/1"}`))
	})
	_, _ = c.GetWebhookEndpoint(context.Background(), "ep/special")
	assertPathContains(t, capturedURL, "ep%2Fspecial")
}

// ---------------------------------------------------------------------------
// CreateWebhookEndpoint
// ---------------------------------------------------------------------------

func TestCreateWebhookEndpointRequiresName(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.CreateWebhookEndpoint(context.Background(), WebhookEndpointCreateInput{
		CallbackURL: "https://example.com/wh",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty Name, got %T: %v", err, err)
	}
}

func TestCreateWebhookEndpointRequiresCallbackURL(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.CreateWebhookEndpoint(context.Background(), WebhookEndpointCreateInput{
		Name: "My Hook",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty CallbackURL, got %T: %v", err, err)
	}
}

func TestCreateWebhookEndpointSendsCorrectBody(t *testing.T) {
	var capturedBody []byte
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"ep1","name":"Prod Hook"}`))
	})
	active := true
	_, err := c.CreateWebhookEndpoint(context.Background(), WebhookEndpointCreateInput{
		Name:        "Prod Hook",
		CallbackURL: "https://prod.example.com/webhook",
		IsActive:    &active,
		EventTypes:  []string{"contract.awarded"},
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("failed to decode captured body: %v", err)
	}
	if body["name"] != "Prod Hook" {
		t.Errorf("body.name: want %q, got %v", "Prod Hook", body["name"])
	}
	if body["callback_url"] != "https://prod.example.com/webhook" {
		t.Errorf("body.callback_url mismatch: %v", body["callback_url"])
	}
	if body["is_active"] != true {
		t.Errorf("body.is_active: want true, got %v", body["is_active"])
	}
}

func TestCreateWebhookEndpointBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"ep1"}`))
	})
	_, _ = c.CreateWebhookEndpoint(context.Background(), WebhookEndpointCreateInput{
		Name:        "Hook",
		CallbackURL: "https://example.com",
	})
	assertPathContains(t, capturedURL, "/api/webhooks/endpoints/")
}

// ---------------------------------------------------------------------------
// UpdateWebhookEndpoint
// ---------------------------------------------------------------------------

func TestUpdateWebhookEndpointRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.UpdateWebhookEndpoint(context.Background(), "", WebhookEndpointUpdateInput{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty id, got %T: %v", err, err)
	}
}

func TestUpdateWebhookEndpointSendsPatch(t *testing.T) {
	var method string
	var capturedBody []byte
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"ep1","name":"Updated"}`))
	})
	newName := "Updated Hook"
	active := false
	_, err := c.UpdateWebhookEndpoint(context.Background(), "ep-uuid", WebhookEndpointUpdateInput{
		Name:     &newName,
		IsActive: &active,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if method != "PATCH" {
		t.Errorf("expected PATCH, got %s", method)
	}
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["name"] != "Updated Hook" {
		t.Errorf("body.name: want %q, got %v", "Updated Hook", body["name"])
	}
}

func TestUpdateWebhookEndpointBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"ep1"}`))
	})
	_, _ = c.UpdateWebhookEndpoint(context.Background(), "ep-uuid-99", WebhookEndpointUpdateInput{})
	assertPathContains(t, capturedURL, "/api/webhooks/endpoints/ep-uuid-99/")
}

// ---------------------------------------------------------------------------
// DeleteWebhookEndpoint
// ---------------------------------------------------------------------------

func TestDeleteWebhookEndpointRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	err := c.DeleteWebhookEndpoint(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty id, got %T: %v", err, err)
	}
}

func TestDeleteWebhookEndpointSendsDelete(t *testing.T) {
	var method string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.WriteHeader(204)
	})
	err := c.DeleteWebhookEndpoint(context.Background(), "ep-uuid")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if method != "DELETE" {
		t.Errorf("expected DELETE, got %s", method)
	}
}

func TestDeleteWebhookEndpointBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.WriteHeader(204)
	})
	_ = c.DeleteWebhookEndpoint(context.Background(), "ep-to-delete")
	assertPathContains(t, capturedURL, "/api/webhooks/endpoints/ep-to-delete/")
}

// ---------------------------------------------------------------------------
// TestWebhookEndpoint
// ---------------------------------------------------------------------------

func TestTestWebhookEndpointRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.TestWebhookEndpoint(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty endpointID, got %T: %v", err, err)
	}
}

func TestTestWebhookEndpointSendsCorrectBody(t *testing.T) {
	var capturedBody []byte
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true,"status_code":200}`))
	})
	_, err := c.TestWebhookEndpoint(context.Background(), "endpoint-id-abc")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["endpoint"] != "endpoint-id-abc" {
		t.Errorf("body.endpoint: want %q, got %v", "endpoint-id-abc", body["endpoint"])
	}
}

// ---------------------------------------------------------------------------
// GetWebhookSamplePayload
// ---------------------------------------------------------------------------

func TestGetWebhookSamplePayloadBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})
	_, _ = c.GetWebhookSamplePayload(context.Background(), "")
	assertPathContains(t, capturedURL, "/api/webhooks/endpoints/sample-payload/")
}

func TestGetWebhookSamplePayloadWithEventType(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"event_type":"contract.awarded"}`))
	})
	_, _ = c.GetWebhookSamplePayload(context.Background(), "contract.awarded")
	assertQueryContains(t, capturedURL, map[string]string{"event_type": "contract.awarded"}, nil)
}

func TestGetWebhookSamplePayloadNoEventTypeNoQParam(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})
	_, _ = c.GetWebhookSamplePayload(context.Background(), "")
	assertQueryContains(t, capturedURL, nil, []string{"event_type"})
}

// ---------------------------------------------------------------------------
// ListWebhookAlerts
// ---------------------------------------------------------------------------

func TestListWebhookAlertsBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListWebhookAlerts(context.Background(), nil)
	assertPathContains(t, capturedURL, "/api/webhooks/alerts/")
}

// ---------------------------------------------------------------------------
// GetWebhookAlert
// ---------------------------------------------------------------------------

func TestGetWebhookAlertRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetWebhookAlert(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetWebhookAlertBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"alert_id":"al1"}`))
	})
	_, _ = c.GetWebhookAlert(context.Background(), "alert-uuid-99")
	assertPathContains(t, capturedURL, "/api/webhooks/alerts/alert-uuid-99/")
}

// ---------------------------------------------------------------------------
// CreateWebhookAlert
// ---------------------------------------------------------------------------

func TestCreateWebhookAlertRequiresName(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.CreateWebhookAlert(context.Background(), WebhookAlertCreateInput{
		QueryType: "contract",
		Filters:   map[string]any{"awarding_agency": "9700"},
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty Name, got %T: %v", err, err)
	}
}

func TestCreateWebhookAlertRequiresQueryType(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.CreateWebhookAlert(context.Background(), WebhookAlertCreateInput{
		Name:    "My Alert",
		Filters: map[string]any{"awarding_agency": "9700"},
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty QueryType, got %T: %v", err, err)
	}
}

func TestCreateWebhookAlertRequiresFilters(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.CreateWebhookAlert(context.Background(), WebhookAlertCreateInput{
		Name:      "My Alert",
		QueryType: "contract",
	})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty Filters, got %T: %v", err, err)
	}
}

func TestCreateWebhookAlertSendsCorrectBody(t *testing.T) {
	var capturedBody []byte
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"alert_id":"al1","name":"DoD Contracts"}`))
	})
	_, err := c.CreateWebhookAlert(context.Background(), WebhookAlertCreateInput{
		Name:      "DoD Contracts",
		QueryType: "contract",
		Filters:   map[string]any{"awarding_agency": "9700"},
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["name"] != "DoD Contracts" {
		t.Errorf("body.name: want %q, got %v", "DoD Contracts", body["name"])
	}
	if body["query_type"] != "contract" {
		t.Errorf("body.query_type: want %q, got %v", "contract", body["query_type"])
	}
	filters, ok := body["filters"].(map[string]any)
	if !ok || filters["awarding_agency"] != "9700" {
		t.Errorf("body.filters mismatch: %v", body["filters"])
	}
}

// ---------------------------------------------------------------------------
// UpdateWebhookAlert
// ---------------------------------------------------------------------------

func TestUpdateWebhookAlertRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.UpdateWebhookAlert(context.Background(), "", WebhookAlertUpdateInput{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty id, got %T: %v", err, err)
	}
}

func TestUpdateWebhookAlertBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"alert_id":"al1"}`))
	})
	_, _ = c.UpdateWebhookAlert(context.Background(), "alert-to-update", WebhookAlertUpdateInput{})
	assertPathContains(t, capturedURL, "/api/webhooks/alerts/alert-to-update/")
}

// ---------------------------------------------------------------------------
// DeleteWebhookAlert
// ---------------------------------------------------------------------------

func TestDeleteWebhookAlertRequiresID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	err := c.DeleteWebhookAlert(context.Background(), "")
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty id, got %T: %v", err, err)
	}
}

func TestDeleteWebhookAlertBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.WriteHeader(204)
	})
	_ = c.DeleteWebhookAlert(context.Background(), "alert-to-kill")
	assertPathContains(t, capturedURL, "/api/webhooks/alerts/alert-to-kill/")
}

func TestDeleteWebhookAlertSendsDelete(t *testing.T) {
	var method string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.WriteHeader(204)
	})
	_ = c.DeleteWebhookAlert(context.Background(), "al-uuid")
	if method != "DELETE" {
		t.Errorf("expected DELETE, got %s", method)
	}
}

func TestWebhookAlertPathEscape(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"alert_id":"al/1"}`))
	})
	_, _ = c.GetWebhookAlert(context.Background(), "al/special")
	if !strings.Contains(capturedURL, "al%2Fspecial") {
		t.Errorf("expected path to contain al%%2Fspecial, got %q", capturedURL)
	}
}
