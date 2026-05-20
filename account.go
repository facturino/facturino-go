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

// AccountDeletionResponse is the payload returned when a deletion is
// scheduled.
type AccountDeletionResponse struct {
	Object                string `json:"object"`
	DeletionScheduledAt   string `json:"deletionScheduledAt,omitempty"`
	Message               string `json:"message,omitempty"`
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

// AccountNotificationPreferencesUpdate is the body for PATCH
// /v1/account/notifications. Pointers let callers omit a field by
// leaving it nil — the API merges only the fields explicitly set.
type AccountNotificationPreferencesUpdate struct {
	InvoicePaid     *bool `json:"invoicePaid,omitempty"`
	InvoiceOverdue  *bool `json:"invoiceOverdue,omitempty"`
	QuoteAccepted   *bool `json:"quoteAccepted,omitempty"`
	PAReceived      *bool `json:"paReceived,omitempty"`
	ProductNews     *bool `json:"productNews,omitempty"`
}

// ScheduleDeletion schedules the authenticated account for deletion in
// 30 days (RGPD article 17). Returns 409 if a deletion is already
// scheduled, and 400 while the account still has an active paid
// subscription. Reversible during the grace period via CancelDeletion.
func (s *AccountService) ScheduleDeletion() (*AccountDeletionResponse, error) {
	var out AccountDeletionResponse
	if err := s.client.post("/account/schedule-deletion", nil, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelDeletion cancels a pending account deletion (within the
// 30-day grace window).
func (s *AccountService) CancelDeletion() (*AccountDeletionResponse, error) {
	var out AccountDeletionResponse
	if err := s.client.post("/account/cancel-deletion", nil, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
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

// UpdateNotifications updates the per-channel email-notification
// preferences. Per-event channel preferences live under
// NotificationService.UpdatePreferences instead.
func (s *AccountService) UpdateNotifications(params *AccountNotificationPreferencesUpdate) (*AccountNotificationPreferencesUpdate, error) {
	var out AccountNotificationPreferencesUpdate
	if err := s.client.patch("/account/notifications", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
