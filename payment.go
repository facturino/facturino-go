package facturino

import (
	"encoding/json"
	"fmt"
)

// Payment is a payment recorded against an invoice.
type Payment struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Amount     string `json:"amount"`
	Method     string `json:"method"`
	Reference  string `json:"reference"`
	PaidAt     string `json:"paidAt"`
	RecordedBy string `json:"recorded_by"`
	Created    string `json:"created"`
}

// PaymentParams are the parameters for recording a payment. Amount in centimes.
type PaymentParams struct {
	Amount    int    `json:"amount"`
	Method    string `json:"method"`
	Reference string `json:"reference,omitempty"`
	PaidAt    string `json:"paidAt"`

	IdempotencyKey string `json:"-"`
}

// PaymentService operates on invoice payments.
type PaymentService struct {
	client *httpClient
}

// Create records a payment on an invoice.
func (s *PaymentService) Create(invoiceID string, params *PaymentParams) (*Payment, error) {
	var pay Payment
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post(
		fmt.Sprintf("/invoices/%s/payments", invoiceID),
		params, &pay, opts,
	)
	if err != nil {
		return nil, err
	}
	return &pay, nil
}

// List returns a paginated iterator over payments for the given invoice.
func (s *PaymentService) List(invoiceID string, params *ListParams) *PaymentIterator {
	path := fmt.Sprintf("/invoices/%s/payments", invoiceID)
	iter := newIterator[*Payment](s.client, path, params, decodePayment)
	return &PaymentIterator{iter: iter}
}

// PaymentIterator iterates over payments.
type PaymentIterator struct {
	iter *Iterator[*Payment]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *PaymentIterator) Next() bool { return it.iter.Next() }

// Payment returns the most recently fetched payment.
func (it *PaymentIterator) Payment() *Payment { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *PaymentIterator) Err() error { return it.iter.Err() }

func decodePayment(raw json.RawMessage) (*Payment, error) {
	var p Payment
	err := json.Unmarshal(raw, &p)
	return &p, err
}
