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

// PlatformInvoice represents a subscription invoice issued by Facturino
// (INTEK CENTER) to the account.
type PlatformInvoice struct {
	Object   string                `json:"object"`
	ID       string                `json:"id"`
	Status   string                `json:"status"`
	Number   string                `json:"number,omitempty"`
	Totals   map[string]string     `json:"totals"`
	Dates    map[string]string     `json:"dates"`
	Items    []PlatformInvoiceItem `json:"items"`
	Metadata map[string]any        `json:"metadata,omitempty"`
	Created  string                `json:"created"`
}

// PlatformInvoiceItem is a single line of a platform invoice.
type PlatformInvoiceItem struct {
	Description string `json:"description"`
	LineTotal   int    `json:"lineTotal"`
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
	Object    string `json:"object"`
	URL       string `json:"url"`
	ExpiresIn int    `json:"expires_in"`
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
// to this account. Pass a ListParams with Limit / StartingAfter to
// iterate beyond the first page.
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
