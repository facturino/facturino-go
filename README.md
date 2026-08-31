# facturino-go

Go client library for the [Facturino](https://facturino.com) API — developer-first e-invoicing for French businesses.

[![Go Reference](https://pkg.go.dev/badge/github.com/facturino/facturino-go/v2.svg)](https://pkg.go.dev/github.com/facturino/facturino-go/v2)
[![Test](https://github.com/facturino/facturino-go/actions/workflows/test.yml/badge.svg)](https://github.com/facturino/facturino-go/actions/workflows/test.yml)

## Installation

```bash
go get github.com/facturino/facturino-go/v2
```

Requires Go 1.21+. No external dependencies.

## Quick Start

The recommended path is decision-first: identity → final tax decision →
create the decision-backed draft immediately → your chosen collection flow.
Facturino imposes no payment service provider and no payment method: an
immediate capture, a bank transfer, a direct debit or payment on agreed terms
all fit the same contract.

```go
package main

import (
    "fmt"
    "log"

    facturino "github.com/facturino/facturino-go/v2"
)

func main() {
    client := facturino.New("fac_test_xxx")

    // 1. Decide before the final amount is presented, the invoice is issued,
    //    or collection starts.
    decision, err := client.TaxDecisions.Create(&facturino.TaxDecisionParams{
        TaxSource:   "facturino", // or "integration" to supply your own VAT
        Customer:    "cus_8f2k4m9n",
        EffectiveAt: "2026-09-15",
        Currency:    "eur",
        PriceMode:   "tax_exclusive",
        Lines: []*facturino.TaxDecisionLineParams{{
            Reference:    "abo-pro",
            Description:  "Abonnement Pro",
            Category:     "electronically_supplied_services",
            RateCategory: "standard",
            UnitAmount:   2900, // integer centimes
            Quantity:     "1",  // decimal STRING, never a float
        }},
        // Required, 255 characters at most — checked before anything is sent.
        IdempotencyKey: "order-4711",
    })
    if err != nil {
        log.Fatal(err)
    }

    // 2. Act only on a final decision. "pending_verification" does not mean
    //    "nothing to charge": Totals and AmountToCharge are nil, not zero.
    if !decision.IsFinal() {
        log.Fatalf("decision not final: %v", decision.Issues)
    }

    // 3. Create the decision-backed draft immediately: no VAT is restated.
    invoice, err := client.Invoices.Create(&facturino.InvoiceParams{
        Customer:      decision.CustomerID,
        TaxDecisionID: decision.ID,
        DecisionLines: []*facturino.DecisionLineParams{{TaxLineRef: "abo-pro", Unit: "month"}},
        Buyer:         buyer,
        Dates:         &facturino.InvoiceDatesParams{Issued: "2026-09-15", Due: "2026-10-15"},
        Payment:       payment,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 4. Choose your collection flow — see the two variants below.
    fmt.Println(invoice.ID, *decision.AmountToCharge)
}
```

**Immediate collection** — capture the decided amount, verify, then finalize:

```go
// Capture exactly AmountToCharge through your payment provider, payment
// processor, bank transfer or external collection flow. Carry decision.ID in
// the provider metadata, order reference or custom reference. The settlement
// keeps its OWN financial reference (charge id, transfer wording…): the two
// identifiers are different things and must stay distinct.
settlement, err := yourCollectionProcess.Capture(
    *decision.AmountToCharge, decision.Currency,
    map[string]string{"taxDecisionId": decision.ID},
)
if err != nil {
    log.Fatal(err)
}

// Re-read the decision by its own id and verify what was actually captured.
source, err := client.TaxDecisions.Retrieve(decision.ID)
if err != nil {
    log.Fatal(err)
}
if settlement.Amount != *source.AmountToCharge {
    log.Fatal("amount mismatch")
}
if settlement.Currency != source.Currency {
    log.Fatal("currency mismatch")
}

if _, err := client.Invoices.Finalize(invoice.ID); err != nil {
    log.Fatal(err)
}

// Record the REAL payment — its real date, method and the settlement's
// financial reference, never the decision id
// (transfer, card, check, cash, direct_debit, sepa, paypal or other).
if _, err := client.Payments.Create(invoice.ID, &facturino.PaymentParams{
    Amount:    settlement.Amount,
    Method:    settlement.Method,
    Reference: settlement.Reference,
    PaidAt:    settlement.PaidAt,
}); err != nil {
    log.Fatal(err)
}

// Send to the platform only on the channel the FROZEN decision states.
if source.InvoiceChannel != nil && *source.InvoiceChannel == "einvoicing" {
    if _, err := client.Invoices.Send(invoice.ID); err != nil {
        log.Fatal(err)
    }
}
```

**Payment on terms** — finalize and deliver now, collect later:

```go
if _, err := client.Invoices.Finalize(invoice.ID); err != nil {
    log.Fatal(err)
}
if decision.InvoiceChannel != nil && *decision.InvoiceChannel == "einvoicing" {
    if _, err := client.Invoices.Send(invoice.ID); err != nil {
        log.Fatal(err)
    }
}

// …once the transfer arrives, record the REAL collection date.
if _, err := client.Payments.Create(invoice.ID, &facturino.PaymentParams{
    Amount:    *decision.AmountToCharge,
    Method:    "transfer",
    Reference: "VIR-2026-000871",
    PaidAt:    "2026-10-12",
}); err != nil {
    log.Fatal(err)
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

## Tax decisions

The full walkthrough lives in [Quick Start](#quick-start). A decision is
immutable: it fixes the VAT, the exact `AmountToCharge` and the reporting
obligations of one commercial operation, then never changes. Only a `final`
decision carries amounts, and the amount always comes from the decision —
never from a locally computed total.

`Create` answers `201` on creation and `200` when the same key already produced
that decision — both return the decision, so your code reads one shape either
way. Reusing the same key with a different body answers `409`, surfaced as a
`*facturino.Error` with `HTTPStatusCode == 409`; it is never retried.

### Optional: carrying the decision id through a PSP

These are examples, not requirements. If you collect through a PSP, keep the
decision id on the payment so step 4 can verify what was actually captured:
Stripe carries it in `metadata` (`facturino_tax_decision_id`), PayPal in
`custom_id` — and PayPal wants decimal units, so divide the centimes by 100.

### What a decision states

| Field | Meaning |
|---|---|
| `Status` | `final`, `pending_verification` or `unsupported`. Only `final` carries amounts. |
| `AmountToCharge` | Exact amount to debit, integer centimes. `nil` unless final. |
| `Totals` | `TotalHT` / `TotalVAT` / `TotalTTC`, integer centimes. `nil` unless final. |
| `InvoiceChannel` | `einvoicing` or `none` — whether the invoice travels the network. |
| `TransactionReporting` | `ereporting`, `none` or `outside_scope`. |
| `PaymentReporting` | `fr212`, `ereporting` or `none`. |
| `ForeignTaxReviewRequired` | A foreign tax may apply; review it outside Facturino. |
| `Vies` | VIES status only (`valid`, `invalid`, `unavailable`, `invalid_format`). |
| `Issues` | What is missing, when the decision is not final. |
| `ObligationReasons` | Why each axis carries the obligation it does. |
| `ExpiresAt` / `Expired` | Past this instant the decision no longer opens a payment. |

Facturino decides **French VAT and the matching French obligations**. It does
not provide worldwide tax compliance: when a foreign tax may apply, the decision
says so through `ForeignTaxReviewRequired`. An operation whose `InvoiceChannel`
is `none` is not deposited on a certified platform — its obligation, if any,
goes through e-reporting.

### Missing evidence, then a retry

Supply the evidence and retry the SAME operation. Send the territorial
**signal**, never the raw one: a country and, where the territory needs it, a
postal code — not an IP address, a PSP payload or bank account details.

```go
params.RetryOfTaxDecisionID = pending.ID
params.LocationEvidence = []*facturino.LocationEvidenceParams{{
    Kind:        "billing_address",
    Country:     "FR",
    PostalCode:  "75002",
    ThirdParty:  false,
    Source:      "declared",
    CollectedAt: "2026-09-15",
}}
```

## Three status axes

A document has three states that do not follow from one another. The historical
`Status` field stays populated as their projection.

```go
inv.DocumentStatus     // draft | finalized | cancelled
inv.TransmissionStatus // not_applicable | pending | sending | deposited | transmitted | approved | rejected
inv.TransmissionDetail // available | received | suspended | refused
inv.PaymentStatus      // unpaid | partially_paid | paid | partially_refunded | refunded
```

Recording a payment never moves the transmission axis, and a refund does not
erase the collection that happened.

## Supplying your own VAT (`TaxSource: "integration"`)

If your system already determines the VAT — an ERP, a marketplace engine, an
in-house rules service — declare it on the decision instead of asking
Facturino to determine it. Each line then carries the VAT you supply:
`VatRate` (a pointer to an integer centipercent — zero is a real rate),
`VatCode` (S, Z, E, AE, K, G or O) and, when the rate is zero, the
`VatexCode` and `PlaceOfSupply` justifying it. Facturino validates the
coherence of the whole (a positive rate with an exemption code, a franchise
seller charging VAT, a reverse charge to a consumer… are refused with
`integration_vat_incoherent`) and never silently corrects a rate. The
decision, the invoice and the reporting obligations then work exactly as with
`TaxSource: "facturino"` — the two journeys are equals.

```go
standardRate := 2000
decision, err := client.TaxDecisions.Create(&facturino.TaxDecisionParams{
    TaxSource:   "integration",
    Customer:    "cus_8f2k4m9n",
    EffectiveAt: "2026-09-15",
    Currency:    "eur",
    PriceMode:   "tax_exclusive",
    Lines: []*facturino.TaxDecisionLineParams{{
        Reference:   "conseil",
        Description: "Prestation de conseil",
        Category:    "services",
        UnitAmount:  10000,
        Quantity:    "1",
        VatRate:     &standardRate, // 20.00 % — supplied by YOUR system
        VatCode:     "S",
    }},
    IdempotencyKey: "order-4712",
})

invoice, err := client.Invoices.Create(&facturino.InvoiceParams{
    Customer:      decision.CustomerID,
    TaxDecisionID: decision.ID,
    DecisionLines: []*facturino.DecisionLineParams{{TaxLineRef: "conseil", Unit: "unit"}},
    Buyer:         buyer,
    Dates:         &facturino.InvoiceDatesParams{Issued: "2026-09-15", Due: "2026-10-15"},
    Payment:       payment,
})
```

## Amounts and Rates

All monetary amounts are **integers in centimes** (10000 = 100.00 EUR).
VAT rates are **integers in centipercent** (2000 = 20.00%).

```go
line := &facturino.TaxDecisionLineParams{
    UnitAmount: 15000, // 150.00 EUR
    Quantity:   "1",   // decimal string
}
```

## Resources

### Invoices

```go
inv, _ := client.Invoices.Create(&facturino.InvoiceParams{
    Customer:      "cus_xxx",
    TaxDecisionID: "taxdec_xxx",
    DecisionLines: []*facturino.DecisionLineParams{{TaxLineRef: "abo-pro", Unit: "month"}},
})
// TaxDecisionID and DecisionLines are required — the SDK refuses their
// absence locally, before any HTTP call. Deposits and Schedule travel
// alongside the decision; both settle server-side against the decided amount.
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
    Name:    "ACME Corp",
    Type:    "company",
    SIRET:   "73282932000074",
    Address: &facturino.Address{Line1: "10 rue de la Paix", PostalCode: "75002", City: "Paris", Country: "FR"},
})

// Lookup resolves company details from the INSEE Sirene registry
// (not a stored customer): use the result to prefill CustomerParams.
lookup, _ := client.Customers.Lookup(&facturino.CustomerLookupParams{
    SIRET: "73282932000074",
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

// Convert, decide, bind, finalize: ONE invoice throughout. A converted quote
// yields a COMMERCIAL draft — it states the operation and no VAT
// (TaxSource empty). Bind a final decision to that same invoice, then
// finalize it; never create a second one.
converted, _ := client.Quotes.Convert(q.ID)
decision, _ := client.TaxDecisions.Create(decisionInput)
client.Invoices.BindTaxDecision(converted.ID, &facturino.BindTaxDecisionParams{
    TaxDecisionID: decision.ID,
    DecisionLines: []*facturino.DecisionLineParams{{TaxLineRef: "l1", Unit: "unit"}},
})
client.Invoices.Finalize(converted.ID)
```

### Credit Notes

```go
cn, _ := client.CreditNotes.Create(&facturino.CreditNoteParams{
    RelatedInvoiceID: "inv_xxx",
    CreditNoteType:   "total",
    ReasonCode:       "duplicate",
    // The VAT is inherited from the invoice's frozen snapshot, never restated.
    CreditedLines:    []*facturino.CreditedLineParams{{TaxLineRef: "abo-pro"}},
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
    // Required: each occurrence gets its OWN decision on its generation date.
    TaxInputs: &facturino.RecurringTaxInputsParams{
        TaxSource: "facturino",
        PriceMode: "tax_exclusive",
        Lines:     []*facturino.RecurringTaxLineParams{{ /* ... */ }},
    },
    TemplateInvoice: &facturino.RecurringTemplateParams{PaymentTermsDays: 30},
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

### Reference & Health

```go
forms, _ := client.Reference.ListLegalForms(&facturino.ReferenceListParams{Search: "SAS"})
providers, _ := client.Reference.ListPaProviders() // supported Plateformes Agréées (BYOPA)
status, _ := client.Health.Check()                 // public liveness probe
```

> **Public token endpoints** — the recipient-facing portals (`/pay/:token`,
> `/portal/:token`, `/quote-portal/:token`) are intentionally not exposed by the
> SDK: they are opened by the end recipient through a hosted page, not called
> with an API key.

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

An `Idempotency-Key` protects the **replay of one request**. It is not a
deduplicator: the API never decides on its own that two requests "mean the same
thing".

- **Same key + same canonical body** — the first 2xx response is replayed
  verbatim, and the operation is not executed a second time.
- **Same key + different body** — `409 idempotency_error`. A key belongs to a
  request, not to an endpoint.
- **Different keys** — two distinct operations, even with byte-identical bodies.
  Two requests describing the same operation are **not** deduplicated
  automatically; the key, and only the key, declares that two sends are the same
  attempt.
- **Canonical body** — JSON object keys are compared in a stable order, so
  reordering them does not change the request. Changing a value, adding or
  removing a field does. Array order is significant: two lines swapped are two
  different documents.
- **Failure before execution** (validation, read-only field, sanitisation)
  releases the key, so a corrected retry with the same key runs.
- **Business refusal during execution** is stored and replayed; the operation is
  not re-executed.
- **Scope** — 24 hours, per API key. `POST /v1/tax-decisions` additionally
  carries a durable business idempotency that never expires.

```go
// Same key + same body -> the first response, replayed.
inv, err := client.Invoices.Create(&facturino.InvoiceParams{
    Customer:       "cus_xxx",
    TaxDecisionID:  "taxdec_xxx",
    DecisionLines:  []*facturino.DecisionLineParams{{ /* ... */ }},
    IdempotencyKey: "order-4821",
})

// Retry with new evidence is NOT idempotency. It takes a NEW decision on the
// same commercial operation: use a NEW key and link the previous decision.
params.RetryOfTaxDecisionID = suspended.ID
params.IdempotencyKey = "order-4821-retry-1"
decision, err := client.TaxDecisions.Create(params)
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
