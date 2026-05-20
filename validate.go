package facturino

// ValidateParams is the body for POST /v1/validate.
//
// The shape depends on Kind:
//   - "siret" / "vat" / "iban" / "bic": fill Value.
//   - "invoice": fill Invoice with a full invoice payload — the server
//     runs Schematron against EN16931 + CIUS-FR.
type ValidateParams struct {
	Kind    string                 `json:"kind"`
	Value   string                 `json:"value,omitempty"`
	Invoice map[string]interface{} `json:"invoice,omitempty"`
}

// ValidateResponse carries the validation outcome plus any error /
// warning the server raised.
type ValidateResponse struct {
	Object   string             `json:"object"`
	Valid    bool               `json:"valid"`
	Kind     string             `json:"kind,omitempty"`
	Errors   []ValidationIssue  `json:"errors,omitempty"`
	Warnings []ValidationIssue  `json:"warnings,omitempty"`
}

// ValidationIssue describes a single error or warning emitted by the
// validator.
type ValidationIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

// ValidateService exposes synchronous structural validation of
// business identifiers (SIRET, VAT, IBAN, BIC) and document payloads
// (invoice line items against EN16931 / CIUS-FR Schematron rules).
//
// Validate calls never mutate any resource — they are safe to call
// client-side to surface errors before hitting the write endpoints.
type ValidateService struct {
	client *httpClient
}

// Run executes a single validation request.
func (s *ValidateService) Run(params *ValidateParams) (*ValidateResponse, error) {
	var out ValidateResponse
	if err := s.client.post("/validate", params, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}
