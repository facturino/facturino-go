package facturino

import (
	"encoding/json"
	"fmt"
)

// Quote is a quotation (devis).
type Quote struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Livemode bool   `json:"livemode"`

	Customer *CustomerRef `json:"customer"`
	Status   string       `json:"status"`
	Number   string       `json:"number"`
	Currency string       `json:"currency"`

	VersionNumber    int    `json:"versionNumber"`
	ParentQuoteID    string `json:"parentQuoteId,omitempty"`
	IsLatestVersion  bool   `json:"isLatestVersion"`

	Items  []*LineItem `json:"items"`
	Totals *Totals     `json:"totals"`

	Dates *QuoteDates `json:"dates"`
	Notes string      `json:"notes,omitempty"`

	ViewedAt  string `json:"viewedAt,omitempty"`
	ViewCount int    `json:"viewCount"`
	AcceptedAt string `json:"acceptedAt,omitempty"`

	Signature *QuoteSignature `json:"signature,omitempty"`

	ConvertedInvoiceID string `json:"convertedInvoiceId,omitempty"`

	Files *QuoteFiles `json:"files,omitempty"`

	Created string `json:"created"`
	Updated string `json:"updated"`
}

// QuoteDates holds quote dates.
type QuoteDates struct {
	Issued     string `json:"issued"`
	Sent       string `json:"sent,omitempty"`
	ValidUntil string `json:"validUntil"`
}

// QuoteSignature holds e-signature data.
type QuoteSignature struct {
	SignedAt     string `json:"signedAt,omitempty"`
	SignerEmail  string `json:"signerEmail,omitempty"`
	SignerIP     string `json:"signerIp,omitempty"`
	DocumentHash string `json:"documentHash,omitempty"`
}

// QuoteFiles holds generated document paths.
type QuoteFiles struct {
	PDFPath string `json:"pdfPath,omitempty"`
}

// QuoteParams are the parameters for creating a quote.
type QuoteParams struct {
	Customer    string        `json:"customerId"`
	Items       []*ItemParams `json:"lines"`
	Notes       string        `json:"notes,omitempty"`
	ValidUntil  string        `json:"validUntil,omitempty"`
	Currency    string        `json:"currency,omitempty"`

	IdempotencyKey string `json:"-"`
}

// QuoteUpdateParams are the parameters for updating a draft quote.
type QuoteUpdateParams struct {
	Items      []*ItemParams `json:"lines,omitempty"`
	Notes      string        `json:"notes,omitempty"`
	ValidUntil string        `json:"validUntil,omitempty"`
}

// QuoteService operates on quotes.
type QuoteService struct {
	client *httpClient
}

// Create creates a new draft quote.
func (s *QuoteService) Create(params *QuoteParams) (*Quote, error) {
	var q Quote
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post("/quotes", params, &q, opts)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// Get retrieves a quote by ID.
func (s *QuoteService) Get(id string) (*Quote, error) {
	var q Quote
	err := s.client.get(fmt.Sprintf("/quotes/%s", id), nil, &q)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// Update updates a draft quote.
func (s *QuoteService) Update(id string, params *QuoteUpdateParams) (*Quote, error) {
	var q Quote
	err := s.client.patch(fmt.Sprintf("/quotes/%s", id), params, &q)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// Delete soft-deletes a draft quote.
func (s *QuoteService) Delete(id string) error {
	return s.client.del(fmt.Sprintf("/quotes/%s", id))
}

// List returns a paginated iterator over quotes.
func (s *QuoteService) List(params *ListParams) *QuoteIterator {
	iter := newIterator[*Quote](s.client, "/quotes", params, decodeQuote)
	return &QuoteIterator{iter: iter}
}

// Send emails the quote to the customer.
func (s *QuoteService) Send(id string) (*Quote, error) {
	var q Quote
	err := s.client.post(fmt.Sprintf("/quotes/%s/send", id), nil, &q, nil)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// Accept marks the quote as accepted by the buyer.
func (s *QuoteService) Accept(id string) (*Quote, error) {
	var q Quote
	err := s.client.post(fmt.Sprintf("/quotes/%s/accept", id), nil, &q, nil)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// Refuse marks the quote as refused by the buyer.
func (s *QuoteService) Refuse(id string) (*Quote, error) {
	var q Quote
	err := s.client.post(fmt.Sprintf("/quotes/%s/refuse", id), nil, &q, nil)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// Convert creates a draft invoice from an accepted quote.
func (s *QuoteService) Convert(id string) (*Invoice, error) {
	var inv Invoice
	err := s.client.post(fmt.Sprintf("/quotes/%s/convert", id), nil, &inv, nil)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// GetPDF retrieves a cached PDF URL or triggers async generation.
func (s *QuoteService) GetPDF(id string) (*DocumentResponse, error) {
	var resp DocumentResponse
	err := s.client.get(fmt.Sprintf("/quotes/%s/pdf", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSignatureProof retrieves the signature proof for an accepted quote.
func (s *QuoteService) GetSignatureProof(id string) (*DocumentResponse, error) {
	var resp DocumentResponse
	err := s.client.get(fmt.Sprintf("/quotes/%s/signature-proof", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// QuoteIterator iterates over quotes.
type QuoteIterator struct {
	iter *Iterator[*Quote]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *QuoteIterator) Next() bool { return it.iter.Next() }

// Quote returns the most recently fetched quote.
func (it *QuoteIterator) Quote() *Quote { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *QuoteIterator) Err() error { return it.iter.Err() }

func decodeQuote(raw json.RawMessage) (*Quote, error) {
	var q Quote
	err := json.Unmarshal(raw, &q)
	return &q, err
}
