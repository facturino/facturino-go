package facturino

import "encoding/json"

type InvoiceWebhookData struct {
	ID                 *string                `json:"id,omitempty"`
	Object             *string                `json:"object,omitempty"`
	Status             *string                `json:"status,omitempty"`
	PreviousStatus     *string                `json:"previous_status,omitempty"`
	Livemode           *bool                  `json:"livemode,omitempty"`
	Number             *string                `json:"number,omitempty"`
	DocumentStatus     *string                `json:"documentStatus,omitempty"`
	TransmissionStatus *string                `json:"transmissionStatus,omitempty"`
	TransmissionDetail *string                `json:"transmissionDetail,omitempty"`
	PaymentStatus      *string                `json:"paymentStatus,omitempty"`
	PAErrorCode        *string                `json:"paErrorCode,omitempty"`
	RejectionReason    *string                `json:"rejectionReason,omitempty"`
	RejectionCategory  *string                `json:"rejectionCategory,omitempty"`
	RejectionCode      *string                `json:"rejectionCode,omitempty"`
	RejectionSource    *string                `json:"rejectionSource,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

type IncomingInvoiceWebhookData struct {
	ID          *string `json:"id,omitempty"`
	Object      *string `json:"object,omitempty"`
	PaInvoiceId *string `json:"pa_invoice_id,omitempty"`
	SenderSiret *string `json:"sender_siret,omitempty"`
	SenderName  *string `json:"sender_name,omitempty"`
	Number      *string `json:"number,omitempty"`
	TotalHt     *string `json:"total_ht,omitempty"`
	TotalTva    *string `json:"total_tva,omitempty"`
	TotalTtc    *string `json:"total_ttc,omitempty"`
}

type QuoteWebhookData struct {
	ID             *string                `json:"id,omitempty"`
	Object         *string                `json:"object,omitempty"`
	Status         *string                `json:"status,omitempty"`
	PreviousStatus *string                `json:"previous_status,omitempty"`
	Livemode       *bool                  `json:"livemode,omitempty"`
	Number         *string                `json:"number,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type CreditNoteWebhookData struct {
	PAStatus             *string                `json:"paStatus,omitempty"`
	PAInvoiceID          *string                `json:"paInvoiceId,omitempty"`
	ID                   *string                `json:"id,omitempty"`
	Object               *string                `json:"object,omitempty"`
	Status               *string                `json:"status,omitempty"`
	PreviousStatus       *string                `json:"previous_status,omitempty"`
	Livemode             *bool                  `json:"livemode,omitempty"`
	Number               *string                `json:"number,omitempty"`
	DocumentStatus       *string                `json:"documentStatus,omitempty"`
	TransmissionStatus   *string                `json:"transmissionStatus,omitempty"`
	TransmissionDetail   *string                `json:"transmissionDetail,omitempty"`
	PaymentStatus        *string                `json:"paymentStatus,omitempty"`
	PAErrorCode          *string                `json:"paErrorCode,omitempty"`
	RejectionReason      *string                `json:"rejectionReason,omitempty"`
	RejectionCategory    *string                `json:"rejectionCategory,omitempty"`
	RejectionCode        *string                `json:"rejectionCode,omitempty"`
	RejectionSource      *string                `json:"rejectionSource,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	RelatedInvoiceId     *string                `json:"relatedInvoiceId,omitempty"`
	RelatedInvoiceNumber *string                `json:"relatedInvoiceNumber,omitempty"`
}

type CustomerWebhookData struct {
	ID       *string `json:"id,omitempty"`
	Object   *string `json:"object,omitempty"`
	Livemode *bool   `json:"livemode,omitempty"`
}

type PaymentCreatedWebhookData struct {
	InvoiceId *string `json:"invoiceId,omitempty"`
	PaymentId *string `json:"paymentId,omitempty"`
	Amount    *string `json:"amount,omitempty"`
	Method    *string `json:"method,omitempty"`
}

type PaymentReceivedWebhookData struct {
	ID                 *string                `json:"id,omitempty"`
	Object             *string                `json:"object,omitempty"`
	Status             *string                `json:"status,omitempty"`
	PreviousStatus     *string                `json:"previous_status,omitempty"`
	Livemode           *bool                  `json:"livemode,omitempty"`
	Number             *string                `json:"number,omitempty"`
	DocumentStatus     *string                `json:"documentStatus,omitempty"`
	TransmissionStatus *string                `json:"transmissionStatus,omitempty"`
	TransmissionDetail *string                `json:"transmissionDetail,omitempty"`
	PaymentStatus      *string                `json:"paymentStatus,omitempty"`
	PAErrorCode        *string                `json:"paErrorCode,omitempty"`
	RejectionReason    *string                `json:"rejectionReason,omitempty"`
	RejectionCategory  *string                `json:"rejectionCategory,omitempty"`
	RejectionCode      *string                `json:"rejectionCode,omitempty"`
	RejectionSource    *string                `json:"rejectionSource,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	Amount             *string                `json:"amount,omitempty"`
	TotalPaid          *string                `json:"total_paid,omitempty"`
	TotalDue           *string                `json:"total_due,omitempty"`
	Total              *string                `json:"total,omitempty"`
	AmountDue          *string                `json:"amountDue,omitempty"`
}

type EreportingWebhookData struct {
	ID      *string `json:"id,omitempty"`
	Object  *string `json:"object,omitempty"`
	Status  *string `json:"status,omitempty"`
	Type    *string `json:"type,omitempty"`
	Period  *string `json:"period,omitempty"`
	Attempt *int    `json:"attempt,omitempty"`
}

type RecurringGeneratedWebhookData struct {
	ID                 *string `json:"id,omitempty"`
	Object             *string `json:"object,omitempty"`
	RecurringInvoiceId *string `json:"recurringInvoiceId,omitempty"`
	Livemode           *bool   `json:"livemode,omitempty"`
}

type RecurringFailedWebhookData struct {
	ID     *string `json:"id,omitempty"`
	Object *string `json:"object,omitempty"`
	Error  *string `json:"error,omitempty"`
}

type ExportWebhookData struct {
	ID     *string `json:"id,omitempty"`
	Object *string `json:"object,omitempty"`
	Count  *int    `json:"count,omitempty"`
}

type SubscriptionWebhookData struct {
	Plan                 *string `json:"plan,omitempty"`
	Status               *string `json:"status,omitempty"`
	StripeSubscriptionId *string `json:"stripeSubscriptionId,omitempty"`
	PausedUntil          *string `json:"pausedUntil,omitempty"`
	Reason               *string `json:"reason,omitempty"`
}

func decodeEventPayload(eventType string, data map[string]interface{}) (interface{}, error) {
	var target interface{}
	switch eventType {
	case "invoice.created", "invoice.finalized", "invoice.sending", "invoice.sent", "invoice.deposited", "invoice.transmitted", "invoice.available", "invoice.received", "invoice.approved", "invoice.refused", "invoice.rejected", "invoice.suspended", "invoice.paid", "invoice.partially_paid", "invoice.overdue":
		target = &InvoiceWebhookData{}
	case "invoice.incoming.received":
		target = &IncomingInvoiceWebhookData{}
	case "quote.created", "quote.sent", "quote.viewed", "quote.accepted", "quote.refused", "quote.expired", "quote.converted":
		target = &QuoteWebhookData{}
	case "credit_note.created", "credit_note.finalized", "credit_note.credit_deposited", "credit_note.credit_transmitted", "credit_note.credit_approved", "credit_note.credit_refused", "credit_note.sent":
		target = &CreditNoteWebhookData{}
	case "customer.created", "customer.updated", "customer.deleted":
		target = &CustomerWebhookData{}
	case "payment.created":
		target = &PaymentCreatedWebhookData{}
	case "payment.received":
		target = &PaymentReceivedWebhookData{}
	case "ereporting.submitted":
		target = &EreportingWebhookData{}
	case "recurring_invoice.generated":
		target = &RecurringGeneratedWebhookData{}
	case "recurring_invoice.failed":
		target = &RecurringFailedWebhookData{}
	case "export.ready":
		target = &ExportWebhookData{}
	case "subscription.created", "subscription.cancelled", "subscription.renewed", "subscription.paused":
		target = &SubscriptionWebhookData{}
	default:
		return data, nil
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return nil, err
	}
	return target, nil
}

// TypedData selects the payload model by event type, retaining Data as the raw view.
func (e *Event) TypedData() (interface{}, error)        { return decodeEventPayload(e.Type, e.Data) }
func (e *WebhookEvent) TypedData() (interface{}, error) { return decodeEventPayload(e.Type, e.Data) }
