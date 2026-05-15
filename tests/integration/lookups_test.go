//go:build integration

package integration

import (
	"testing"

	tango "github.com/makegov/tango-go"
)

func TestLiveListAgencies(t *testing.T) {
	c := newLiveClient(t)
	page, err := c.ListAgencies(ctx(t), &tango.ListAgenciesOptions{Limit: 5})
	if err != nil {
		t.Fatalf("ListAgencies: %v", err)
	}
	if len(page.Results) == 0 {
		t.Error("expected at least one agency")
	}
}

func TestLiveGetAgency(t *testing.T) {
	c := newLiveClient(t)
	agency, err := c.GetAgency(ctx(t), "9700")
	if err != nil {
		t.Fatalf("GetAgency(9700): %v", err)
	}
	if agency == nil {
		t.Fatal("expected non-nil AgencyRecord")
	}
	if agency.Code == nil || *agency.Code != "9700" {
		t.Errorf("expected code=9700, got %v", agency.Code)
	}
}

func TestLiveListNAICS(t *testing.T) {
	c := newLiveClient(t)
	page, err := c.ListNAICS(ctx(t), &tango.ListNAICSOptions{
		ListOptions: tango.ListOptions{Limit: 5},
	})
	if err != nil {
		t.Fatalf("ListNAICS: %v", err)
	}
	if len(page.Results) == 0 {
		t.Error("expected at least one NAICS entry")
	}
}

func TestLiveGetNAICS(t *testing.T) {
	c := newLiveClient(t)
	rec, err := c.GetNAICS(ctx(t), "541512")
	if err != nil {
		t.Fatalf("GetNAICS(541512): %v", err)
	}
	if rec == nil {
		t.Error("expected non-nil result")
	}
}

func TestLiveGetVersion(t *testing.T) {
	c := newLiveClient(t)
	rec, err := c.GetVersion(ctx(t))
	if err != nil {
		t.Fatalf("GetVersion: %v", err)
	}
	if rec == nil {
		t.Error("expected non-nil version record")
	}
}
