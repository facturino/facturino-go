package facturino

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Invoice is a Facturino invoice.
type Invoice struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Type   string `json:"type"`
	// Status is the summary projection of the three axes below. It stays
	// populated and supported; prefer the axes when you need to tell
	// transmission from collection.
	Status string `json:"status"`
	// DocumentStatus is the documentary axis: "draft", "finalized" or
	// "cancelled".
	DocumentStatus string `json:"documentStatus,omitempty"`
	// TransmissionStatus is the transmission axis: "not_applicable", "pending",
	// "sending", "deposited", "transmitted", "approved" or "rejected". A
	// collection never moves it.
	TransmissionStatus string `json:"transmissionStatus,omitempty"`
	// TransmissionDetail is the DGFiP detail inside "transmitted" / "rejected":
	// "available", "received", "suspended" or "refused".
	TransmissionDetail string `json:"transmissionDetail,omitempty"`
	// PaymentStatus is the collection axis: "unpaid", "partially_paid", "paid",
	// "partially_refunded" or "refunded". A refund does not erase the
	// collection that happened.
	PaymentStatus string `json:"paymentStatus,omitempty"`
	// TaxSource says who determined the VAT: "facturino" or "integration".
	// Empty when the commercial draft is not decided yet (taxSource: null).
	TaxSource string `json:"taxSource,omitempty"`
	// TaxDecisionID names the decision backing this invoice, when it has one.
	TaxDecisionID string `json:"taxDecisionId,omitempty"`
	// TaxSnapshot is the frozen fiscal position copied from the decision.
	TaxSnapshot map[string]interface{} `json:"taxSnapshot,omitempty"`
	// CommercialDraft is the operation an UNDECIDED draft states — typically
	// one produced by Quotes.Convert. Present only while TaxSource is empty; it
	// disappears the moment the invoice is bound to a decision.
	//
	// Read its lines to build the decision that fiscalises this draft: the line
	// references are assigned server-side at conversion, and the decision must
	// state exactly the operation the draft carries.
	CommercialDraft *CommercialDraft `json:"commercialDraft,omitempty"`
	Number          string           `json:"number"`
	Currency        string           `json:"currency"`
	Livemode        bool             `json:"livemode"`

	Customer *CustomerRef `json:"customer"`
	Items    []*LineItem  `json:"items"`
	Totals   *Totals      `json:"totals"`

	Dates       *InvoiceDates        `json:"dates"`
	PaymentInfo *InvoicePaymentTerms `json:"paymentInfo"`
	Einvoicing  *InvoiceEinvoicing   `json:"einvoicing"`
	Archive     *InvoiceArchive      `json:"archive"`
	Files       *InvoiceFiles        `json:"files"`
	Portal      *InvoicePortal       `json:"portal"`

	Notes         string `json:"notes"`
	LegalMentions string `json:"legalMentions"`

	Lifecycle []*LifecycleEntry `json:"lifecycle"`

	Metadata map[string]interface{} `json:"metadata"`

	// Expanded holds related resources inlined when requested via the
	// expand parameter on Get (see InvoiceGetParams.Expand). It is nil
	// unless expansion was requested.
	Expanded *InvoiceExpanded `json:"expanded,omitempty"`

	Created string `json:"created"`
	Updated string `json:"updated"`
}

// InvoiceExpanded holds related resources inlined on an invoice when
// requested through the expand parameter.
type InvoiceExpanded struct {
	// Customer is populated when expand includes "customer".
	Customer *Customer `json:"customer,omitempty"`

	// CreditNotes lists the credit notes issued against the invoice. It
	// is populated when expand includes "credit_notes".
	CreditNotes []*CreditNote `json:"credit_notes,omitempty"`

	// NetBalance is the invoice total less the sum of its credit notes,
	// as integer centimes. It is populated when expand includes
	// "credit_notes".
	NetBalance int `json:"net_balance,omitempty"`
}

// InvoiceDates holds invoice dates.
type InvoiceDates struct {
	Issued       string `json:"issued"`
	Due          string `json:"due"`
	ServiceStart string `json:"serviceStart,omitempty"`
	ServiceEnd   string `json:"serviceEnd,omitempty"`
	FinalizedAt  string `json:"finalizedAt,omitempty"`
	SentAt       string `json:"sentAt,omitempty"`
	// PaidAt is the actual settlement date: the paidAt of the payment that
	// cleared the balance. Empty while a balance remains, cleared again when
	// a reversal reopens one.
	PaidAt string `json:"paidAt,omitempty"`
}

// InvoicePaymentTerms holds payment terms.
type InvoicePaymentTerms struct {
	Terms                string `json:"terms"`
	TermsDays            int    `json:"termsDays"`
	Method               string `json:"method"`
	IBAN                 string `json:"iban,omitempty"`
	BIC                  string `json:"bic,omitempty"`
	EarlyPaymentDiscount string `json:"earlyPaymentDiscount,omitempty"`
	LatePaymentRate      string `json:"latePaymentRate"`
	CollectionFee        string `json:"collectionFee"`
}

// InvoiceSubmissionArtefact is the CII regenerated for a deposit when the
// frozen original no longer satisfies a CIUS-FR rule: same number, same
// amounts, stored under InvoiceFiles.CorrectedXMLPath; the archived Factur-X
// is never rewritten.
type InvoiceSubmissionArtefact struct {
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	GeneratedAt string `json:"generatedAt"`
	// CorrectedRules lists the rules the regeneration satisfied (e.g. BR-FR-08).
	CorrectedRules []string `json:"correctedRules"`
}

// InvoicePreviousSubmission is a closed attempt of the same document on the
// platform: a rejected deposit resent under the same number opens a new
// attempt whose identifiers are the live ones on InvoiceEinvoicing.
type InvoicePreviousSubmission struct {
	PAID             string `json:"paId"`
	PATransactionID  string `json:"paTransactionId"`
	PAIdempotencyKey string `json:"paIdempotencyKey"`
	PAStatus         string `json:"paStatus"`
	PAStatusCode     string `json:"paStatusCode"`
	PAErrorCode      string `json:"paErrorCode"`
	RejectionReason  string `json:"rejectionReason"`
	SentAt           string `json:"sentAt"`
	// ClosedAt is when the next attempt was opened.
	ClosedAt string `json:"closedAt"`
}

// InvoiceEinvoicing holds e-invoicing (PA) status.
type InvoiceEinvoicing struct {
	PAID     string `json:"paId"`
	PAStatus string `json:"paStatus"`
	// PAStatusCode is the raw platform status code (e.g. "fr:200").
	PAStatusCode    string `json:"paStatusCode,omitempty"`
	PATransactionID string `json:"paTransactionId"`
	PAErrorCode     string `json:"paErrorCode"`
	// RejectionReason is the platform's reason for a rejection, whatever
	// channel it arrived through; RefusalReason is the buyer's reason for a refusal.
	RejectionReason  string `json:"rejectionReason,omitempty"`
	RefusalReason    string `json:"refusalReason,omitempty"`
	PAIdempotencyKey string `json:"paIdempotencyKey"`
	PeppolDeliveryID string `json:"peppolDeliveryId"`
	EreportingID     string `json:"ereportingId"`
	SentAt           string `json:"sentAt"`
	TrackingID       string `json:"trackingId"`

	SubmissionArtefact *InvoiceSubmissionArtefact `json:"submissionArtefact,omitempty"`
	// PreviousSubmissions lists the closed attempts, oldest first; absent until
	// a rejected deposit is resent.
	PreviousSubmissions []*InvoicePreviousSubmission `json:"previousSubmissions,omitempty"`
}

// InvoiceArchive holds hash chain data.
type InvoiceArchive struct {
	Hash             string `json:"hash"`
	PreviousHash     string `json:"previousHash"`
	ArchivedAt       string `json:"archivedAt"`
	ArchivedFilePath string `json:"archivedFilePath,omitempty"`
}

// InvoiceFiles holds storage paths for generated documents (PDF, Factur-X, XML).
type InvoiceFiles struct {
	PDFPath     string `json:"pdfPath,omitempty"`
	FacturXPath string `json:"facturxPath,omitempty"`
	XMLPath     string `json:"xmlPath,omitempty"`
	// CorrectedXMLPath is the CII regenerated for the deposit, see InvoiceEinvoicing.SubmissionArtefact.
	CorrectedXMLPath string `json:"correctedXmlPath,omitempty"`
}

// InvoicePortal holds payment portal data.
type InvoicePortal struct {
	Token         string `json:"token"`
	ExpiresAt     string `json:"expiresAt"`
	FirstViewedAt string `json:"firstViewedAt,omitempty"`
}

// CustomerRef is a customer reference embedded in invoices, quotes, and credit notes.
type CustomerRef struct {
	Ref      string            `json:"ref"`
	Snapshot *CustomerSnapshot `json:"snapshot"`
}

// CustomerSnapshot is a frozen copy of customer data captured at finalization.
type CustomerSnapshot struct {
	Name      string   `json:"name"`
	SIRET     string   `json:"siret,omitempty"`
	VATNumber string   `json:"vatNumber,omitempty"`
	Address   *Address `json:"address"`
}

// Address is a postal address.
type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	PostalCode string `json:"postalCode"`
	City       string `json:"city"`
	Country    string `json:"country"`
}

// Contact is a contact person.
type Contact struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Role      string `json:"role,omitempty"`
}

// LineItem is a single line on an invoice, quote, or credit note.
type LineItem struct {
	ID              string `json:"id"`
	Description     string `json:"description"`
	Quantity        string `json:"quantity"` // decimal-string count, not money
	Unit            string `json:"unit"`
	UnitPrice       int    `json:"unitPrice"`       // integer centimes
	DiscountPercent int    `json:"discountPercent"` // centièmes de pourcent
	LineAmount      int    `json:"lineAmount"`
	VATRate         int    `json:"vatRate"` // centièmes de pourcent
	VATCode         string `json:"vatCode"`
	VATexCode       string `json:"vatexCode,omitempty"` // specific VATEX code (BT-121), when set
	VATAmount       int    `json:"vatAmount"`
	LineTotal       int    `json:"lineTotal"`
	Product         string `json:"product"`
}

// VATBreakdown is a VAT subtotal grouped by rate.
type VATBreakdown struct {
	Rate      int    `json:"rate"` // centièmes de pourcent
	Code      string `json:"code"`
	VATexCode string `json:"vatexCode,omitempty"`
	Base      int    `json:"base"` // integer centimes
	Amount    int    `json:"amount"`
}

// Totals is the financial summary of a document.
type Totals struct {
	TotalHT        int             `json:"totalHT"` // integer centimes
	DiscountAmount int             `json:"discountAmount"`
	VATBreakdown   []*VATBreakdown `json:"vatBreakdown"`
	TotalVAT       int             `json:"totalVAT"`
	TotalTTC       int             `json:"totalTTC"`
	AmountDue      int             `json:"amountDue"`
	AmountPaid     int             `json:"amountPaid"`
}

// LifecycleEntry records a status change in the audit trail.
type LifecycleEntry struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Source    string `json:"source"`
	Details   string `json:"details,omitempty"`
}

// ItemParams defines a line item. Quantity is a decimal string (e.g. "2.5").
// UnitPrice in centimes, VATRate/DiscountPercent in centipercent.
type ItemParams struct {
	Description string `json:"description"`
	Quantity    string `json:"quantity"`
	Unit        string `json:"unit"`
	UnitPrice   int    `json:"unitPrice"`
	VATRate     int    `json:"vatRate"`
	VATCode     string `json:"vatCode"`
	// VATexCode is an optional specific VATEX exemption code (BT-121), e.g.
	// "VATEX-FR-261", used when the basis differs from the VAT category default.
	VATexCode       string `json:"vatexCode,omitempty"`
	DiscountPercent int    `json:"discountPercent,omitempty"`
	Product         string `json:"product,omitempty"`
}

// BuyerParams identifies the buyer (B2B recipient) on an invoice.
// CompanyName and Address are required (CIUS-FR BT-44/BG-8).
type BuyerParams struct {
	CompanyName     string   `json:"companyName"`
	Siret           string   `json:"siret,omitempty"`
	VATNumber       string   `json:"vatNumber,omitempty"`
	Address         *Address `json:"address"`
	DeliveryAddress *Address `json:"deliveryAddress,omitempty"`
}

// InvoiceParams are the parameters for creating a draft invoice.
type InvoiceParams struct {
	Customer string       `json:"customerId"`
	Type     string       `json:"type,omitempty"`
	Buyer    *BuyerParams `json:"buyer"`
	// TaxDecisionID backs the invoice with an immutable tax decision.
	// Required: every invoice is created from a decision, and the VAT comes
	// from it — the invoice never restates a rate.
	TaxDecisionID string `json:"taxDecisionId"`
	// DecisionLines carries presentation only — unit and catalogue product. The
	// rate, category, VATEX code and legal mention all come from the decision.
	// Required, one entry per decided line.
	DecisionLines       []*DecisionLineParams `json:"decisionLines"`
	Dates               *InvoiceDatesParams   `json:"dates"`
	Payment             *PaymentTermsParams   `json:"payment"`
	PurchaseOrderNumber string                `json:"purchaseOrderNumber,omitempty"`
	Notes               string                `json:"notes,omitempty"`
	// Deposits links fully paid deposit invoices (386) whose TTC is deducted
	// from this balance invoice (BT-113, CGI art. 289). Max 20. Settled
	// server-side against the DECIDED amount due.
	Deposits []*DepositParam `json:"deposits,omitempty"`
	// Schedule is a payment schedule of 2 to 12 instalments. Validated
	// server-side: they must distribute exactly the decided amount due.
	Schedule []*ScheduleParam       `json:"schedule,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// AutoFinalize finalizes the invoice in the same call (assigns its number).
	AutoFinalize bool `json:"autoFinalize,omitempty"`
	// AutoSend finalizes then sends the invoice in the same call: by email to the
	// customer and/or by deposit to the connected PA.
	AutoSend *AutoSendParams `json:"autoSend,omitempty"`

	IdempotencyKey string `json:"-"`
}

// DecisionLineParams is a presentation-only line of a decision-backed document.
//
// It carries no VAT: the rate, the category, the VATEX code and the legal
// mention all come from the decided line it references.
type DecisionLineParams struct {
	// TaxLineRef is the reference of the decided line this document line renders.
	TaxLineRef string `json:"taxLineRef"`
	Unit       string `json:"unit"`
	Product    string `json:"product,omitempty"`
}

// AutoSendParams selects the one-shot delivery channels for invoice creation.
type AutoSendParams struct {
	Email bool `json:"email,omitempty"`
	PA    bool `json:"pa,omitempty"`
}

// DepositParam links a deposit invoice to a balance invoice.
type DepositParam struct {
	InvoiceID string `json:"invoiceId"`
}

// ScheduleParam is a payment-schedule instalment. Amount in integer centimes.
type ScheduleParam struct {
	Amount  int    `json:"amount"`
	DueDate string `json:"dueDate"`
	Label   string `json:"label,omitempty"`
}

// InvoiceDatesParams are date overrides for invoice creation.
type InvoiceDatesParams struct {
	Issued       string `json:"issued"`
	Due          string `json:"due"`
	ServiceStart string `json:"serviceStart,omitempty"`
	ServiceEnd   string `json:"serviceEnd,omitempty"`
}

// PaymentTermsParams are payment terms for an invoice.
// LatePaymentRate, CollectionFee and EarlyPaymentDiscount are decimal strings.
type PaymentTermsParams struct {
	Terms                string `json:"terms"`
	TermsDays            int    `json:"termsDays"`
	Method               string `json:"method"`
	IBAN                 string `json:"iban,omitempty"`
	BIC                  string `json:"bic,omitempty"`
	EarlyPaymentDiscount string `json:"earlyPaymentDiscount,omitempty"`
	LatePaymentRate      string `json:"latePaymentRate"`
	CollectionFee        string `json:"collectionFee"`
}

// InvoiceUpdateParams are the parameters for updating a draft's non-fiscal
// fields. The commercial operation and its VAT belong to the decision, which
// is immutable: to change the operation, take a new decision and create a new
// draft.
type InvoiceUpdateParams struct {
	Dates   *InvoiceDatesParams `json:"dates,omitempty"`
	Payment *PaymentTermsParams `json:"payment,omitempty"`
	// Notes is a pointer so that an empty string clears the notes instead of
	// being dropped by omitempty. Leave nil to keep the current notes.
	Notes               *string                `json:"notes,omitempty"`
	PurchaseOrderNumber string                 `json:"purchaseOrderNumber,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// InvoiceStatusResponse is a lightweight status check.
type InvoiceStatusResponse struct {
	Status     string `json:"status"`
	Einvoicing struct {
		PAStatus string `json:"paStatus"`
		PAID     string `json:"paId"`
	} `json:"einvoicing"`
	Dates struct {
		Due         string `json:"due"`
		FinalizedAt string `json:"finalizedAt"`
	} `json:"dates"`
}

// InvoiceVerifyResponse is the result of archive hash chain verification.
type InvoiceVerifyResponse struct {
	ID          string          `json:"id"`
	Object      string          `json:"object"`
	Verified    bool            `json:"verified"`
	Archive     *InvoiceArchive `json:"archive"`
	ChainLength int             `json:"chain_length"`
	Details     string          `json:"details"`
}

// DocumentResponse is returned by document generation endpoints.
// URL is set for cached documents; job fields are set for async generation (HTTP 202).
type DocumentResponse struct {
	URL       string `json:"url,omitempty"`
	ExpiresIn int    `json:"expires_in,omitempty"`

	ID        string `json:"id,omitempty"`
	Object    string `json:"object,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
	InvoiceID string `json:"invoiceId,omitempty"`
}

// PaymentLinkResponse is returned when creating a Stripe payment link.
type PaymentLinkResponse struct {
	Object    string `json:"object"`
	URL       string `json:"url"`
	SessionID string `json:"session_id"`
}

// PaymentLinkParams configures a payment link.
type PaymentLinkParams struct {
	SuccessURL string `json:"success_url,omitempty"`
	CancelURL  string `json:"cancel_url,omitempty"`
}

// PaymentTokenResponse is returned when generating a payment token.
type PaymentTokenResponse struct {
	Object    string `json:"object"`
	Token     string `json:"token"`
	PayURL    string `json:"pay_url"`
	ExpiresAt string `json:"expires_at"`
}

// InvoiceService operates on invoices.
type InvoiceService struct {
	client *httpClient
}

// Create creates a new draft invoice from an immutable tax decision.
//
// TaxDecisionID and DecisionLines are required and checked locally, before
// any HTTP call: every invoice is created from a decision, and the invoice
// never restates a rate. Deposits and Schedule travel alongside the decision;
// both are settled server-side against the DECIDED amount due.
func (s *InvoiceService) Create(params *InvoiceParams) (*Invoice, error) {
	if params == nil {
		return nil, errors.New("facturino: invoice params are required")
	}
	if params.TaxDecisionID == "" {
		return nil, errors.New("facturino: TaxDecisionID is required; every invoice is created from an immutable tax decision (TaxDecisions.Create)")
	}
	if len(params.DecisionLines) == 0 {
		return nil, errors.New("facturino: DecisionLines is required and non-empty; each entry references a decision line by TaxLineRef and completes the document-only details (unit, product)")
	}
	var inv Invoice
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post("/invoices", params, &inv, opts)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// InvoiceGetParams tunes a single-invoice retrieval. Expand inlines
// related resources into Invoice.Expanded; supported values are
// "customer", "items.product" and "credit_notes". Requesting
// "credit_notes" also populates Expanded.NetBalance.
type InvoiceGetParams struct {
	Expand []string
}

// Get retrieves an invoice by ID. Pass an optional InvoiceGetParams to
// expand related resources (for example expand "credit_notes" to inline
// the invoice's credit notes and net balance under Invoice.Expanded).
func (s *InvoiceService) Get(id string, params ...*InvoiceGetParams) (*Invoice, error) {
	var vals url.Values
	if len(params) > 0 && params[0] != nil && len(params[0].Expand) > 0 {
		vals = url.Values{"expand": {strings.Join(params[0].Expand, ",")}}
	}
	var inv Invoice
	err := s.client.get(fmt.Sprintf("/invoices/%s", id), vals, &inv)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// Update updates a draft invoice.
func (s *InvoiceService) Update(id string, params *InvoiceUpdateParams) (*Invoice, error) {
	var inv Invoice
	err := s.client.patch(fmt.Sprintf("/invoices/%s", id), params, &inv)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// Delete soft-deletes a draft invoice.
func (s *InvoiceService) Delete(id string) error {
	return s.client.del(fmt.Sprintf("/invoices/%s", id))
}

// InvoiceListParams adds invoice-specific filters to ListParams.
type InvoiceListParams struct {
	ListParams

	// ConvertedFrom filters invoices to those converted from the given
	// quote (a "quo_"-prefixed identifier).
	ConvertedFrom string
}

// List returns a paginated iterator over invoices.
func (s *InvoiceService) List(params *InvoiceListParams) *InvoiceIterator {
	var lp *ListParams
	if params != nil {
		lp = &params.ListParams
	}
	iter := newIterator[*Invoice](s.client, "/invoices", lp, decodeInvoice)
	if params != nil && params.ConvertedFrom != "" {
		iter.extraParams = url.Values{"convertedFrom": {params.ConvertedFrom}}
	}
	return &InvoiceIterator{iter: iter}
}

// CommercialDraftLine is one line of a commercial draft: the operation as
// stated, with NO VAT. UnitPrice is in integer cents, in the draft's price
// mode; Quantity is a decimal string. RateCategory is the band the seller asks
// for — the decision concludes the actual rate.
type CommercialDraftLine struct {
	// Reference is assigned server-side; the decision reuses it.
	Reference      string               `json:"reference"`
	Description    string               `json:"description"`
	Quantity       string               `json:"quantity"`
	Unit           string               `json:"unit"`
	UnitPrice      int                  `json:"unitPrice"`
	SupplyCategory string               `json:"supplyCategory"`
	RateCategory   string               `json:"rateCategory"`
	Discount       *TaxDecisionDiscount `json:"discount,omitempty"`
	Product        string               `json:"product,omitempty"`
}

// CommercialDraft is the operation an undecided draft states. TotalCents is a
// COMMERCIAL total: neither a decided net nor a decided gross amount, because
// nothing has been decided yet.
type CommercialDraft struct {
	PriceMode  string                 `json:"priceMode"`
	Lines      []*CommercialDraftLine `json:"lines"`
	TotalCents int                    `json:"totalCents"`
}

// BindTaxDecisionParams binds a FINAL tax decision to a commercial draft that
// already exists. It carries the decision and the presentation of its lines,
// and nothing else: the draft already states the buyer, the dates and the
// payment terms, and the decision states the whole fiscal content.
type BindTaxDecisionParams struct {
	// TaxDecisionID is the FINAL decision to freeze onto the draft.
	TaxDecisionID string `json:"taxDecisionId"`
	// DecisionLines carries presentation only — unit and catalogue product —
	// one entry per decided line, matched by TaxLineRef.
	DecisionLines []*DecisionLineParams `json:"decisionLines"`
	// IdempotencyKey is sent as the Idempotency-Key header.
	IdempotencyKey string `json:"-"`
}

// BindTaxDecision freezes a FINAL tax decision onto a commercial draft that
// already exists — typically the one Quotes.Convert produced.
//
// It closes the quote cycle on ONE document:
//
//	converted, _ := client.Quotes.Convert(quoteID)
//	decision, _ := client.TaxDecisions.Create(input)
//	client.Invoices.BindTaxDecision(converted.InvoiceID, &facturino.BindTaxDecisionParams{
//		TaxDecisionID: decision.ID,
//		DecisionLines: []*facturino.DecisionLineParams{{TaxLineRef: "l1", Unit: "unit"}},
//	})
//	client.Invoices.Finalize(converted.InvoiceID)
//
// The invoice stays a DRAFT: binding freezes the VAT, Finalize issues it.
// Idempotent on the decision — replaying the same call returns the same invoice.
func (s *InvoiceService) BindTaxDecision(id string, params *BindTaxDecisionParams) (*Invoice, error) {
	if params == nil {
		return nil, errors.New("facturino: bind params are required")
	}
	if params.TaxDecisionID == "" {
		return nil, errors.New("facturino: TaxDecisionID is required; a commercial draft is fiscalised by binding a FINAL tax decision to it")
	}
	if len(params.DecisionLines) == 0 {
		return nil, errors.New("facturino: DecisionLines is required and non-empty; each entry references a decision line by TaxLineRef and completes the document-only details (unit, product)")
	}
	var inv Invoice
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post(fmt.Sprintf("/invoices/%s/bind-tax-decision", id), params, &inv, opts)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// Finalize finalizes a draft invoice, assigning it a number.
func (s *InvoiceService) Finalize(id string) (*Invoice, error) {
	var inv Invoice
	err := s.client.post(fmt.Sprintf("/invoices/%s/finalize", id), nil, &inv, nil)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// finalizeBody carries the optional collection of a finalization.
type finalizeBody struct {
	Payment *PaymentParams `json:"payment,omitempty"`
}

// FinalizeWithPayment finalizes a draft invoice that was ALREADY COLLECTED
// before issuance: the numbering and the collection are applied in the same
// transaction, so the original PDF and Factur-X are rendered on a settled
// invoice and say so.
//
// payment is the very *PaymentParams that Payments.Create takes (amount in
// integer centimes), and the resulting payment is indistinguishable from one
// recorded afterwards.
//
// All or nothing: a collection beyond the amount due is refused
// (422 payment_exceeds_amount_due) and the invoice stays a draft — no number is
// burned. Use Finalize when no collection accompanies the issuance.
func (s *InvoiceService) FinalizeWithPayment(id string, payment *PaymentParams) (*Invoice, error) {
	var inv Invoice
	opts := &requestOption{}
	if payment != nil && payment.IdempotencyKey != "" {
		opts.idempotencyKey = payment.IdempotencyKey
	}
	err := s.client.post(
		fmt.Sprintf("/invoices/%s/finalize", id),
		finalizeBody{Payment: payment}, &inv, opts,
	)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// Send submits a finalized invoice to the PA for e-invoicing.
func (s *InvoiceService) Send(id string) (*Invoice, error) {
	var inv Invoice
	err := s.client.post(fmt.Sprintf("/invoices/%s/send", id), nil, &inv, nil)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// InvoiceEmailParams contains optional overrides for the email send.
// All fields are optional; with a nil params the email is dispatched
// to the customer's saved address with the default subject line.
type InvoiceEmailParams struct {
	RecipientEmail string `json:"recipientEmail,omitempty"`
	CustomMessage  string `json:"customMessage,omitempty"`
	IncludeXML     bool   `json:"includeXml,omitempty"`
	CustomSubject  string `json:"customSubject,omitempty"`
}

// InvoiceEmailResponse is returned by Email. When the PDF is still being
// generated server-side the response carries Status="pending" and a
// JobID to poll; otherwise Status="sent" along with the recipient and
// the dispatch timestamp.
type InvoiceEmailResponse struct {
	Status    string `json:"status"`
	InvoiceID string `json:"invoiceId,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	SentAt    string `json:"sentAt,omitempty"`
	JobID     string `json:"jobId,omitempty"`
	PollURL   string `json:"pollUrl,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// Email sends a finalized invoice to its customer by email with the
// PDF (and optionally the Factur-X XML CII) attached.
func (s *InvoiceService) Email(id string, params *InvoiceEmailParams) (*InvoiceEmailResponse, error) {
	var resp InvoiceEmailResponse
	err := s.client.post(fmt.Sprintf("/invoices/%s/email", id), params, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Cancel cancels a draft invoice (status `draft` only). Finalized
// invoices are immutable under French law (CGI art. 289) — to refund
// or correct a finalized invoice, issue a credit note via
// CreditNotes.Create() instead.
func (s *InvoiceService) Cancel(id string) error {
	return s.client.post(fmt.Sprintf("/invoices/%s/cancel", id), nil, nil, nil)
}

// InvoiceRemindParams optionally tunes the reminder. Level can be
// 1 (friendly), 2 (firm) or 3 (formal). Omitting both fields sends
// a level-1 reminder with the default body.
type InvoiceRemindParams struct {
	Level   int    `json:"level,omitempty"`
	Message string `json:"message,omitempty"`
}

// Remind sends a payment reminder email for an overdue invoice.
func (s *InvoiceService) Remind(id string, params *InvoiceRemindParams) error {
	var resp map[string]interface{}
	return s.client.post(fmt.Sprintf("/invoices/%s/remind", id), params, &resp, nil)
}

// Clone creates a new draft invoice by cloning an existing one.
func (s *InvoiceService) Clone(id string) (*Invoice, error) {
	var inv Invoice
	err := s.client.post(fmt.Sprintf("/invoices/%s/clone", id), nil, &inv, nil)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// GetPDF retrieves a cached PDF URL or triggers async generation.
func (s *InvoiceService) GetPDF(id string) (*DocumentResponse, error) {
	var resp DocumentResponse
	err := s.client.get(fmt.Sprintf("/invoices/%s/pdf", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFacturX retrieves a cached Factur-X URL or triggers async generation.
func (s *InvoiceService) GetFacturX(id string) (*DocumentResponse, error) {
	var resp DocumentResponse
	err := s.client.get(fmt.Sprintf("/invoices/%s/facturx", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetXML retrieves the invoice XML. Defaults to CII; pass "ubl" for UBL format.
func (s *InvoiceService) GetXML(id string, format string) ([]byte, error) {
	params := url.Values{}
	if format != "" {
		params.Set("format", format)
	}
	path := fmt.Sprintf("/invoices/%s/xml", id)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	data, _, err := s.client.doRaw("GET", path, nil, nil)
	return data, err
}

// GetStatus returns a lightweight status check.
func (s *InvoiceService) GetStatus(id string) (*InvoiceStatusResponse, error) {
	var resp InvoiceStatusResponse
	err := s.client.get(fmt.Sprintf("/invoices/%s/status", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Verify checks the archive hash chain integrity.
func (s *InvoiceService) Verify(id string) (*InvoiceVerifyResponse, error) {
	var resp InvoiceVerifyResponse
	err := s.client.get(fmt.Sprintf("/invoices/%s/verify", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListEvents returns lifecycle events for the invoice.
func (s *InvoiceService) ListEvents(id string) (*ListResponse, error) {
	var resp ListResponse
	err := s.client.get(fmt.Sprintf("/invoices/%s/events", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAuditTrail returns audit log entries for the invoice.
func (s *InvoiceService) GetAuditTrail(id string, params *ListParams) (*ListResponse, error) {
	var resp ListResponse
	var vals url.Values
	if params != nil {
		vals = params.toValues()
	}
	err := s.client.get(fmt.Sprintf("/invoices/%s/audit-trail", id), vals, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GenerateAuditTrailPDF triggers async audit trail PDF generation.
func (s *InvoiceService) GenerateAuditTrailPDF(id string) (*DocumentResponse, error) {
	var resp DocumentResponse
	err := s.client.post(fmt.Sprintf("/invoices/%s/audit-trail/pdf", id), nil, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePaymentLink creates a Stripe payment link (Pro+ plan).
func (s *InvoiceService) CreatePaymentLink(id string, params *PaymentLinkParams) (*PaymentLinkResponse, error) {
	var resp PaymentLinkResponse
	err := s.client.post(fmt.Sprintf("/invoices/%s/payment-link", id), params, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePaymentToken generates a secure payment token (Pro+ plan).
func (s *InvoiceService) CreatePaymentToken(id string) (*PaymentTokenResponse, error) {
	var resp PaymentTokenResponse
	err := s.client.post(fmt.Sprintf("/invoices/%s/payment-token", id), nil, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// InvoicePortalLinkResponse carries the signed client-portal URL and
// the token embedded in it.
type InvoicePortalLinkResponse struct {
	URL       string `json:"url"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// CreatePortalLink generates a signed client-portal link for a
// finalized invoice. The URL points to a public, branded portal where
// the customer can view the invoice, download the PDF and trigger the
// payment flow. The embedded token grants read-only access to a single
// invoice and expires after the configured lifetime. Returns
// invalid_status_transition if the invoice is still in draft.
func (s *InvoiceService) CreatePortalLink(id string) (*InvoicePortalLinkResponse, error) {
	var resp InvoicePortalLinkResponse
	if err := s.client.post(fmt.Sprintf("/invoices/%s/portal-link", id), nil, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// IncomingInvoiceParams records a supplier invoice received outside the
// platform (manual entry), so it appears in the inbound register alongside
// e-invoices delivered via the PA.
type IncomingInvoiceParams struct {
	SenderName  string `json:"senderName"`
	SenderSiret string `json:"senderSiret,omitempty"`
	// Amount is the total incl. VAT, in integer cents.
	Amount    int    `json:"amount"`
	Reference string `json:"reference,omitempty"`
	Notes     string `json:"notes,omitempty"`

	IdempotencyKey string `json:"-"`
}

// CreateIncoming records a supplier invoice received outside the platform.
func (s *InvoiceService) CreateIncoming(params *IncomingInvoiceParams) (*ReceivedInvoice, error) {
	var inv ReceivedInvoice
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post("/invoices/incoming", params, &inv, opts)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// ListIncoming returns a paginated iterator over received (incoming) invoices.
// These are supplier invoices received via a PA, so each item is a
// ReceivedInvoice — iterate with ReceivedInvoiceIterator.ReceivedInvoice().
func (s *InvoiceService) ListIncoming(params *ListParams) *ReceivedInvoiceIterator {
	iter := newIterator[*ReceivedInvoice](s.client, "/invoices/incoming", params, decodeReceivedInvoice)
	return &ReceivedInvoiceIterator{iter: iter}
}

// InvoiceIterator iterates over invoices.
type InvoiceIterator struct {
	iter *Iterator[*Invoice]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *InvoiceIterator) Next() bool { return it.iter.Next() }

// Invoice returns the most recently fetched invoice.
func (it *InvoiceIterator) Invoice() *Invoice { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *InvoiceIterator) Err() error { return it.iter.Err() }

func decodeInvoice(raw json.RawMessage) (*Invoice, error) {
	var inv Invoice
	err := json.Unmarshal(raw, &inv)
	return &inv, err
}
