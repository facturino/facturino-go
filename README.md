# facturino-go

Go client library for the [Facturino](https://facturino.com) API — developer-first e-invoicing for French businesses.

[![Go Reference](https://pkg.go.dev/badge/github.com/facturino/facturino-go/v2.svg)](https://pkg.go.dev/github.com/facturino/facturino-go/v2)
[![Test](https://github.com/facturino/facturino-go/actions/workflows/test.yml/badge.svg)](https://github.com/facturino/facturino-go/actions/workflows/test.yml)

## Installation

```bash
go get github.com/facturino/facturino-go/v2@v2.2.0
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
| `SettledObligations` | The axes French law settles DESPITE a non-final decision. `nil` when final — the three axes above are then the settled ones. Each axis inside is `nil` when it depends on the treatment that could not be concluded. It authorises nothing. |
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

### What this source is not

`VatRate`, `VatCode` and `VatexCode` describe **French VAT**. The contract has no
local-tax jurisdiction, no local tax scheme and no withholding, so this source is
not a way to pass one through.

Where a local tax of a French overseas collectivity or the TAAF
(Saint-Pierre-et-Miquelon, Saint-Barthélemy, Saint-Martin, French Polynesia, New
Caledonia, Wallis-and-Futuna, TAAF) can change what you invoice — or what you
actually collect — the decision is **not final under either source**, with the
same issue code and no amount:

| Issue code | When |
|---|---|
| `com_taaf_local_tax_not_determined` | Non-taxable buyer, operation located in the collectivity (electronically supplied service, CGI art. 259 D). |
| `com_taaf_local_regime_not_sourced` | Taxable buyer, but no official act of the collectivity states who bears its tax. |
| `com_taaf_payment_withholding_not_modelled` | French Polynesia: the client withholds part of the payment at source. |
| `seller_com_taaf_local_tax_not_determined` | The seller itself is established in one of the seven. |

The one sourced exception stays final under both sources: a B2B service located
in **New Caledonia** supplied by a seller **not established in New Caledonia**,
where art. Lp. 507-1 makes the taxable customer account for the taxe générale sur
la consommation itself. That article reaches only a supplier established outside
the territory. A seller established in New Caledonia is the ordinary collector of
the taxe générale sur la consommation on its own sales, at a rate this contract
does not carry, so its decision is not final either
(`seller_com_taaf_local_tax_not_determined`).

Because of this, `PlaceOfSupply` is **required** on every line as soon as the
buyer is established in one of the seven — the place of the operation is what
says whether the local tax is at stake, and it is never assumed
(`422 integration_vat_incoherent` otherwise). A place located in France (goods
that never leave the territory, a general B2C service) keeps the decision final
as anywhere else.

## B2C sales to consumers in other member states

An **electronically supplied service** to a consumer established in another
member state (Directive 2006/112/EC art. 58) and an **intra-EU distance sale** of
goods (art. 33(a)) follow one common regime. A **general** B2C service does not:
it stays taxed where the supplier is established (art. 45), and nothing below
concerns it.

Four questions are answered separately, in this order. Collapsing any two of them
produces a wrong rate:

1. is the operation covered by a destination rule;
2. does the common **EUR 10,000** threshold still allow taxation at origin
   (art. 59c(1));
3. did the seller **opt** for taxation at destination (art. 59c(3));
4. how is the tax due at destination **declared** — Union one-stop shop, or a
   local VAT registration in that member state?

The one-stop shop answers the **last** of the four: it is a way of declaring and
paying a tax, not a rule of place. Not registering never restores the seller's
own national VAT — it leaves the decision without an amount.

That does not make the questions watertight in fact. For a **French** seller,
registering for the Union scheme is how the option of art. 59c(3) is exercised:
an **active** registration therefore settles the place at destination on its own,
and the threshold has nothing left to decide (`basis: "oss_union_registration"`,
`threshold: null`). The registration is dated — one opened in October decides
nothing for a September sale, and one that ended decides nothing any more. It is
sourced for France only: the way the option is exercised is fixed by the member
state where it is exercised, so a seller established elsewhere keeps the ordinary
threshold path and states its option explicitly.

The threshold is an **inclusive** cap of `1000000` centimes excluding VAT, open
only to a seller established in a **single** member state; the operation that
carries the running total past it is itself taxed at destination.

That running total lives in an **annual ledger** — `/v1/eu-threshold-ledgers`,
one per company, per mode and per calendar year — and NOT on the fiscal profile.
A profile revision is an immutable rule that decisions freeze; a turnover total
moves with every sale and gets corrected, so the two are kept apart. Facturino
keeps the register of the operations it receives; the sales made on your other
channels must be brought in by an adjustment, or by the opening declaration:

- **opening a year** declares four figures — the previous year's total and the
  total already made this year, each with its *services* part (see the two
  counters below) — plus the coverage mode: `facturino_only` (every covered sale
  goes through Facturino) or `mixed_channels`;
- **an adjustment** adds the turnover of another channel and moves forward the
  day those channels are declared complete through. Under `mixed_channels` a
  decision is served only up to that day.

**Two counters, strictly apart — and independent.** The same movements feed two
thresholds that do not measure the same thing: the common EUR 10,000 threshold
above, and the EUR 100,000 threshold of Reg. 282/2011 art. 24b that governs how
many items of location evidence are required. The second counts only telecom,
broadcasting and electronically supplied services, **domestic ones included**;
the first counts only **cross-border** supplies. A distance sale of goods raises
the first and never the second — and a domestic electronic service raises the
second and never the first.

Neither bounds the other, in either direction: a publisher selling mostly at home
legitimately declares far more on the evidence counter than on the common one.
That is why every figure comes in a pair (`amount` / `evidenceAmount`,
`acquiredMin` / `acquiredEvidenceMin`, …) rather than as a total and a share of
it, and why the single-evidence relaxation is **computed by the engine** on that
second counter rather than declared by the seller.

**Acquired and reserved are published apart, and never summed.** `acquiredMin`
is what the year has certainly made; `reservedMin` is the slices held right now
by operations still being decided. A held slice may still disappear, and one
combined "total" would hide exactly that.

Nothing is assumed: no year starts at zero on its own, and no sale made elsewhere
is presumed absent. A decision reserves its slice of the total in a transaction
and consumes it with the decision itself, so two concurrent operations never read
the same figure as certain and a replay counts nothing twice. A verdict is frozen
only when it holds at **both** bounds — with every concurrent slice counted and
with none of them — which is what makes an abandoned operation simply disappear
instead of staying in the total as turnover that never existed.

**Giving an amount back is a qualified correction, never a negative
adjustment.** Directive 2006/112/EC art. 90(1) reduces the taxable amount of a
supply on cancellation, refusal or a price reduction after the supply, and the
thresholds count the VALUE of the supplies — so a correction names the movement
it corrects, its qualification, the resource it rests on and its evidence.

A movement gives back what it brought in **once**, whatever the number of
corrections: the ledger keeps each movement's balance inside the same
transaction, and every entry publishes its `remainingMin`. `correctsEntryId` is
restricted to the ids the ledger itself mints — it reaches a document path, and
free text must not.

What cannot be qualified that way is not subtracted at all: the ledger goes
**under review** and stops deciding, rather than freeze verdicts on a total
nobody stands behind. A review is settled by **reconciliation**, never by a
comment: you state the version you checked and the two acquired totals you
verified, and only an exact agreement reopens the ledger.

**A decision is never frozen without its slice.** If the slice it held has
disappeared when the decision is about to be written, the transaction is
abandoned — no decision, no audit entry, no settled claim
(`409 eu_threshold_reservation_lost`) — and the ledger goes under review on a
path of its own, so the review survives that abandonment. Freezing a decision the
running total does not carry would leave the sale inside a decision and outside
the year the next operation reads.

Movements are paginated with a cursor (`limit`, `starting_after`): the ledger
keeps them all, a page shows some.

| Issue code | When |
| --- | --- |
| `eu_threshold_state_missing` | No ledger is open for the year of the operation. Open it: nothing is assumed to be zero. |
| `eu_threshold_external_coverage_incomplete` | Other channels exist and are not declared complete through the operation date. Record an adjustment — even a zero one, which simply states that nothing happened. |
| `eu_threshold_backdated_operation` | The operation predates one already counted, and decisions were frozen on that running total. It is refused rather than silently recomputed. |
| `eu_threshold_concurrent_decision_pending` | Other operations of the company are being decided at this very moment, and the cap falls between "all of them confirmed" and "all abandoned". Nothing is frozen on that: decide again once they conclude. |
| `eu_threshold_review_required` | The ledger is under review — its running total is known to be wrong, and no verdict rests on it until the review is settled. |
| `location_evidence_relief_undetermined` | One third-party item of evidence, and the art. 24b relaxation could not be established: open the year's ledger and declare the services figures, or supply a second item. |
| `eu_threshold_reservation_lost` | The slice this decision held is gone. The decision is refused rather than frozen without it, and the ledger goes under review. |
| `destination_threshold_operation_value_missing` | A line value cannot be sized, so no slice of the total can be taken for it. |
| `destination_threshold_price_mode_ambiguous` | Tax-inclusive price: the VAT-exclusive value depends on the rate the threshold has to decide, and the bounds fall on both sides of the cap. |
| `destination_option_period_invalid` | The option is declared over less than its minimum binding period. |
| `destination_option_scope_not_sourced` | Seller established outside France: the binding period is fixed by the member state where the option is exercised, and only the French one is sourced here. |
| `destination_establishment_in_member_state` | The seller declares an establishment in the destination member state: which establishment supplies then decides both the place and the mechanism, and no fact of the contract names it. A local VAT registration alone is not an establishment. |
| `destination_regional_scope_undetermined` | The member state publishes regional standard rates and the address places none of them — state the customer's postal code. |
| `destination_mechanism_missing` | Destination taxation is due with neither the Union scheme nor a local registration valid for that state on that date. |
| `franchise_destination_taxation_not_modelled` | Seller under the French small-enterprise exemption whose operation is taxed by another member state. |

Rates come from a **local, dated, versioned registry**: no network call during a
decision, and a decision replayed years later reproduces the same rate. A rate
change is a new period, never a rewrite. Only the **standard** rate of the 27
member states is tabulated — the scope of the reduced rates follows a national
classification this contract does not hold, and an ordinary electronically
supplied service is never granted the rate an electronic publication may benefit
from (`destination_rate_band_not_available`). An effect date before the registry
answers `destination_rate_not_sourced_for_date`.

A **region publishing its own standard rate** never blocks a whole member state.
What governs the region decides the answer:

- **the rate follows the place of the operation.** The region is then a territory
  of its own and the address decides: Portugal mainland 23%, Madeira 22%, Azores
  16% (CIVA art. 18, CTT postal ranges). Only an address placing no region at all
  is refused — `destination_regional_scope_undetermined`, naming the missing fact
  rather than falling back on the mainland rate;
- **the rate is reserved to operations carried out in the zone by a supplier
  established there.** A supplier at distance does not acquire it from the
  consumer's address, so the national rate is the final answer — and no postal
  cartography of the zone is needed to say so: Austrian Jungholz and
  Kleinwalsertal (§ 10(4) UStG — 20%, not 19%), and the Greek island regime FOR
  SERVICES, which AADE reserves to a supplier established on the island for an
  operation performed there. An electronically supplied service from France is
  therefore taxed at 24% in Greece, in Athens and in Kalymnos alike;
- **the rate follows the destination and the zone is not cartographied here.**
  Neither the regional rate nor the national one can be asserted, so the
  operation is refused: `destination_regional_regime_not_sourced`, non-final and
  without an amount. This is Greece FOR GOODS: since 2026-01-01 the islands of
  fewer than 20,000 inhabitants apply a reduced standard rate to the goods
  delivered there, intra-community acquisitions included, and the list of those
  islands is not cartographied by postal code here. An intra-EU distance sale of
  goods to Greece is therefore refused — in Athens as in Kalymnos — and never
  taxed at 24% by default. The answer is given per FAMILY of operation: the same
  member state can be settled for services and left open for goods.

### The annual ledger

```go
// Open the year — nothing starts at zero on its own.
_, _ = client.EuThresholdLedgers.Open(&facturino.OpenEuThresholdLedgerParams{
    Year:                        "2026",
    PreviousYearAmount:          250000, // centimes, VAT excluded
    CurrentYearOpening:          100000,
    CoverageMode:                "mixed_channels",
    ExternalCompleteThroughDate: "2026-01-01",
})

// Bring in what was sold elsewhere. Append-only, idempotent on Reference.
_, _ = client.EuThresholdLedgers.Adjust("2026", &facturino.EuThresholdAdjustmentParams{
    Reference:                   "marketplace-2026-08",
    Amount:                      40000,
    ExternalCompleteThroughDate: "2026-09-15",
    Reason:                      "Marketplace sales, August",
})

ledger, _ := client.EuThresholdLedgers.Get("2026")
_ = ledger.CumulativeMin // total already acquired, in centimes
_ = ledger.RemainingMin  // what is left before the cap
```

### The same rule under `taxSource: 'integration'`

An integration that concludes its own VAT does not get a different territoriality.
The `integration` source traverses the same coverage, the same threshold, the same
option, the same evidence and the same declarative mechanism; what differs is the
outcome. Where `facturino` **produces** the rate, `integration` **compares** the
one you supply to the legal result:

- equal — decision `final`;
- a category no B2C supply taxed at destination can carry (`AE`, `K`, `G`, `O`),
  or a `placeOfSupply` the rule contradicts — `422 integration_vat_incoherent`;
- a rate neither the destination **standard** rate nor the published bands of the
  seller's own territory confirm — non-final, with
  `eu_b2c_rate_supplied_mismatch`. Facturino holds only the standard rate of
  another member state, and only the bands it publishes for a French territory,
  so it can neither confirm a reduced rate nor correct yours. At origin that
  refusal asserts NO foreign tax: the operation is taxed in France;
- a rule that could not conclude — blocked exactly as under `facturino`, with
  the same meaning of `pending_verification` and `unsupported`.

The confrontation reaches BOTH places the rule can settle, and the territorial
frontier is shared too: a seller established outside the French VAT territory, or
a buyer sitting in a territory excluded from the EU VAT territory, raises under
`integration` exactly the obstacle it raises under `facturino`.

Because of this, `goodsMovement` is **required** on a goods line as soon as the
buyer is a consumer of another member state: that movement decides whether the
distance-sale rule applies, and it is never assumed.

### What the decision freezes

`TaxDecision.EuB2cDestination` carries what the rule concluded, as data rather
than as a sentence — the verdict and its basis, the threshold figures it was
decided on, the declarative mechanism, and the rate entry with its registry
version, its source, its verification date, its period and its region. It is
present as soon as the rule covers a line — including on a decision that is NOT
final, where it states exactly what is missing — and `nil` on every operation
the rule does not reach.


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
