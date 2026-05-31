package facturino

import (
	"fmt"
	"net/url"
)

// Cabinet groups managed client companies under a single billing entity
// — used by chartered-accountant cabinets (plans cabinet_50,
// cabinet_200, cabinet_500).
type Cabinet struct {
	Object   string                 `json:"object"`
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Plan     string                 `json:"plan"`
	Branding *CabinetBranding       `json:"branding,omitempty"`
	Created  string                 `json:"created,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CabinetBranding contains the white-label branding for a cabinet
// (logo, colors, custom domain).
type CabinetBranding struct {
	LogoURL      string `json:"logoUrl,omitempty"`
	PrimaryColor string `json:"primaryColor,omitempty"`
	CustomDomain string `json:"customDomain,omitempty"`
}

// CabinetCreateParams is the body for POST /v1/cabinets.
type CabinetCreateParams struct {
	Name  string `json:"name"`
	Siret string `json:"siret"`
	Plan  string `json:"plan"`
}

// CabinetBrandingUpdate is the body for PATCH /v1/cabinets/:id/branding.
type CabinetBrandingUpdate struct {
	LogoURL      string `json:"logoUrl,omitempty"`
	PrimaryColor string `json:"primaryColor,omitempty"`
	CustomDomain string `json:"customDomain,omitempty"`
}

// CabinetDashboard is the cross-company KPI dashboard.
type CabinetDashboard struct {
	Object           string `json:"object"`
	TotalRevenue     string `json:"totalRevenue"`
	OverdueAmount    string `json:"overdueAmount,omitempty"`
	EReportingStatus string `json:"eReportingStatus,omitempty"`
	CompanyCount     int    `json:"companyCount"`
}

// CabinetDashboardParams scopes the dashboard query to a date range.
type CabinetDashboardParams struct {
	PeriodStart string
	PeriodEnd   string
}

// CabinetActivity is one entry of the reverse-chronological activity
// feed across managed companies.
type CabinetActivity struct {
	Object      string `json:"object"`
	ID          string `json:"id"`
	CompanyID   string `json:"companyId"`
	CompanyName string `json:"companyName,omitempty"`
	Type        string `json:"type"`
	Message     string `json:"message"`
	Created     string `json:"created"`
}

// CabinetActivityList is the paginated response for
// GET /v1/cabinets/:id/activity.
type CabinetActivityList struct {
	Object  string            `json:"object"`
	Data    []CabinetActivity `json:"data"`
	HasMore bool              `json:"has_more"`
}

// CabinetList is the paginated response for GET /v1/cabinets.
type CabinetList struct {
	Object  string    `json:"object"`
	Data    []Cabinet `json:"data"`
	HasMore bool      `json:"has_more"`
}

// CabinetBillingSplit is the per-company breakdown of the subscription
// invoice — useful for cabinets that rebill subscription costs to
// their clients.
type CabinetBillingSplit struct {
	Object string                 `json:"object"`
	Total  int                    `json:"total"`
	Split  []CabinetBillingShare  `json:"split"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// CabinetBillingShare is one row of the per-company billing breakdown.
type CabinetBillingShare struct {
	CompanyID   string `json:"companyId"`
	CompanyName string `json:"companyName,omitempty"`
	Amount      int    `json:"amount"`
	SharePct    string `json:"sharePct,omitempty"`
}

// CabinetCompanySummary describes a managed company attached to a
// cabinet.
type CabinetCompanySummary struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Siret     string `json:"siret,omitempty"`
	AttachedAt string `json:"attachedAt,omitempty"`
}

// CabinetCompanyList is the paginated response for
// GET /v1/cabinets/:id/companies.
type CabinetCompanyList struct {
	Object  string                  `json:"object"`
	Data    []CabinetCompanySummary `json:"data"`
	HasMore bool                    `json:"has_more"`
}

// CabinetAddCompanyParams is the body for attaching an existing
// company to a cabinet (by SIRET or by ID).
type CabinetAddCompanyParams struct {
	CompanyID   string `json:"companyId,omitempty"`
	Siret       string `json:"siret,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
}

// CabinetMemberInviteParams is the body for inviting a member onto a
// cabinet (admin / accountant / viewer role).
type CabinetMemberInviteParams struct {
	Email       string `json:"email"`
	Role        string `json:"role"`
	DisplayName string `json:"displayName,omitempty"`
}

// CabinetMemberInviteResponse is the acknowledgement returned on
// invitation.
type CabinetMemberInviteResponse struct {
	Object string `json:"object"`
	ID     string `json:"id"`
	Status string `json:"status"`
}

// CabinetService manages cabinets and their cross-company surfaces.
// Calls require a cabinet_* plan; integrations on pro or below receive
// plan_limit_error.
type CabinetService struct {
	client *httpClient
}

// List returns the paginated list of cabinets owned by the
// authenticated account.
func (s *CabinetService) List(params *ListParams) (*CabinetList, error) {
	var out CabinetList
	var q url.Values
	if params != nil {
		q = params.toValues()
	}
	if err := s.client.get("/cabinets", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a cabinet by id.
func (s *CabinetService) Retrieve(id string) (*Cabinet, error) {
	var out Cabinet
	if err := s.client.get(fmt.Sprintf("/cabinets/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create creates a new cabinet — only allowed under a cabinet_* plan.
func (s *CabinetService) Create(params *CabinetCreateParams) (*Cabinet, error) {
	var out Cabinet
	if err := s.client.post("/cabinets", params, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateBranding sets the white-label branding (logo, colors, custom
// domain).
func (s *CabinetService) UpdateBranding(id string, params *CabinetBrandingUpdate) (*Cabinet, error) {
	var out Cabinet
	if err := s.client.patch(fmt.Sprintf("/cabinets/%s/branding", id), params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Dashboard returns the cross-company KPI dashboard (revenue, overdue,
// e-reporting status). Pass a CabinetDashboardParams to restrict the
// figures to a date range.
func (s *CabinetService) Dashboard(id string, params *CabinetDashboardParams) (*CabinetDashboard, error) {
	var out CabinetDashboard
	q := url.Values{}
	if params != nil {
		if params.PeriodStart != "" {
			q.Set("period_start", params.PeriodStart)
		}
		if params.PeriodEnd != "" {
			q.Set("period_end", params.PeriodEnd)
		}
	}
	if err := s.client.get(fmt.Sprintf("/cabinets/%s/dashboard", id), q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Activity returns the reverse-chronological activity feed across every
// managed company.
func (s *CabinetService) Activity(id string, params *ListParams) (*CabinetActivityList, error) {
	var out CabinetActivityList
	var q url.Values
	if params != nil {
		q = params.toValues()
	}
	if err := s.client.get(fmt.Sprintf("/cabinets/%s/activity", id), q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BillingSplit returns the per-company breakdown of the active
// subscription invoice.
func (s *CabinetService) BillingSplit(id string) (*CabinetBillingSplit, error) {
	var out CabinetBillingSplit
	if err := s.client.get(fmt.Sprintf("/cabinets/%s/billing-split", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListCompanies returns the paginated list of companies managed under
// a cabinet.
func (s *CabinetService) ListCompanies(id string, params *ListParams) (*CabinetCompanyList, error) {
	var out CabinetCompanyList
	var q url.Values
	if params != nil {
		q = params.toValues()
	}
	if err := s.client.get(fmt.Sprintf("/cabinets/%s/companies", id), q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddCompany attaches an existing company (by SIRET or by ID) under the
// cabinet.
func (s *CabinetService) AddCompany(id string, params *CabinetAddCompanyParams) (*CabinetCompanySummary, error) {
	var out CabinetCompanySummary
	if err := s.client.post(fmt.Sprintf("/cabinets/%s/companies", id), params, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// InviteMember invites a team member onto the cabinet (admin /
// accountant / viewer role).
func (s *CabinetService) InviteMember(id string, params *CabinetMemberInviteParams) (*CabinetMemberInviteResponse, error) {
	var out CabinetMemberInviteResponse
	if err := s.client.post(fmt.Sprintf("/cabinets/%s/members", id), params, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}
