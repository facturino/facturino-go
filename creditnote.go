package facturino

import (
	"encoding/json"
	"fmt"
)

// CreditNote is a credit note (avoir).
type CreditNote struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Livemode bool   `json:"livemode"`

	Customer         *CustomerRef `json:"customer"`
	RelatedInvoiceID string       `json:"relatedInvoiceId"`
	Status           string       `json:"status"`
	CreditNoteType   string       `json:"creditNoteType"`
	Number           string       `json:"number"`
	Currency         string       `json:"currency"`

	ReasonCode string `json:"reasonCode"`
	Reason     string `json:"reason,omitempty"`

	Items  []*LineItem `json:"items"`
	Totals *Totals     `json:"totals"`

	Dates *CreditNoteDates `json:"dates"`

	Einvoicing *CreditNoteEinvoicing `json:"einvoicing,omitempty"`
	Files      *CreditNoteFiles      `json:"files,omitempty"`
	Notes      string                `json:"notes,omitempty"`

	Archive *CreditNoteArchive `json:"archive,omitempty"`

	Created string `json:"created"`
	Updated string `json:"updated"`
}

// CreditNoteDates holds credit note dates.
type CreditNoteDates struct {
	Issued      string `json:"issued"`
	FinalizedAt string `json:"finalizedAt,omitempty"`
	SentAt      string `json:"sentAt,omitempty"`
}

// CreditNoteEinvoicing holds e-invoicing status.
type CreditNoteEinvoicing struct {
	PAID        string `json:"paId"`
	PAStatus    string `json:"paStatus"`
	DepositedAt string `json:"depositedAt"`
}

// CreditNoteFiles holds generated document paths.
type CreditNoteFiles struct {
	PDFPath     string `json:"pdfPath,omitempty"`
	FacturXPath string `json:"facturxPath,omitempty"`
}

// CreditNoteArchive holds hash chain data.
type CreditNoteArchive struct {
	Hash         string `json:"hash"`
	PreviousHash string `json:"previousHash"`
	ArchivedAt   string `json:"archivedAt"`
}

// CreditNoteParams are the parameters for creating a credit note.
type CreditNoteParams struct {
	Customer         string           `json:"customerId"`
	RelatedInvoiceID string           `json:"relatedInvoiceId"`
	CreditNoteType   string           `json:"creditNoteType"`
	ReasonCode       string           `json:"reasonCode"`
	Reason           string           `json:"reason,omitempty"`
	Items            []*ItemParams    `json:"items"`
	Dates            *CreditNoteDates `json:"dates"`
	Notes            string           `json:"notes,omitempty"`

	IdempotencyKey string `json:"-"`
}

// CreditNoteUpdateParams are the parameters for updating a draft credit note.
type CreditNoteUpdateParams struct {
	Items []*ItemParams `json:"items,omitempty"`
	// ReasonCode is set only on create; the update endpoint rejects it.
	Reason string `json:"reason,omitempty"`
	Notes  string `json:"notes,omitempty"`
}

// CreditNoteService operates on credit notes.
type CreditNoteService struct {
	client *httpClient
}

// Create creates a new draft credit note.
func (s *CreditNoteService) Create(params *CreditNoteParams) (*CreditNote, error) {
	var cn CreditNote
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post("/credit-notes", params, &cn, opts)
	if err != nil {
		return nil, err
	}
	return &cn, nil
}

// Get retrieves a credit note by ID.
func (s *CreditNoteService) Get(id string) (*CreditNote, error) {
	var cn CreditNote
	err := s.client.get(fmt.Sprintf("/credit-notes/%s", id), nil, &cn)
	if err != nil {
		return nil, err
	}
	return &cn, nil
}

// Update updates a draft credit note.
func (s *CreditNoteService) Update(id string, params *CreditNoteUpdateParams) (*CreditNote, error) {
	var cn CreditNote
	err := s.client.patch(fmt.Sprintf("/credit-notes/%s", id), params, &cn)
	if err != nil {
		return nil, err
	}
	return &cn, nil
}

// Delete soft-deletes a draft credit note.
func (s *CreditNoteService) Delete(id string) error {
	return s.client.del(fmt.Sprintf("/credit-notes/%s", id))
}

// List returns a paginated iterator over credit notes.
func (s *CreditNoteService) List(params *ListParams) *CreditNoteIterator {
	iter := newIterator[*CreditNote](s.client, "/credit-notes", params, decodeCreditNote)
	return &CreditNoteIterator{iter: iter}
}

// Finalize finalizes a draft credit note.
func (s *CreditNoteService) Finalize(id string) (*CreditNote, error) {
	var cn CreditNote
	err := s.client.post(fmt.Sprintf("/credit-notes/%s/finalize", id), nil, &cn, nil)
	if err != nil {
		return nil, err
	}
	return &cn, nil
}

// Send submits a finalized credit note to the PA for e-invoicing.
func (s *CreditNoteService) Send(id string) (*CreditNote, error) {
	var cn CreditNote
	err := s.client.post(fmt.Sprintf("/credit-notes/%s/send", id), nil, &cn, nil)
	if err != nil {
		return nil, err
	}
	return &cn, nil
}

// CreditNoteRefundParams are the optional parameters for Refund. Amount is in
// integer centimes and defaults to the full credit-note total when nil.
type CreditNoteRefundParams struct {
	Amount     *int   `json:"amount,omitempty"`
	Method     string `json:"method,omitempty"`
	RefundedAt string `json:"refundedAt,omitempty"`

	IdempotencyKey string `json:"-"`
}

// CreditNoteRefund is the result of recording a credit-note refund: a negative
// `refund` payment written on the linked invoice.
type CreditNoteRefund struct {
	ID           string `json:"id"`
	Object       string `json:"object"`
	CreditNoteID string `json:"creditNoteId"`
	InvoiceID    string `json:"invoiceId"`
	Amount       int    `json:"amount"`
}

// Refund records the disbursement of a finalized credit note back to the customer.
func (s *CreditNoteService) Refund(id string, params *CreditNoteRefundParams) (*CreditNoteRefund, error) {
	var resp CreditNoteRefund
	opts := &requestOption{}
	if params != nil && params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post(fmt.Sprintf("/credit-notes/%s/refund", id), params, &resp, opts)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreditNoteEmailParams contains optional overrides for the email send.
type CreditNoteEmailParams struct {
	RecipientEmail string `json:"recipientEmail,omitempty"`
	CustomMessage  string `json:"customMessage,omitempty"`
	IncludeXML     bool   `json:"includeXml,omitempty"`
	CustomSubject  string `json:"customSubject,omitempty"`
}

// CreditNoteEmailResponse is returned by Email. When the PDF is still being
// generated server-side the response carries `Status = "pending"` and a
// `JobID` to poll; otherwise `Status = "sent"`.
type CreditNoteEmailResponse struct {
	Status       string `json:"status"`
	CreditNoteID string `json:"creditNoteId,omitempty"`
	Recipient    string `json:"recipient,omitempty"`
	SentAt       string `json:"sentAt,omitempty"`
	JobID        string `json:"jobId,omitempty"`
	PollURL      string `json:"pollUrl,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

// Email sends the credit note to its customer by email with the PDF
// (and optionally the Factur-X XML) attached.
func (s *CreditNoteService) Email(id string, params *CreditNoteEmailParams) (*CreditNoteEmailResponse, error) {
	var resp CreditNoteEmailResponse
	err := s.client.post(fmt.Sprintf("/credit-notes/%s/email", id), params, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPDF retrieves a cached PDF URL or triggers async generation.
func (s *CreditNoteService) GetPDF(id string) (*DocumentResponse, error) {
	var resp DocumentResponse
	err := s.client.get(fmt.Sprintf("/credit-notes/%s/pdf", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFacturX retrieves a cached Factur-X URL or triggers async generation.
func (s *CreditNoteService) GetFacturX(id string) (*DocumentResponse, error) {
	var resp DocumentResponse
	err := s.client.get(fmt.Sprintf("/credit-notes/%s/facturx", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetXML retrieves the credit note XML. Defaults to CII; pass "ubl" for UBL format.
func (s *CreditNoteService) GetXML(id string, format string) ([]byte, error) {
	path := fmt.Sprintf("/credit-notes/%s/xml", id)
	if format != "" {
		path += "?format=" + format
	}
	data, _, err := s.client.doRaw("GET", path, nil, nil)
	return data, err
}

// CreditNoteIterator iterates over credit notes.
type CreditNoteIterator struct {
	iter *Iterator[*CreditNote]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *CreditNoteIterator) Next() bool { return it.iter.Next() }

// CreditNote returns the most recently fetched credit note.
func (it *CreditNoteIterator) CreditNote() *CreditNote { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *CreditNoteIterator) Err() error { return it.iter.Err() }

func decodeCreditNote(raw json.RawMessage) (*CreditNote, error) {
	var cn CreditNote
	err := json.Unmarshal(raw, &cn)
	return &cn, err
}
