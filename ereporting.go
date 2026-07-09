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
	Status    string `json:"status"`
	Type      string `json:"type"`

	Period string `json:"period"`

	TotalHT  int `json:"totalHT"` // integer centimes
	TotalTVA int `json:"totalTVA"`
	TotalTTC int `json:"totalTTC"`

	Lines []*EReportingLine `json:"lines"`

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
}

// EReportingParams are the parameters for creating a declaration. Amounts in centimes.
type EReportingParams struct {
	Type   string                  `json:"type"`
	Period string                  `json:"period"`
	Lines  []*EReportingLineParams `json:"lines"`

	IdempotencyKey string `json:"-"`
}

// EReportingLineParams defines a single e-reporting line.
type EReportingLineParams struct {
	Category  string `json:"category"`
	Amount    int    `json:"amount"`
	VATRate   int    `json:"vatRate"`
	VATAmount int    `json:"vatAmount"`
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
