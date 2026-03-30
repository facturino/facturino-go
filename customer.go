package facturino

import (
	"encoding/json"
	"fmt"
)

// Customer is a buyer.
type Customer struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Livemode bool   `json:"livemode"`

	Name      string `json:"name"`
	SIRET     string `json:"siret,omitempty"`
	SIREN     string `json:"siren,omitempty"`
	VATNumber string `json:"vatNumber,omitempty"`
	LegalForm string `json:"legalForm,omitempty"`
	NAFCode   string `json:"nafCode,omitempty"`

	Address         *Address `json:"address"`
	DeliveryAddress *Address `json:"deliveryAddress,omitempty"`

	Type     string     `json:"type"`
	Contacts []*Contact `json:"contacts,omitempty"`

	PaymentTerms int    `json:"paymentTerms,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Notes        string `json:"notes,omitempty"`

	Balance  string `json:"balance"`
	Currency string `json:"currency"`

	SIRETVerified bool `json:"siretVerified"`
	VATVerified   bool `json:"vatVerified"`

	PAIdentifier         string `json:"paIdentifier,omitempty"`
	PreferredFormat      string `json:"preferredFormat,omitempty"`
	ReceivingPAID        string `json:"receivingPaId,omitempty"`
	RecipientServiceCode string `json:"recipientServiceCode,omitempty"`

	Active  bool   `json:"active"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// CustomerParams are the parameters for creating a customer.
type CustomerParams struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Email     string `json:"email,omitempty"`
	SIRET     string `json:"siret,omitempty"`
	VATNumber string `json:"vatNumber,omitempty"`
	LegalForm string `json:"legalForm,omitempty"`
	NAFCode   string `json:"nafCode,omitempty"`

	Address         *Address `json:"address,omitempty"`
	DeliveryAddress *Address `json:"deliveryAddress,omitempty"`

	Contacts     []*Contact `json:"contacts,omitempty"`
	PaymentTerms int        `json:"paymentTerms,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	Notes        string     `json:"notes,omitempty"`
	Currency     string     `json:"currency,omitempty"`

	PAIdentifier         string `json:"paIdentifier,omitempty"`
	PreferredFormat      string `json:"preferredFormat,omitempty"`
	ReceivingPAID        string `json:"receivingPaId,omitempty"`
	RecipientServiceCode string `json:"recipientServiceCode,omitempty"`

	IdempotencyKey string `json:"-"`
}

// CustomerUpdateParams are the parameters for updating a customer.
type CustomerUpdateParams struct {
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	SIRET     string `json:"siret,omitempty"`
	VATNumber string `json:"vatNumber,omitempty"`
	LegalForm string `json:"legalForm,omitempty"`
	NAFCode   string `json:"nafCode,omitempty"`

	Address         *Address `json:"address,omitempty"`
	DeliveryAddress *Address `json:"deliveryAddress,omitempty"`

	Contacts     []*Contact `json:"contacts,omitempty"`
	PaymentTerms int        `json:"paymentTerms,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	Notes        string     `json:"notes,omitempty"`

	PAIdentifier         string `json:"paIdentifier,omitempty"`
	PreferredFormat      string `json:"preferredFormat,omitempty"`
	ReceivingPAID        string `json:"receivingPaId,omitempty"`
	RecipientServiceCode string `json:"recipientServiceCode,omitempty"`

	Active *bool `json:"active,omitempty"`
}

// CustomerLookupParams find a customer by SIRET or VAT number.
type CustomerLookupParams struct {
	SIRET     string `json:"siret,omitempty"`
	VATNumber string `json:"vatNumber,omitempty"`
}

// CustomerService operates on customers.
type CustomerService struct {
	client *httpClient
}

// Create creates a new customer.
func (s *CustomerService) Create(params *CustomerParams) (*Customer, error) {
	var cus Customer
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post("/customers", params, &cus, opts)
	if err != nil {
		return nil, err
	}
	return &cus, nil
}

// Get retrieves a customer by ID.
func (s *CustomerService) Get(id string) (*Customer, error) {
	var cus Customer
	err := s.client.get(fmt.Sprintf("/customers/%s", id), nil, &cus)
	if err != nil {
		return nil, err
	}
	return &cus, nil
}

// Update updates a customer.
func (s *CustomerService) Update(id string, params *CustomerUpdateParams) (*Customer, error) {
	var cus Customer
	err := s.client.patch(fmt.Sprintf("/customers/%s", id), params, &cus)
	if err != nil {
		return nil, err
	}
	return &cus, nil
}

// Delete soft-deletes a customer.
func (s *CustomerService) Delete(id string) error {
	return s.client.del(fmt.Sprintf("/customers/%s", id))
}

// List returns a paginated iterator over customers.
func (s *CustomerService) List(params *ListParams) *CustomerIterator {
	iter := newIterator[*Customer](s.client, "/customers", params, decodeCustomer)
	return &CustomerIterator{iter: iter}
}

// Lookup finds a customer by SIRET or VAT.
func (s *CustomerService) Lookup(params *CustomerLookupParams) (*Customer, error) {
	var cus Customer
	err := s.client.post("/customers/lookup", params, &cus, nil)
	if err != nil {
		return nil, err
	}
	return &cus, nil
}

// ImportCSV imports customers from base64-encoded CSV content.
func (s *CustomerService) ImportCSV(content string) (*ImportResult, error) {
	var result ImportResult
	body := map[string]string{"csv": content}
	err := s.client.post("/customers/import", body, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExportCSV triggers a CSV export.
func (s *CustomerService) ExportCSV() (*ExportResult, error) {
	var result ExportResult
	err := s.client.get("/customers/export", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ImportResult holds CSV import results.
type ImportResult struct {
	Object   string `json:"object"`
	Imported int    `json:"imported"`
	Skipped  int    `json:"skipped"`
	Errors   int    `json:"errors"`
}

// ExportResult holds CSV export results.
type ExportResult struct {
	Object    string `json:"object"`
	URL       string `json:"url,omitempty"`
	ExpiresIn int    `json:"expires_in,omitempty"`
	JobID     string `json:"id,omitempty"`
	Status    string `json:"status,omitempty"`
}

// CustomerIterator iterates over customers.
type CustomerIterator struct {
	iter *Iterator[*Customer]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *CustomerIterator) Next() bool { return it.iter.Next() }

// Customer returns the most recently fetched customer.
func (it *CustomerIterator) Customer() *Customer { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *CustomerIterator) Err() error { return it.iter.Err() }

func decodeCustomer(raw json.RawMessage) (*Customer, error) {
	var c Customer
	err := json.Unmarshal(raw, &c)
	return &c, err
}
