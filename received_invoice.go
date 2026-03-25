package facturino

import (
	"encoding/json"
	"fmt"
)

// ReceivedInvoice is an incoming invoice received via a PA (Plateforme Agreee).
type ReceivedInvoice struct {
	ID           string `json:"id"`
	Object       string `json:"object"`
	Livemode     bool   `json:"livemode"`
	PAInvoiceID  string `json:"paInvoiceId"`
	SourcePA     string `json:"sourcePA"`
	SourceFormat string `json:"sourceFormat"`
	Status       string `json:"status"`

	SenderSiret string `json:"senderSiret"`
	SenderName  string `json:"senderName"`

	Number   string `json:"number"`
	IssuedAt string `json:"issuedAt"`
	DueAt    string `json:"dueAt"`
	TotalHT  string `json:"totalHT"`
	TotalTVA string `json:"totalTVA"`
	TotalTTC string `json:"totalTTC"`

	XMLPath string `json:"xmlPath"`
	PDFPath string `json:"pdfPath,omitempty"`

	ApprovedAt    string `json:"approvedAt,omitempty"`
	RefusedAt     string `json:"refusedAt,omitempty"`
	RefusalReason string `json:"refusalReason,omitempty"`

	Einvoicing *ReceivedInvoiceEinvoicing `json:"einvoicing,omitempty"`

	Reconciled           bool   `json:"reconciled"`
	ReconciledPaymentID  string `json:"reconciledPaymentId,omitempty"`

	Lifecycle []*LifecycleEntry          `json:"lifecycle"`
	Metadata  map[string]interface{}     `json:"metadata"`

	Created string `json:"created"`
	Updated string `json:"updated"`
}

// ReceivedInvoiceEinvoicing holds PA lifecycle tracking for a received invoice.
type ReceivedInvoiceEinvoicing struct {
	FR204Sent        bool   `json:"fr204Sent,omitempty"`
	DirectoryEntryID string `json:"directoryEntryId,omitempty"`
}

// RefuseParams are the parameters for refusing a received invoice.
type RefuseParams struct {
	Reason string `json:"reason"`
}

// RecordPaymentParams are the parameters for recording a payment on a received invoice. Amount in centimes.
type RecordPaymentParams struct {
	Amount    int    `json:"amount"`
	Method    string `json:"method,omitempty"`
	Reference string `json:"reference,omitempty"`
	PaidAt    string `json:"paidAt,omitempty"`
}

// ReceivedInvoiceActionResponse is returned by approve, refuse, and suspend actions.
type ReceivedInvoiceActionResponse struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Status     string `json:"status,omitempty"`
	Reconciled bool   `json:"reconciled,omitempty"`
}

// ReceivedInvoiceService operates on received invoices (factures entrantes).
type ReceivedInvoiceService struct {
	client *httpClient
}

// List returns a paginated iterator over received invoices.
func (s *ReceivedInvoiceService) List(params *ListParams) *ReceivedInvoiceIterator {
	iter := newIterator[*ReceivedInvoice](s.client, "/received-invoices", params, decodeReceivedInvoice)
	return &ReceivedInvoiceIterator{iter: iter}
}

// Get retrieves a received invoice by ID.
func (s *ReceivedInvoiceService) Get(id string) (*ReceivedInvoice, error) {
	var ri ReceivedInvoice
	err := s.client.get(fmt.Sprintf("/received-invoices/%s", id), nil, &ri)
	if err != nil {
		return nil, err
	}
	return &ri, nil
}

// Approve approves a received invoice (sends fr:205 lifecycle event to PA).
func (s *ReceivedInvoiceService) Approve(id string) (*ReceivedInvoiceActionResponse, error) {
	var resp ReceivedInvoiceActionResponse
	err := s.client.post(fmt.Sprintf("/received-invoices/%s/approve", id), nil, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Refuse refuses a received invoice with a mandatory reason (sends fr:206 lifecycle event to PA).
func (s *ReceivedInvoiceService) Refuse(id string, params *RefuseParams) (*ReceivedInvoiceActionResponse, error) {
	var resp ReceivedInvoiceActionResponse
	err := s.client.post(fmt.Sprintf("/received-invoices/%s/refuse", id), params, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Suspend suspends a received invoice (sends fr:207 lifecycle event to PA).
func (s *ReceivedInvoiceService) Suspend(id string) (*ReceivedInvoiceActionResponse, error) {
	var resp ReceivedInvoiceActionResponse
	err := s.client.post(fmt.Sprintf("/received-invoices/%s/suspend", id), nil, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// RecordPayment records a payment on a received invoice (sends fr:212 lifecycle event to PA).
func (s *ReceivedInvoiceService) RecordPayment(id string, params *RecordPaymentParams) (*ReceivedInvoiceActionResponse, error) {
	var resp ReceivedInvoiceActionResponse
	err := s.client.post(fmt.Sprintf("/received-invoices/%s/record-payment", id), params, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ReceivedInvoiceIterator iterates over received invoices.
type ReceivedInvoiceIterator struct {
	iter *Iterator[*ReceivedInvoice]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *ReceivedInvoiceIterator) Next() bool { return it.iter.Next() }

// ReceivedInvoice returns the most recently fetched received invoice.
func (it *ReceivedInvoiceIterator) ReceivedInvoice() *ReceivedInvoice { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *ReceivedInvoiceIterator) Err() error { return it.iter.Err() }

func decodeReceivedInvoice(raw json.RawMessage) (*ReceivedInvoice, error) {
	var ri ReceivedInvoice
	err := json.Unmarshal(raw, &ri)
	return &ri, err
}
