package facturino

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestInvoiceCreate(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/v1/invoices" {
			t.Errorf("Path = %q, want /v1/invoices", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{
			"id":"inv_abc123",
			"object":"invoice",
			"type":"standard",
			"status":"draft",
			"number":"",
			"currency":"eur",
			"livemode":false,
			"customer":{"ref":"cus_xyz","snapshot":{"name":"ACME Corp","address":{"line1":"1 rue Test","postalCode":"75001","city":"Paris","country":"FR"}}},
			"items":[{"id":"li_1","description":"Consulting","quantity":"1","unit":"heure","unitPrice":"100.00","discountPercent":"0.00","lineAmount":"100.00","vatRate":"20.00","vatCode":"S","vatAmount":"20.00","lineTotal":"120.00","product":""}],
			"totals":{"totalHT":"100.00","discountAmount":"0.00","vatBreakdown":[{"rate":"20.00","code":"S","base":"100.00","amount":"20.00"}],"totalVAT":"20.00","totalTTC":"120.00","amountDue":"120.00","amountPaid":"0.00"},
			"created":"2026-03-15T10:00:00Z",
			"updated":"2026-03-15T10:00:00Z"
		}`)
	})

	inv, err := client.Invoices.Create(&InvoiceParams{
		Customer: "cus_xyz",
		Items: []*ItemParams{
			{Description: "Consulting", Quantity: 1, UnitPrice: 10000, VATRate: 2000},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.ID != "inv_abc123" {
		t.Errorf("ID = %q, want %q", inv.ID, "inv_abc123")
	}
	if inv.Status != "draft" {
		t.Errorf("Status = %q, want %q", inv.Status, "draft")
	}
	if inv.Totals.TotalTTC != "120.00" {
		t.Errorf("TotalTTC = %q, want %q", inv.Totals.TotalTTC, "120.00")
	}
	if inv.Customer.Ref != "cus_xyz" {
		t.Errorf("Customer.Ref = %q, want %q", inv.Customer.Ref, "cus_xyz")
	}
}

func TestInvoiceGet(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/v1/invoices/inv_123" {
			t.Errorf("Path = %q, want /v1/invoices/inv_123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice","status":"finalized","number":"FAC-2026-0001"}`)
	})

	inv, err := client.Invoices.Get("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Number != "FAC-2026-0001" {
		t.Errorf("Number = %q, want %q", inv.Number, "FAC-2026-0001")
	}
}

func TestInvoiceUpdate(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Method = %q, want PATCH", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice","status":"draft","notes":"Updated notes"}`)
	})

	inv, err := client.Invoices.Update("inv_123", &InvoiceUpdateParams{
		Notes: "Updated notes",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Notes != "Updated notes" {
		t.Errorf("Notes = %q, want %q", inv.Notes, "Updated notes")
	}
}

func TestInvoiceFinalize(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/v1/invoices/inv_123/finalize" {
			t.Errorf("Path = %q, want /v1/invoices/inv_123/finalize", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice","status":"finalized","number":"FAC-2026-0001"}`)
	})

	inv, err := client.Invoices.Finalize("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != "finalized" {
		t.Errorf("Status = %q, want %q", inv.Status, "finalized")
	}
}

func TestInvoiceSend(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/invoices/inv_123/send" {
			t.Errorf("Path = %q, want /v1/invoices/inv_123/send", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice","status":"sending"}`)
	})

	inv, err := client.Invoices.Send("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != "sending" {
		t.Errorf("Status = %q, want %q", inv.Status, "sending")
	}
}

func TestInvoiceClone(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/invoices/inv_123/clone" {
			t.Errorf("Path = %q, want /v1/invoices/inv_123/clone", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"inv_456","object":"invoice","status":"draft"}`)
	})

	inv, err := client.Invoices.Clone("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.ID != "inv_456" {
		t.Errorf("ID = %q, want %q", inv.ID, "inv_456")
	}
}

func TestInvoiceGetPDF(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/invoices/inv_123/pdf" {
			t.Errorf("Path = %q, want /v1/invoices/inv_123/pdf", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"url":"https://storage.example.com/inv_123.pdf","expires_in":900}`)
	})

	resp, err := client.Invoices.GetPDF("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.URL == "" {
		t.Error("expected URL to be set")
	}
	if resp.ExpiresIn != 900 {
		t.Errorf("ExpiresIn = %d, want 900", resp.ExpiresIn)
	}
}

func TestInvoiceGetXML(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/invoices/inv_123/xml" {
			t.Errorf("Path = %q, want /v1/invoices/inv_123/xml", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		fmt.Fprint(w, `<?xml version="1.0"?><CrossIndustryInvoice/>`)
	})

	data, err := client.Invoices.GetXML("inv_123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty XML response")
	}
}

func TestInvoiceGetStatus(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/invoices/inv_123/status" {
			t.Errorf("Path = %q, want /v1/invoices/inv_123/status", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"deposited","einvoicing":{"paStatus":"accepted","paId":"pa_123"},"dates":{"due":"2026-04-15","finalizedAt":"2026-03-15T10:00:00Z"}}`)
	})

	resp, err := client.Invoices.GetStatus("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "deposited" {
		t.Errorf("Status = %q, want %q", resp.Status, "deposited")
	}
}

func TestInvoiceVerify(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice","verified":true,"chain_length":5,"details":"Chain is valid"}`)
	})

	resp, err := client.Invoices.Verify("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Verified {
		t.Error("expected Verified = true")
	}
	if resp.ChainLength != 5 {
		t.Errorf("ChainLength = %d, want 5", resp.ChainLength)
	}
}

func TestInvoiceCreatePaymentLink(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/invoices/inv_123/payment-link" {
			t.Errorf("Path = %q, want /v1/invoices/inv_123/payment-link", r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["success_url"] != "https://example.com/success" {
			t.Errorf("success_url = %q", body["success_url"])
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"payment_link","url":"https://checkout.stripe.com/xxx","session_id":"cs_test_123"}`)
	})

	resp, err := client.Invoices.CreatePaymentLink("inv_123", &PaymentLinkParams{
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.URL == "" {
		t.Error("expected URL to be set")
	}
	if resp.SessionID != "cs_test_123" {
		t.Errorf("SessionID = %q, want %q", resp.SessionID, "cs_test_123")
	}
}

func TestInvoiceCreatePaymentToken(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"payment_token","token":"abc123","pay_url":"/pay/abc123","expires_at":"2026-03-16T10:00:00Z"}`)
	})

	resp, err := client.Invoices.CreatePaymentToken("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token != "abc123" {
		t.Errorf("Token = %q, want %q", resp.Token, "abc123")
	}
}

func TestPaymentCreate(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/v1/invoices/inv_123/payments" {
			t.Errorf("Path = %q, want /v1/invoices/inv_123/payments", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"pay_abc","object":"payment","amount":"100.00","method":"transfer","paidAt":"2026-03-15T10:00:00Z"}`)
	})

	pay, err := client.Payments.Create("inv_123", &PaymentParams{
		Amount: 10000,
		Method: "transfer",
		PaidAt: "2026-03-15T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pay.ID != "pay_abc" {
		t.Errorf("ID = %q, want %q", pay.ID, "pay_abc")
	}
}

func TestCustomerCreateAndGet(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			w.WriteHeader(201)
		}
		fmt.Fprint(w, `{
			"id":"cus_abc",
			"object":"customer",
			"name":"ACME Corp",
			"type":"company",
			"siret":"12345678901234",
			"balance":"0.00",
			"currency":"eur",
			"active":true,
			"siretVerified":true,
			"vatVerified":false,
			"created":"2026-03-15T10:00:00Z",
			"updated":"2026-03-15T10:00:00Z"
		}`)
	})

	cus, err := client.Customers.Create(&CustomerParams{
		Name:  "ACME Corp",
		Type:  "company",
		SIRET: "12345678901234",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cus.Name != "ACME Corp" {
		t.Errorf("Name = %q, want %q", cus.Name, "ACME Corp")
	}
	if cus.SIRET != "12345678901234" {
		t.Errorf("SIRET = %q, want %q", cus.SIRET, "12345678901234")
	}
}

func TestProductCreateAndGet(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			w.WriteHeader(201)
		}
		fmt.Fprint(w, `{
			"id":"prod_abc",
			"object":"product",
			"name":"Widget",
			"unitPrice":"50.00",
			"vatRate":"20.00",
			"vatCode":"S",
			"unit":"piece",
			"active":true,
			"tags":["hardware"]
		}`)
	})

	prod, err := client.Products.Create(&ProductParams{
		Name:      "Widget",
		UnitPrice: 5000,
		VATRate:   2000,
		VATCode:   "S",
		Unit:      "piece",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prod.Name != "Widget" {
		t.Errorf("Name = %q, want %q", prod.Name, "Widget")
	}
}

func TestSandboxReset(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sandbox/reset" {
			t.Errorf("Path = %q, want /v1/sandbox/reset", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"sandbox_reset","deleted_count":15,"fixtures_created":35}`)
	})

	resp, err := client.Sandbox.ResetData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DeletedCount != 15 {
		t.Errorf("DeletedCount = %d, want 15", resp.DeletedCount)
	}
}

func TestSandboxSimulateStatus(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sandbox/simulate-status/inv_123" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice","status":"deposited","simulated":true}`)
	})

	resp, err := client.Sandbox.SimulateStatus("inv_123", &SimulateStatusParams{Status: "deposited"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Simulated {
		t.Error("expected Simulated = true")
	}
}
