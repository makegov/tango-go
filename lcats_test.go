package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListLcatsRequiresOwner(t *testing.T) {
	cases := []struct {
		name string
		opts *ListLcatsOptions
	}{
		{"nil opts", nil},
		{"empty opts", &ListLcatsOptions{}},
		{"both blank", &ListLcatsOptions{UEI: "", IDVKey: ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := newTestClient(t, emptyListHandler)
			_, err := c.ListLcats(context.Background(), tc.opts)
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected *ValidationError, got %T: %v", err, err)
			}
		})
	}
}

func TestListLcatsDispatchesByOwner(t *testing.T) {
	t.Run("UEI dispatches to entity endpoint", func(t *testing.T) {
		var capturedURL string
		c, _ := newTestClient(t, captureURLHandler(&capturedURL))
		_, _ = c.ListLcats(context.Background(), &ListLcatsOptions{UEI: "UEI123"})
		assertPathContains(t, capturedURL, "/api/entities/UEI123/lcats/")
	})

	t.Run("IDVKey dispatches to IDV endpoint", func(t *testing.T) {
		var capturedURL string
		c, _ := newTestClient(t, captureURLHandler(&capturedURL))
		_, _ = c.ListLcats(context.Background(), &ListLcatsOptions{IDVKey: "IDV-001"})
		assertPathContains(t, capturedURL, "/api/idvs/IDV-001/lcats/")
	})

	t.Run("both set: UEI wins", func(t *testing.T) {
		var capturedURL string
		c, _ := newTestClient(t, captureURLHandler(&capturedURL))
		_, _ = c.ListLcats(context.Background(), &ListLcatsOptions{UEI: "U1", IDVKey: "I1"})
		assertPathContains(t, capturedURL, "/api/entities/U1/lcats/")
	})
}

func TestListLcatsForwardsFilters(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLHandler(&capturedURL))
	_, _ = c.ListLcats(context.Background(), &ListLcatsOptions{
		UEI: "UEI123",
		EntityLcatsOptions: EntityLcatsOptions{
			Ordering: "labor_category",
			Search:   "software engineer",
		},
	})
	assertQueryContains(t, capturedURL,
		map[string]string{
			"ordering": "labor_category",
			"search":   "software engineer",
		},
		nil,
	)
}

func TestIterateLcatsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateLcats(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
	// Iterator.Next() should surface the validation error.
	if it.Next() {
		t.Error("expected Next() to return false on missing owner")
	}
	var ve *ValidationError
	if !errors.As(it.Err(), &ve) {
		t.Errorf("expected *ValidationError from iter.Err(), got %T: %v", it.Err(), it.Err())
	}
}
