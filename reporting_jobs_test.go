package facturino

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
)

// Reports, invoice export jobs and the audit trail PDF, against the shapes the
// API sends (camelCase reports; invoice_export jobs; 501 on the audit trail PDF).

const vatReportBody = `{"object":"vat_report","period":{"start":"2026-01-01","end":"2026-03-31"},
"vatBreakdown":[{"rate":"20.00","taxableAmount":10000,"vatAmount":2000}],
"totalHT":10000,"totalVAT":2000,"totalTTC":12000,"invoiceCount":3}`

const revenueReportBody = `{"object":"revenue_report","period":{"start":"2026-01-01","end":"2026-03-31"},
"revenue":{"invoiced":12000,"creditNotes":-2400,"net":9600},
"payments":{"received":6000,"outstanding":3600},"invoiceCount":3,"creditNoteCount":1,
"breakdown":[{"period":"2026-Q1","revenue":{"invoiced":12000,"creditNotes":-2400,"net":9600},
"payments":{"received":6000,"outstanding":3600},"invoiceCount":3,"creditNoteCount":1}]}`

func TestReportsDecodeCamelCaseFields(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/reporting/vat":
			fmt.Fprint(w, vatReportBody)
		case "/v1/reporting/revenue":
			fmt.Fprint(w, revenueReportBody)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	vat, err := client.Reporting.VAT(&VATReportParams{PeriodStart: "2026-01-01", PeriodEnd: "2026-03-31"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vat.VATBreakdown) != 1 || vat.VATBreakdown[0].Rate != "20.00" ||
		vat.VATBreakdown[0].TaxableAmount != 10000 || vat.VATBreakdown[0].VATAmount != 2000 {
		t.Errorf("VATBreakdown = %+v, want one 20.00 entry of 10000/2000", vat.VATBreakdown)
	}
	if vat.TotalHT != 10000 || vat.TotalVAT != 2000 || vat.TotalTTC != 12000 || vat.InvoiceCount != 3 {
		t.Errorf("VAT totals = %d/%d/%d count %d, want 10000/2000/12000 count 3",
			vat.TotalHT, vat.TotalVAT, vat.TotalTTC, vat.InvoiceCount)
	}

	rev, err := client.Reporting.Revenue(&RevenueReportParams{PeriodStart: "2026-01-01", PeriodEnd: "2026-03-31", GroupBy: "quarter"})
	if err != nil {
		t.Fatal(err)
	}
	if rev.Revenue == nil || rev.Revenue.Invoiced != 12000 || rev.Revenue.CreditNotes != -2400 || rev.Revenue.Net != 9600 {
		t.Errorf("Revenue = %+v, want 12000/-2400/9600", rev.Revenue)
	}
	if rev.Payments == nil || rev.Payments.Received != 6000 || rev.Payments.Outstanding != 3600 {
		t.Errorf("Payments = %+v, want 6000/3600", rev.Payments)
	}
	if rev.InvoiceCount != 3 || rev.CreditNoteCount != 1 {
		t.Errorf("counts = %d/%d, want 3/1", rev.InvoiceCount, rev.CreditNoteCount)
	}
	if len(rev.Breakdown) != 1 || rev.Breakdown[0].Revenue.CreditNotes != -2400 ||
		rev.Breakdown[0].InvoiceCount != 3 || rev.Breakdown[0].CreditNoteCount != 1 {
		t.Errorf("Breakdown = %+v, want one 2026-Q1 group with -2400 credit notes", rev.Breakdown)
	}
}

func TestJobDecodesInvoiceExportFields(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"job_export","object":"job","type":"invoice_export","status":"completed_with_errors",
"progress":100,"result":"exports/job_export.zip","error":null,"included":4998,"failed":2,"truncated":true,
"created":"2026-11-13T10:00:00.000Z","updatedAt":"2026-11-13T10:05:00.000Z","expireAt":"2026-11-20T10:05:00.000Z"}`)
	})

	job, err := client.Jobs.Get("job_export")
	if err != nil {
		t.Fatal(err)
	}
	if job.Type != "invoice_export" || job.ExpireAt != "2026-11-20T10:05:00.000Z" {
		t.Errorf("Type = %q, ExpireAt = %q, want invoice_export and the ISO string", job.Type, job.ExpireAt)
	}
	if job.Included == nil || *job.Included != 4998 || job.Failed == nil || *job.Failed != 2 ||
		job.Truncated == nil || !*job.Truncated {
		t.Errorf("Included/Failed/Truncated = %v/%v/%v, want 4998/2/true", job.Included, job.Failed, job.Truncated)
	}
}

func TestGenerateAuditTrailPDFNotImplementedIsAnAPIErrorWithoutRetry(t *testing.T) {
	var calls int32
	client, _ := testServerWithRetries(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprint(w, `{"error":{"type":"api_error","code":"not_implemented","message":"Audit trail PDF is not available yet.","request_id":"req_501"}}`)
	}, 3)

	_, err := client.Invoices.GenerateAuditTrailPDF("inv_1")

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if apiErr.HTTPStatusCode != http.StatusNotImplemented || apiErr.Code != "not_implemented" || apiErr.Type != ErrorTypeAPI {
		t.Errorf("HTTPStatusCode/Code/Type = %d/%q/%q, want 501/not_implemented/api_error",
			apiErr.HTTPStatusCode, apiErr.Code, apiErr.Type)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1 (a 501 is not retried)", got)
	}
}
