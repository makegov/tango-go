package tango

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestValidateRequiresValue(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.Validate(context.Background(), ValidateInput{Type: ValidatePIID})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError for empty Value, got %T: %v", err, err)
	}
}

func TestValidateSendsCorrectBody(t *testing.T) {
	var capturedBody []byte
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result":"valid","type":"piid","value":"W15P7T19C0001"}`))
	})
	result, err := c.Validate(context.Background(), ValidateInput{
		Type:  ValidatePIID,
		Value: "W15P7T19C0001",
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if result.Result != "valid" {
		t.Errorf("expected result=valid, got %q", result.Result)
	}
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["type"] != "piid" {
		t.Errorf("body.type: want piid, got %v", body["type"])
	}
	if body["value"] != "W15P7T19C0001" {
		t.Errorf("body.value mismatch: %v", body["value"])
	}
}

func TestValidateAllInputTypes(t *testing.T) {
	cases := []ValidateInputType{
		ValidatePIID,
		ValidateSolicitation,
		ValidateUEI,
	}
	for _, typ := range cases {
		t.Run(string(typ), func(t *testing.T) {
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"result":"valid"}`))
			})
			result, err := c.Validate(context.Background(), ValidateInput{Type: typ, Value: "test-val"})
			if err != nil {
				t.Fatalf("unexpected error for type %q: %v", typ, err)
			}
			if result.Result != "valid" {
				t.Errorf("expected result=valid, got %q", result.Result)
			}
		})
	}
}

func TestValidateServerErrorPropagated(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	})
	_, err := c.Validate(context.Background(), ValidateInput{Type: ValidateUEI, Value: "UEI123"})
	var ae *AuthError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *AuthError, got %T: %v", err, err)
	}
}
