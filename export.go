package facturino

import "fmt"

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

// ExportStatusResponse is returned by export status endpoints.
type ExportStatusResponse struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	CompanyID   string `json:"company_id"`
	DownloadURL string `json:"download_url,omitempty"`
}

// RGPDExportResponse is returned by RGPD data export.
type RGPDExportResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CompanyID string `json:"company_id"`
}

// ExportInvoicesResponse is returned by the bulk invoice export endpoint.
type ExportInvoicesResponse struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// ExportService operates on data exports (FEC, RGPD, invoices).
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

// GetFECStatus returns the status of a FEC export job. Returns download_url when completed.
func (s *ExportService) GetFECStatus(jobID string) (*ExportStatusResponse, error) {
	var resp ExportStatusResponse
	err := s.client.get(fmt.Sprintf("/exports/fec/%s", jobID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetExportStatus returns the status of any export job. Returns download_url when completed.
func (s *ExportService) GetExportStatus(jobID string) (*ExportStatusResponse, error) {
	var resp ExportStatusResponse
	err := s.client.get(fmt.Sprintf("/exports/%s", jobID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ExportInvoices triggers a bulk export of all finalized invoices as ZIP (Factur-X PDF + CII XML).
func (s *ExportService) ExportInvoices() (*ExportInvoicesResponse, error) {
	var resp ExportInvoicesResponse
	err := s.client.post("/exports/invoices", nil, &resp, nil)
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
