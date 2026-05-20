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

// BillingSubscriptionUpdateParams is the body for
// PATCH /v1/billing/subscription. Set CancelAtPeriodEnd to schedule a
// cancellation at the end of the current cycle.
type BillingSubscriptionUpdateParams struct {
	Plan              string `json:"plan,omitempty"`
	Cycle             string `json:"cycle,omitempty"`
	CancelAtPeriodEnd *bool  `json:"cancelAtPeriodEnd,omitempty"`
}

// BillingCheckoutParams is the body for POST /v1/billing/checkout.
type BillingCheckoutParams struct {
	Plan       string `json:"plan"`
	Cycle      string `json:"cycle,omitempty"`
	SuccessURL string `json:"successUrl"`
	CancelURL  string `json:"cancelUrl"`
}

// BillingCheckoutResponse carries the Stripe Checkout URL.
type BillingCheckoutResponse struct {
	URL       string `json:"url"`
	SessionID string `json:"sessionId,omitempty"`
}

// BillingPortalParams is the body for POST /v1/billing/portal.
type BillingPortalParams struct {
	ReturnURL string `json:"returnUrl"`
}

// BillingPortalResponse carries the Stripe Customer-Portal URL.
type BillingPortalResponse struct {
	URL string `json:"url"`
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

// BillingService manages the Stripe-backed subscription, lists and
// downloads platform invoices, and opens Customer-Portal / Checkout
// sessions.
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

// UpdateSubscription changes plan / cycle or schedules a cancellation
// at the end of the period.
func (s *BillingService) UpdateSubscription(params *BillingSubscriptionUpdateParams) (*BillingSubscription, error) {
	var out BillingSubscription
	if err := s.client.patch("/billing/subscription", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Checkout creates a Stripe Checkout session for the first paid
// subscription.
func (s *BillingService) Checkout(params *BillingCheckoutParams) (*BillingCheckoutResponse, error) {
	var out BillingCheckoutResponse
	if err := s.client.post("/billing/checkout", params, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Portal creates a Stripe Customer-Portal session for self-service
// changes.
func (s *BillingService) Portal(params *BillingPortalParams) (*BillingPortalResponse, error) {
	var out BillingPortalResponse
	if err := s.client.post("/billing/portal", params, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Pause pauses the active subscription (Pro+ plans).
func (s *BillingService) Pause() (*BillingSubscription, error) {
	var out BillingSubscription
	if err := s.client.post("/billing/pause", nil, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Resume resumes a paused subscription.
func (s *BillingService) Resume() (*BillingSubscription, error) {
	var out BillingSubscription
	if err := s.client.post("/billing/resume", nil, &out, nil); err != nil {
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
