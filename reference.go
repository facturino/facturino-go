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
