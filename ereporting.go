package facturino

import (
	"encoding/json"
	"fmt"
)

// EReporting is an e-reporting declaration.
type EReporting struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Livemode  bool   `json:"livemode"`
	CompanyID string `json:"companyId"`
	// Status is the SUMMARY derived from State (draft, submitted, accepted,
	// rejected, skipped) — never an authority of its own.
	Status string `json:"status"`
	// State is the actual state: reserved, submitting, queued, submitted,
	// acknowledged, accepted, rejected, blocked, reconciliation_required.
	// "reserved" means NOTHING has been transmitted yet.
	State string `json:"state,omitempty"`
	Type  string `json:"type"`
	Volet string `json:"volet,omitempty"` // transaction | payment

	Period      string `json:"period"`
	PeriodStart string `json:"periodStart,omitempty"`
	PeriodEnd   string `json:"periodEnd,omitempty"`

	TotalHT  int `json:"totalHT"` // integer centimes
	TotalTVA int `json:"totalTVA"`
	TotalTTC int `json:"totalTTC"`

	Lines []*EReportingLine `json:"lines"`

	// Attempt is 1 for the first declaration of these facts; incremented only by a retry.
	Attempt       int    `json:"attempt,omitempty"`
	BlockedReason string `json:"blockedReason,omitempty"`
	// PARejectionReason is the refusal reason transmitted by the platform,
	// whatever channel it arrived through.
	PARejectionReason         string `json:"paRejectionReason,omitempty"`
	ReconciliationReason      string `json:"reconciliationReason,omitempty"`
	SupersedesDeclarationID   string `json:"supersedesDeclarationId,omitempty"`
	SupersededByDeclarationID string `json:"supersededByDeclarationId,omitempty"`

	SubmittedAt  string `json:"submittedAt,omitempty"`
	PAResponseID string `json:"paResponseId,omitempty"`

	Created string `json:"created"`
	Updated string `json:"updated"`
}

// EReportingLine is a line in an e-reporting declaration.
type EReportingLine struct {
	Category  string `json:"category"`
	Amount    int    `json:"amount"`  // integer centimes
	VATRate   int    `json:"vatRate"` // centièmes de pourcent
	VATAmount int    `json:"vatAmount"`

	// Date is the transaction day (B2C) or the invoice date (unit lines), YYYY-MM-DD.
	Date string `json:"date,omitempty"`
	// IssueDate, on payment lines of a unit invoice, is its issue date (distinct from Date).
	IssueDate     string `json:"issueDate,omitempty"`
	InvoiceNumber string `json:"invoiceNumber,omitempty"`
	Country       string `json:"country,omitempty"`
	PartnerVAT    string `json:"partnerVat,omitempty"`
	// PartnerName is the buyer's legal name frozen by the tax decision (G2.19).
	PartnerName string `json:"partnerName,omitempty"`
	// VATCategoryCode is the decided UNTDID 5305 category (S, Z, E, AE, K, G, O).
	VATCategoryCode string `json:"vatCategoryCode,omitempty"`
	// VATExCode is the decided exemption code; an explicit null means no VATEX code.
	VATExCode *string `json:"vatexCode,omitempty"`
	// Count, B2C only, is the number of distinct invoices aggregated on Date.
	Count int `json:"count,omitempty"`
	// DocumentType, on unit lines, is "380" (invoice) or "381" (credit note).
	DocumentType string `json:"documentType,omitempty"`
	// OriginalInvoiceNumber and OriginalInvoiceDate, on a credit note ("381"),
	// name the invoice corrected (BT-25 / BT-26). Both are required to transmit
	// a unit credit note (DGFiP G1.32).
	OriginalInvoiceNumber string `json:"originalInvoiceNumber,omitempty"`
	OriginalInvoiceDate   string `json:"originalInvoiceDate,omitempty"`
}

// EReportingParams are the parameters for creating a declaration. Amounts in centimes.
type EReportingParams struct {
	Type   string                  `json:"type"`
	Period string                  `json:"period"`
	Lines  []*EReportingLineParams `json:"lines"`

	IdempotencyKey string `json:"-"`
}

// EReportingLineParams defines a single e-reporting line. The optional fields
// carry the same meaning as on EReportingLine.
type EReportingLineParams struct {
	Category  string `json:"category"`
	Amount    int    `json:"amount"`
	VATRate   int    `json:"vatRate"`
	VATAmount int    `json:"vatAmount"`

	Date            string `json:"date,omitempty"`
	IssueDate       string `json:"issueDate,omitempty"`
	InvoiceNumber   string `json:"invoiceNumber,omitempty"`
	Country         string `json:"country,omitempty"`
	PartnerVAT      string `json:"partnerVat,omitempty"`
	PartnerName     string `json:"partnerName,omitempty"`
	VATCategoryCode string `json:"vatCategoryCode,omitempty"`
	// VATExCode is the decided exemption code. A nil pointer omits the field;
	// to send an explicit `"vatexCode": null` (no exemption code), set
	// VATExCodeNull instead.
	VATExCode     *string `json:"vatexCode,omitempty"`
	VATExCodeNull bool    `json:"-"`
	Count         int     `json:"count,omitempty"`
	DocumentType  string  `json:"documentType,omitempty"`
	// OriginalInvoiceNumber and OriginalInvoiceDate are both required to
	// transmit a unit credit note ("381", DGFiP G1.32).
	OriginalInvoiceNumber string `json:"originalInvoiceNumber,omitempty"`
	OriginalInvoiceDate   string `json:"originalInvoiceDate,omitempty"`
}

// MarshalJSON emits an explicit null VATEX code when VATExCodeNull is set;
// otherwise the field follows VATExCode (omitted when nil).
func (p EReportingLineParams) MarshalJSON() ([]byte, error) {
	type alias EReportingLineParams
	if !p.VATExCodeNull {
		return json.Marshal(alias(p))
	}
	return json.Marshal(struct {
		alias
		VATExCode *string `json:"vatexCode"`
	}{alias: alias(p)})
}

// EReportingService operates on e-reporting declarations.
type EReportingService struct {
	client *httpClient
}

// List returns a paginated iterator over e-reporting declarations.
func (s *EReportingService) List(params *ListParams) *EReportingIterator {
	iter := newIterator[*EReporting](s.client, "/ereporting/declarations", params, decodeEReporting)
	return &EReportingIterator{iter: iter}
}

// Get retrieves an e-reporting declaration by ID.
func (s *EReportingService) Get(id string) (*EReporting, error) {
	var e EReporting
	err := s.client.get(fmt.Sprintf("/ereporting/declarations/%s", id), nil, &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// CreateDeclaration creates a new e-reporting declaration.
func (s *EReportingService) CreateDeclaration(params *EReportingParams) (*EReporting, error) {
	var e EReporting
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post("/ereporting/declarations", params, &e, opts)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// SubmitDeclaration submits a draft to the PA.
func (s *EReportingService) SubmitDeclaration(id string) (*EReporting, error) {
	var e EReporting
	err := s.client.post(fmt.Sprintf("/ereporting/declarations/%s/submit", id), nil, &e, nil)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// EReportingIterator iterates over e-reporting declarations.
type EReportingIterator struct {
	iter *Iterator[*EReporting]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *EReportingIterator) Next() bool { return it.iter.Next() }

// EReporting returns the most recently fetched e-reporting declaration.
func (it *EReportingIterator) EReporting() *EReporting { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *EReportingIterator) Err() error { return it.iter.Err() }

func decodeEReporting(raw json.RawMessage) (*EReporting, error) {
	var e EReporting
	err := json.Unmarshal(raw, &e)
	return &e, err
}
