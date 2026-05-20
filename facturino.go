// Package facturino is a Go client for the Facturino e-invoicing API.
//
// Amounts are integers in centimes (10000 = 100.00 EUR).
// VAT rates are integers in centipercent (2000 = 20.00%).
package facturino

import "net/http"

// Client is the entry point for the Facturino API.
type Client struct {
	Account           *AccountService
	Invoices          *InvoiceService
	Payments          *PaymentService
	Customers         *CustomerService
	Products          *ProductService
	Quotes            *QuoteService
	CreditNotes       *CreditNoteService
	Events            *EventService
	WebhookEndpoints  *WebhookEndpointService
	RecurringInvoices *RecurringInvoiceService
	ReceivedInvoices  *ReceivedInvoiceService
	Companies         *CompanyService
	Members           *MemberService
	APIKeys           *APIKeyService
	Exports           *ExportService
	Archives          *ArchiveService
	EReporting        *EReportingService
	Reporting         *ReportingService
	Mfa               *MfaService
	Jobs              *JobService
	Sandbox           *SandboxService

	client *httpClient
}

// ClientOption configures a Client.
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) ClientOption {
	return func(c *clientConfig) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *clientConfig) {
		c.httpClient = client
	}
}

// WithMaxRetries sets the maximum number of retries for retryable errors (429, 5xx).
// Default is 3. Use 0 to disable retries (single attempt only).
func WithMaxRetries(n int) ClientOption {
	return func(c *clientConfig) {
		c.maxRetries = n
	}
}

// New creates a Facturino API client. Use fac_test_ keys for sandbox, fac_live_ for production.
func New(apiKey string, opts ...ClientOption) *Client {
	cfg := &clientConfig{
		maxRetries: defaultMaxRetries,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	hc := newHTTPClient(apiKey, cfg.baseURL, cfg.httpClient, cfg.maxRetries)

	c := &Client{client: hc}
	c.Invoices = &InvoiceService{client: hc}
	c.Payments = &PaymentService{client: hc}
	c.Customers = &CustomerService{client: hc}
	c.Products = &ProductService{client: hc}
	c.Quotes = &QuoteService{client: hc}
	c.CreditNotes = &CreditNoteService{client: hc}
	c.Events = &EventService{client: hc}
	c.WebhookEndpoints = &WebhookEndpointService{client: hc}
	c.RecurringInvoices = &RecurringInvoiceService{client: hc}
	c.ReceivedInvoices = &ReceivedInvoiceService{client: hc}
	c.Companies = &CompanyService{client: hc}
	c.Members = &MemberService{client: hc}
	c.APIKeys = &APIKeyService{client: hc}
	c.Exports = &ExportService{client: hc}
	c.Archives = &ArchiveService{client: hc}
	c.EReporting = &EReportingService{client: hc}
	c.Reporting = &ReportingService{client: hc}
	c.Mfa = &MfaService{client: hc}
	c.Jobs = &JobService{client: hc}
	c.Sandbox = &SandboxService{client: hc}
	c.Account = &AccountService{client: hc}

	return c
}
