package tango

import (
	"context"
	"net/url"
)

// ListVehicleAwardees lists the awardees (entities holding child IDVs)
// under a contracting vehicle (/api/vehicles/{uuid}/awardees/).
func (c *Client) ListVehicleAwardees(ctx context.Context, uuid string, opts *ListOptions) (*PaginatedResponse[Record], error) {
	if uuid == "" {
		return nil, &ValidationError{&APIError{Message: "vehicle uuid is required"}}
	}
	q := url.Values{}
	if opts != nil {
		opts.applyTo(q)
	}
	return listGeneric[Record](ctx, c, "/api/vehicles/"+pathEscape(uuid)+"/awardees/", q)
}

// ListVehicleOrders lists task orders placed under a contracting
// vehicle's child IDVs (/api/vehicles/{uuid}/orders/).
//
// This endpoint exists in the Python SDK (list_vehicle_orders) but not
// the Node SDK; it's included here for full parity.
func (c *Client) ListVehicleOrders(ctx context.Context, uuid string, opts *ListOptions) (*PaginatedResponse[Record], error) {
	if uuid == "" {
		return nil, &ValidationError{&APIError{Message: "vehicle uuid is required"}}
	}
	q := url.Values{}
	if opts != nil {
		opts.applyTo(q)
	}
	return listGeneric[Record](ctx, c, "/api/vehicles/"+pathEscape(uuid)+"/orders/", q)
}
