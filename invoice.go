package facturino

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Invoice is a Facturino invoice.
type Invoice struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Number   string `json:"number"`
	Currency string `json:"currency"`
	Livemode bool   `json:"livemode"`

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

	Created string `json:"created"`
	Updated string `json:"updated"`
}

// InvoiceDates holds invoice dates.
type InvoiceDates struct {
	Issued       string `json:"issued"`
	Due          string `json:"due"`
	ServiceStart string `json:"serviceStart,omitempty"`
	ServiceEnd   string `json:"serviceEnd,omitempty"`
	FinalizedAt  string `json:"finalizedAt,omitempty"`
	SentAt       string `json:"sentAt,omitempty"`
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

// InvoiceEinvoicing holds e-invoicing (PA) status.
type InvoiceEinvoicing struct {
	PAID              string `json:"paId"`
	PAStatus          string `json:"paStatus"`
	PATransactionID   string `json:"paTransactionId"`
	PAErrorCode       string `json:"paErrorCode"`
	PAIdempotencyKey  string `json:"paIdempotencyKey"`
	PeppolDeliveryID  string `json:"peppolDeliveryId"`
	EreportingID      string `json:"ereportingId"`
	SentAt            string `json:"sentAt"`
	TrackingID        string `json:"trackingId"`
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
	PDFPath      string `json:"pdfPath,omitempty"`
	FacturXPath  string `json:"facturxPath,omitempty"`
	XMLPath      string `json:"xmlPath,omitempty"`
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
	Quantity        string `json:"quantity"`
	Unit            string `json:"unit"`
	UnitPrice       string `json:"unitPrice"`
	DiscountPercent string `json:"discountPercent"`
	LineAmount      string `json:"lineAmount"`
	VATRate         string `json:"vatRate"`
	VATCode         string `json:"vatCode"`
	VATAmount       string `json:"vatAmount"`
	LineTotal       string `json:"lineTotal"`
	Product         string `json:"product"`
}

// VATBreakdown is a VAT subtotal grouped by rate.
type VATBreakdown struct {
	Rate   string `json:"rate"`
	Code   string `json:"code"`
	Base   string `json:"base"`
	Amount string `json:"amount"`
}

// Totals is the financial summary of a document.
type Totals struct {
	TotalHT        string          `json:"totalHT"`
	DiscountAmount string          `json:"discountAmount"`
	VATBreakdown   []*VATBreakdown `json:"vatBreakdown"`
	TotalVAT       string          `json:"totalVAT"`
	TotalTTC       string          `json:"totalTTC"`
	AmountDue      string          `json:"amountDue"`
	AmountPaid     string          `json:"amountPaid"`
}

// LifecycleEntry records a status change in the audit trail.
type LifecycleEntry struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Source    string `json:"source"`
	Details   string `json:"details,omitempty"`
}

// ItemParams defines a line item. UnitPrice in centimes, VATRate/DiscountPercent in centipercent.
type ItemParams struct {
	Description     string `json:"description"`
	Quantity        int    `json:"quantity"`
	Unit            string `json:"unit,omitempty"`
	UnitPrice       int    `json:"unitPrice"`
	VATRate         int    `json:"vatRate"`
	VATCode         string `json:"vatCode,omitempty"`
	DiscountPercent int    `json:"discountPercent,omitempty"`
	Product         string `json:"product,omitempty"`
}

// InvoiceParams are the parameters for creating a draft invoice.
type InvoiceParams struct {
	Customer string        `json:"customerId"`
	Type     string        `json:"type,omitempty"`
	Items    []*ItemParams `json:"lines"`
	Dates    *InvoiceDatesParams  `json:"dates,omitempty"`
	Payment  *PaymentTermsParams  `json:"payment,omitempty"`
	Notes    string        `json:"notes,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	IdempotencyKey string `json:"-"`
}

// InvoiceDatesParams are date overrides for invoice creation.
type InvoiceDatesParams struct {
	Issued       string `json:"issued"`
	Due          string `json:"due"`
	ServiceStart string `json:"serviceStart,omitempty"`
	ServiceEnd   string `json:"serviceEnd,omitempty"`
}

// PaymentTermsParams are payment terms for an invoice.
type PaymentTermsParams struct {
	Terms                string `json:"terms,omitempty"`
	TermsDays            int    `json:"termsDays,omitempty"`
	Method               string `json:"method,omitempty"`
	IBAN                 string `json:"iban,omitempty"`
	BIC                  string `json:"bic,omitempty"`
	EarlyPaymentDiscount int    `json:"earlyPaymentDiscount,omitempty"`
	LatePaymentRate      int    `json:"latePaymentRate,omitempty"`
	CollectionFee        int    `json:"collectionFee,omitempty"`
}

// InvoiceUpdateParams are the parameters for updating a draft.
type InvoiceUpdateParams struct {
	Items    []*ItemParams          `json:"lines,omitempty"`
	Dates    *InvoiceDatesParams    `json:"dates,omitempty"`
	Payment  *PaymentTermsParams    `json:"payment,omitempty"`
	Notes    string                 `json:"notes,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
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
	InvoiceID string `json:"invoice_id,omitempty"`
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

// Create creates a new draft invoice.
func (s *InvoiceService) Create(params *InvoiceParams) (*Invoice, error) {
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

// Get retrieves an invoice by ID.
func (s *InvoiceService) Get(id string) (*Invoice, error) {
	var inv Invoice
	err := s.client.get(fmt.Sprintf("/invoices/%s", id), nil, &inv)
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

// List returns a paginated iterator over invoices.
func (s *InvoiceService) List(params *ListParams) *InvoiceIterator {
	iter := newIterator[*Invoice](s.client, "/invoices", params, decodeInvoice)
	return &InvoiceIterator{iter: iter}
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

// Send submits a finalized invoice to the PA for e-invoicing.
func (s *InvoiceService) Send(id string) (*Invoice, error) {
	var inv Invoice
	err := s.client.post(fmt.Sprintf("/invoices/%s/send", id), nil, &inv, nil)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// Cancel attempts to cancel an invoice.
// Note: Under French e-invoicing regulations, this returns an error.
// Use CreditNotes.Create() to issue a corrective document instead.
func (s *InvoiceService) Cancel(id string) error {
	return s.client.post(fmt.Sprintf("/invoices/%s/cancel", id), nil, nil, nil)
}

// Remind sends a payment reminder for the invoice.
func (s *InvoiceService) Remind(id string) error {
	var resp map[string]interface{}
	return s.client.post(fmt.Sprintf("/invoices/%s/remind", id), nil, &resp, nil)
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

// CreateIncoming creates an incoming invoice (received from a supplier).
func (s *InvoiceService) CreateIncoming(params *InvoiceParams) (*Invoice, error) {
	var inv Invoice
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

// ListIncoming returns a paginated iterator over incoming invoices.
func (s *InvoiceService) ListIncoming(params *ListParams) *InvoiceIterator {
	iter := newIterator[*Invoice](s.client, "/invoices/incoming", params, decodeInvoice)
	return &InvoiceIterator{iter: iter}
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
