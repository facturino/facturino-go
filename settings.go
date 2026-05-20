package facturino

import "fmt"

// AccountingSettings is the per-company accounting configuration
// (FEC accounts, journal codes, VAT regime) returned by
// GET /v1/companies/:id/settings/accounting.
type AccountingSettings struct {
	Object       string            `json:"object"`
	VATRegime    string            `json:"vatRegime,omitempty"`
	JournalCode  string            `json:"journalCode,omitempty"`
	Accounts     map[string]string `json:"accounts,omitempty"`
}

// AccountingSettingsUpdate is the body for PATCH
// /v1/companies/:id/settings/accounting.
type AccountingSettingsUpdate struct {
	VATRegime   string            `json:"vatRegime,omitempty"`
	JournalCode string            `json:"journalCode,omitempty"`
	Accounts    map[string]string `json:"accounts,omitempty"`
}

// ReminderSettings is the per-company automatic dunning schedule
// (J+7 / J+15 / J+30 by default).
type ReminderSettings struct {
	Object    string `json:"object"`
	Enabled   bool   `json:"enabled"`
	Intervals []int  `json:"intervals"`
}

// ReminderSettingsUpdate is the body for PATCH
// /v1/companies/:id/settings/reminders.
type ReminderSettingsUpdate struct {
	Enabled   *bool `json:"enabled,omitempty"`
	Intervals []int `json:"intervals,omitempty"`
}

// SettingsService exposes the per-company configuration that does not
// belong to a single resource: accounting accounts and the automatic
// reminder schedule. Every method takes the target companyID so a
// single API key can administer settings on each of the companies it
// is scoped to.
type SettingsService struct {
	client *httpClient
}

// RetrieveAccounting returns the accounting configuration for the
// given company.
func (s *SettingsService) RetrieveAccounting(companyID string) (*AccountingSettings, error) {
	var out AccountingSettings
	if err := s.client.get(fmt.Sprintf("/companies/%s/settings/accounting", companyID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAccounting updates the accounting configuration.
func (s *SettingsService) UpdateAccounting(companyID string, params *AccountingSettingsUpdate) (*AccountingSettings, error) {
	var out AccountingSettings
	if err := s.client.patch(fmt.Sprintf("/companies/%s/settings/accounting", companyID), params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RetrieveReminders returns the automatic reminder schedule.
func (s *SettingsService) RetrieveReminders(companyID string) (*ReminderSettings, error) {
	var out ReminderSettings
	if err := s.client.get(fmt.Sprintf("/companies/%s/settings/reminders", companyID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateReminders updates the reminder schedule.
func (s *SettingsService) UpdateReminders(companyID string, params *ReminderSettingsUpdate) (*ReminderSettings, error) {
	var out ReminderSettings
	if err := s.client.patch(fmt.Sprintf("/companies/%s/settings/reminders", companyID), params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
