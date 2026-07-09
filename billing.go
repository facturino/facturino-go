package facturino

import (
	"fmt"
	"net/url"
)

// BillingSubscription is the active Stripe-backed subscription that
// drives the platform billing. Returned by Billing.RetrieveSubscription.
type BillingSubscription struct {
	Object             string `json:"object"`
	ID                 string `json:"id,omitempty"`
	Plan               string `json:"plan"`
	Cycle              string `json:"cycle"`
	Status             string `json:"status"`
	CurrentPeriodStart string `json:"currentPeriodStart,omitempty"`
	CurrentPeriodEnd   string `json:"currentPeriodEnd,omitempty"`
	CancelAtPeriodEnd  bool   `json:"cancelAtPeriodEnd,omitempty"`
	Created            string `json:"created,omitempty"`
}

// PlatformInvoice represents a billing invoice issued by Facturino to
// the account.
type PlatformInvoice struct {
	Object  string `json:"object"`
	ID      string `json:"id"`
	Number  string `json:"number,omitempty"`
	Amount  int    `json:"amount"`
	Status  string `json:"status"`
	PaidAt  string `json:"paidAt,omitempty"`
	Created string `json:"created"`
}

// PlatformInvoiceList is the paginated response for
// GET /v1/billing/invoices.
type PlatformInvoiceList struct {
	Object  string            `json:"object"`
	Data    []PlatformInvoice `json:"data"`
	HasMore bool              `json:"has_more"`
}

// PlatformInvoicePDF is the short-lived signed-URL response for
// GET /v1/billing/invoices/:id/pdf.
type PlatformInvoicePDF struct {
	URL       string `json:"url"`
	ExpiresAt string `json:"expiresAt,omitempty"`
}

// BillingService reads the active subscription and lists / downloads
// platform invoices issued by Facturino to the account. Subscription
// changes (plan, checkout, portal) are handled in the dashboard, not the
// API.
type BillingService struct {
	client *httpClient
}

// RetrieveSubscription returns the active subscription (plan, cycle,
// status, period bounds).
func (s *BillingService) RetrieveSubscription() (*BillingSubscription, error) {
	var out BillingSubscription
	if err := s.client.get("/billing/subscription", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListInvoices returns the paginated list of platform invoices issued
// to this account. Pass a ListParams with Limit / StartingAfter /
// EndingBefore to iterate beyond the first page.
func (s *BillingService) ListInvoices(params *ListParams) (*PlatformInvoiceList, error) {
	var out PlatformInvoiceList
	var q url.Values
	if params != nil {
		q = params.toValues()
	}
	if err := s.client.get("/billing/invoices", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetInvoicePDF returns a short-lived signed URL to download a platform
// invoice PDF.
func (s *BillingService) GetInvoicePDF(invoiceID string) (*PlatformInvoicePDF, error) {
	var out PlatformInvoicePDF
	if err := s.client.get(fmt.Sprintf("/billing/invoices/%s/pdf", invoiceID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
