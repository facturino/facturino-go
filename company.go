package facturino

import "fmt"

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
	PAID                   string `json:"paId,omitempty"`
	PAConnectedAt          string `json:"paConnectedAt,omitempty"`

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
type CreditNoteSettings struct {
	Prefix     string `json:"prefix"`
	NextNumber int    `json:"nextNumber"`
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

// CompanyService operates on companies.
type CompanyService struct {
	client *httpClient
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

// DeleteCGV deletes the CGV PDF.
func (s *CompanyService) DeleteCGV(id string) error {
	return s.client.del(fmt.Sprintf("/companies/%s/cgv", id))
}
