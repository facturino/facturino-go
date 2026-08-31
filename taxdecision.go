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
type TaxDecision struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Livemode bool   `json:"livemode"`

	CompanyID string `json:"companyId"`
	// Status is "final", "pending_verification" or "unsupported".
	Status string `json:"status"`
	// TaxSource says who determined the VAT: "facturino" (Facturino decided
	// from the commercial lines and the territorial evidence) or "integration"
	// (the integration supplied the VAT, validated for coherence).
	TaxSource  string               `json:"taxSource"`
	CustomerID string               `json:"customerId"`
	Customer   *TaxDecisionCustomer `json:"customer"`

	SellerProfileID       string                 `json:"sellerProfileId"`
	SellerProfileRevision int                    `json:"sellerProfileRevision"`
	SellerProfile         map[string]interface{} `json:"sellerProfile,omitempty"`

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

	// ForeignTaxReviewRequired reports that a foreign tax may apply. Facturino
	// decides French VAT and the matching French obligations; this case must be
	// reviewed outside Facturino.
	ForeignTaxReviewRequired bool `json:"foreignTaxReviewRequired"`

	Vies                  *ViesResult                  `json:"vies"`
	LocationEvidence      []*LocationEvidenceResult    `json:"locationEvidence"`
	NonEuBusinessEvidence *NonEuBusinessEvidenceResult `json:"nonEuBusinessEvidence"`

	// Issues says what is missing when the decision is not final.
	Issues []*TaxDecisionIssue `json:"issues"`
	// ObligationReasons says why each axis carries the obligation it does.
	ObligationReasons []*TaxDecisionObligationReason `json:"obligationReasons"`

	// RetryOfTaxDecisionID names the decision this one retries after the
	// missing facts were supplied.
	RetryOfTaxDecisionID string `json:"retryOfTaxDecisionId,omitempty"`

	Created string `json:"created"`
	Updated string `json:"updated"`
}

// TaxDecisionTotals holds the decided totals, in integer centimes.
type TaxDecisionTotals struct {
	TotalHT  int `json:"totalHT"`
	TotalVAT int `json:"totalVAT"`
	TotalTTC int `json:"totalTTC"`
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
	// Nature is "business" or "consumer".
	Nature                   string `json:"nature"`
	NatureBasis              string `json:"natureBasis"`
	TerritoryID              string `json:"territoryId"`
	TerritoryKind            string `json:"territoryKind"`
	DeclaredCountry          string `json:"declaredCountry"`
	DeclaredPostalCode       string `json:"declaredPostalCode,omitempty"`
	LegalRegistrationID      string `json:"legalRegistrationId,omitempty"`
	VatNumber                string `json:"vatNumber,omitempty"`
	CrossBorderTaxableStatus string `json:"crossBorderTaxableStatus"`
	BusinessStatusBasis      string `json:"businessStatusBasis"`
}

// TaxDecisionLine is one decided line. Amounts are integer centimes and are nil
// unless the decision is final.
type TaxDecisionLine struct {
	Reference   string `json:"reference"`
	Description string `json:"description"`
	// Category is the fiscal nature: "goods", "services",
	// "electronically_supplied_services", "deposit" or "ancillary_costs".
	Category string `json:"category"`
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
	VatexCode       string `json:"vatexCode,omitempty"`
	LegalMention    string `json:"legalMention,omitempty"`

	RateCentipercent       *int   `json:"rateCentipercent"`
	RateBasis              string `json:"rateBasis,omitempty"`
	PlaceOfSupply          string `json:"placeOfSupply,omitempty"`
	PlaceOfSupplyReference string `json:"placeOfSupplyReference,omitempty"`
	TreatmentReference     string `json:"treatmentReference,omitempty"`

	AmountHT  *int `json:"amountHT"`
	AmountVAT *int `json:"amountVAT"`
	AmountTTC *int `json:"amountTTC"`

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
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
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
	Kind               string `json:"kind"`
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
	// Category is "goods", "services", "electronically_supplied_services",
	// "deposit" or "ancillary_costs".
	Category string `json:"category"`
	// RelatedCategory is required when Category is "deposit" or
	// "ancillary_costs": it names the principal supply they follow.
	RelatedCategory string `json:"relatedCategory,omitempty"`
	// RateCategory is the band requested. The engine decides whether it applies.
	RateCategory string `json:"rateCategory"`
	// PlaceOfSupplyRule claims a territoriality rule. Only "general" is
	// implemented; anything else is reported as unsupported rather than guessed.
	PlaceOfSupplyRule string `json:"placeOfSupplyRule,omitempty"`
	// GoodsMovement is "stays_in_seller_territory",
	// "dispatched_to_buyer_territory" or "unknown".
	GoodsMovement string `json:"goodsMovement,omitempty"`
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
	// VatexCode justifies a zero rate ("VATEX-EU-G", "VATEX-EU-AE",
	// "VATEX-FR-FRANCHISE"…). "integration" source only.
	VatexCode string `json:"vatexCode,omitempty"`
	// PlaceOfSupply is the territory the supplied VAT concluded on
	// ("FR-MET", "DE", "GB"…). "integration" source only.
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
// Two equal fiscal sources share the resource, declared with TaxSource:
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
