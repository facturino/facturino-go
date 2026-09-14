package facturino

// ObligationFollowUp is read-only technical follow-up, separate from accounting
// and the fiscal verdict. Nil values never invent historical ownership.
type ObligationFollowUp struct {
	State         string  `json:"state"`
	ReasonCode    *string `json:"reasonCode"`
	ReasonSource  *string `json:"reasonSource"`
	Owner         *string `json:"owner"`
	Action        *string `json:"action"`
	NextAttemptAt *string `json:"nextAttemptAt"`
}

type ObligationWebhookData struct {
	ID         string                   `json:"id"`
	Object     string                   `json:"object"`
	Status     *string                  `json:"status,omitempty"`
	Obligation *ObligationFollowUp      `json:"obligation"`
	Reason     *string                  `json:"reason"`
	InvoiceID  *string                  `json:"invoiceId,omitempty"`
	FR212      *PaymentCollectionStatus `json:"fr212,omitempty"`
}

// EReportingBlock exposes preparation incidents without reclassifying frozen facts.
type EReportingBlock struct {
	Volet       *string             `json:"volet,omitempty"`
	Code        string              `json:"code"`
	Remedy      *string             `json:"remedy,omitempty"`
	Obligation  *ObligationFollowUp `json:"obligation,omitempty"`
	PeriodStart *string             `json:"periodStart,omitempty"`
	PeriodEnd   *string             `json:"periodEnd,omitempty"`
	Updated     *string             `json:"updated,omitempty"`
}
