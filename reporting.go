package facturino

import "net/url"

// VATBreakdownEntry is a single VAT rate entry in a VAT report.
type VATBreakdownEntry struct {
	Rate          string `json:"rate"`
	TaxableAmount int    `json:"taxable_amount"`
	VATAmount     int    `json:"vat_amount"`
}

// VATReport is a VAT summary for a given period. Amounts in centimes.
type VATReport struct {
	Object       string               `json:"object"`
	Period       *ReportPeriod        `json:"period"`
	VATBreakdown []*VATBreakdownEntry `json:"vat_breakdown"`
	TotalHT      int                  `json:"total_ht"`
	TotalVAT     int                  `json:"total_vat"`
	TotalTTC     int                  `json:"total_ttc"`
	InvoiceCount int                  `json:"invoice_count"`
}

// RevenueReport is a revenue summary for a given period. Amounts in centimes.
type RevenueReport struct {
	Object          string               `json:"object"`
	Period          *ReportPeriod        `json:"period"`
	Revenue         *RevenueBreakdown    `json:"revenue"`
	Payments        *PaymentsBreakdown   `json:"payments"`
	InvoiceCount    int                  `json:"invoice_count"`
	CreditNoteCount int                  `json:"credit_note_count"`
	Breakdown       []*RevenuePeriodGroup `json:"breakdown,omitempty"`
}

// ReportPeriod defines the start and end dates of a report.
type ReportPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// RevenueBreakdown holds invoiced, credit note, and net revenue totals.
type RevenueBreakdown struct {
	Invoiced    int `json:"invoiced"`
	CreditNotes int `json:"credit_notes"`
	Net         int `json:"net"`
}

// PaymentsBreakdown holds received and outstanding payment totals.
type PaymentsBreakdown struct {
	Received    int `json:"received"`
	Outstanding int `json:"outstanding"`
}

// RevenuePeriodGroup is a revenue breakdown for a single sub-period (month or quarter).
type RevenuePeriodGroup struct {
	Period          string             `json:"period"`
	Revenue         *RevenueBreakdown  `json:"revenue"`
	Payments        *PaymentsBreakdown `json:"payments"`
	InvoiceCount    int                `json:"invoice_count"`
	CreditNoteCount int                `json:"credit_note_count"`
}

// VATReportParams are the query parameters for a VAT report.
type VATReportParams struct {
	PeriodStart string
	PeriodEnd   string
}

// RevenueReportParams are the query parameters for a revenue report.
type RevenueReportParams struct {
	PeriodStart string
	PeriodEnd   string
	GroupBy     string // "month" or "quarter" (optional)
}

// ReportingService operates on financial reports (VAT, revenue).
type ReportingService struct {
	client *httpClient
}

// VAT returns a VAT report for the given period. Requires Essential+ plan.
func (s *ReportingService) VAT(params *VATReportParams) (*VATReport, error) {
	v := url.Values{}
	if params != nil {
		if params.PeriodStart != "" {
			v.Set("period_start", params.PeriodStart)
		}
		if params.PeriodEnd != "" {
			v.Set("period_end", params.PeriodEnd)
		}
	}
	var resp VATReport
	err := s.client.get("/reporting/vat", v, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Revenue returns a revenue report for the given period. Requires Essential+ plan.
func (s *ReportingService) Revenue(params *RevenueReportParams) (*RevenueReport, error) {
	v := url.Values{}
	if params != nil {
		if params.PeriodStart != "" {
			v.Set("period_start", params.PeriodStart)
		}
		if params.PeriodEnd != "" {
			v.Set("period_end", params.PeriodEnd)
		}
		if params.GroupBy != "" {
			v.Set("group_by", params.GroupBy)
		}
	}
	var resp RevenueReport
	err := s.client.get("/reporting/revenue", v, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
