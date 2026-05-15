package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListGsaElibraryContractsFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListGsaElibraryContractsOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"schedule", "piid", "search"},
		},
		{
			name: "all filters",
			opts: &ListGsaElibraryContractsOptions{
				Schedule:       "MAS",
				ContractNumber: "47QSMA22D0014",
				Key:            "key-abc",
				PIID:           "47QSMA22D0014",
				UEI:            "UEI12345",
				SIN:            "54151S",
				Search:         "cloud services",
				Ordering:       "contract_number",
			},
			wantQS: map[string]string{
				"schedule":        "MAS",
				"contract_number": "47QSMA22D0014",
				"key":             "key-abc",
				"piid":            "47QSMA22D0014",
				"uei":             "UEI12345",
				"sin":             "54151S",
				"search":          "cloud services",
				"ordering":        "contract_number",
			},
		},
		{
			name:    "zero values omitted",
			opts:    &ListGsaElibraryContractsOptions{},
			notInQS: []string{"schedule", "piid", "search"},
		},
		{
			name: "extra map",
			opts: &ListGsaElibraryContractsOptions{Extra: map[string]any{"active": "true"}},
			wantQS: map[string]string{"active": "true"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListGsaElibraryContracts(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetGsaElibraryContractRequiresUUID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetGsaElibraryContract(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetGsaElibraryContractBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetGsaElibraryContract(context.Background(), "uuid-gsa-1234", nil)
	assertPathContains(t, capturedURL, "/api/gsa_elibrary_contracts/uuid-gsa-1234/")
}

func TestGetGsaElibraryContractWithOptions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetGsaElibraryContract(context.Background(), "uuid1", &GetEntityOptions{
		Flat: true, FlatLists: true,
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"flat":       "true",
		"flat_lists": "true",
	}, nil)
}

func TestIterateGsaElibraryContractsNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateGsaElibraryContracts(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}
