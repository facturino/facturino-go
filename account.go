package facturino

import "fmt"

// Account is the payload returned by the GET /v1/account endpoint.
//
// It exposes the user, company, plan and key scopes that the current API
// key is bound to. Integrations should call AccountService.Retrieve()
// on startup so they can display the connected company and the
// environment (fac_test_ vs fac_live_).
type Account struct {
	Object       string          `json:"object"`
	UserID       string          `json:"userId"`
	CompanyID    string          `json:"companyId"`
	Plan         string          `json:"plan"`
	Livemode     bool            `json:"livemode"`
	APIKeyPrefix string          `json:"apiKeyPrefix"`
	Permissions  []string        `json:"permissions"`
	Company      *AccountCompany `json:"company,omitempty"`
	User         *AccountUser    `json:"user,omitempty"`
}

// AccountCompany is a snapshot of the active company. May be nil when
// the authenticated user has not created one yet.
type AccountCompany struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Siret     string `json:"siret,omitempty"`
	VATRegime string `json:"vatRegime,omitempty"`
}

// AccountUser carries the verification status of the authenticated user.
type AccountUser struct {
	EmailVerified bool `json:"emailVerified"`
}

// AccountService provides access to the GET /v1/account endpoint.
//
// Use this on integration startup to display the connected account and
// the environment so users can confirm they are about to operate
// against the right company.
type AccountService struct {
	client *httpClient
}

// Retrieve returns the account context (user, company, plan, livemode,
// key scopes) associated with the API key carried by this client.
func (s *AccountService) Retrieve() (*Account, error) {
	var account Account
	if err := s.client.get("/account", nil, &account); err != nil {
		return nil, err
	}
	return &account, nil
}

// AccountExportResponse is the ack returned by POST /v1/account/export.
type AccountExportResponse struct {
	Object   string `json:"object"`
	ExportID string `json:"exportId"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
}

// AccountExportDownload carries the short-lived signed URL for a
// previously-prepared RGPD export.
type AccountExportDownload struct {
	URL       string `json:"url"`
	ExpiresAt string `json:"expiresAt,omitempty"`
}

// RequestExport starts a full data export (RGPD article 20). Returns
// immediately with an ExportID; the file is prepared asynchronously
// and the user receives an in-app notification when ready. Pass the
// returned id to DownloadExport to fetch a signed URL.
func (s *AccountService) RequestExport() (*AccountExportResponse, error) {
	var out AccountExportResponse
	if err := s.client.post("/account/export", nil, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// DownloadExport returns a short-lived (5 min) signed URL for a
// previously-prepared RGPD export.
func (s *AccountService) DownloadExport(exportID string) (*AccountExportDownload, error) {
	var out AccountExportDownload
	if err := s.client.get(fmt.Sprintf("/account/exports/%s/download", exportID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
