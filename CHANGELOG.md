# Changelog

All notable changes to the Facturino Go SDK are documented here. This project
adheres to [Semantic Versioning](https://semver.org/) and
[Keep a Changelog](https://keepachangelog.com/).

## [2.3.0] - 2026-09-04

### Added
- `Invoices.FinalizeWithPayment(id, *PaymentParams)` — an optional collection
  applied in the SAME transaction as the numbering, so an invoice already paid
  before issuance is ISSUED settled: the original PDF and Factur-X are rendered
  on a settled document. The parameter is the very `*PaymentParams` that
  `Payments.Create` takes (integer centimes), and the resulting payment is
  indistinguishable from one recorded afterwards. All or nothing: a collection
  beyond the amount due is refused (`422 payment_exceeds_amount_due`) and the
  invoice stays a draft. `Invoices.Finalize(id)` is unchanged and keeps sending
  no body — Go has no optional parameters, so the new behaviour is a new method
  rather than a signature change.
- `InvoiceDates.PaidAt` — the actual settlement date, empty until the invoice
  is fully paid.

### Documented
- `Job.Status` lists its values, including the TERMINAL `superseded`: the
  render was produced for a ledger revision that a collection, a cancellation
  or a refund has since overtaken. It is not a failure — call
  `Invoices.GetPDF(id)` again and a current render is published. Tolerate
  unknown values: more may be added.

## [2.2.0] - 2026-09-03

### Added
- `Error.Issues` (`[]ErrorIssue`) — the detailed reasons behind one refusal,
  each with a stable `Code`, the `Param` in cause and a `Message`. An EMPTY,
  non-nil slice when the API sent none, so reading it never needs a nil check.

### Changed
- `LocationEvidenceResult.TerritoryID` is empty when the API resolved the item
  at COUNTRY level only — a network kind (`ip_geolocation`, `bank_details`,
  `sim_mobile_country`, `fixed_line`) supplied without a postal code, published
  as JSON `null`. `DeclaredCountry` carries the country. The field keeps its
  `string` type, so no caller changes.

### Documented
- **`issues` is ADDITIVE.** The main `code` is unchanged and stays the value to
  branch on; `param` still points at the first field in cause. A refusal that has
  nothing more to say carries no `issues` key at all — an empty array would make
  "no detail available" indistinguishable from "detail computed, and empty".
- **The published code is the most PRECISE stable one.** Where the main code
  covers several distinct facts, the entry names the actual one: a buyer
  territory that would not resolve answers `validation_error` as the main code
  and `invalid_postal_code` or `territory_conflict` as the detail.
- **`param` is never a guess.** It is filled only where the mapping holds for
  every request — `invalid_postal_code` → `customer.address.postalCode`,
  `territory_conflict` → `customer.address.country`, a reason carried by a line →
  `lines[<index>]` — and is absent otherwise: an approximate pointer would send
  the caller to correct a field that was right.
- The reasons come from the tax decision engine, so they reach the refusals of
  `POST /v1/tax-decisions` and of every endpoint that takes a decision on the
  way — invoices, credit notes, quotes, recurring invoices. Messages carry no
  personal data.

## [2.1.0] - 2026-09-02

### Added
- `EuThresholdLedgers` service — the annual ledger the two EU B2C thresholds are
  assessed on: `Open`, `Get` / `Retrieve`, `ListEntries`, `Adjust`, `Correct`,
  `Review` and `ResolveReview` on `/v1/eu-threshold-ledgers`. The running totals
  are a LEDGER — per company, per mode and per calendar year — not figures on the
  fiscal profile. Types: `EuThresholdLedger`, `EuThresholdLedgerOpening`,
  `EuThresholdLedgerReview`, `EuThresholdLedgerEntry`,
  `EuThresholdLedgerEntryCorrection`, `EuThresholdLedgerEntryList`,
  `EuThresholdReservation`, `OpenEuThresholdLedgerParams`,
  `EuThresholdAdjustmentParams`, `EuThresholdCorrectionParams`,
  `EuThresholdReviewParams`, `EuThresholdReviewResolutionParams`,
  `EuThresholdEntryListParams`.

### Documented
- **Two counters, strictly apart and INDEPENDENT.** The same movements feed the
  common EUR 10,000 threshold (art. 59c(1) — intra-EU distance sales of goods AND
  cross-border services to consumers) and the EUR 100,000 location-evidence
  threshold (Reg. 282/2011 art. 24b — electronically supplied services, domestic
  ones included). A distance sale of goods raises the first and never the second;
  a domestic electronic service raises the second and never the first. Neither
  bounds the other, so every figure comes in a PAIR (`amount` /
  `evidenceAmount`, `acquiredMin` / `acquiredEvidenceMin`…) rather than as a
  total and a share of it.
- **Acquired and reserved are published apart, and never summed.** `acquiredMin`
  is what the year has certainly made; `reservedMin` is the slices held right now
  by operations still being decided. A held slice may still disappear, and one
  combined figure would hide exactly that.
- **An abandoned operation leaves nothing behind.** A verdict is frozen only when
  it holds at BOTH bounds of the running total — with every concurrent slice
  counted, and with none of them — so a released slice simply disappears
  (`reservation_released`, amount zero) instead of being kept as turnover that
  never existed. When the cap falls between the two bounds, nothing is frozen:
  `eu_threshold_concurrent_decision_pending`, and the same operation decides
  cleanly once the concurrent ones conclude.
- **A decision is never frozen without its slice.** If the slice has disappeared
  when the decision is about to be written, the transaction is abandoned — no
  decision, no audit entry, no settled claim (`409
  eu_threshold_reservation_lost`) — and the ledger goes under review on a path of
  its own, which is what makes the review survive that abandonment.
- **Giving an amount back is a QUALIFIED correction, never a negative
  adjustment.** Directive 2006/112/EC art. 90(1) reduces the taxable amount of a
  supply on cancellation, refusal or a price reduction after the supply, so a
  correction names the movement it corrects, its qualification, the resource it
  rests on and its evidence. A movement gives back what it brought in ONCE,
  whatever the number of corrections: the balance is kept inside the same
  transaction, and every entry publishes its `remainingMin`. `correctsEntryId` is
  restricted to the ids the ledger mints — it reaches a document path, and free
  text must not. `eu_threshold_correction_target_unknown`,
  `eu_threshold_correction_target_not_correctable`,
  `eu_threshold_correction_exceeds_counted`.
- **A ledger can stop deciding, and only a reconciliation reopens it.**
  `status: "review_required"` blocks every new reservation
  (`eu_threshold_review_required`) when the running total is known to be wrong.
  Settling it is a RECONCILIATION, never a comment: you state the version you
  checked and the two acquired totals you verified, and only an exact agreement
  reactivates the ledger (`eu_threshold_reconciliation_stale`,
  `eu_threshold_reconciliation_mismatch`). What was verified is written into the
  immutable `review_resolved` movement, with its evidence reference.
- **An ACTIVE Union one-stop-shop registration settles the place.** For a French
  seller, registering for the scheme is how the option of art. 59c(3) is
  exercised: the place is then the destination and the threshold has nothing left
  to decide (`basis: "oss_union_registration"`, `threshold: null`). The
  registration is DATED — one opened in October decides nothing for a September
  sale, one that ended decides nothing any more. Sourced for France only: the way
  the option is exercised is fixed by the member state where it is exercised.
- **The art. 24b single-evidence relaxation is COMPUTED, never declared.** It is
  read off the ledger's EUR 100,000 counter and published on the decision as
  `euB2cDestination.evidenceRelief`, with the totals it rested on.
  `undeterminable` is a first-class answer — two items of evidence are then
  required, and `location_evidence_relief_undetermined` says which fact is
  missing. The seller profile carries no relief flag: a figure is not a checkbox.
- **The seller profile carries DATED registrations.** `euB2cDestination` holds
  the option for taxation at destination, the one-stop-shop registrations
  (`ossRegistrations`) and the local VAT registrations, all with their periods.
  There are no undated `oss` booleans and no IOSS flag: a scheme without a period
  cannot say whether it covered the day of the operation, and no rule of this
  engine reads an import scheme.
- **Completeness dates are checked, never clamped**: a date must belong to the
  ledger's own year (`eu_threshold_date_outside_year`) and must never be in the
  future (`eu_threshold_date_in_future`). `opening.declaredAt` is the ISO instant
  the declaration was made, not the day it declares complete. The calendar year
  itself is bounded by one shared check (`eu_threshold_year_invalid`).
- **Movements are append-only and paginated with a cursor** (`limit`,
  `starting_after`); an adjustment is idempotent on your own reference, compared
  through a canonical fingerprint of the WHOLE body
  (`eu_threshold_entry_conflict`). The reference never becomes a document id: it
  is hashed and kept verbatim as data.
- A profile revision is an immutable RULE that decisions freeze; a turnover total
  moves with every sale and gets corrected — which is why the two are kept apart.
  A revision written earlier may still carry an old `threshold` snapshot, the
  undated `oss` booleans or a `locationEvidenceRelief` flag: all three are
  IGNORED, never read as a counter, a scheme or a figure. Such a revision stays
  valid for every operation the EU B2C rule does not reach; only a covered one
  asks for the facts it lacks.
- Opening a year declares four figures — the previous year's total and the total
  already made this year, each on its own counter — plus the COVERAGE MODE. `facturino_only` is a
  declaration that every covered sale goes through Facturino; `mixed_channels`
  says other channels exist, and the ledger then serves a decision only up to
  the day those channels are declared complete through
  (`eu_threshold_external_coverage_incomplete`). Sales made elsewhere are never
  assumed absent, and no year is ever assumed to start at zero
  (`eu_threshold_state_missing`).
- A decision RESERVES its slice of the running total in a transaction, then
  CONSUMES it in the same transaction that writes the decision: two concurrent
  operations never start from the same figure, a replay of the same
  `Idempotency-Key` counts nothing twice, and a decision that does not conclude
  gives its slice back. An operation dated before one already counted is refused
  (`eu_threshold_backdated_operation`) rather than silently recomputed.
- Adjustments are append-only and idempotent on the caller's own reference. They
  never lower a total: giving an amount back is a QUALIFIED correction naming the
  movement it corrects (Directive 2006/112/EC art. 90(1)), and what cannot be
  qualified that way puts the ledger under review rather than being subtracted.
- `TaxDecision.euB2cDestination.threshold` now freezes the ledger slice itself:
  which ledger, at which version, at which position in its total order, under
  which coverage mode, with the opening figures, the adjustments folded in, the
  total before, the operation's own value and the total after.
- The removed non-final codes (`destination_threshold_declaration_missing`,
  `…_stale`, `…_not_exhaustive`, `…_incomplete`) are replaced by the ledger's
  own: `eu_threshold_state_missing`,
  `eu_threshold_external_coverage_incomplete`, `eu_threshold_backdated_operation`.
- The `integration` source is not a channel for a non-French tax. `VatRate`,
  `VatCode` and `VatexCode` describe French VAT; the contract carries no
  local-tax jurisdiction, no local tax scheme and no withholding. Where a local
  tax of a French overseas collectivity or the TAAF can change what is invoiced
  or what is collected, the decision is not final under EITHER source
  (`com_taaf_local_tax_not_determined`, `com_taaf_local_regime_not_sourced`,
  `com_taaf_payment_withholding_not_modelled`,
  `seller_com_taaf_local_tax_not_determined`). A B2B service located in New
  Caledonia stays final under both sources: art. Lp. 507-1 makes the taxable
  customer account for the local tax itself.
- `PlaceOfSupply` is required on every `integration` line as soon as the buyer is
  established in one of those seven territories; omitting it answers
  `422 integration_vat_incoherent`. It was previously required only when the
  SELLER was established there.
- The exception for New Caledonia reaches only a supplier NOT established in New
  Caledonia: art. Lp. 507-1 puts the taxe générale sur la consommation on the
  taxable customer of a supplier established outside the territory. A seller
  established there is the ordinary collector of that tax on its own sales, and
  its decision is not final either.
- B2C sales to consumers of other member states are decided by FOUR separate
  questions: is the operation covered by a destination rule, does the common
  EUR 10,000 threshold still allow taxation at origin, did the seller opt for
  destination, and how is the tax declared. The one-stop shop is the fourth
  question only: registering never moves a place of taxation, and not registering
  never restores the seller's own national VAT — it leaves the decision without
  an amount (`destination_mechanism_missing`).
- The threshold figures live in the annual LEDGER, never on the fiscal profile
  (superseded above), because the same threshold counts the sales invoiced
  elsewhere. Non-final codes of the earlier design, all replaced by the ledger's:
  `destination_threshold_declaration_missing`,
  `destination_threshold_declaration_stale`,
  `destination_threshold_declaration_not_exhaustive`,
  `destination_threshold_operation_value_missing`,
  `destination_threshold_price_mode_ambiguous`,
  `destination_threshold_declaration_incomplete`,
  `destination_option_period_invalid`,
  `destination_option_scope_not_sourced`,
  `destination_establishment_in_member_state`,
  `franchise_destination_taxation_not_modelled`.
- A declared figure carries the day through which it is COMPLETE. A total closed
  on 1 March says nothing about the sales of 2 March: it can still prove the
  threshold is passed, never that it is not
  (`destination_threshold_declaration_incomplete`).
- Under a tax-inclusive price the threshold bounds the VAT-exclusive value with
  the only rates the operation can legally bear — the seller's own and the
  destination's — instead of the extremes of the whole Union.
- Destination rates come from a local, dated, versioned registry: no network call
  during a decision, and only the STANDARD rate of the 27 member states is
  tabulated (`destination_rate_band_not_available`,
  `destination_rate_not_sourced_for_date`). A region publishing its own standard
  rate never blocks a whole member state; what governs it decides the answer:
  where the rate follows the PLACE, the region is a territory of its own and the
  address decides (Portugal: mainland 23%, Madeira 22%, Azores 16%; an address
  placing no region answers `destination_regional_scope_undetermined`); where the
  rate is reserved to operations carried out in the zone by a supplier
  ESTABLISHED there, a supplier at distance is taxed at the national rate,
  finally — Austrian Jungholz and Kleinwalsertal (20%, not 19%) and the Greek
  island regime FOR SERVICES, where an electronically supplied service from
  France is taxed at 24%, in Athens and in Kalymnos alike, with no postal
  cartography needed to answer; and where the rate follows the DESTINATION over a
  zone this registry does not cartography, the operation is REFUSED with
  `destination_regional_regime_not_sourced` — this is Greece FOR GOODS, whose
  islands of fewer than 20,000 inhabitants apply a reduced standard rate to the
  goods delivered there since 2026-01-01, intra-community acquisitions included.
  An intra-EU distance sale of goods to Greece is therefore refused, in Athens as
  in Kalymnos, and never taxed at 24% by default. The answer is given per FAMILY
  of operation: the same member state can be settled for services and left open
  for goods.
- The `integration` source traverses the SAME destination rule as `facturino`:
  same coverage, threshold, option, evidence and declarative mechanism, and the
  SAME territorial frontier — a seller outside the French VAT territory or a
  buyer in a territory excluded from the EU VAT territory raises there exactly
  the obstacles it raises here. It compares the supplied rate to the legal result
  instead of producing it, at BOTH places the rule can settle: the destination
  standard rate, or the published bands of the seller's own territory. Equal
  means final; a category no covered B2C supply can carry (`AE`, `K`, `G`, `O`)
  or a contradicted `placeOfSupply` answers `422 integration_vat_incoherent`; a
  rate neither confirms leaves the decision non-final with
  `eu_b2c_rate_supplied_mismatch`. At origin that refusal asserts NO foreign tax:
  the operation is taxed in France. Obstacles keep their meaning —
  `pending_verification` when a fact is missing, `unsupported` when no rule
  covers the case — and `foreignTaxReviewRequired` is asserted only where a tax
  outside French VAT may really be due.

### Added
- `TaxDecision.EuB2cDestination` (`*TaxDecisionEuB2cDestination`): what the EU
  B2C destination rule concluded, frozen as DATA — verdict and basis, the
  threshold figures it was decided on, the declarative mechanism, and the rate
  entry with its registry version, its source, its verification date, its period
  and its region. A reference string cannot be audited years later; this can.
  `nil` on every operation the rule does not reach. New types:
  `TaxDecisionEuB2cDestination`, `TaxDecisionThresholdTrace`,
  `TaxDecisionDestinationOption`, `TaxDecisionDestinationMechanism`,
  `TaxDecisionDestinationRate`.

### Documented
- `TaxDecisionLineParams.GoodsMovement` is REQUIRED on a goods line as soon as
  the buyer is a consumer established in another member state, under BOTH
  sources: that movement decides whether the intra-EU distance sale rule applies
  (Directive 2006/112/EC art. 33(a)), and it is never assumed.

### Contract note

- The dated API contract stays `2026-09-01`. The `SellerTaxProfile` schema of that
  first publication carried two fields, `oss` and `locationEvidenceRelief`, which
  are replaced by `euB2cDestination`. `oss` carried a declarative mechanism alone,
  without the place option and the local registrations that decide alongside it;
  `locationEvidenceRelief` was a checkbox where the art. 24b relaxation is a
  FIGURE, computed by the engine on the EUR 100,000 counter of the annual ledger.
  The date is kept because the correction lands before any external consumer: no
  integration ever received the two removed fields, so nothing is broken and there
  is no dated version to keep alongside. A correction of this kind is admissible
  only before a consumer exists; afterwards only a new date would be.

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
- `TaxDecision.SettledObligations` (`*TaxDecisionSettledObligations`): the
  obligation axes French law settles even when the decision is NOT final. The
  three top-level axes stay `nil` on such a decision — an axis is never read off
  a decision that did not conclude — while an obligation that stands is carried
  as a value rather than lost with the amount. `nil` on a final decision, where
  the top-level axes are the settled ones. It authorises nothing: a non-final
  decision is not invoiceable and opens no payment.
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
