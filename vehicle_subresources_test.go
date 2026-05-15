package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListVehicleAwardeesRequiresUUID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListVehicleAwardees(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListVehicleAwardeesBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListVehicleAwardees(context.Background(), "vehicle-uuid-001", nil)
	assertPathContains(t, capturedURL, "/api/vehicles/vehicle-uuid-001/awardees/")
}

func TestListVehicleAwardeesWithPagination(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListVehicleAwardees(context.Background(), "uuid1", &ListOptions{
		Limit: 25, Cursor: "cursor123",
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"limit":  "25",
		"cursor": "cursor123",
	}, nil)
}

func TestListVehicleOrdersRequiresUUID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.ListVehicleOrders(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestListVehicleOrdersBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListVehicleOrders(context.Background(), "vehicle-uuid-002", nil)
	assertPathContains(t, capturedURL, "/api/vehicles/vehicle-uuid-002/orders/")
}

func TestListVehicleOrdersNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	resp, err := c.ListVehicleOrders(context.Background(), "uuid1", nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}
