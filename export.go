package facturino

// FECParams are the parameters for generating an FEC export.
type FECParams struct {
	PeriodStart      string `json:"period_start"`
	PeriodEnd        string `json:"period_end"`
	SendToAccountant bool   `json:"send_to_accountant,omitempty"`
	AccountantEmail  string `json:"accountant_email,omitempty"`
}

// FECResponse is returned by FEC generation.
type FECResponse struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	CompanyID   string `json:"company_id"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

// RGPDExportResponse is returned by RGPD data export.
type RGPDExportResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CompanyID string `json:"company_id"`
}

// ExportService operates on data exports (FEC, RGPD).
type ExportService struct {
	client *httpClient
}

// GenerateFEC triggers an async FEC export (Pro/Cabinet plan required).
func (s *ExportService) GenerateFEC(params *FECParams) (*FECResponse, error) {
	var resp FECResponse
	err := s.client.post("/exports/fec", params, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ExportRGPD triggers an async full RGPD data export.
func (s *ExportService) ExportRGPD() (*RGPDExportResponse, error) {
	var resp RGPDExportResponse
	err := s.client.post("/exports/full", nil, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
