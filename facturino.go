// Package facturino is a Go client for the Facturino e-invoicing API.
//
// Amounts are integers in centimes (10000 = 100.00 EUR).
// VAT rates are integers in centipercent (2000 = 20.00%).
package facturino

import (
	"context"
	"net/http"
	"time"
)

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
	Exports           *ExportService
	Archives          *ArchiveService
	EReporting        *EReportingService
	Reporting         *ReportingService
	Jobs              *JobService
	Sandbox           *SandboxService
	Billing           *BillingService
	Reference         *ReferenceService
	TaxDecisions      *TaxDecisionService
	// EuThresholdLedgers maintains the annual ledger of the EUR 10,000 threshold.
	EuThresholdLedgers *EuThresholdLedgerService
	Usage              *UsageService
	Validate           *ValidateService
	Health             *HealthService

	client *httpClient
}

// ClientOption configures a Client.
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL         string
	httpClient      *http.Client
	maxRetries      int
	autoIdempotency bool
	retryBudget     time.Duration
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

// WithAutoIdempotency controls automatic stable POST keys (default true).
// POST calls without a key are never automatically retried.
func WithAutoIdempotency(enabled bool) ClientOption {
	return func(c *clientConfig) { c.autoIdempotency = enabled }
}

// WithRetryBudget sets the maximum cumulative retry waiting time (default 60s).
// A Retry-After exceeding the remaining budget returns the HTTP error immediately.
func WithRetryBudget(budget time.Duration) ClientOption {
	return func(c *clientConfig) {
		if budget < 0 {
			budget = 0
		}
		c.retryBudget = budget
	}
}

// New creates a Facturino API client. Use fac_test_ keys for sandbox, fac_live_ for production.
func New(apiKey string, opts ...ClientOption) *Client {
	cfg := &clientConfig{
		maxRetries:      defaultMaxRetries,
		autoIdempotency: true,
		retryBudget:     60 * time.Second,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	hc := newHTTPClient(apiKey, cfg.baseURL, cfg.httpClient, cfg.maxRetries)
	hc.autoIdempotency = cfg.autoIdempotency
	hc.retryBudget = cfg.retryBudget
	return buildClient(hc)
}

// WithContext returns a shallow copy of the client whose requests use ctx as
// their default context, so a deadline or cancellation applies across a group
// of calls:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	inv, err := client.WithContext(ctx).Invoices.Get(id)
//
// The original client is unchanged; both share the same HTTP transport.
func (c *Client) WithContext(ctx context.Context) *Client {
	return buildClient(c.client.withContext(ctx))
}

// buildClient wires every service to the given httpClient.
func buildClient(hc *httpClient) *Client {
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
	c.Exports = &ExportService{client: hc}
	c.Archives = &ArchiveService{client: hc}
	c.EReporting = &EReportingService{client: hc}
	c.Reporting = &ReportingService{client: hc}
	c.Jobs = &JobService{client: hc}
	c.Sandbox = &SandboxService{client: hc}
	c.Account = &AccountService{client: hc}
	c.Billing = &BillingService{client: hc}
	c.Reference = &ReferenceService{client: hc}
	c.TaxDecisions = &TaxDecisionService{client: hc}
	c.EuThresholdLedgers = &EuThresholdLedgerService{client: hc}
	c.Usage = &UsageService{client: hc}
	c.Validate = &ValidateService{client: hc}
	c.Health = &HealthService{client: hc}

	return c
}
