# Changelog

All notable changes to the Facturino Go SDK are documented here. This project
adheres to [Semantic Versioning](https://semver.org/) and
[Keep a Changelog](https://keepachangelog.com/).

## [2.0.0] - 2026-08-31

The first STABLE tax determination contract. Every invoice is now created from
an immutable tax decision, taken before the document exists, whichever of the
two fiscal journeys takes it. The module moves to its v2 major path:
`github.com/facturino/facturino-go/v2`.

### Added
- `Client.TaxDecisions` with `Create` and `Get` (`Retrieve` is an alias):
  `POST /v1/tax-decisions` and `GET /v1/tax-decisions/{id}`. A decision fixes
  the VAT, the exact `AmountToCharge` and the three reporting axes for one
  operation, then never changes.
- Two equal fiscal sources, declared with the required
  `TaxDecisionParams.TaxSource`: `"facturino"` (Facturino determines the VAT
  from the commercial lines and the territorial evidence) and `"integration"`
  (your system supplies the VAT of every line through
  `TaxDecisionLineParams.VatRate` — a pointer, zero being a real rate —
  `VatCode`, and where the rate is zero a `VatexCode` and `PlaceOfSupply`).
  Supplied VAT is validated for coherence — contradictions answer
  `integration_vat_incoherent`, never a silent correction.
- Tax-decision types: `TaxDecision` (with `TaxSource`), `TaxDecisionParams`,
  `TaxDecisionLineParams`, `TaxDecisionLine`, `TaxDecisionTotals`,
  `TaxDecisionDiscount`, `TaxDecisionIssue`, `TaxDecisionObligationReason`,
  `ViesResult`, `LocationEvidenceParams`, `NonEuBusinessEvidenceParams` and
  their result counterparts. `TaxDecision.IsFinal()` reports whether the
  decision carries amounts.
- `Invoice` and `CreditNote` expose the three status axes — `DocumentStatus`,
  `TransmissionStatus`, `TransmissionDetail`, `PaymentStatus` — plus
  `TaxSource`, `TaxDecisionID` / `OriginalTaxDecisionID` and `TaxSnapshot`.
  The `Status` field stays populated as their summary projection.
- `InvoiceParams.Deposits` and `InvoiceParams.Schedule` travel alongside the
  decision: both are settled server-side against the DECIDED amount due.
- `Invoices.BindTaxDecision(id, *BindTaxDecisionParams)`:
  `POST /v1/invoices/{id}/bind-tax-decision`. Freezes a FINAL decision onto a
  commercial draft that already exists — the one `Quotes.Convert` produced — so
  the quote cycle runs on ONE document: convert → decide → bind → finalize. The
  invoice stays a draft; `Finalize` issues it. Idempotent on the decision.

### Changed — BREAKING
- The module path is `github.com/facturino/facturino-go/v2`.
- `InvoiceService.Create` requires `TaxDecisionID` + `DecisionLines`, checked
  locally before any HTTP call: an invoice never states its own VAT.
  `InvoiceParams.Items` is removed with the explicit-tax contract.
- `InvoiceUpdateParams` carries non-fiscal fields only — `Dates`, `Payment`,
  `Notes`, `PurchaseOrderNumber`, `Metadata`. `Items`, `Deposits` and
  `Schedule` are removed from updates: to change the operation, take a new
  decision and create a new draft.
- `CreditNoteParams.CreditedLines` is required and `CreditNoteParams.Items` is
  removed (also from `CreditNoteUpdateParams`): a credit note strictly
  inherits the VAT of the invoice it credits.
- `RecurringInvoiceParams.TaxInputs` is required and now carries its own
  `TaxSource`; `RecurringTemplateParams.Items` is removed. Each occurrence
  gets its OWN decision on its generation date.
- `TaxSource` values are `"facturino"` and `"integration"` — the two equal
  journeys, and the only two. A commercial draft that has not been decided yet
  reads `taxSource: null`, which unmarshals to the empty string in Go: it states
  the operation and no fiscal conclusion.
- The `Facturino-Version` header sent by the client is `2026-09-01`, the
  single stable contract; the API serves no earlier version.

### Notes
- `TaxDecisions.Create` requires a non-empty `IdempotencyKey` of at most
  `MaxIdempotencyKeyLength` (255) characters, checked locally so an over-long
  key fails immediately instead of after a round trip. A decision is never
  created through an anonymous retryable POST.
- `TaxDecisions.Create` returns the decision on both `201` (created) and `200`
  (the same key already produced it). Reusing a key with a different body
  answers `409`, surfaced as a `*facturino.Error` with `HTTPStatusCode == 409`.
- A tax decision is immutable: `TaxDecisionService` exposes no update, patch
  or delete. To re-decide the same operation after supplying missing evidence,
  create a new decision with `RetryOfTaxDecisionID`.

## [1.1.0] - 2026-08-05

### Added
- Document `paypal` as an accepted payment method (`PaymentParams.Method`). The
  field is a plain string, so the value already passes through; the API may add
  further payment method values over time — tolerate unknown values.

## [1.0.0] - 2026-07-25 — Initial release
