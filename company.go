package facturino

import (
	"fmt"
	"net/url"
)

// Company is a seller/tenant company.
type Company struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`

	Name         string `json:"name"`
	SIRET        string `json:"siret"`
	SIREN        string `json:"siren"`
	VATNumber    string `json:"vatNumber"`
	LegalForm    string `json:"legalForm"`
	APECode      string `json:"apeCode,omitempty"`
	RCS          string `json:"rcs,omitempty"`
	CapitalSocial string `json:"capitalSocial,omitempty"`
	TVAIntracom  string `json:"tvaIntracom"`

	Address *Address `json:"address"`
	Email   string   `json:"email,omitempty"`
	Phone   string   `json:"phone,omitempty"`
	Website string   `json:"website,omitempty"`

	LogoPath string `json:"logoPath,omitempty"`

	VATRegime string `json:"vatRegime"`

	BillingEmail         string       `json:"billingEmail,omitempty"`
	BankDetails          *BankDetails `json:"bankDetails"`
	DefaultPaymentTerms  int          `json:"defaultPaymentTerms"`
	DefaultPaymentMethod string       `json:"defaultPaymentMethod"`

	InvoiceSettings    *InvoiceSettings    `json:"invoiceSettings"`
	QuoteSettings      *QuoteSettings      `json:"quoteSettings"`
	CreditNoteSettings *CreditNoteSettings `json:"creditNoteSettings"`

	ReminderConfig   *ReminderConfig   `json:"reminderConfig,omitempty"`
	AccountingConfig *AccountingConfig `json:"accountingConfig,omitempty"`

	CGVUrl       string `json:"cgvUrl,omitempty"`
	CGVPdfPath   string `json:"cgvPdfPath,omitempty"`
	CGVUpdatedAt string `json:"cgvUpdatedAt,omitempty"`

	StripeConnectAccountID string `json:"stripeConnectAccountId,omitempty"`
	Einvoicing *EinvoicingConfig `json:"einvoicing,omitempty"`

	MonthlyInvoiceCount    int `json:"monthlyInvoiceCount,omitempty"`
	MonthlyQuoteCount      int `json:"monthlyQuoteCount,omitempty"`
	MonthlyAPIRequestCount int `json:"monthlyApiRequestCount,omitempty"`

	Active              bool `json:"active"`
	OnboardingCompleted bool `json:"onboardingCompleted"`

	Created string `json:"created"`
	Updated string `json:"updated"`
}

// BankDetails holds bank account info.
type BankDetails struct {
	IBAN     string `json:"iban,omitempty"`
	BIC      string `json:"bic,omitempty"`
	BankName string `json:"bankName,omitempty"`
}

// InvoiceSettings holds numbering and defaults for invoices.
type InvoiceSettings struct {
	Prefix         string `json:"prefix"`
	NextNumber     int    `json:"nextNumber"`
	DefaultVATRate string `json:"defaultVatRate"`
	LegalMentions  string `json:"legalMentions,omitempty"`
	YearlyReset    bool   `json:"yearlyReset"`
}

// QuoteSettings holds numbering defaults for quotes.
type QuoteSettings struct {
	Prefix              string `json:"prefix"`
	NextNumber          int    `json:"nextNumber"`
	DefaultValidityDays int    `json:"defaultValidityDays"`
}

// CreditNoteSettings holds numbering defaults for credit notes.
//
// NumberingMode controls how credit notes are numbered: "separate"
// (default) gives credit notes their own number series, while "unified"
// makes them share the invoice number series.
type CreditNoteSettings struct {
	Prefix        string `json:"prefix"`
	NextNumber    int    `json:"nextNumber"`
	NumberingMode string `json:"numberingMode,omitempty"`
}

// ReminderConfig configures automatic payment reminders (Pro plan).
type ReminderConfig struct {
	Enabled       bool   `json:"enabled"`
	Schedule      []int  `json:"schedule"`
	CustomSubject string `json:"customSubject,omitempty"`
	CustomBody    string `json:"customBody,omitempty"`
}

// AccountingConfig holds FEC export settings.
type AccountingConfig struct {
	JournalCode         string            `json:"journalCode"`
	ClientAccountPrefix string            `json:"clientAccountPrefix"`
	ServiceAccount      string            `json:"serviceAccount"`
	GoodsAccount        string            `json:"goodsAccount"`
	VATAccounts         map[string]string `json:"vatAccounts"`
	BankAccount         string            `json:"bankAccount"`
	BankJournalCode     string            `json:"bankJournalCode"`
	CustomMappings      map[string]string `json:"customMappings"`
}

// CompanyCreateParams is the body for POST /v1/companies.
//
// Subject to the per-plan company quota (free / essential: 1, pro: 3,
// cabinet_*: 50+); exceeding the quota returns a 402 plan_limit_error.
type CompanyCreateParams struct {
	Name          string       `json:"name"`
	SIRET         string       `json:"siret"`
	Address       *Address     `json:"address"`
	VATNumber     string       `json:"vatNumber,omitempty"`
	LegalForm     interface{}  `json:"legalForm,omitempty"`
	NAF           interface{}  `json:"naf,omitempty"`
	TVAIntracom   string       `json:"tvaIntracom,omitempty"`
	RCS           string       `json:"rcs,omitempty"`
	CapitalSocial string       `json:"capitalSocial,omitempty"`
	VATRegime     string       `json:"vatRegime,omitempty"`
	Email         string       `json:"email,omitempty"`
	Phone         string       `json:"phone,omitempty"`
	Website       string       `json:"website,omitempty"`
	BankDetails   *BankDetails `json:"bankDetails,omitempty"`
}

// CompanyInvoicingSettingsUpdate is the body for PATCH
// /v1/companies/:id/invoicing-settings.
type CompanyInvoicingSettingsUpdate struct {
	VATRegime       string           `json:"vatRegime,omitempty"`
	InvoiceSettings *InvoiceSettings `json:"invoiceSettings,omitempty"`
	QuoteSettings   *QuoteSettings   `json:"quoteSettings,omitempty"`
	CreditNoteSettings *CreditNoteSettings `json:"creditNoteSettings,omitempty"`
}

// CompanyMilestoneResponse is the acknowledgement returned by
// POST /v1/companies/:id/milestones.
type CompanyMilestoneResponse struct {
	Object    string `json:"object"`
	Milestone string `json:"milestone"`
	ReachedAt string `json:"reachedAt"`
}

// CompanyUpdateParams are the parameters for updating a company.
type CompanyUpdateParams struct {
	Name         string   `json:"name,omitempty"`
	SIRET        string   `json:"siret,omitempty"`
	VATNumber    string   `json:"vatNumber,omitempty"`
	LegalForm    string   `json:"legalForm,omitempty"`
	APECode      string   `json:"apeCode,omitempty"`
	RCS          string   `json:"rcs,omitempty"`
	CapitalSocial string  `json:"capitalSocial,omitempty"`

	Address *Address `json:"address,omitempty"`
	Email   string   `json:"email,omitempty"`
	Phone   string   `json:"phone,omitempty"`
	Website string   `json:"website,omitempty"`

	VATRegime string `json:"vatRegime,omitempty"`

	BillingEmail         string       `json:"billingEmail,omitempty"`
	BankDetails          *BankDetails `json:"bankDetails,omitempty"`
	DefaultPaymentTerms  int          `json:"defaultPaymentTerms,omitempty"`
	DefaultPaymentMethod string       `json:"defaultPaymentMethod,omitempty"`

	InvoiceSettings    *InvoiceSettings    `json:"invoiceSettings,omitempty"`
	QuoteSettings      *QuoteSettings      `json:"quoteSettings,omitempty"`
	CreditNoteSettings *CreditNoteSettings `json:"creditNoteSettings,omitempty"`

	ReminderConfig   *ReminderConfig   `json:"reminderConfig,omitempty"`
	AccountingConfig *AccountingConfig `json:"accountingConfig,omitempty"`
}

// CGVResponse is returned by CGV (Terms & Conditions) endpoints.
type CGVResponse struct {
	Object    string `json:"object"`
	URL       string `json:"url,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
	Deleted   bool   `json:"deleted,omitempty"`
}

// StripeConnectParams are the parameters for connecting a Stripe account.
type StripeConnectParams struct {
	ReturnURL  string `json:"return_url,omitempty"`
	RefreshURL string `json:"refresh_url,omitempty"`
}

// StripeConnectResponse is returned by Stripe connect operations.
type StripeConnectResponse struct {
	Object    string `json:"object"`
	URL       string `json:"url,omitempty"`
	Connected bool   `json:"connected,omitempty"`
	AccountID string `json:"account_id,omitempty"`
}

// StripeDashboardResponse is returned by the Stripe dashboard endpoint.
type StripeDashboardResponse struct {
	Object string `json:"object"`
	URL    string `json:"url"`
}

// CompanyService operates on companies.
type CompanyService struct {
	client *httpClient
}

// List returns a paginated list of companies.
func (s *CompanyService) List(params *ListParams) (*ListResponse, error) {
	var resp ListResponse
	var vals url.Values
	if params != nil {
		vals = params.toValues()
	}
	err := s.client.get("/companies", vals, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Create creates a new company under the authenticated user.
//
// Subject to the per-plan company quota (free / essential: 1, pro: 3,
// cabinet_*: 50+); exceeding the quota returns a 402 plan_limit_error.
func (s *CompanyService) Create(params *CompanyCreateParams) (*Company, error) {
	var c Company
	if err := s.client.post("/companies", params, &c, nil); err != nil {
		return nil, err
	}
	return &c, nil
}

// Get retrieves a company by ID.
func (s *CompanyService) Get(id string) (*Company, error) {
	var c Company
	err := s.client.get(fmt.Sprintf("/companies/%s", id), nil, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Update updates a company.
func (s *CompanyService) Update(id string, params *CompanyUpdateParams) (*Company, error) {
	var c Company
	err := s.client.patch(fmt.Sprintf("/companies/%s", id), params, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UploadCGV uploads a base64-encoded CGV PDF.
func (s *CompanyService) UploadCGV(id string, base64Content string) (*CGVResponse, error) {
	var resp CGVResponse
	body := map[string]string{"content": base64Content}
	err := s.client.post(fmt.Sprintf("/companies/%s/cgv", id), body, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCGV returns the CGV download URL.
func (s *CompanyService) GetCGV(id string) (*CGVResponse, error) {
	var resp CGVResponse
	err := s.client.get(fmt.Sprintf("/companies/%s/cgv", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateInvoicingSettings updates the invoicing settings (numbering
// format, default payment terms, default VAT rate, footer mentions…)
// and the VAT regime for the company.
func (s *CompanyService) UpdateInvoicingSettings(id string, params *CompanyInvoicingSettingsUpdate) (*Company, error) {
	var c Company
	if err := s.client.patch(fmt.Sprintf("/companies/%s/invoicing-settings", id), params, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// AddMilestone marks an onboarding milestone as reached (for example
// "first_invoice_sent", "pa_connected", "bank_added"). Used by the
// dashboard to compute the onboarding progress and surface remaining
// steps.
func (s *CompanyService) AddMilestone(id, milestone string) (*CompanyMilestoneResponse, error) {
	var resp CompanyMilestoneResponse
	body := map[string]string{"milestone": milestone}
	if err := s.client.post(fmt.Sprintf("/companies/%s/milestones", id), body, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteCGV deletes the CGV PDF.
func (s *CompanyService) DeleteCGV(id string) error {
	return s.client.del(fmt.Sprintf("/companies/%s/cgv", id))
}

// EinvoicingConfig holds the BYOPA PA connection configuration.
type EinvoicingConfig struct {
	Enabled      bool   `json:"enabled"`
	Provider     string `json:"provider,omitempty"`
	ConnectedAt  string `json:"connectedAt,omitempty"`
	HealthStatus string `json:"healthStatus,omitempty"`
}

// PAConnectionParams are the parameters for connecting a PA.
type PAConnectionParams struct {
	Provider     string `json:"provider"`
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	APIKey       string `json:"apiKey,omitempty"`
	CustomBaseURL string `json:"customBaseUrl,omitempty"`
}

// PAConnectionResult is returned by ConnectPA.
type PAConnectionResult struct {
	Provider    string `json:"provider"`
	Status      string `json:"status"`
	ConnectedAt string `json:"connectedAt"`
}

// PATestResult is returned by TestPAConnection. Healthy reflects whether the PA
// is reachable with valid credentials. ErrorCode is set only when Healthy is
// false: "pa_credentials_invalid" (fix credentials), "pa_unreachable"
// (PA/network outage), "pa_not_supported" (the PA exposes no directory lookup —
// a capability gap, not a misconfiguration), or "pa_error".
type PATestResult struct {
	Object    string `json:"object"`
	Healthy   bool   `json:"healthy"`
	LatencyMs int    `json:"latencyMs"`
	Details   string `json:"details"`
	Provider  string `json:"provider"`
	TestedAt  string `json:"testedAt"`
	ErrorCode string `json:"errorCode,omitempty"`
}

// ConnectPA connects a PA to the company. The client provides their own PA account credentials.
func (s *CompanyService) ConnectPA(id string, params *PAConnectionParams) (*PAConnectionResult, error) {
	var r PAConnectionResult
	err := s.client.post(fmt.Sprintf("/companies/%s/pa-connection", id), params, &r, nil)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// DisconnectPA disconnects the PA from a company.
func (s *CompanyService) DisconnectPA(id string) error {
	return s.client.del(fmt.Sprintf("/companies/%s/pa-connection", id))
}

// TestPAConnection tests the PA connection (health check + credential validation).
func (s *CompanyService) TestPAConnection(id string) (*PATestResult, error) {
	var r PATestResult
	err := s.client.post(fmt.Sprintf("/companies/%s/pa-connection/test", id), nil, &r, nil)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ConnectStripe initiates Stripe Connect onboarding for online payments (Pro+ plan).
func (s *CompanyService) ConnectStripe(params *StripeConnectParams) (*StripeConnectResponse, error) {
	var resp StripeConnectResponse
	err := s.client.post("/companies/stripe-connect", params, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStripeDashboard returns the Stripe dashboard URL for the connected account.
func (s *CompanyService) GetStripeDashboard() (*StripeDashboardResponse, error) {
	var resp StripeDashboardResponse
	err := s.client.get("/companies/stripe-dashboard", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DisconnectStripe disconnects the Stripe account.
func (s *CompanyService) DisconnectStripe() error {
	return s.client.del("/companies/stripe-connect")
}
