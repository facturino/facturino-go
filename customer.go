package facturino

import (
	"encoding/json"
	"fmt"
)

// Customer is a buyer.
type Customer struct {
	RawResponse `json:"-"`
	CompanyId   string                `json:"companyId,omitempty"`
	Warnings    []*BuyerNatureWarning `json:"warnings,omitempty"`
	ID          string                `json:"id"`
	Object      string                `json:"object"`
	Livemode    bool                  `json:"livemode"`

	Name      string     `json:"name"`
	SIRET     *string    `json:"siret,omitempty"`
	SIREN     string     `json:"siren,omitempty"`
	VATNumber *string    `json:"vatNumber,omitempty"`
	LegalForm *LegalForm `json:"legalForm,omitempty"`
	NAF       *NafCode   `json:"naf,omitempty"`

	Address         *Address `json:"address"`
	DeliveryAddress *Address `json:"deliveryAddress,omitempty"`

	Type     string     `json:"type"`
	Contacts []*Contact `json:"contacts,omitempty"`

	PaymentTerms int      `json:"paymentTerms,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Notes        string   `json:"notes,omitempty"`

	Balance  int    `json:"balance"` // integer centimes
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
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Email     string          `json:"email,omitempty"`
	SIRET     string          `json:"siret,omitempty"`
	VATNumber string          `json:"vatNumber,omitempty"`
	LegalForm *LegalFormInput `json:"legalForm,omitempty"`
	NAF       *NafInput       `json:"naf,omitempty"`

	Address         *Address `json:"address,omitempty"`
	DeliveryAddress *Address `json:"deliveryAddress,omitempty"`

	Contacts     []*Contact `json:"contacts,omitempty"`
	PaymentTerms int        `json:"paymentTerms,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	Notes        string     `json:"notes,omitempty"`

	PAIdentifier    string `json:"paIdentifier,omitempty"`
	PreferredFormat string `json:"preferredFormat,omitempty"`
	ReceivingPAID   string `json:"receivingPaId,omitempty"`

	IdempotencyKey string `json:"-"`
}

// CustomerUpdateParams are the parameters for updating a customer.
type CustomerUpdateParams struct {
	Name      string          `json:"name,omitempty"`
	Type      string          `json:"type,omitempty"`
	SIRET     string          `json:"siret,omitempty"`
	VATNumber string          `json:"vatNumber,omitempty"`
	LegalForm *LegalFormInput `json:"legalForm,omitempty"`
	NAF       *NafInput       `json:"naf,omitempty"`

	Address         *Address `json:"address,omitempty"`
	DeliveryAddress *Address `json:"deliveryAddress,omitempty"`

	Contacts     []*Contact `json:"contacts,omitempty"`
	PaymentTerms int        `json:"paymentTerms,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	Notes        string     `json:"notes,omitempty"`

	PAIdentifier    string `json:"paIdentifier,omitempty"`
	PreferredFormat string `json:"preferredFormat,omitempty"`
	ReceivingPAID   string `json:"receivingPaId,omitempty"`

	Active *bool `json:"active,omitempty"`
}

// CustomerLookupParams looks up a company in the INSEE Sirene registry.
// Provide exactly one of SIRET (exact match) or Query (name search).
type CustomerLookupParams struct {
	SIRET string `json:"siret,omitempty"`
	Query string `json:"query,omitempty"`
}

// RegistryDisclosure explains protected lookup fields; it does not certify identity.
type RegistryDisclosure struct {
	Status         *string  `json:"status"`
	WithheldFields []string `json:"withheldFields"`
}

// SireneCompany holds registry details, not a stored customer.
// Feed the disclosed fields into CustomerParams to create one.
type SireneCompany struct {
	Disclosure *RegistryDisclosure `json:"disclosure,omitempty"`
	Name       string              `json:"name"`
	SIRET      string              `json:"siret"`
	SIREN      string              `json:"siren"`
	VATNumber  string              `json:"vatNumber"`
	LegalForm  *LegalForm          `json:"legalForm"`
	NAF        *NafCode            `json:"naf"`
	Address    *Address            `json:"address"`
	Active     bool                `json:"active"`
}

// SireneLookupResult is returned by CustomerService.Lookup. For a SIRET lookup,
// Found reports whether the registry had a match and Data carries the details
// (nil when not found, with Warning explaining why). For a name Query the API
// returns object "sirene_search" with Results and Total instead.
type SireneLookupResult struct {
	Object  string           `json:"object"`
	Found   bool             `json:"found"`
	Data    *SireneCompany   `json:"data,omitempty"`
	Warning string           `json:"warning,omitempty"`
	Results []*SireneCompany `json:"results,omitempty"`
	Total   int              `json:"total,omitempty"`
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

// Lookup resolves a company from the INSEE Sirene registry by SIRET or name.
// The result is registry data (not a stored customer): pass the fields you want
// into Customers.Create to persist a customer.
func (s *CustomerService) Lookup(params *CustomerLookupParams) (*SireneLookupResult, error) {
	var result SireneLookupResult
	err := s.client.post("/customers/lookup", params, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ImportCSV imports customers from raw CSV text (not base64-encoded).
func (s *CustomerService) ImportCSV(content string) (*ImportResult, error) {
	var result ImportResult
	body := map[string]string{"csv": content}
	err := s.client.post("/customers/import", body, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExportCSV returns the full customer list as raw CSV (Content-Type text/csv).
func (s *CustomerService) ExportCSV() (string, error) {
	data, _, err := s.client.doRaw("GET", "/customers/export", nil, nil)
	return string(data), err
}

// ImportResult holds CSV import results.
type ImportResult struct {
	Object   string `json:"object"`
	Imported int    `json:"imported"`
	Skipped  int    `json:"skipped"`
	Errors   int    `json:"errors"`
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

// BuyerNatureWarning is advisory: customer/decision creation still succeeds.
type BuyerNatureWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param"`
}

const BuyerNatureSuspect = "buyer_nature_suspect"
