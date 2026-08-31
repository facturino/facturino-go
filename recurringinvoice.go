package facturino

import (
	"encoding/json"
	"fmt"
)

// RecurringInvoice is a recurring invoice schedule.
type RecurringInvoice struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Livemode bool   `json:"livemode"`

	CustomerID string `json:"customerId"`
	Currency   string `json:"currency"`

	Frequency          string `json:"frequency"`
	CustomInterval     int    `json:"customInterval,omitempty"`
	CustomUnit         string `json:"customUnit,omitempty"`
	StartDate          string `json:"startDate"`
	NextGenerationDate string `json:"nextGenerationDate"`
	EndDate            string `json:"endDate,omitempty"`

	TemplateInvoice *RecurringTemplate `json:"templateInvoice"`

	AutoFinalize bool `json:"autoFinalize"`
	AutoSend     bool `json:"autoSend"`

	LastGeneratedAt        string `json:"lastGeneratedAt,omitempty"`
	LastGeneratedInvoiceID string `json:"lastGeneratedInvoiceId,omitempty"`
	GenerationCount        int    `json:"generationCount"`

	Active  bool   `json:"active"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// RecurringTemplate holds the template for generated invoices.
type RecurringTemplate struct {
	Lines            []*LineItem `json:"lines"`
	Notes            string      `json:"notes,omitempty"`
	PaymentMethod    string      `json:"paymentMethod,omitempty"`
	PaymentTermsDays int         `json:"paymentTermsDays,omitempty"`
	Currency         string      `json:"currency"`
}

// RecurringInvoiceParams are the parameters for creating a recurring invoice.
type RecurringInvoiceParams struct {
	CustomerID string `json:"customerId"`

	Frequency      string `json:"frequency"`
	CustomInterval int    `json:"customInterval,omitempty"`
	CustomUnit     string `json:"customUnit,omitempty"`
	StartDate      string `json:"startDate"`
	EndDate        string `json:"endDate,omitempty"`
	// NextGenerationDate is when the first invoice is generated (required).
	NextGenerationDate string `json:"nextGenerationDate"`

	// TaxInputs is required: the fiscal inputs are re-decided at EVERY
	// occurrence, on its own generation date. A recurrence never stores a
	// decision — one taken today would apply this quarter's rules to next
	// year's invoice.
	TaxInputs *RecurringTaxInputsParams `json:"taxInputs,omitempty"`

	TemplateInvoice *RecurringTemplateParams `json:"templateInvoice"`

	AutoFinalize bool `json:"autoFinalize,omitempty"`
	AutoSend     bool `json:"autoSend,omitempty"`

	IdempotencyKey string `json:"-"`
}

// RecurringTaxInputsParams are the inputs a per-occurrence decision needs.
//
// With TaxSource "facturino", no rate, category code, VATEX code or legal
// mention is accepted: the engine decides at every occurrence. With
// "integration", each line carries the supplied VAT (VatRate, VatCode, and
// where the rate is zero a VatexCode and PlaceOfSupply), re-validated for
// coherence at every occurrence.
type RecurringTaxInputsParams struct {
	// TaxSource is "facturino" or "integration".
	TaxSource string `json:"taxSource"`
	// PriceMode is "tax_exclusive" or "tax_inclusive".
	PriceMode string                    `json:"priceMode"`
	Lines     []*RecurringTaxLineParams `json:"lines"`
}

// RecurringTaxLineParams is one commercial line of a decision-backed
// recurrence. It carries its own presentation because no decision is stored.
type RecurringTaxLineParams struct {
	TaxDecisionLineParams
	Unit    string `json:"unit"`
	Product string `json:"product,omitempty"`
}

// RecurringTemplateParams are the document-only template parameters for the
// generated invoices (payment terms, notes); never VAT.
type RecurringTemplateParams struct {
	Notes            string `json:"notes,omitempty"`
	PaymentMethod    string `json:"paymentMethod,omitempty"`
	PaymentTermsDays int    `json:"paymentTermsDays,omitempty"`
}

// RecurringInvoiceUpdateParams are the parameters for updating a recurring invoice.
type RecurringInvoiceUpdateParams struct {
	// TaxInputs replaces the fiscal inputs wholesale.
	TaxInputs      *RecurringTaxInputsParams `json:"taxInputs,omitempty"`
	Frequency      string                    `json:"frequency,omitempty"`
	CustomInterval int                       `json:"customInterval,omitempty"`
	CustomUnit     string                    `json:"customUnit,omitempty"`
	EndDate        string                    `json:"endDate,omitempty"`

	TemplateInvoice *RecurringTemplateParams `json:"templateInvoice,omitempty"`

	AutoFinalize *bool `json:"autoFinalize,omitempty"`
	AutoSend     *bool `json:"autoSend,omitempty"`
}

// RecurringInvoiceService operates on recurring invoices.
type RecurringInvoiceService struct {
	client *httpClient
}

// Create creates a new recurring invoice schedule.
func (s *RecurringInvoiceService) Create(params *RecurringInvoiceParams) (*RecurringInvoice, error) {
	var ri RecurringInvoice
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post("/recurring-invoices", params, &ri, opts)
	if err != nil {
		return nil, err
	}
	return &ri, nil
}

// Get retrieves a recurring invoice by ID.
func (s *RecurringInvoiceService) Get(id string) (*RecurringInvoice, error) {
	var ri RecurringInvoice
	err := s.client.get(fmt.Sprintf("/recurring-invoices/%s", id), nil, &ri)
	if err != nil {
		return nil, err
	}
	return &ri, nil
}

// Update updates a recurring invoice.
func (s *RecurringInvoiceService) Update(id string, params *RecurringInvoiceUpdateParams) (*RecurringInvoice, error) {
	var ri RecurringInvoice
	err := s.client.patch(fmt.Sprintf("/recurring-invoices/%s", id), params, &ri)
	if err != nil {
		return nil, err
	}
	return &ri, nil
}

// Delete deletes a recurring invoice schedule.
func (s *RecurringInvoiceService) Delete(id string) error {
	return s.client.del(fmt.Sprintf("/recurring-invoices/%s", id))
}

// List returns a paginated iterator over recurring invoices.
func (s *RecurringInvoiceService) List(params *ListParams) *RecurringInvoiceIterator {
	iter := newIterator[*RecurringInvoice](s.client, "/recurring-invoices", params, decodeRecurringInvoice)
	return &RecurringInvoiceIterator{iter: iter}
}

// Resume resumes a paused recurring invoice schedule.
func (s *RecurringInvoiceService) Resume(id string) (*RecurringInvoice, error) {
	var ri RecurringInvoice
	err := s.client.post(fmt.Sprintf("/recurring-invoices/%s/resume", id), nil, &ri, nil)
	if err != nil {
		return nil, err
	}
	return &ri, nil
}

// Pause pauses an active recurring invoice schedule.
func (s *RecurringInvoiceService) Pause(id string) (*RecurringInvoice, error) {
	var ri RecurringInvoice
	err := s.client.post(fmt.Sprintf("/recurring-invoices/%s/pause", id), nil, &ri, nil)
	if err != nil {
		return nil, err
	}
	return &ri, nil
}

// RecurringInvoiceIterator iterates over recurring invoices.
type RecurringInvoiceIterator struct {
	iter *Iterator[*RecurringInvoice]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *RecurringInvoiceIterator) Next() bool { return it.iter.Next() }

// RecurringInvoice returns the most recently fetched recurring invoice.
func (it *RecurringInvoiceIterator) RecurringInvoice() *RecurringInvoice { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *RecurringInvoiceIterator) Err() error { return it.iter.Err() }

func decodeRecurringInvoice(raw json.RawMessage) (*RecurringInvoice, error) {
	var ri RecurringInvoice
	err := json.Unmarshal(raw, &ri)
	return &ri, err
}
