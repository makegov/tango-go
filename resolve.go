package tango

import "context"

// ResolveTargetType selects the resource the resolver searches against.
type ResolveTargetType string

const (
	// ResolveEntity targets the entity (SAM.gov vendor) resolver.
	ResolveEntity ResolveTargetType = "entity"
	// ResolveOrganization targets the federal organization resolver
	// (departments / agencies / sub-agencies / offices).
	ResolveOrganization ResolveTargetType = "organization"
)

// ResolveInput is the body POSTed to /api/resolve/.
type ResolveInput struct {
	Name       string            `json:"name"`
	TargetType ResolveTargetType `json:"target_type"`
	State      string            `json:"state,omitempty"`
	City       string            `json:"city,omitempty"`
	Context    string            `json:"context,omitempty"`
}

// ResolveCandidate is a single candidate returned by Resolve.
type ResolveCandidate struct {
	// Identifier is the canonical UEI (when TargetType is "entity") or
	// organization key (when TargetType is "organization").
	Identifier string `json:"identifier,omitempty"`

	// DisplayName is the human-readable name of the candidate.
	DisplayName string `json:"display_name,omitempty"`

	// MatchTier is the confidence label ("low" | "medium" | "high").
	// Pro+ tiers only — Free responses omit this.
	MatchTier string `json:"match_tier,omitempty"`

	// Extra holds any additional server-supplied metadata.
	Extra map[string]any `json:"extra,omitempty"`
}

// ResolveResult is the response from /api/resolve/.
type ResolveResult struct {
	// Count is the number of candidates returned (capped by tier:
	// Free=3, Pro+=5).
	Count      int                `json:"count,omitempty"`
	Candidates []ResolveCandidate `json:"candidates"`
}

// Resolve POSTs to /api/resolve/ to fuzzy-match a name to entity or
// organization candidates.
func (c *Client) Resolve(ctx context.Context, input ResolveInput) (*ResolveResult, error) {
	if input.Name == "" {
		return nil, &ValidationError{&APIError{Message: "Resolve: Name is required"}}
	}
	if input.TargetType != ResolveEntity && input.TargetType != ResolveOrganization {
		return nil, &ValidationError{&APIError{Message: "Resolve: TargetType must be \"entity\" or \"organization\""}}
	}
	result, err := postGeneric[*ResolveResult](ctx, c, "/api/resolve/", input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ValidateInputType is the identifier type checked by Validate.
type ValidateInputType string

const (
	// ValidatePIID validates the format of a PIID / contract identifier.
	ValidatePIID ValidateInputType = "piid"
	// ValidateSolicitation validates the format of a solicitation number.
	ValidateSolicitation ValidateInputType = "solicitation"
	// ValidateUEI validates the format and check digits of a UEI.
	ValidateUEI ValidateInputType = "uei"
)

// ValidateInput is the body POSTed to /api/validate/.
type ValidateInput struct {
	Type  ValidateInputType `json:"type"`
	Value string            `json:"value"`
}

// ValidateResult is the response from /api/validate/.
type ValidateResult struct {
	// Result is "valid", "invalid", or "low_confidence".
	Result string `json:"result"`
	Type   string `json:"type,omitempty"`
	Value  string `json:"value,omitempty"`
	// Errors carries structured failures when Result is non-valid.
	Errors []map[string]any `json:"errors,omitempty"`
}

// Validate POSTs to /api/validate/ to check whether an identifier is
// well-formed and known.
func (c *Client) Validate(ctx context.Context, input ValidateInput) (*ValidateResult, error) {
	if input.Value == "" {
		return nil, &ValidationError{&APIError{Message: "Validate: Value is required"}}
	}
	result, err := postGeneric[*ValidateResult](ctx, c, "/api/validate/", input)
	if err != nil {
		return nil, err
	}
	return result, nil
}
