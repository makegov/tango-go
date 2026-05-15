package tango

// DefaultBaseURL is the public Tango SaaS endpoint.
const DefaultBaseURL = "https://tango.makegov.com"

// Shape presets mirror the canonical defaults in tango-node and tango-python.
// Pass any of these (or a custom comma-separated field list) via the Shape
// option on a list/get call to control which fields the API returns.
const (
	// ShapeContractsMinimal — default for ListContracts.
	ShapeContractsMinimal = "key,piid,award_date,recipient(display_name),description,total_contract_value"

	// ShapeEntitiesMinimal — default for ListEntities.
	ShapeEntitiesMinimal = "uei,legal_business_name,cage_code,business_types"

	// ShapeEntitiesComprehensive — default for GetEntity.
	ShapeEntitiesComprehensive = "uei,legal_business_name,dba_name,cage_code," +
		"business_types,primary_naics,naics_codes,psc_codes," +
		"email_address,entity_url,description,capabilities,keywords," +
		"physical_address,mailing_address," +
		"federal_obligations(*),congressional_district"

	// ShapeForecastsMinimal — default for ListForecasts.
	ShapeForecastsMinimal = "id,title,anticipated_award_date,fiscal_year,naics_code,status"

	// ShapeOpportunitiesMinimal — default for ListOpportunities.
	ShapeOpportunitiesMinimal = "opportunity_id,title,solicitation_number,response_deadline,active"

	// ShapeNoticesMinimal — default for ListNotices.
	ShapeNoticesMinimal = "notice_id,title,solicitation_number,posted_date"

	// ShapeProtestsMinimal — default for ListProtests.
	ShapeProtestsMinimal = "case_id,case_number,title,source_system,outcome,filed_date"

	// ShapeGrantsMinimal — default for ListGrants.
	ShapeGrantsMinimal = "grant_id,opportunity_number,title,status(*),agency_code"

	// ShapeIDVsMinimal — default for ListIDVs.
	ShapeIDVsMinimal = "key,piid,award_date,recipient(display_name,uei),description,total_contract_value,obligated,idv_type"

	// ShapeIDVsComprehensive — default for GetIDV.
	ShapeIDVsComprehensive = "key,piid,award_date,description,fiscal_year,total_contract_value,obligated," +
		"idv_type,multiple_or_single_award_idv,type_of_idc,period_of_performance(start_date,last_date_to_order)," +
		"recipient(display_name,legal_business_name,uei,cage)," +
		"awarding_office(*),funding_office(*),place_of_performance(*),parent_award(key,piid)," +
		"competition(*),legislative_mandates(*),transactions(*),subawards_summary(*)"

	// ShapeVehiclesMinimal — default for ListVehicles.
	ShapeVehiclesMinimal = "uuid,solicitation_identifier,is_synthetic_solicitation,program_acronym," +
		"organization_id,organization,vehicle_type,description," +
		"idv_count,awardee_count,order_count,total_obligated," +
		"vehicle_obligations,vehicle_contracts_value,latest_award_date," +
		"solicitation_title,solicitation_date"

	// ShapeVehiclesComprehensive — default for GetVehicle.
	ShapeVehiclesComprehensive = "uuid,solicitation_identifier,is_synthetic_solicitation,agency_id,program_acronym," +
		"organization_id,organization(*),vehicle_type,who_can_use," +
		"solicitation_title,solicitation_description,solicitation_date,opportunity_id," +
		"naics_code,psc_code,set_aside," +
		"fiscal_year,award_date,latest_award_date,last_date_to_order," +
		"description,idv_count,awardee_count,order_count,total_obligated," +
		"vehicle_obligations,vehicle_contracts_value," +
		"type_of_idc,contract_type,metrics(*)"

	// ShapeVehicleAwardeesMinimal — default for ListVehicleAwardees.
	ShapeVehicleAwardeesMinimal = "uuid,key,piid,award_date,title,order_count,idv_obligations,idv_contracts_value,recipient(display_name,uei)"

	// ShapeVehicleOrdersMinimal — default for ListVehicleOrders.
	ShapeVehicleOrdersMinimal = "key,piid,award_date,obligated,total_contract_value,description,recipient(display_name,uei)"

	// ShapeOrganizationsMinimal — default for ListOrganizations.
	ShapeOrganizationsMinimal = "key,fh_key,name,level,type,short_name"

	// ShapeOTAsMinimal — default for ListOTAs.
	ShapeOTAsMinimal = "key,piid,award_date,recipient(display_name,uei),description,total_contract_value,obligated"

	// ShapeOTIDVsMinimal — default for ListOTIDVs.
	ShapeOTIDVsMinimal = "key,piid,award_date,recipient(display_name,uei),description,total_contract_value,obligated,idv_type"

	// ShapeSubawardsMinimal — default for ListSubawards. The API rejects "id"
	// and "amount" in subaward shapes; stick to the canonical fields.
	ShapeSubawardsMinimal = "award_key,prime_recipient(uei,display_name),subaward_recipient(uei,display_name)"

	// ShapeGsaElibraryContractsMinimal — default for ListGsaElibraryContracts.
	ShapeGsaElibraryContractsMinimal = "uuid,contract_number,schedule,recipient(display_name,uei),idv(key,award_date)"

	// ShapeItdashboardInvestmentsMinimal — default for ListItdashboardInvestments.
	// Matches the API's INVESTMENT_LIST_DEFAULT_SHAPE.
	ShapeItdashboardInvestmentsMinimal = "uii,agency_name,bureau_name,investment_title," +
		"type_of_investment,part_of_it_portfolio,updated_time,url"

	// ShapeItdashboardInvestmentsComprehensive — default for GetItdashboardInvestment.
	// Matches the API's INVESTMENT_RETRIEVE_DEFAULT_SHAPE.
	ShapeItdashboardInvestmentsComprehensive = "uii,agency_code,agency_name,bureau_code,bureau_name," +
		"investment_title,type_of_investment,part_of_it_portfolio," +
		"updated_time,url"
)
