package facturino

import (
	"fmt"
	"strings"
)

// MaxIdempotencyKeyLength is the longest Idempotency-Key the API accepts.
const MaxIdempotencyKeyLength = 255

// TaxDecision is an immutable fiscal position.
//
// It fixes the VAT, the exact amount to charge and the three reporting axes for
// ONE commercial operation, then never changes. Ask for a decision BEFORE
// charging anything: the amount to debit is AmountToCharge, not a total
// computed locally.
//
// Only a final decision carries amounts. On pending_verification or
// unsupported, Totals and AmountToCharge are nil — never zero: absent is not
// "nothing to charge".
//
// A missing fact leaves pending_verification, with the fact and remedy in
// Issues. Unmodelled local regimes, excluded EU VAT territories and the goods
// regime of a non-taxable legal person carrying a VAT number are unsupported
// under "facturino"; under "integration" they are final with a notice and
// ForeignTaxReviewRequired, and the integrator declares and assumes the tax.
// Both sources refuse a document mixing category O with any other category
// (vat_category_o_split_required); issue two documents. Both also refuse an EU
// address missing the postal code needed to distinguish VAT territories
// (destination_regional_scope_undetermined). Act only on a final decision.
type TaxDecision struct {
	RawResponse `json:"-"`
	Warnings    []*BuyerNatureWarning `json:"warnings,omitempty"`
	ID          string                `json:"id"`
	Object      string                `json:"object"`
	Livemode    bool                  `json:"livemode"`

	CompanyID string `json:"companyId"`
	// Status is "final", "pending_verification" or "unsupported"; an open string.
	Status string `json:"status"`
	// TaxSource says who determined the VAT: "facturino" (Facturino decided
	// from the commercial lines and the territorial evidence) or "integration"
	// (the integration supplied the VAT, validated for coherence).
	TaxSource  string               `json:"taxSource"`
	CustomerID string               `json:"customerId"`
	Customer   *TaxDecisionCustomer `json:"customer"`

	SellerProfileID       string `json:"sellerProfileId"`
	SellerProfileRevision int    `json:"sellerProfileRevision"`
	// SellerProfile may include "localTaxRegistrations": a list of objects with
	// territory (an open string, currently "MF"), reference, effectiveFrom
	// (YYYY-MM-DD) and effectiveTo (YYYY-MM-DD or null).
	SellerProfile map[string]interface{} `json:"sellerProfile,omitempty"`

	Currency string `json:"currency"`
	// PriceMode is "tax_exclusive" or "tax_inclusive".
	PriceMode string `json:"priceMode"`

	// EffectiveAt is the civil date the operation takes effect on.
	EffectiveAt string `json:"effectiveAt"`
	DecidedAt   string `json:"decidedAt"`
	// ExpiresAt is the instant past which the decision may no longer open a
	// payment. It stays readable and immutable.
	ExpiresAt              string `json:"expiresAt"`
	CheckoutValidityPolicy string `json:"checkoutValidityPolicy"`
	// Expired is derived from the server clock at read time, never stored.
	Expired bool `json:"expired"`

	RulesVersion      string `json:"rulesVersion"`
	ReportingCalendar string `json:"reportingCalendar"`
	RoundingPolicy    string `json:"roundingPolicy"`
	// RequestFingerprint is the SHA-256 of the canonical request. The raw
	// idempotency key is never stored nor returned.
	RequestFingerprint string `json:"requestFingerprint"`
	// OperationFingerprint is the SHA-256 of the commercial operation, used to
	// control retries.
	OperationFingerprint string `json:"operationFingerprint"`

	Lines []*TaxDecisionLine `json:"lines"`
	// Totals is nil unless Status is "final". Amounts are integer centimes.
	Totals       *TaxDecisionTotals         `json:"totals"`
	VatBreakdown []*TaxDecisionVatBreakdown `json:"vatBreakdown"`
	// AmountToCharge is the exact amount to debit, in integer centimes. It is
	// nil unless Status is "final".
	AmountToCharge *int `json:"amountToCharge"`

	// InvoiceChannel is "einvoicing" or "none". An operation outside the
	// e-invoicing channel is not deposited on a certified platform.
	InvoiceChannel *string `json:"invoiceChannel"`
	// TransactionReporting is "ereporting", "none" or "outside_scope".
	TransactionReporting *string `json:"transactionReporting"`
	// PaymentReporting is "fr212", "ereporting" or "none".
	PaymentReporting *string `json:"paymentReporting"`

	// SettledObligations carries the axes French law settles DESPITE a
	// non-final decision. It is nil when Status is "final": the three axes
	// above are then the settled ones and nothing duplicates them. It
	// authorises nothing — a non-final decision is not invoiceable, never
	// reaches a certified platform and never opens a payment.
	SettledObligations *TaxDecisionSettledObligations `json:"settledObligations"`

	// EuB2cDestination carries what the EU B2C destination rule concluded,
	// frozen as DATA: the verdict and its basis, the threshold figures it was
	// decided on, the declarative mechanism and the rate entry with its
	// registry version, its source, its verification date, its period and its
	// region. A reference string cannot be audited years later; this can. It is
	// nil on every operation the rule does not reach, and present as soon as it
	// covers a line — including on a decision that is NOT final, where it says
	// exactly what is missing.
	EuB2cDestination *TaxDecisionEuB2cDestination `json:"euB2cDestination"`

	// ForeignTaxReviewRequired reports that a foreign tax needs review outside
	// Facturino. Unmodelled cases accepted under "integration" carry a notice;
	// the integrator declares and assumes the tax. Saint-Martin TGCA is decided
	// separately in LocalTax under both sources.
	ForeignTaxReviewRequired bool `json:"foreignTaxReviewRequired"`

	Vies *ViesResult `json:"vies"`
	// LegalCategory is the SIRENE consultation frozen with the decision; nil
	// when no consultation was due, including historical responses.
	LegalCategory         *TaxDecisionLegalCategory    `json:"legalCategory,omitempty"`
	LocationEvidence      []*LocationEvidenceResult    `json:"locationEvidence"`
	NonEuBusinessEvidence *NonEuBusinessEvidenceResult `json:"nonEuBusinessEvidence"`

	// Issues names the fact and remedy on pending_verification, or explains the
	// refusal on unsupported. Empty on a final decision.
	Issues []*TaxDecisionIssue `json:"issues"`
	// Notices says what the decision retained and what the seller can do next.
	// Always present, empty when there is nothing to report. Notices never stand
	// for a missing fact: show them to the seller, and act only on a final decision.
	Notices []*TaxDecisionNotice `json:"notices"`
	// ObligationReasons says why each axis carries the obligation it does.
	ObligationReasons []*TaxDecisionObligationReason `json:"obligationReasons"`

	// RetryOfTaxDecisionID names the decision this one retries after the
	// missing facts were supplied.
	RetryOfTaxDecisionID *string `json:"retryOfTaxDecisionId,omitempty"`

	Created string `json:"created"`
	Updated string `json:"updated"`
}

// TaxDecisionLegalCategory is the SIRENE legal-category consultation frozen
// with the decision, without other registry data.
type TaxDecisionLegalCategory struct {
	// Source is the registry consulted, currently "sirene".
	Source string `json:"source"`
	// SIREN is the nine-digit registration number consulted.
	SIREN string `json:"siren"`
	// Status is an open string: "found", "not_found" or "unavailable".
	// The latter two leave legal_category_unresolved pending; they do not
	// establish that the buyer is non-taxable.
	Status string `json:"status"`
	// LegalCategoryCode is the INSEE code returned by the registry, or nil for
	// JSON null. Codes stay open strings.
	LegalCategoryCode *string `json:"legalCategoryCode"`
	// CheckedAt is the civil consultation date (YYYY-MM-DD), or that of the
	// cached registry response used by the decision.
	CheckedAt string `json:"checkedAt"`
}

// TaxDecisionTotals holds the decided totals, in integer centimes.
type TaxDecisionTotals struct {
	TotalHT  int `json:"totalHT"`
	TotalVAT int `json:"totalVAT"`
	TotalTTC int `json:"totalTTC"`
	// TotalLocalTax is the sum of line TGCA, in integer centimes, included in
	// TotalTTC and AmountToCharge; nil when absent, distinct from zero.
	TotalLocalTax *int `json:"totalLocalTax,omitempty"`
}

// TaxDecisionVatBreakdown is one VAT group of a decision.
type TaxDecisionVatBreakdown struct {
	RateCentipercent int    `json:"rateCentipercent"`
	CategoryCode     string `json:"categoryCode"`
	VatexCode        string `json:"vatexCode,omitempty"`
	Base             int    `json:"base"`
	Amount           int    `json:"amount"`
}

// TaxDecisionCustomer is the buyer identity and qualification, frozen when the
// decision was taken.
type TaxDecisionCustomer struct {
	CustomerID string `json:"customerId"`
	Name       string `json:"name"`
	// Nature is "business" or "consumer": the nature the decision is taken on.
	// A missing status fact leaves the decision pending, except for the legal
	// presumption described by NatureBasis below.
	Nature string `json:"nature"`
	// NatureBasis is how the nature was established: "declared_consumer",
	// "declared_business_with_registration",
	// "declared_business_with_valid_vat_number",
	// "declared_business_without_identification", "declared_non_taxable" or
	// "declared_business_not_evidenced". Values stay open strings.
	//
	// "declared_business_not_evidenced" applies only under art. 18(2), second
	// subparagraph, of Implementing Regulation (EU) 282/2011: telecommunications,
	// broadcasting and electronically supplied services to a buyer established
	// in the EU VAT territory (Monaco included) who communicated no VAT number.
	// The buyer is treated as a non-taxable person, with a notice. A communicated
	// but unvalidated number leaves eu_b2b_vat_number_not_validated pending;
	// INSEE categories 7, 8 and 9 without taxStatus keep their refusal.
	//
	// "declared_non_taxable" means the customer declared taxStatus: non_taxable.
	// A legal person carrying a VAT number is checked through VIES: a valid
	// number makes its services follow taxable-person rules (CGI art. 259-0);
	// an unreachable VIES leaves vies_unavailable pending.
	NatureBasis         string `json:"natureBasis"`
	TerritoryID         string `json:"territoryId"`
	TerritoryKind       string `json:"territoryKind"`
	DeclaredCountry     string `json:"declaredCountry"`
	DeclaredPostalCode  string `json:"declaredPostalCode,omitempty"`
	LegalRegistrationID string `json:"legalRegistrationId,omitempty"`
	// VatNumber is the VAT number as DECLARED on the customer record, kept for
	// audit whatever became of it.
	VatNumber string `json:"vatNumber,omitempty"`
	// RetainedVatNumber is the VAT number RETAINED by the decision: the only one
	// the document, its electronic invoice (UBL, CII, Factur-X) and the
	// e-reporting carry. It is nil without a declared number, when the number is
	// set aside because the national registration establishes the taxable status
	// (notice detail code "vat_number_ignored_national_registration"), or when
	// the buyer is treated as a non-taxable person. It is also nil on a decision
	// taken before the field existed, where the declared number stands for the
	// retained one.
	RetainedVatNumber        *string `json:"retainedVatNumber,omitempty"`
	CrossBorderTaxableStatus string  `json:"crossBorderTaxableStatus"`
	BusinessStatusBasis      string  `json:"businessStatusBasis"`
}

// TaxDecisionLine is one decided line. Amounts are integer centimes and are nil
// unless the decision is final.
type TaxDecisionLine struct {
	Reference   string `json:"reference"`
	Description string `json:"description"`
	// Category is the fiscal nature: "goods", "services",
	// "electronically_supplied_services", "deposit" or "ancillary_costs".
	// Nil only on a pending decision for a quote/draft line whose supply
	// category was not declared (supply_category_not_declared), without amounts.
	// A service line whose two readings give the same complete tax result is
	// decided as "services" with a notice. POST /v1/tax-decisions still requires
	// Category on every input line under both sources.
	Category *string `json:"category"`
	// RelatedCategory names the principal supply a deposit or an ancillary
	// cost follows.
	RelatedCategory   string `json:"relatedCategory,omitempty"`
	EffectiveCategory string `json:"effectiveCategory,omitempty"`
	// Quantity is a decimal string, never a float.
	Quantity   string               `json:"quantity"`
	UnitAmount int                  `json:"unitAmount"`
	Discount   *TaxDecisionDiscount `json:"discount,omitempty"`
	// RateCategory is the band requested: "standard", "intermediate",
	// "reduced", "super_reduced" or "specific".
	RateCategory      string `json:"rateCategory"`
	PlaceOfSupplyRule string `json:"placeOfSupplyRule,omitempty"`
	GoodsMovement     string `json:"goodsMovement,omitempty"`

	Treatment       string `json:"treatment,omitempty"`
	VatCategoryCode string `json:"vatCategoryCode,omitempty"`
	// VatexCode is an open string from the official CIUS-FR list (88 codes),
	// e.g. "VATEX-FR-CGI261-4", "VATEX-FR-CGI295" or "VATEX-EU-132-1I".
	VatexCode    string `json:"vatexCode,omitempty"`
	LegalMention string `json:"legalMention,omitempty"`

	RateCentipercent       *int   `json:"rateCentipercent"`
	RateBasis              string `json:"rateBasis,omitempty"`
	PlaceOfSupply          string `json:"placeOfSupply,omitempty"`
	PlaceOfSupplyReference string `json:"placeOfSupplyReference,omitempty"`
	TreatmentReference     string `json:"treatmentReference,omitempty"`

	AmountHT  *int `json:"amountHT"`
	AmountVAT *int `json:"amountVAT"`
	AmountTTC *int `json:"amountTTC"`
	// LocalTax is the decided Saint-Martin TGCA, separate from French VAT.
	// Under both sources, a registration in force gives 4%; a declared exemption
	// gives zero with the mention. Without either the decision stays pending
	// (com_taaf_supplier_local_tax_liable).
	LocalTax *LocalTax `json:"localTax,omitempty"`

	InvoiceChannel       string `json:"invoiceChannel,omitempty"`
	TransactionReporting string `json:"transactionReporting,omitempty"`
	PaymentReporting     string `json:"paymentReporting,omitempty"`
}

// TaxDecisionDiscount is an explicit discount. Percent is expressed in
// centi-percent (2500 is 25.00 %); Amount is in integer centimes.
type TaxDecisionDiscount struct {
	// Type is "percent" or "amount".
	Type  string `json:"type"`
	Value int    `json:"value"`
}

// TaxDecisionIssue reports what prevents a decision from being final.
type TaxDecisionIssue struct {
	// Code is an open string. Missing facts include
	// "business_status_not_evidenced", "eu_b2b_vat_number_not_validated",
	// "non_eu_business_status_not_evidenced", "vies_unavailable",
	// "location_evidence_insufficient", "location_evidence_contradictory",
	// "location_evidence_relief_undetermined", "eu_threshold_state_missing",
	// "goods_movement_unknown", "supply_category_not_declared" and
	// "com_taaf_supplier_local_tax_liable". "supply_category_not_declared" asks
	// for a quote/draft line's fiscal nature: no nature was declared, and the
	// decision cannot retain one complete tax result. Category is nil and no
	// amounts are decided.
	// "vat_category_o_split_required" refuses mixed category O documents under
	// both sources (EN 16931 BR-O-11 to BR-O-14): issue two documents.
	// "non_taxable_legal_person_goods_regime_not_modelled" refuses, under
	// "facturino", goods dispatched to another member state or XI for a
	// non-taxable legal person that communicated a VAT number; "integration"
	// retains the supplied treatment with a notice and ForeignTaxReviewRequired.
	Code string `json:"code"`
	// DetailCode is a more precise, equally stable reason, when Code alone does
	// not say what to complete. "einvoicing_buyer_registration_required": a buyer
	// declared as a business without a SIRET, while the electronic invoice the
	// seller must issue requires the buyer's SIREN (CGI art. 242 nonies A).
	// "tax_status_not_declared": INSEE categories 7, 8 or 9 without taxStatus.
	// "legal_category_unresolved": SIRENE could not give the legal category.
	// "french_vat_and_tgca_cumulation": a general service to a Saint-Martin
	// consumer without LocalTaxExemption; com_taaf_local_regime_not_sourced is
	// unsupported under "facturino", a final notice under "integration".
	// Open string; empty when the code says it all.
	DetailCode string `json:"detailCode,omitempty"`
	Message    string `json:"message,omitempty"`
	// Reference is the normative reference.
	Reference string `json:"reference,omitempty"`
	// LineID is the reference of the line concerned; empty when the reason covers
	// the whole operation.
	LineID string `json:"lineId,omitempty"`
}

// TaxDecisionNotice says what the decision retained and what the seller can do
// next. Notices never stand for a missing fact: Issues alone lists what keeps a
// decision from being final. Show notices to the seller; act only on final.
type TaxDecisionNotice struct {
	// Code is an open string. Examples: "business_status_not_evidenced" within
	// the art. 18(2) presumption described by TaxDecisionCustomer.NatureBasis;
	// "destination_mechanism_missing" (destination VAT is invoiced, with
	// registration guidance); "location_evidence_contradictory" (two concordant
	// items of different kinds establish the consumer territory, or every
	// possible territory gives the same complete tax result); a VAT number set
	// aside because national registration establishes status.
	// "supply_category_not_declared" retains "services" for a service line
	// without a declared supply category when both "services" and
	// "electronically_supplied_services" give the same complete tax result
	// (VAT, category, mentions, local tax and the three obligation axes).
	//
	// Under "integration", "com_taaf_local_regime_not_sourced",
	// "com_taaf_local_tax_not_determined",
	// "com_taaf_payment_withholding_not_modelled",
	// "eu_special_territory_not_implemented" and
	// "non_taxable_legal_person_goods_regime_not_modelled" carry the unmodelled
	// case accepted with ForeignTaxReviewRequired; the integrator declares and
	// assumes the tax. Under "facturino" these cases are unsupported issues.
	// Treat an unknown code as a generic notice.
	Code string `json:"code"`
	// DetailCode is a more precise, equally stable reason, when Code alone does
	// not say what to do. On "destination_mechanism_missing": "threshold_exceeded"
	// (the one-stop-shop registration period runs from the first operation above
	// the threshold), "destination_option" or
	// "multi_member_state_establishment" (from the first operation taxed at
	// destination). On "vat_number_territory_mismatch" or
	// "eu_b2b_vat_number_not_validated":
	// "vat_number_ignored_national_registration" (number set aside because the
	// national registration establishes the business status: the buyer stays a
	// business, with no VAT number on the document). Empty when the code says it
	// all. "concordant_evidence" and "same_treatment" name the location rule;
	// "french_vat_and_tgca_cumulation" names a general service to a Saint-Martin
	// consumer without LocalTaxExemption under "integration". Values stay open.
	DetailCode string `json:"detailCode,omitempty"`
	// Message is a diagnostic in English, free of personal data.
	Message string `json:"message"`
	// Reference is the normative reference of the notice.
	Reference string `json:"reference,omitempty"`
	// LineID is the reference of the line concerned; empty when the notice covers
	// the whole operation.
	LineID string `json:"lineId,omitempty"`
}

// TaxDecisionEuB2cDestination is what the EU B2C destination rule concluded.
type TaxDecisionEuB2cDestination struct {
	// CoveredLineIDs names the lines the rule covers.
	CoveredLineIDs []string `json:"coveredLineIds"`
	// RuleKinds holds "tbe_services" and/or "intra_eu_distance_sale".
	RuleKinds []string `json:"ruleKinds"`
	// DestinationMemberState is the consumer's member state, ISO 3166-1 alpha-2.
	DestinationMemberState string `json:"destinationMemberState"`
	// DestinationTerritoryID is the consumer's canonical territory, region
	// included where the member state publishes one ("PT-MA"…).
	DestinationTerritoryID string `json:"destinationTerritoryId"`
	// Place is "origin" or "destination"; nil while the rule did not settle it.
	Place *string `json:"place"`
	// Basis is what settled the place: "multi_member_state_establishment",
	// "destination_option", "oss_union_registration", "threshold_exceeded" or
	// "below_threshold".
	//
	// "oss_union_registration": the seller holds an ACTIVE Union one-stop-shop
	// registration. For a French seller, registering IS how the option of
	// art. 59c(3) is exercised, so the threshold has nothing left to decide and
	// Threshold stays nil, exactly as on an explicit option. Sourced for France
	// only: the way the option is exercised is fixed by the member state where
	// it is exercised.
	Basis     *string `json:"basis"`
	Reference string  `json:"reference"`
	Detail    string  `json:"detail"`

	Threshold *TaxDecisionThresholdTrace       `json:"threshold"`
	Option    *TaxDecisionDestinationOption    `json:"option"`
	Mechanism *TaxDecisionDestinationMechanism `json:"mechanism"`
	Rate      *TaxDecisionDestinationRate      `json:"rate"`
	// EvidenceRelief is how the art. 24b relaxation was settled; nil when the
	// location-evidence rule did not apply.
	EvidenceRelief *TaxDecisionEvidenceRelief `json:"evidenceRelief"`
}

// TaxDecisionThresholdTrace is the slice of the ANNUAL LEDGER this decision
// took, frozen with it.
//
// The running total is not a photograph left on the fiscal profile: it is a
// transactional ledger per company, per mode and per year, and the decision
// freezes which ledger, at which version, and at which position in the total
// order of movements it occupied. Under a tax-inclusive price the operation's
// VAT-exclusive value depends on the rate the threshold has to decide, so it is
// an interval; under a tax-exclusive price both bounds coincide.
type TaxDecisionThresholdTrace struct {
	// DecidedOn is "previous_year" or "ledger_cumulative". On "previous_year"
	// the previous calendar year closes the question on its own.
	DecidedOn string `json:"decidedOn"`
	// CapCents is the cap of art. 59c(1), in integer centimes.
	CapCents int `json:"capCents"`
	// StateID names the annual ledger ("2026_live", "2026_test").
	StateID string `json:"stateId"`
	Year    string `json:"year"`
	// StateVersion is the ledger version the slice was taken at.
	StateVersion int `json:"stateVersion"`
	// Sequence is the position in the ledger's total order of movements.
	Sequence int `json:"sequence"`
	// ReservationID is the decision's durable idempotency claim.
	ReservationID string `json:"reservationId"`
	// CoverageMode is "facturino_only" or "mixed_channels".
	CoverageMode            string `json:"coverageMode"`
	PreviousYearAmountCents int    `json:"previousYearAmountCents"`
	CurrentYearOpeningCents int    `json:"currentYearOpeningCents"`
	OpeningDeclaredAt       string `json:"openingDeclaredAt"`
	// ExternalCompleteThroughDate is the day the channels other than Facturino
	// are declared complete through.
	ExternalCompleteThroughDate string `json:"externalCompleteThroughDate"`
	AdjustmentTotalCents        int    `json:"adjustmentTotalCents"`
	AdjustmentCount             int    `json:"adjustmentCount"`
	// CumulativeBeforeMinCents is what CERTAINLY precedes the operation:
	// settled movements only. CumulativeBeforeMaxCents adds every slice held by
	// an operation decided at the same moment — the gap between the two IS that
	// concurrency, and a verdict is frozen only when it holds at both bounds.
	CumulativeBeforeMinCents int `json:"cumulativeBeforeMinCents"`
	CumulativeBeforeMaxCents int `json:"cumulativeBeforeMaxCents"`
	// PendingPredecessorCount is how many concurrent operations the upper bound
	// accounts for.
	PendingPredecessorCount int `json:"pendingPredecessorCount"`
	OperationValueMinCents  int `json:"operationValueMinCents"`
	OperationValueMaxCents  int `json:"operationValueMaxCents"`
	CumulativeAfterMinCents int `json:"cumulativeAfterMinCents"`
	CumulativeAfterMaxCents int `json:"cumulativeAfterMaxCents"`
}

// TaxDecisionDestinationOption is the option period, when it settled the place.
type TaxDecisionDestinationOption struct {
	EffectiveFrom string  `json:"effectiveFrom"`
	EffectiveTo   *string `json:"effectiveTo"`
}

// TaxDecisionDestinationMechanism says how the tax due at destination is
// declared. It never decides a place of taxation.
// The registration is DATED: a one-stop shop opened in October does not
// declare a September sale.
type TaxDecisionDestinationMechanism struct {
	// Kind is "oss_union" or "local_registration".
	Kind        string `json:"kind"`
	MemberState string `json:"memberState"`
	Reference   string `json:"reference"`
	// MemberStateOfIdentification is the state the scheme is filed in; nil for
	// a local registration.
	MemberStateOfIdentification *string `json:"memberStateOfIdentification"`
	EffectiveFrom               string  `json:"effectiveFrom"`
	EffectiveTo                 *string `json:"effectiveTo"`
}

// TaxDecisionEvidenceRelief is how the art. 24b single-evidence relaxation was
// settled, with the figures it rested on.
//
// It is COMPUTED by the engine on the ledger's EUR 100,000 counter — the one
// that never counts a distance sale of goods — and never declared by the
// seller. "undeterminable" is a first-class answer: two items of evidence are
// then required, and the issue says which fact is missing.
type TaxDecisionEvidenceRelief struct {
	// Status is "available", "unavailable" or "undeterminable".
	Status string `json:"status"`
	// CapCents is the cap of art. 24b, 2nd subparagraph (10,000,000).
	CapCents                int     `json:"capCents"`
	StateID                 *string `json:"stateId"`
	Year                    *string `json:"year"`
	PreviousYearAmountCents *int    `json:"previousYearAmountCents"`
	CumulativeAfterMinCents *int    `json:"cumulativeAfterMinCents"`
	CumulativeAfterMaxCents *int    `json:"cumulativeAfterMaxCents"`
	// UndeterminedCode is "ledger_not_consulted", "ledger_unavailable" or
	// "amount_interval_straddles_cap".
	UndeterminedCode *string `json:"undeterminedCode"`
}

// TaxDecisionDestinationRate is the rate entry the decision was taken under,
// with everything that justifies it.
type TaxDecisionDestinationRate struct {
	RegistryVersion string `json:"registryVersion"`
	MemberState     string `json:"memberState"`
	TerritoryID     string `json:"territoryId"`
	// RegionID names the region when the rate is a regional one.
	RegionID     *string `json:"regionId"`
	Centipercent int     `json:"centipercent"`
	ValidFrom    string  `json:"validFrom"`
	ValidTo      *string `json:"validTo"`
	Source       string  `json:"source"`
	VerifiedAt   string  `json:"verifiedAt"`
}

// TaxDecisionSettledObligations carries the axes French law settles on its own,
// available on a decision that is NOT final. An axis is nil when it depends on
// the treatment the engines could not conclude; it is never guessed.
type TaxDecisionSettledObligations struct {
	// InvoiceChannel is "einvoicing" or "none".
	InvoiceChannel *string `json:"invoiceChannel"`
	// TransactionReporting is "ereporting", "none" or "outside_scope".
	TransactionReporting *string `json:"transactionReporting"`
	// PaymentReporting is "fr212", "ereporting" or "none".
	PaymentReporting *string `json:"paymentReporting"`
}

// TaxDecisionObligationReason explains one reporting axis.
type TaxDecisionObligationReason struct {
	// Axis is "invoiceChannel", "transactionReporting" or "paymentReporting".
	Axis      string `json:"axis"`
	Code      string `json:"code"`
	Reference string `json:"reference,omitempty"`
	Message   string `json:"message,omitempty"`
}

// ViesResult is the VIES consultation outcome. The status is kept, never the
// raw response.
type ViesResult struct {
	// Status is "valid", "invalid", "unavailable" or "invalid_format".
	Status              string `json:"status"`
	CheckedAt           string `json:"checkedAt,omitempty"`
	NormalizedVatNumber string `json:"normalizedVatNumber,omitempty"`
	Source              string `json:"source,omitempty"`
	ReturnedName        string `json:"returnedName,omitempty"`
	ConsultationNumber  string `json:"consultationNumber,omitempty"`
}

// LocationEvidenceParams is one piece of location evidence (Implementing
// Regulation (EU) 282/2011).
//
// Send the territorial SIGNAL, never the raw one: a country — and a postal code
// where the territory needs one — not an IP address, a PSP payload or bank
// account details. Reference is a bounded opaque identifier such as a PSP
// charge id, not the signal itself.
type LocationEvidenceParams struct {
	// Kind is "billing_address", "ip_geolocation", "bank_details",
	// "sim_mobile_country", "fixed_line" or "other_commercial".
	Kind string `json:"kind"`
	// Country is an ISO 3166-1 alpha-2 code.
	Country    string `json:"country"`
	PostalCode string `json:"postalCode,omitempty"`
	// ThirdParty reports whether the evidence comes from a party independent of
	// both seller and buyer.
	ThirdParty bool `json:"thirdParty"`
	// Source is "psp", "network", "bank", "declared" or "other".
	Source string `json:"source"`
	// CollectedAt is a civil date (YYYY-MM-DD).
	CollectedAt string `json:"collectedAt"`
	Reference   string `json:"reference,omitempty"`
}

// LocationEvidenceResult is normalized territorial evidence kept with the
// decision. No raw signal is exposed.
type LocationEvidenceResult struct {
	Kind string `json:"kind"`
	// TerritoryID is the canonical territory the evidence points to. It is
	// EMPTY when the API resolved the item at country level only — a network
	// kind (ip_geolocation, bank_details, sim_mobile_country, fixed_line)
	// supplied without a postal code, published as JSON null. DeclaredCountry
	// then carries the country.
	TerritoryID        string `json:"territoryId"`
	DeclaredCountry    string `json:"declaredCountry"`
	DeclaredPostalCode string `json:"declaredPostalCode,omitempty"`
	ThirdParty         bool   `json:"thirdParty"`
	Source             string `json:"source"`
	CollectedAt        string `json:"collectedAt"`
	Reference          string `json:"reference,omitempty"`
}

// NonEuBusinessEvidenceParams is non-EU business-status evidence
// (282/2011 art. 18-3).
type NonEuBusinessEvidenceParams struct {
	// Kind is "tax_authority_certificate", "vat_or_similar_number" or
	// "other_commercial_evidence".
	Kind      string `json:"kind"`
	Reference string `json:"reference"`
	// IssuedByCountry is an ISO 3166-1 alpha-2 code.
	IssuedByCountry                 string `json:"issuedByCountry"`
	IssuedByPostalCode              string `json:"issuedByPostalCode,omitempty"`
	ReasonableVerificationPerformed bool   `json:"reasonableVerificationPerformed"`
	CollectedAt                     string `json:"collectedAt"`
}

// NonEuBusinessEvidenceResult is the normalized non-EU business-status evidence.
type NonEuBusinessEvidenceResult struct {
	Kind                            string `json:"kind"`
	Reference                       string `json:"reference"`
	IssuedByTerritoryID             string `json:"issuedByTerritoryId"`
	ReasonableVerificationPerformed bool   `json:"reasonableVerificationPerformed"`
	CollectedAt                     string `json:"collectedAt"`
}

// TaxDecisionLineParams describes one line of the operation to decide.
//
// The caller describes the operation; it never states the outcome. No rate, VAT
// category code, VATEX code, legal mention or reporting axis is accepted here.
type TaxDecisionLineParams struct {
	// Reference is a stable caller-side identifier, echoed on the decided line.
	Reference   string `json:"reference"`
	Description string `json:"description"`
	// Category is required under both sources: "goods", "services",
	// "electronically_supplied_services", "deposit" or "ancillary_costs".
	Category string `json:"category"`
	// RelatedCategory is required when Category is "deposit" or
	// "ancillary_costs": it names the principal supply they follow.
	RelatedCategory string `json:"relatedCategory,omitempty"`
	// RateCategory is the band requested. The engine decides whether it applies.
	//
	// Omitted when empty: a line whose VAT the INTEGRATION supplies states a
	// rate, not a band, and the API refuses the key outright on that source.
	RateCategory string `json:"rateCategory,omitempty"`
	// PlaceOfSupplyRule is the territoriality rule of the operation; "general"
	// when empty, under both sources. On the "integration" source it is the rule
	// under which the declared PlaceOfSupply is checked against the place
	// Facturino decides on the same facts. A derogation the engine does not
	// implement (immovable property, event…) makes the decision unsupported,
	// exactly as under the "facturino" source: declaring it never skips that
	// check.
	PlaceOfSupplyRule string `json:"placeOfSupplyRule,omitempty"`
	// GoodsMovement is "stays_in_seller_territory",
	// "dispatched_to_buyer_territory" or "unknown".
	//
	// Required on a goods line — under BOTH sources — as soon as the buyer is a
	// consumer established in another member state: that movement decides
	// whether the intra-EU distance sale rule applies (Directive 2006/112/EC
	// art. 33(a)), and it is never assumed.
	// An unknown movement leaves goods_movement_unknown pending, without amounts.
	GoodsMovement string `json:"goodsMovement,omitempty"`
	// LocalTaxExemption is an open string: "cgism_253" and
	// "not_used_in_saint_martin" are currently documented. Omitted when empty.
	// Under both sources, a declared exemption gives zero TGCA with the mention;
	// otherwise a registration in force gives 4%. Without either, the decision
	// stays pending (com_taaf_supplier_local_tax_liable).
	LocalTaxExemption string `json:"localTaxExemption,omitempty"`
	// UnitAmount is the unit price in integer centimes, in the request's
	// PriceMode.
	UnitAmount int `json:"unitAmount"`
	// Quantity is a decimal STRING, never a float: 0.1 + 0.2 is not 0.3, and a
	// quantity is a financial figure.
	Quantity string               `json:"quantity"`
	Discount *TaxDecisionDiscount `json:"discount,omitempty"`

	// VatRate carries the supplied rate in integer centipercent, on the
	// "integration" source only. A pointer: zero is a real rate (an exempt or
	// zero-rated line), absent means "not supplied".
	VatRate *int `json:"vatRate,omitempty"`
	// VatCode is the supplied VAT category code: "S", "Z", "E", "AE", "K",
	// "G" or "O". "integration" source only.
	VatCode string `json:"vatCode,omitempty"`
	// VatexCode justifies a zero rate using the official CIUS-FR list (88 codes,
	// e.g. "VATEX-FR-CGI261-4", "VATEX-FR-AE", "VATEX-EU-I", "VATEX-EU-J").
	// Open string, "integration" source only. Legacy "VATEX-FR-261" and
	// "VATEX-FR-AUTOLIQ" are accepted; new writes store and return the official
	// codes "VATEX-FR-CGI261-4" and "VATEX-FR-AE". Issued documents keep their code.
	VatexCode string `json:"vatexCode,omitempty"`
	// PlaceOfSupply is the territory the supplied VAT concluded on
	// ("FR-MET", "DE", "GB"…), 2 to 16 characters. "integration" source only.
	//
	// It must be the place Facturino decides on the same facts, otherwise the
	// request is refused with 422 integration_vat_incoherent; it feeds the
	// obligation axes, never a VAT calculation.
	//
	// Required as soon as the buyer is established in a French overseas
	// collectivity or the TAAF ("PM", "BL", "MF", "PF", "NC", "WF", "TF"):
	// the place is what says whether the local tax of that collectivity is at
	// stake, and it is never assumed. An operation located in one of those seven
	// is outside the scope of French VAT: category "O" at 0. Unmodelled local
	// taxes give an integration notice with ForeignTaxReviewRequired; cases
	// settled by a collectivity's text stay final under both sources. Saint-Martin
	// TGCA is decided separately in LocalTax under both sources. VatRate, VatCode
	// and VatexCode describe French VAT only.
	PlaceOfSupply string `json:"placeOfSupply,omitempty"`
}

// TaxDecisionParams are the parameters for taking a tax decision.
type TaxDecisionParams struct {
	// TaxSource is required: "facturino" asks Facturino to determine the VAT;
	// "integration" supplies the VAT of every line (VatRate, VatCode, and
	// where the rate is zero a VatexCode and the line's PlaceOfSupply).
	// Supplied VAT is validated for coherence — contradictions answer
	// integration_vat_incoherent, never a silent correction.
	TaxSource string `json:"taxSource"`
	Customer  string `json:"customerId"`
	// EffectiveAt is a civil date (YYYY-MM-DD). A timestamp is refused: turning
	// an instant into a civil date is a timezone call that belongs to you.
	EffectiveAt string `json:"effectiveAt"`
	// Currency is "eur" only in this ruleset. Any other currency is refused,
	// never converted.
	Currency string `json:"currency"`
	// PriceMode is "tax_exclusive" or "tax_inclusive".
	PriceMode string                   `json:"priceMode"`
	Lines     []*TaxDecisionLineParams `json:"lines"`

	LocationEvidence      []*LocationEvidenceParams    `json:"locationEvidence,omitempty"`
	NonEuBusinessEvidence *NonEuBusinessEvidenceParams `json:"nonEuBusinessEvidence,omitempty"`

	// RetryOfTaxDecisionID names the previous decision this one retries after
	// the missing facts were supplied. The commercial operation must be
	// identical; only the evidence may change.
	RetryOfTaxDecisionID string `json:"retryOfTaxDecisionId,omitempty"`

	IdempotencyKey string `json:"-"`
}

// TaxDecisionService handles the tax-decisions resource.
//
// A decision is immutable: there is no update and no delete. To re-decide the
// same operation after supplying missing evidence, create a new decision with
// RetryOfTaxDecisionID.
//
// Two fiscal sources share the resource, declared with TaxSource:
// "facturino" (Facturino determines the VAT) and "integration" (the caller
// supplies the VAT of every line, validated for coherence). Both produce the
// same decision object, feed the same reporting obligations engine and back
// invoices the same way.
type TaxDecisionService struct {
	client *httpClient
}

// Create takes a decision on a commercial operation.
//
// Business idempotency is durable here, beyond the 24-hour transport window,
// and it is keyed on the KEY - not on the operation. The same IdempotencyKey
// with the same canonical body always replays the same decision, so a retried
// request after a lost response replays that decision instead of taking a
// second one. Two DIFFERENT keys describing the same operation produce TWO
// decisions: nothing matches them up.
//
// The API answers 201 on creation and 200 when the same key already produced
// this decision — both return the decision, so the caller reads one shape
// either way. Reusing the key with a different body answers 409, surfaced as a
// *facturino.Error with HTTPStatusCode 409; it is never retried.
func (s *TaxDecisionService) Create(params *TaxDecisionParams) (*TaxDecision, error) {
	if params == nil {
		return nil, fmt.Errorf("facturino: tax decision params are required")
	}
	key := strings.TrimSpace(params.IdempotencyKey)
	if key == "" {
		return nil, fmt.Errorf("facturino: Idempotency-Key is required for tax decisions")
	}
	// The API caps the key at 255 characters. Checking it here turns a round
	// trip and a 400 into an immediate, local error.
	if len(key) > MaxIdempotencyKeyLength {
		return nil, fmt.Errorf(
			"facturino: Idempotency-Key must be at most %d characters (received %d)",
			MaxIdempotencyKeyLength, len(key),
		)
	}
	var d TaxDecision
	opts := &requestOption{}
	opts.idempotencyKey = key
	if err := s.client.post("/tax-decisions", params, &d, opts); err != nil {
		return nil, err
	}
	return &d, nil
}

// Get reads a decision back — typically after a payment capture, to check that
// the captured amount, currency and buyer match what was decided.
func (s *TaxDecisionService) Get(id string) (*TaxDecision, error) {
	var d TaxDecision
	if err := s.client.get(fmt.Sprintf("/tax-decisions/%s", id), nil, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// Retrieve is an alias of Get, matching the name used in the published
// documentation and the other SDKs.
func (s *TaxDecisionService) Retrieve(id string) (*TaxDecision, error) {
	return s.Get(id)
}

// IsFinal reports whether the decision carries amounts. Only a final decision
// does; on any other status Totals and AmountToCharge are nil.
func (d *TaxDecision) IsFinal() bool {
	return d.Status == "final"
}
