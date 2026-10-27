package tango

import "encoding/json"

// AssistanceAwardRecord is the typed return of GetAssistanceGrant, GetAssistanceLoan, GetAssistanceDirectPayment and GetAssistanceOther.
//
// Its named fields are the endpoints' default shape.
// Every field is a pointer (or a slice) so nil means "absent": a shape that leaves a field out, or a null the server sent, never reads as "" or 0.
// Fields a shape adds beyond the default (summary, listings(...), transactions, obligations_by_fiscal_year(...) and the rest) land in Extra.
type AssistanceAwardRecord struct {
	// Key is the canonical award identifier every detail and sub-route URL takes.
	Key *string `json:"key,omitempty"`
	// FAIN is set on non-aggregate records and URI on aggregate ones.
	FAIN *string `json:"fain,omitempty"`
	URI  *string `json:"uri,omitempty"`

	// AwardDate is YYYY-MM-DD.
	AwardDate  *string `json:"award_date,omitempty"`
	FiscalYear *int    `json:"fiscal_year,omitempty"`

	AssistanceType *AssistanceTypeRef `json:"assistance_type,omitempty"`
	// RecordType is "1" (aggregate), "2" (non-aggregate) or "3" (non-aggregate to an individual, PII-redacted).
	RecordType *string `json:"record_type,omitempty"`

	// CFDANumber is the primary assistance listing number; AssistanceListing adds its catalog title and is nil when the catalog does not list the number.
	CFDANumber        *string               `json:"cfda_number,omitempty"`
	AssistanceListing *AssistanceListingRef `json:"assistance_listing,omitempty"`

	Title *string `json:"title,omitempty"`

	// PragmaticObligations is the money column comparable across the four endpoints: the obligation on grants and payments, the subsidy cost on loans.
	Obligated            *float64 `json:"obligated,omitempty"`
	PragmaticObligations *float64 `json:"pragmatic_obligations,omitempty"`
	FaceValueOfLoan      *float64 `json:"face_value_of_loan,omitempty"`
	LoanSubsidyCost      *float64 `json:"loan_subsidy_cost,omitempty"`

	AwardingOffice *AwardOffice `json:"awarding_office,omitempty"`
	FundingOffice  *AwardOffice `json:"funding_office,omitempty"`

	// Recipient is nil on aggregate and PII-redacted records (record types "1" and "3").
	Recipient *AssistanceRecipient `json:"recipient,omitempty"`

	PlaceOfPerformance  *AssistancePlaceOfPerformance  `json:"place_of_performance,omitempty"`
	PeriodOfPerformance *AssistancePeriodOfPerformance `json:"period_of_performance,omitempty"`

	// Extra captures any field outside the default shape. Not serialized on the way out.
	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AssistanceAwardRecord) UnmarshalJSON(data []byte) error {
	type alias AssistanceAwardRecord
	return unmarshalWithExtra(data, (*alias)(a))
}

// AssistanceTypeRef is an award's GSDM assistance type, such as {"code":"F001","description":"Grant"}.
type AssistanceTypeRef struct {
	Code        *string `json:"code,omitempty"`
	Description *string `json:"description,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON accepts the expand's object and also the bare code string a shape naming the assistance_type leaf returns, so that shape fills only Code.
func (a *AssistanceTypeRef) UnmarshalJSON(data []byte) error {
	var code string
	if err := json.Unmarshal(data, &code); err == nil {
		a.Code = &code
		return nil
	}
	type alias AssistanceTypeRef
	return unmarshalWithExtra(data, (*alias)(a))
}

// AssistanceListingRef is an assistance listing (CFDA) number with its title from the catalog ListAssistanceListings serves.
type AssistanceListingRef struct {
	Number *string `json:"number,omitempty"`
	Title  *string `json:"title,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AssistanceListingRef) UnmarshalJSON(data []byte) error {
	type alias AssistanceListingRef
	return unmarshalWithExtra(data, (*alias)(a))
}

// AwardOffice is an award's awarding or funding office, resolved to Tango's organization tree.
type AwardOffice struct {
	OrganizationID *string `json:"organization_id,omitempty"`
	OfficeCode     *string `json:"office_code,omitempty"`
	OfficeName     *string `json:"office_name,omitempty"`
	AgencyCode     *string `json:"agency_code,omitempty"`
	AgencyName     *string `json:"agency_name,omitempty"`
	DepartmentCode *string `json:"department_code,omitempty"`
	DepartmentName *string `json:"department_name,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AwardOffice) UnmarshalJSON(data []byte) error {
	type alias AwardOffice
	return unmarshalWithExtra(data, (*alias)(a))
}

// AssistanceRecipient is an assistance award's recipient: one shared record per recipient, shaped like an entity, whether or not it has a UEI.
// Its values all come from the recipient's latest filing.
type AssistanceRecipient struct {
	// UUID is Tango's recipient identifier; for a recipient with a UEI it is the same UUID the entity carries.
	UUID                          *string                           `json:"uuid,omitempty"`
	UEI                           *string                           `json:"uei,omitempty"`
	DUNS                          *string                           `json:"duns,omitempty"`
	LegalBusinessName             *string                           `json:"legal_business_name,omitempty"`
	LegalBusinessNameRaw          *string                           `json:"legal_business_name_raw,omitempty"`
	County                        *string                           `json:"county,omitempty"`
	CongressionalDistrict         *string                           `json:"congressional_district,omitempty"`
	CongressionalDistrictOriginal *string                           `json:"congressional_district_original,omitempty"`
	HighestOwner                  *AssistanceRecipientOwner         `json:"highest_owner,omitempty"`
	BusinessTypes                 []AssistanceRecipientBusinessType `json:"business_types,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AssistanceRecipient) UnmarshalJSON(data []byte) error {
	type alias AssistanceRecipient
	return unmarshalWithExtra(data, (*alias)(a))
}

// AssistanceRecipientOwner is a recipient's parent organization.
type AssistanceRecipientOwner struct {
	UEI                  *string `json:"uei,omitempty"`
	DUNS                 *string `json:"duns,omitempty"`
	LegalBusinessName    *string `json:"legal_business_name,omitempty"`
	LegalBusinessNameRaw *string `json:"legal_business_name_raw,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AssistanceRecipientOwner) UnmarshalJSON(data []byte) error {
	type alias AssistanceRecipientOwner
	return unmarshalWithExtra(data, (*alias)(a))
}

// AssistanceRecipientBusinessType is one of a recipient's business types, in USAspending's recipient-type codes rather than SAM.gov's.
type AssistanceRecipientBusinessType struct {
	Code        *string `json:"code,omitempty"`
	Description *string `json:"description,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AssistanceRecipientBusinessType) UnmarshalJSON(data []byte) error {
	type alias AssistanceRecipientBusinessType
	return unmarshalWithExtra(data, (*alias)(a))
}

// AssistancePlaceOfPerformance is where an assistance award is performed.
// Code and Scope tell a county-wide place from a state-wide one.
type AssistancePlaceOfPerformance struct {
	// Country is an ISO 3166 alpha-3 code.
	Country         *string `json:"country,omitempty"`
	StateCode       *string `json:"state_code,omitempty"`
	StateName       *string `json:"state_name,omitempty"`
	StateFIPS       *string `json:"state_fips,omitempty"`
	CountyFIPS      *string `json:"county_fips,omitempty"`
	CountyName      *string `json:"county_name,omitempty"`
	CityName        *string `json:"city_name,omitempty"`
	ZipCode         *string `json:"zip_code,omitempty"`
	ZipPlusFour     *string `json:"zip_plus_four,omitempty"`
	Code            *string `json:"code,omitempty"`
	Scope           *string `json:"scope,omitempty"`
	ForeignLocation *string `json:"foreign_location,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AssistancePlaceOfPerformance) UnmarshalJSON(data []byte) error {
	type alias AssistancePlaceOfPerformance
	return unmarshalWithExtra(data, (*alias)(a))
}

// AssistancePeriodOfPerformance is an assistance award's period of performance (YYYY-MM-DD).
type AssistancePeriodOfPerformance struct {
	StartDate      *string `json:"start_date,omitempty"`
	CurrentEndDate *string `json:"current_end_date,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AssistancePeriodOfPerformance) UnmarshalJSON(data []byte) error {
	type alias AssistancePeriodOfPerformance
	return unmarshalWithExtra(data, (*alias)(a))
}

// AssistanceTransaction is one action on an assistance award, as the transactions sub-routes return it.
type AssistanceTransaction struct {
	Key                *string `json:"key,omitempty"`
	ModificationNumber *string `json:"modification_number,omitempty"`
	// TransactionDate is YYYY-MM-DD.
	TransactionDate *string `json:"transaction_date,omitempty"`
	// ActionType is a GSDM action type code such as "A1" (new award); it says what the action was, not whether money moved.
	ActionType                     *string  `json:"action_type,omitempty"`
	AssistanceTypeCode             *string  `json:"assistance_type_code,omitempty"`
	CorrectionDeleteIndicator      *string  `json:"correction_delete_indicator,omitempty"`
	CFDANumber                     *string  `json:"cfda_number,omitempty"`
	Obligated                      *float64 `json:"obligated,omitempty"`
	PragmaticObligations           *float64 `json:"pragmatic_obligations,omitempty"`
	NonFederalFundingAmount        *float64 `json:"non_federal_funding_amount,omitempty"`
	IndirectCostFederalShareAmount *float64 `json:"indirect_cost_federal_share_amount,omitempty"`
	FaceValueOfLoan                *float64 `json:"face_value_of_loan,omitempty"`
	LoanSubsidyCost                *float64 `json:"loan_subsidy_cost,omitempty"`
	AwardAmountBasis               *string  `json:"award_amount_basis,omitempty"`
	AwardRecipientBasis            *string  `json:"award_recipient_basis,omitempty"`
	ActionDateFiscalYear           *int     `json:"action_date_fiscal_year,omitempty"`
	LastModifiedDate               *string  `json:"last_modified_date,omitempty"`
	InitialReportDate              *string  `json:"initial_report_date,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AssistanceTransaction) UnmarshalJSON(data []byte) error {
	type alias AssistanceTransaction
	return unmarshalWithExtra(data, (*alias)(a))
}

// AssistanceAwardListing is an assistance award's money under one assistance listing, as the listings sub-routes return it.
// An award funded under several programs has one per program; IsPrimary marks the one the award's cfda_number names.
type AssistanceAwardListing struct {
	CFDANumber                     *string  `json:"cfda_number,omitempty"`
	IsPrimary                      *bool    `json:"is_primary,omitempty"`
	Obligated                      *float64 `json:"obligated,omitempty"`
	PragmaticObligations           *float64 `json:"pragmatic_obligations,omitempty"`
	NonFederalFundingAmount        *float64 `json:"non_federal_funding_amount,omitempty"`
	IndirectCostFederalShareAmount *float64 `json:"indirect_cost_federal_share_amount,omitempty"`
	FaceValueOfLoan                *float64 `json:"face_value_of_loan,omitempty"`
	LoanSubsidyCost                *float64 `json:"loan_subsidy_cost,omitempty"`
	TransactionsCount              *int     `json:"transactions_count,omitempty"`
	// FirstActionDate and LastActionDate are YYYY-MM-DD.
	FirstActionDate *string `json:"first_action_date,omitempty"`
	LastActionDate  *string `json:"last_action_date,omitempty"`

	Extra map[string]any `json:"-"`
}

// UnmarshalJSON catches forward-compatible fields into Extra.
func (a *AssistanceAwardListing) UnmarshalJSON(data []byte) error {
	type alias AssistanceAwardListing
	return unmarshalWithExtra(data, (*alias)(a))
}
