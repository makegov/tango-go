package tango

import (
	"encoding/json"
	"testing"
)

// TestAgencyRecordUnmarshalKnownFields verifies that known fields decode
// into named struct fields.
func TestAgencyRecordUnmarshalKnownFields(t *testing.T) {
	raw := []byte(`{"agency_id":"A001","name":"Test Agency","abbreviation":"TA","code":"9700","department":{"name":"DoD"}}`)
	var rec AgencyRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	assertStrPtr(t, "agency_id", rec.AgencyID, "A001")
	assertStrPtr(t, "name", rec.Name, "Test Agency")
	assertStrPtr(t, "abbreviation", rec.Abbreviation, "TA")
	assertStrPtr(t, "code", rec.Code, "9700")
	if rec.Department["name"] != "DoD" {
		t.Errorf("department name: want DoD, got %v", rec.Department["name"])
	}
}

// TestAgencyRecordUnmarshalExtra verifies that unknown fields land in Extra.
func TestAgencyRecordUnmarshalExtra(t *testing.T) {
	raw := []byte(`{"agency_id":"A001","name":"Test","unknown_future_field":"future_val","nested_new":{"a":1}}`)
	var rec AgencyRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rec.Extra == nil {
		t.Fatal("expected Extra to be populated with unknown fields")
	}
	if rec.Extra["unknown_future_field"] != "future_val" {
		t.Errorf("Extra[unknown_future_field]: want %q, got %v", "future_val", rec.Extra["unknown_future_field"])
	}
	if rec.Extra["nested_new"] == nil {
		t.Error("expected Extra[nested_new] to be present")
	}
}

// TestAgencyRecordKnownFieldsNotInExtra verifies that known fields do NOT
// appear in Extra.
func TestAgencyRecordKnownFieldsNotInExtra(t *testing.T) {
	raw := []byte(`{"agency_id":"A001","name":"Test"}`)
	var rec AgencyRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rec.Extra != nil {
		for k := range rec.Extra {
			if k == "agency_id" || k == "name" {
				t.Errorf("known field %q should not appear in Extra", k)
			}
		}
	}
}

// TestWebhookEndpointUnmarshalKnownFields tests the WebhookEndpoint model.
func TestWebhookEndpointUnmarshalKnownFields(t *testing.T) {
	raw := []byte(`{"id":"ep1","name":"My Hook","callback_url":"https://example.com/wh","secret":"s3cr3t","is_active":true,"created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-02T00:00:00Z"}`)
	var ep WebhookEndpoint
	if err := json.Unmarshal(raw, &ep); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	assertStrPtr(t, "id", ep.ID, "ep1")
	assertStrPtr(t, "name", ep.Name, "My Hook")
	assertStrPtr(t, "callback_url", ep.CallbackURL, "https://example.com/wh")
	if ep.IsActive == nil || !*ep.IsActive {
		t.Error("expected IsActive=true")
	}
}

// TestWebhookEndpointUnmarshalExtra verifies forward-compat Extra capture.
func TestWebhookEndpointUnmarshalExtra(t *testing.T) {
	raw := []byte(`{"id":"ep1","name":"hook","new_field_v2":"surprise"}`)
	var ep WebhookEndpoint
	if err := json.Unmarshal(raw, &ep); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ep.Extra == nil || ep.Extra["new_field_v2"] != "surprise" {
		t.Errorf("Extra[new_field_v2]: want %q, got %v", "surprise", ep.Extra["new_field_v2"])
	}
}

// TestWebhookEventTypeUnmarshalExtra verifies Extra on WebhookEventType.
func TestWebhookEventTypeUnmarshalExtra(t *testing.T) {
	raw := []byte(`{"event_type":"contract.awarded","description":"A contract was awarded","schema_version":2,"beta_flag":true}`)
	var et WebhookEventType
	if err := json.Unmarshal(raw, &et); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	assertStrPtr(t, "event_type", et.EventType, "contract.awarded")
	if et.SchemaVersion == nil || *et.SchemaVersion != 2 {
		t.Errorf("expected SchemaVersion=2")
	}
	if et.Extra == nil || et.Extra["beta_flag"] == nil {
		t.Error("expected Extra[beta_flag] to be present")
	}
}

// TestWebhookAlertUnmarshalExtra verifies Extra on WebhookAlert.
func TestWebhookAlertUnmarshalExtra(t *testing.T) {
	raw := []byte(`{"alert_id":"al1","name":"My Alert","query_type":"contract","future_key":"fv"}`)
	var a WebhookAlert
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	assertStrPtr(t, "alert_id", a.AlertID, "al1")
	if a.Extra == nil || a.Extra["future_key"] != "fv" {
		t.Errorf("Extra[future_key]: want %q, got %v", "fv", a.Extra["future_key"])
	}
}

// TestWebhookSampleDeliveryUnmarshalExtra verifies Extra on WebhookSampleDelivery.
func TestWebhookSampleDeliveryUnmarshalExtra(t *testing.T) {
	raw := []byte(`{"timestamp":"2024-01-01T00:00:00Z","events":[],"new_key":"nv"}`)
	var sd WebhookSampleDelivery
	if err := json.Unmarshal(raw, &sd); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sd.Extra == nil || sd.Extra["new_key"] != "nv" {
		t.Errorf("Extra[new_key]: want %q, got %v", "nv", sd.Extra["new_key"])
	}
}

// TestWebhookEventTypesResponseUnmarshalExtra verifies Extra.
func TestWebhookEventTypesResponseUnmarshalExtra(t *testing.T) {
	raw := []byte(`{"event_types":[],"meta_v2":"val"}`)
	var r WebhookEventTypesResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if r.Extra == nil || r.Extra["meta_v2"] != "val" {
		t.Errorf("Extra[meta_v2]: want %q, got %v", "val", r.Extra["meta_v2"])
	}
}

// TestWebhookTestDeliveryResultUnmarshalExtra verifies Extra.
func TestWebhookTestDeliveryResultUnmarshalExtra(t *testing.T) {
	raw := []byte(`{"success":true,"status_code":200,"new_v2":"hello"}`)
	var r WebhookTestDeliveryResult
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if r.Success == nil || !*r.Success {
		t.Error("expected Success=true")
	}
	if r.Extra == nil || r.Extra["new_v2"] != "hello" {
		t.Errorf("Extra[new_v2]: want %q, got %v", "hello", r.Extra["new_v2"])
	}
}

// TestWebhookSamplePayloadResponseUnmarshalExtra verifies Extra.
func TestWebhookSamplePayloadResponseUnmarshalExtra(t *testing.T) {
	raw := []byte(`{"event_type":"contract.awarded","note":"test note","future_x":"y"}`)
	var r WebhookSamplePayloadResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	assertStrPtr(t, "note", r.Note, "test note")
	if r.Extra == nil || r.Extra["future_x"] != "y" {
		t.Errorf("Extra[future_x]: want %q, got %v", "y", r.Extra["future_x"])
	}
}

// TestUnmarshalWithExtraEmptyObject ensures empty object doesn't panic.
func TestUnmarshalWithExtraEmptyObject(t *testing.T) {
	raw := []byte(`{}`)
	var rec AgencyRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestUnmarshalWithExtraRoundTrip encodes and re-decodes to verify field
// round-trip.
func TestUnmarshalWithExtraRoundTrip(t *testing.T) {
	id := "ep99"
	name := "Round Trip Hook"
	active := true
	orig := WebhookEndpoint{
		ID:       &id,
		Name:     &name,
		IsActive: &active,
	}
	encoded, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded WebhookEndpoint
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	assertStrPtr(t, "id", decoded.ID, id)
	assertStrPtr(t, "name", decoded.Name, name)
	if decoded.IsActive == nil || *decoded.IsActive != active {
		t.Errorf("IsActive round-trip failed: got %v", decoded.IsActive)
	}
}

// assertStrPtr checks a *string field has the expected value.
func assertStrPtr(t *testing.T, field string, got *string, want string) {
	t.Helper()
	if got == nil {
		t.Errorf("field %q: want %q, got nil", field, want)
		return
	}
	if *got != want {
		t.Errorf("field %q: want %q, got %q", field, want, *got)
	}
}
