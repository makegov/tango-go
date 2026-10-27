//go:build integration

package integration

import (
	"context"
	"testing"

	tango "github.com/makegov/tango-go"
)

type assistanceGroup struct {
	name         string
	list         func(context.Context, *tango.ListAssistanceAwardsOptions) (*tango.PaginatedResponse[tango.Record], error)
	get          func(context.Context, string, *tango.GetEntityOptions) (*tango.AssistanceAwardRecord, error)
	transactions func(context.Context, string, *tango.AssistanceSubresourceOptions) (*tango.PaginatedResponse[tango.AssistanceTransaction], error)
	listings     func(context.Context, string, *tango.AssistanceSubresourceOptions) (*tango.PaginatedResponse[tango.AssistanceAwardListing], error)
}

func assistanceGroups(c *tango.Client) []assistanceGroup {
	return []assistanceGroup{
		{"grants", c.ListAssistanceGrants, c.GetAssistanceGrant, c.ListAssistanceGrantTransactions, c.ListAssistanceGrantListings},
		{"loans", c.ListAssistanceLoans, c.GetAssistanceLoan, c.ListAssistanceLoanTransactions, c.ListAssistanceLoanListings},
		{"direct_payments", c.ListAssistanceDirectPayments, c.GetAssistanceDirectPayment, c.ListAssistanceDirectPaymentTransactions, c.ListAssistanceDirectPaymentListings},
		{"other", c.ListAssistanceOther, c.GetAssistanceOther, c.ListAssistanceOtherTransactions, c.ListAssistanceOtherListings},
	}
}

func TestLiveAssistanceAwards(t *testing.T) {
	c := newLiveClient(t)
	for _, g := range assistanceGroups(c) {
		t.Run(g.name, func(t *testing.T) {
			page, err := g.list(ctx(t), &tango.ListAssistanceAwardsOptions{
				ListOptions: tango.ListOptions{Limit: 1},
				RecordType:  "2",
			})
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(page.Results) == 0 {
				t.Fatal("expected at least one non-aggregate award")
			}
			key, _ := page.Results[0]["key"].(string)

			award, err := g.get(ctx(t), key, nil)
			if err != nil {
				t.Fatalf("get %s: %v", key, err)
			}
			if award.Key == nil || *award.Key != key {
				t.Errorf("get %s: key round-trip failed, got %v", key, award.Key)
			}
			if award.AssistanceType == nil || award.AssistanceType.Code == nil {
				t.Errorf("get %s: default shape should carry assistance_type(code,description)", key)
			}
			if len(award.Extra) != 0 {
				t.Errorf("get %s: default shape fields missing from the struct: %v", key, award.Extra)
			}

			txs, err := g.transactions(ctx(t), key, &tango.AssistanceSubresourceOptions{Limit: 5})
			if err != nil {
				t.Fatalf("transactions %s: %v", key, err)
			}
			if len(txs.Results) == 0 || txs.Results[0].Key == nil {
				t.Errorf("transactions %s: expected keyed rows, got %d", key, len(txs.Results))
			}

			listings, err := g.listings(ctx(t), key, nil)
			if err != nil {
				t.Fatalf("listings %s: %v", key, err)
			}
			if len(listings.Results) == 0 || listings.Results[0].CFDANumber == nil {
				t.Errorf("listings %s: expected rows with a cfda_number, got %d", key, len(listings.Results))
			}
		})
	}
}

func TestLiveIterateAssistanceGrantsCursorFlow(t *testing.T) {
	c := newLiveClient(t)
	it := c.IterateAssistanceGrants(ctx(t), &tango.ListAssistanceAwardsOptions{
		ListOptions: tango.ListOptions{Limit: 2},
		CFDANumber:  "93.778",
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
	if count < 3 {
		t.Errorf("expected the iterator to cross a page boundary, got %d rows", count)
	}
}
