# facturino-go

Go client library for the [Facturino](https://facturino.com) API — developer-first e-invoicing for French businesses.

[![Go Reference](https://pkg.go.dev/badge/github.com/facturino/facturino-go.svg)](https://pkg.go.dev/github.com/facturino/facturino-go)
[![Test](https://github.com/facturino/facturino-go/actions/workflows/test.yml/badge.svg)](https://github.com/facturino/facturino-go/actions/workflows/test.yml)

## Installation

```bash
go get github.com/facturino/facturino-go
```

Requires Go 1.21+. No external dependencies.

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    facturino "github.com/facturino/facturino-go"
)

func main() {
    client := facturino.New("fac_test_xxx")

    customer, err := client.Customers.Create(&facturino.CustomerParams{
        Name:  "ACME Corp",
        Type:  "company",
        Email: "billing@acme.com",
        SIRET: "12345678901234",
    })
    if err != nil {
        log.Fatal(err)
    }

    invoice, err := client.Invoices.Create(&facturino.InvoiceParams{
        Customer: customer.ID,
        Buyer: &facturino.BuyerParams{
            CompanyName: "Acme SAS",
            Siret:       "55208131766522",
            Address:     &facturino.Address{Line1: "10 rue de la Paix", PostalCode: "75002", City: "Paris", Country: "FR"},
        },
        Items: []*facturino.ItemParams{
            {
                Description: "Consulting - Mars 2026",
                Quantity:    "1",   // decimal string
                Unit:        "unit",
                VATCode:     "S",
                UnitPrice:   10000, // 100.00 EUR (centimes)
                VATRate:     2000,  // 20.00% (centipercent)
            },
        },
        Dates: &facturino.InvoiceDatesParams{Issued: "2026-07-01", Due: "2026-07-31"},
        Payment: &facturino.PaymentTermsParams{
            Terms: "Paiement à 30 jours", TermsDays: 30, Method: "transfer",
            LatePaymentRate: "10.00", CollectionFee: "40.00",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    finalized, err := client.Invoices.Finalize(invoice.ID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Invoice %s finalized: %s\n", finalized.Number, finalized.Status)

    _, err = client.Invoices.Send(finalized.ID)
    if err != nil {
        log.Fatal(err)
    }

    // One-shot: set AutoFinalize (and optionally AutoSend) on InvoiceParams to
    // finalize and deliver in a single Create call:
    //   client.Invoices.Create(&facturino.InvoiceParams{..., AutoFinalize: true,
    //       AutoSend: &facturino.AutoSendParams{Email: true, PA: true}})
}
```

## Configuration

```go
client := facturino.New("fac_test_xxx")

client := facturino.New("fac_live_xxx",
    facturino.WithBaseURL("https://facturino.com/api"),
    facturino.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
    facturino.WithMaxRetries(5),
)
```

## Amounts and Rates

All monetary amounts are **integers in centimes** (10000 = 100.00 EUR).
VAT rates are **integers in centipercent** (2000 = 20.00%).

```go
item := &facturino.ItemParams{
    UnitPrice: 15000, // 150.00 EUR
    VATRate:   2000,  // 20.00%
}
```

## Resources

### Invoices

```go
inv, _ := client.Invoices.Create(&facturino.InvoiceParams{...})
inv, _ = client.Invoices.Get("inv_xxx")
// Inline related resources: expand "customer", "items.product" and/or
// "credit_notes" (also yields Expanded.NetBalance).
inv, _ = client.Invoices.Get("inv_xxx", &facturino.InvoiceGetParams{Expand: []string{"credit_notes"}})
inv, _ = client.Invoices.Update("inv_xxx", &facturino.InvoiceUpdateParams{...})
_ = client.Invoices.Delete("inv_xxx")

inv, _ = client.Invoices.Finalize("inv_xxx")
inv, _ = client.Invoices.Send("inv_xxx")
_ = client.Invoices.Remind("inv_xxx", nil) // or &facturino.InvoiceRemindParams{Level: 2}
sent, _ := client.Invoices.Email("inv_xxx", nil)
_ = sent.Status // "sent" or "pending" (job still rendering the PDF)
clone, _ := client.Invoices.Clone("inv_xxx")

pdf, _ := client.Invoices.GetPDF("inv_xxx")
facturx, _ := client.Invoices.GetFacturX("inv_xxx")
xml, _ := client.Invoices.GetXML("inv_xxx", "cii") // or "ubl"

status, _ := client.Invoices.GetStatus("inv_xxx")
verify, _ := client.Invoices.Verify("inv_xxx")

link, _ := client.Invoices.CreatePaymentLink("inv_xxx", &facturino.PaymentLinkParams{...})
token, _ := client.Invoices.CreatePaymentToken("inv_xxx")
```

### Payments (sub-resource)

```go
pay, _ := client.Payments.Create("inv_xxx", &facturino.PaymentParams{
    Amount: 12000,           // 120.00 EUR
    Method: "transfer",
    PaidAt: "2026-03-15T10:00:00Z",
})

iter := client.Payments.List("inv_xxx", nil)
for iter.Next() {
    fmt.Println(iter.Payment().Amount)
}
```

### Customers

```go
cus, _ := client.Customers.Create(&facturino.CustomerParams{
    Name:  "ACME Corp",
    Type:  "company",
    SIRET: "12345678901234",
})

// Lookup resolves company details from the INSEE Sirene registry
// (not a stored customer): use the result to prefill CustomerParams.
lookup, _ := client.Customers.Lookup(&facturino.CustomerLookupParams{
    SIRET: "12345678901234",
})
if lookup.Found {
    fmt.Println(lookup.Data.Name, lookup.Data.LegalForm.Sigle)
}
```

### Products

```go
prod, _ := client.Products.Create(&facturino.ProductParams{
    Name:      "Consulting",
    UnitPrice: 10000,
    VATRate:   2000,
    Unit:      "heure",
})

// Filter the catalog by name prefix, category and/or active flag.
active := true
iter := client.Products.List(&facturino.ProductListParams{
    Q:        "cons",
    Category: "services",
    Active:   &active,
})
```

### Quotes

```go
q, _ := client.Quotes.Create(&facturino.QuoteParams{
    Customer: "cus_xxx",
    Items:    []*facturino.ItemParams{{...}},
})
q, _ = client.Quotes.Send(q.ID)
q, _ = client.Quotes.Accept(q.ID)
dup, _ := client.Quotes.Clone(q.ID)  // Duplicates the quote into a new draft
inv, _ := client.Quotes.Convert(q.ID) // Creates invoice from accepted quote
```

### Credit Notes

```go
cn, _ := client.CreditNotes.Create(&facturino.CreditNoteParams{
    Customer:         "cus_xxx",
    RelatedInvoiceID: "inv_xxx",
    CreditNoteType:   "total",
    ReasonCode:       "duplicate",
    Items:            []*facturino.ItemParams{{...}},
})
cn, _ = client.CreditNotes.Finalize(cn.ID)
```

### Recurring Invoices

```go
ri, _ := client.RecurringInvoices.Create(&facturino.RecurringInvoiceParams{
    CustomerID: "cus_xxx",
    Frequency:  "monthly",
    StartDate:  "2026-04-01",
    NextGenerationDate: "2026-04-01",
    TemplateInvoice: &facturino.RecurringTemplateParams{
        Items: []*facturino.ItemParams{{...}},
    },
    AutoFinalize: true,
})
_, _ = client.RecurringInvoices.Pause(ri.ID)
ri, _ = client.RecurringInvoices.Resume(ri.ID)
```

### Exports

```go
fec, _ := client.Exports.GenerateFEC(&facturino.FECParams{
    PeriodStart: "2026-01-01",
    PeriodEnd:   "2026-12-31",
})

job, _ := client.Jobs.Get(fec.ID)
```

### E-Reporting

```go
decl, _ := client.EReporting.CreateDeclaration(&facturino.EReportingParams{
    Type:   "b2c",
    Period: "2026-02",
    Lines: []*facturino.EReportingLineParams{{
        Category:  "services",
        Amount:    100000, // 1000.00 EUR
        VATRate:   2000,
        VATAmount: 20000,
    }},
})
decl, _ = client.EReporting.SubmitDeclaration(decl.ID)
```

### Sandbox

```go
reset, _ := client.Sandbox.ResetData()

sim, _ := client.Sandbox.SimulateStatus("inv_xxx", &facturino.SimulateStatusParams{
    Status: "deposited",
})
```

## Pagination

All list endpoints return iterators with automatic pagination.

```go
iter := client.Invoices.List(&facturino.InvoiceListParams{
    ListParams: facturino.ListParams{
        Limit:  10,
        Status: "paid",
    },
    ConvertedFrom: "quo_xxx", // optional: only invoices issued from this quote
})
for iter.Next() {
    inv := iter.Invoice()
    fmt.Printf("%s: %s (%s)\n", inv.ID, inv.Number, inv.Status)
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

## Idempotency

Pass an idempotency key on POST requests to safely retry:

```go
inv, err := client.Invoices.Create(&facturino.InvoiceParams{
    Customer:       "cus_xxx",
    Items:          []*facturino.ItemParams{{...}},
    IdempotencyKey: "unique-request-id-123",
})
```

## Error Handling

API errors are returned as `*facturino.Error`:

```go
inv, err := client.Invoices.Get("inv_nonexistent")
if err != nil {
    if facturino.IsNotFound(err) {
        fmt.Println("Invoice not found")
    } else if facturino.IsRateLimit(err) {
        fmt.Println("Rate limited, retry later")
    } else if facturino.IsAuthentication(err) {
        fmt.Println("Invalid API key")
    } else {
        apiErr, ok := err.(*facturino.Error)
        if ok {
            fmt.Printf("API error: %s (code=%s)\n", apiErr.Message, apiErr.Code)
        }
    }
}
```

## Webhook Verification

Verify webhook signatures with HMAC-SHA256:

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    payload, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("Facturino-Signature")

    event, err := facturino.VerifyWebhookSignature(payload, signature, "whsec_xxx")
    if err != nil {
        http.Error(w, "Invalid signature", http.StatusBadRequest)
        return
    }

    switch event.Type {
    case "invoice.finalized":
    case "invoice.paid":
    case "invoice.approved":
    }

    w.WriteHeader(http.StatusOK)
}
```

## Retries

Retries on 429 (respects `Retry-After`), 500, 502, 503 with exponential backoff (500ms, 1s, 2s). Default: 3 retries. Use `WithMaxRetries(0)` to disable.

## Development

```bash
git clone https://github.com/facturino/facturino-go.git
cd facturino-go
go test ./...
```

## License

MIT License. See [LICENSE](LICENSE).
