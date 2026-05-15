package tango

import (
	"context"
	"errors"
	"testing"
)

func TestListVehiclesFilterMapping(t *testing.T) {
	cases := []struct {
		name    string
		opts    *ListVehiclesOptions
		wantQS  map[string]string
		notInQS []string
	}{
		{
			name:    "nil opts",
			opts:    nil,
			wantQS:  map[string]string{},
			notInQS: []string{"search", "vehicle_type", "ordering"},
		},
		{
			name: "all string filters",
			opts: &ListVehiclesOptions{
				Joiner:                ".",
				Search:                "oasis",
				VehicleType:           "IDC",
				TypeOfIDC:             "GWAC",
				ContractType:          "FFP",
				SetAside:              "8A",
				WhoCanUse:             "All",
				NAICSCode:             "541512",
				PSCCode:               "D302",
				ProgramAcronym:        "OASIS",
				Agency:                "9700",
				OrganizationID:        "ORG001",
				TotalObligatedMin:     "1000000",
				TotalObligatedMax:     "999999999",
				FiscalYear:            "2024",
				AwardDateAfter:        "2020-01-01",
				AwardDateBefore:       "2024-12-31",
				LastDateToOrderAfter:  "2024-01-01",
				LastDateToOrderBefore: "2030-12-31",
				Ordering:              "total_obligated",
			},
			wantQS: map[string]string{
				"joiner":                    ".",
				"search":                    "oasis",
				"vehicle_type":              "IDC",
				"type_of_idc":               "GWAC",
				"contract_type":             "FFP",
				"set_aside":                 "8A",
				"who_can_use":               "All",
				"naics_code":                "541512",
				"psc_code":                  "D302",
				"program_acronym":           "OASIS",
				"agency":                    "9700",
				"organization_id":           "ORG001",
				"total_obligated_min":       "1000000",
				"total_obligated_max":       "999999999",
				"fiscal_year":               "2024",
				"award_date_after":          "2020-01-01",
				"award_date_before":         "2024-12-31",
				"last_date_to_order_after":  "2024-01-01",
				"last_date_to_order_before": "2030-12-31",
				"ordering":                  "total_obligated",
			},
		},
		{
			name: "int count filters",
			opts: &ListVehiclesOptions{
				IDVCountMin:   3,
				IDVCountMax:   50,
				OrderCountMin: 100,
				OrderCountMax: 9999,
			},
			wantQS: map[string]string{
				"idv_count_min":   "3",
				"idv_count_max":   "50",
				"order_count_min": "100",
				"order_count_max": "9999",
			},
		},
		{
			name:    "zero int counts omitted",
			opts:    &ListVehiclesOptions{IDVCountMin: 0, OrderCountMax: 0},
			notInQS: []string{"idv_count_min", "order_count_max"},
		},
		{
			name:   "extra map",
			opts:   &ListVehiclesOptions{Extra: map[string]any{"version": "v2"}},
			wantQS: map[string]string{"version": "v2"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedURL string
			c, _ := newTestClient(t, captureURLHandler(&capturedURL))
			_, _ = c.ListVehicles(context.Background(), tc.opts)
			assertQueryContains(t, capturedURL, tc.wantQS, tc.notInQS)
		})
	}
}

func TestGetVehicleRequiresUUID(t *testing.T) {
	c := NewClient(WithAPIKey("k"), WithBaseURL("http://localhost:0"), WithRetries(0))
	_, err := c.GetVehicle(context.Background(), "", nil)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestGetVehicleBuildsPath(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetVehicle(context.Background(), "vehicle-uuid-1234", nil)
	assertPathContains(t, capturedURL, "/api/vehicles/vehicle-uuid-1234/")
}

func TestGetVehicleWithOptions(t *testing.T) {
	var capturedURL string
	c, _ := newTestClient(t, captureURLRecordHandler(&capturedURL))
	_, _ = c.GetVehicle(context.Background(), "uuid1", &GetEntityOptions{
		Shape: "vehicles(minimal)", Flat: true,
	})
	assertQueryContains(t, capturedURL, map[string]string{
		"shape": "vehicles(minimal)",
		"flat":  "true",
	}, nil)
}

func TestIterateVehiclesNilOpts(t *testing.T) {
	c, _ := newTestClient(t, emptyListHandler)
	it := c.IterateVehicles(context.Background(), nil)
	if it == nil {
		t.Fatal("expected non-nil iterator")
	}
}
