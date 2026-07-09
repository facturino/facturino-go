package facturino

// ValidateService dry-runs an invoice payload against the same EN16931 /
// CIUS-FR rules as Invoices.Create, without persisting anything. Use it to
// surface conformity warnings before creating an invoice. To validate a
// customer's SIRET/VAT, use Customers.Lookup (SIRENE/VIES).
type ValidateService struct {
	client *httpClient
}

// ValidateResponse is the result of a dry-run invoice validation.
type ValidateResponse struct {
	// Valid is true when no conformity warnings were raised.
	Valid bool `json:"valid"`
	// Warnings lists human-readable EN16931 / CIUS-FR conformity warnings.
	Warnings []string `json:"warnings"`
	// SchemaVersion is the version of the validation ruleset applied.
	SchemaVersion string `json:"schemaVersion"`
}

// Run validates an invoice payload and returns {Valid, Warnings}. It accepts
// the same shape as Invoices.Create; nothing is created.
func (s *ValidateService) Run(params *InvoiceParams) (*ValidateResponse, error) {
	var out ValidateResponse
	if err := s.client.post("/validate", params, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}
