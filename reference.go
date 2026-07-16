package facturino

import (
	"net/url"
	"strconv"
)

// LegalForm is one entry of the INSEE legal-form lookup table
// (4-digit code + sigle + label).
type LegalForm struct {
	Object string `json:"object"`
	Code   string `json:"code"`
	Sigle  string `json:"sigle,omitempty"`
	Label  string `json:"label"`
}

// LegalFormList is the paginated response for
// GET /v1/reference/legal-forms.
type LegalFormList struct {
	Object  string      `json:"object"`
	Data    []LegalForm `json:"data"`
	HasMore bool        `json:"has_more"`
}

// NafCode is one entry of the NAF activity-code lookup table (Rev. 2,
// 2008).
type NafCode struct {
	Object string `json:"object"`
	Code   string `json:"code"`
	Label  string `json:"label"`
}

// NafCodeList is the paginated response for
// GET /v1/reference/naf-codes.
type NafCodeList struct {
	Object  string    `json:"object"`
	Data    []NafCode `json:"data"`
	HasMore bool      `json:"has_more"`
}

// PaProvider is one supported Plateforme Agréée (PA) in the public catalogue.
// Facturino is BYOPA: the customer brings their own PA credentials, so
// integrations use this list to render a provider picker and the credential
// fields each PA requires.
type PaProvider struct {
	Slug             string `json:"slug"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	LogoURL          string `json:"logoUrl,omitempty"`
	WebsiteURL       string `json:"websiteUrl,omitempty"`
	DocumentationURL string `json:"documentationUrl,omitempty"`
	SignupURL        string `json:"signupUrl,omitempty"`
	AuthType         string `json:"authType"`
	CredentialLabel1 string `json:"credentialLabel1,omitempty"`
	CredentialLabel2 string `json:"credentialLabel2,omitempty"`
	RequiresBaseURL  bool   `json:"requiresBaseUrl"`
	PricingSummary   string `json:"pricingSummary,omitempty"`
}

// PaProviderList is the response for GET /v1/pa-providers.
type PaProviderList struct {
	Object string       `json:"object"`
	Data   []PaProvider `json:"data"`
}

// LegalFormInput sets a company/customer legal form on create or update.
// Provide either the 4-digit INSEE Code or the Sigle (e.g. "SAS", "SASU");
// the API resolves the canonical sub-object. Do not send a Label — the input
// is strictly validated and unknown keys are rejected.
type LegalFormInput struct {
	Code  string `json:"code,omitempty"`
	Sigle string `json:"sigle,omitempty"`
}

// NafInput sets a company/customer NAF (APE) activity code on create or update.
// Provide the Rev. 2 Code (e.g. "62.01Z" or "6201Z"); the API resolves the
// canonical sub-object. Do not send a Label — unknown keys are rejected.
type NafInput struct {
	Code string `json:"code,omitempty"`
}

// ReferenceListParams carries the standard search / limit filter for
// the lookup endpoints.
type ReferenceListParams struct {
	Search string
	Limit  int
}

// ReferenceService exposes the INSEE legal-form codes and NAF activity
// codes that integrations need to power their own company / customer
// forms. Both lists are stable and cacheable for the lifetime of the
// integration's process.
type ReferenceService struct {
	client *httpClient
}

// ListLegalForms returns INSEE legal forms — pass Search to
// autocomplete a form field.
func (s *ReferenceService) ListLegalForms(params *ReferenceListParams) (*LegalFormList, error) {
	var out LegalFormList
	q := encodeReferenceParams(params)
	if err := s.client.get("/reference/legal-forms", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListNafCodes returns NAF activity codes — pass Search to
// autocomplete on label fragments ("conseil", "transport"…).
func (s *ReferenceService) ListNafCodes(params *ReferenceListParams) (*NafCodeList, error) {
	var out NafCodeList
	q := encodeReferenceParams(params)
	if err := s.client.get("/reference/naf-codes", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPaProviders returns the supported Plateformes Agréées (PA). Public
// catalogue — no filter; fetch once and cache.
func (s *ReferenceService) ListPaProviders() (*PaProviderList, error) {
	var out PaProviderList
	if err := s.client.get("/pa-providers", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func encodeReferenceParams(params *ReferenceListParams) url.Values {
	if params == nil {
		return nil
	}
	q := url.Values{}
	if params.Search != "" {
		q.Set("search", params.Search)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
