//go:build integration

package integration

import (
	"testing"

	tango "github.com/makegov/tango-go"
)

func TestLiveResolveOrganization(t *testing.T) {
	c := newLiveClient(t)
	res, err := c.Resolve(ctx(t), tango.ResolveInput{
		Name:       "Department of Defense",
		TargetType: tango.ResolveOrganization,
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
	if len(res.Candidates) == 0 {
		t.Error("expected at least one candidate for 'Department of Defense'")
	}
}

func TestLiveResolveEntity(t *testing.T) {
	c := newLiveClient(t)
	// Use a well-known large contractor
	res, err := c.Resolve(ctx(t), tango.ResolveInput{
		Name:       "Lockheed Martin",
		TargetType: tango.ResolveEntity,
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
	if len(res.Candidates) == 0 {
		t.Error("expected at least one candidate for 'Lockheed Martin'")
	}
}
