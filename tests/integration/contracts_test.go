//go:build integration

package integration

import (
	"testing"

	tango "github.com/makegov/tango-go"
)

func TestLiveListContracts(t *testing.T) {
	c := newLiveClient(t)
	page, err := c.ListContracts(ctx(t), &tango.ListContractsOptions{
		ListOptions:    tango.ListOptions{Limit: 3},
		AwardingAgency: "9700",
	})
	if err != nil {
		t.Fatalf("ListContracts: %v", err)
	}
	if len(page.Results) == 0 {
		t.Error("expected at least one contract for agency 9700")
	}
	if page.Count == 0 {
		t.Error("expected non-zero count")
	}
}

func TestLiveIterateContractsCursorFlow(t *testing.T) {
	c := newLiveClient(t)
	it := c.IterateContracts(ctx(t), &tango.ListContractsOptions{
		ListOptions:    tango.ListOptions{Limit: 2},
		AwardingAgency: "9700",
	})

	var count int
	for it.Next() {
		_ = it.Item()
		count++
		if count >= 5 {
			break
		}
	}
	if err := it.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if count == 0 {
		t.Error("expected to iterate at least one contract")
	}
	t.Logf("iterated %d contracts via cursor flow", count)
}
