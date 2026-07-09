package facturino

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
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
	PAID             string `json:"paId"`
	PAStatus         string `json:"paStatus"`
	PATransactionID  string `json:"paTransactionId"`
	PAErrorCode      string `json:"paErrorCode"`
	PAIdempotencyKey string `json:"paIdempotencyKey"`
	PeppolDeliveryID string `json:"peppolDeliveryId"`
	EreportingID     string `json:"ereportingId"`
	SentAt           string `json:"sentAt"`
	TrackingID       string `json:"trackingId"`
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
	VATAmount       int    `json:"vatAmount"`
	LineTotal       int    `json:"lineTotal"`
	Product         string `json:"product"`
}

// VATBreakdown is a VAT subtotal grouped by rate.
type VATBreakdown struct {
	Rate   int    `json:"rate"` // centièmes de pourcent
	Code   string `json:"code"`
	Base   int    `json:"base"` // integer centimes
	Amount int    `json:"amount"`
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
	Description     string `json:"description"`
	Quantity        string `json:"quantity"`
	Unit            string `json:"unit"`
	UnitPrice       int    `json:"unitPrice"`
	VATRate         int    `json:"vatRate"`
	VATCode         string `json:"vatCode"`
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
	Customer            string                 `json:"customerId"`
	Type                string                 `json:"type,omitempty"`
	Buyer               *BuyerParams           `json:"buyer"`
	Items               []*ItemParams          `json:"lines"`
	Dates               *InvoiceDatesParams    `json:"dates"`
	Payment             *PaymentTermsParams    `json:"payment"`
	PurchaseOrderNumber string                 `json:"purchaseOrderNumber,omitempty"`
	Notes               string                 `json:"notes,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`

	// AutoFinalize finalizes the invoice in the same call (assigns its number).
	AutoFinalize bool `json:"autoFinalize,omitempty"`
	// AutoSend finalizes then sends the invoice in the same call: by email to the
	// customer and/or by deposit to the connected PA.
	AutoSend *AutoSendParams `json:"autoSend,omitempty"`

	IdempotencyKey string `json:"-"`
}

// AutoSendParams selects the one-shot delivery channels for invoice creation.
type AutoSendParams struct {
	Email bool `json:"email,omitempty"`
	PA    bool `json:"pa,omitempty"`
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

// InvoiceUpdateParams are the parameters for updating a draft.
type InvoiceUpdateParams struct {
	Items   []*ItemParams       `json:"lines,omitempty"`
	Dates   *InvoiceDatesParams `json:"dates,omitempty"`
	Payment *PaymentTermsParams `json:"payment,omitempty"`
	// Notes is a pointer so that an empty string clears the notes instead of
	// being dropped by omitempty. Leave nil to keep the current notes.
	Notes    *string                `json:"notes,omitempty"`
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
