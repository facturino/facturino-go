package facturino

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// Tests for Billing, Reference, Usage and Validate.

func TestBillingRetrieveSubscription(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/v1/billing/subscription" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"subscription","plan":"pro","cycle":"monthly","status":"active"}`)
	})

	sub, err := client.Billing.RetrieveSubscription()
	if err != nil {
		t.Fatal(err)
	}
	if sub.Plan != "pro" {
		t.Errorf("Plan = %q, want pro", sub.Plan)
	}
}

func TestBillingGetInvoicePDF(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/billing/invoices/in_x/pdf" {
			t.Errorf("Path = %q, want /v1/billing/invoices/in_x/pdf", r.URL.Path)
		}
		fmt.Fprint(w, `{"url":"https://signed.example/x.pdf","expiresAt":"2026-05-20T12:05:00Z"}`)
	})

	pdf, err := client.Billing.GetInvoicePDF("in_x")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(pdf.URL, "https://") {
		t.Errorf("URL = %q, want https://...", pdf.URL)
	}
}

func TestReferenceListLegalFormsAndNaf(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/reference/legal-forms":
			if r.URL.Query().Get("search") != "SAS" {
				t.Errorf("search = %q, want SAS", r.URL.Query().Get("search"))
			}
			fmt.Fprint(w, `{"object":"list","data":[{"code":"5710","sigle":"SAS","label":"SAS"}],"has_more":false}`)
		case "/v1/reference/naf-codes":
			fmt.Fprint(w, `{"object":"list","data":[{"code":"62.01Z","label":"Programmation"}],"has_more":false}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	lf, err := client.Reference.ListLegalForms(&ReferenceListParams{Search: "SAS"})
	if err != nil || len(lf.Data) != 1 || lf.Data[0].Code != "5710" {
		t.Fatalf("ListLegalForms failed: %v / %v", err, lf)
	}

	naf, err := client.Reference.ListNafCodes(nil)
	if err != nil || len(naf.Data) != 1 {
		t.Fatalf("ListNafCodes failed: %v / %v", err, naf)
	}
}

func TestUsageRetrieve(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/usage" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"object":"usage","plan":"pro","periodStart":"2026-06-01T00:00:00.000Z","counters":{"invoicesMonth":{"used":12,"limit":1000}}}`)
	})

	u, err := client.Usage.Retrieve()
	if err != nil {
		t.Fatal(err)
	}
	if u.Plan != "pro" || u.Counters["invoicesMonth"].Used != 12 {
		t.Errorf("Usage = %+v", u)
	}
}

func TestValidateRun(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/validate" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		fmt.Fprint(w, `{"valid":true,"warnings":[],"schemaVersion":"2026-02-01"}`)
	})

	res, err := client.Validate.Run(&InvoiceParams{
		Customer: "cus_123",
		Buyer:    &BuyerParams{CompanyName: "Acme", Address: &Address{Line1: "1 rue X", PostalCode: "75001", City: "Paris", Country: "FR"}},
		Items:    []*ItemParams{{Description: "Item", Quantity: "1", UnitPrice: 1000, VATRate: 2000, VATCode: "S"}},
		Dates:    &InvoiceDatesParams{Issued: "2026-01-15", Due: "2026-02-15"},
		Payment:  &PaymentTermsParams{Terms: "Net 30", TermsDays: 30, Method: "transfer", LatePaymentRate: "10.00", CollectionFee: "40.00"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Valid {
		t.Errorf("Valid = false, want true")
	}
}
