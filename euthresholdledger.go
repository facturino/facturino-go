package facturino

import (
	"fmt"
	"net/url"
	"strconv"
)

// EuThresholdLedger is the annual ledger the two EU B2C thresholds are assessed
// on.
//
// It is deliberately separate from the seller's fiscal profile: a profile
// revision is an immutable RULE that decisions freeze, while a turnover total
// moves with every sale and gets corrected. Keeping them apart is what lets a
// correction be recorded without rewriting the rule a frozen decision was taken
// under.
//
// It carries TWO counters, strictly apart and INDEPENDENT: the common
// EUR 10,000 threshold (art. 59c(1) — intra-EU distance sales of goods AND
// cross-border services to consumers) and the EUR 100,000 location-evidence
// threshold (Reg. 282/2011 art. 24b, 2nd subparagraph — electronically supplied
// services, DOMESTIC ones included). A distance sale of goods raises the first
// and never the second; and neither bounds the other, because the common
// threshold counts only cross-border supplies. A seller whose electronic
// services are mostly sold at home legitimately declares more on the second.
//
// Nothing is assumed. Opening a year declares four figures — the two totals and
// their services part — plus whether every covered sale goes through Facturino.
// Sales made elsewhere are never assumed absent: under "mixed_channels" they
// enter through adjustments, and the ledger serves a decision only up to the day
// those channels are declared complete through.
type EuThresholdLedger struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	CompanyID string `json:"companyId"`
	Livemode  bool   `json:"livemode"`
	Year      string `json:"year"`
	// Version is bumped on every write and frozen on the decisions that used it.
	Version int `json:"version"`
	// Status is "open" or "review_required". Under review the ledger refuses
	// every new reservation: its running total is known to be wrong, and a
	// decision is frozen on the figures it reads.
	Status string                   `json:"status"`
	Review *EuThresholdLedgerReview `json:"review"`
	// CapCents is the cap of art. 59c(1); EvidenceCapCents that of art. 24b.
	CapCents         int                       `json:"capCents"`
	EvidenceCapCents int                       `json:"evidenceCapCents"`
	Opening          *EuThresholdLedgerOpening `json:"opening"`
	// ExternalCompleteThroughDate is the day the channels other than Facturino
	// are declared complete through; adjustments move it forward.
	ExternalCompleteThroughDate string `json:"externalCompleteThroughDate"`
	AdjustmentTotal             int    `json:"adjustmentTotal"`
	AdjustmentEvidenceTotal     int    `json:"adjustmentEvidenceTotal"`
	AdjustmentCount             int    `json:"adjustmentCount"`
	// CorrectionTotal is the POSITIVE total of the qualified corrections, which
	// is subtracted from the running total.
	CorrectionTotal         int `json:"correctionTotal"`
	CorrectionEvidenceTotal int `json:"correctionEvidenceTotal"`
	CorrectionCount         int `json:"correctionCount"`
	// AcquiredMin is the total already ACQUIRED: opening + adjustments −
	// corrections + final decisions. AcquiredMax is its upper bound under a
	// tax-inclusive price.
	AcquiredMin         int `json:"acquiredMin"`
	AcquiredMax         int `json:"acquiredMax"`
	AcquiredEvidenceMin int `json:"acquiredEvidenceMin"`
	AcquiredEvidenceMax int `json:"acquiredEvidenceMax"`
	// ReservedMin/Max are the slices held right now by operations being
	// decided. They are NOT acquired and never summed with the figures above:
	// any of them may still disappear.
	ReservedMin         int `json:"reservedMin"`
	ReservedMax         int `json:"reservedMax"`
	ReservedEvidenceMin int `json:"reservedEvidenceMin"`
	ReservedEvidenceMax int `json:"reservedEvidenceMax"`
	// RemainingMin is what is left before each cap, on the figures acquired.
	RemainingMin            int     `json:"remainingMin"`
	EvidenceRemainingMin    int     `json:"evidenceRemainingMin"`
	SettledCount            int     `json:"settledCount"`
	LastConsumedEffectiveAt *string `json:"lastConsumedEffectiveAt"`
	// Reservations are the slices held by decisions in flight.
	Reservations []*EuThresholdReservation `json:"reservations"`
	// Entries is the FIRST page of movements, newest first.
	Entries           []*EuThresholdLedgerEntry `json:"entries"`
	EntriesHasMore    bool                      `json:"entriesHasMore"`
	EntriesNextCursor *string                   `json:"entriesNextCursor"`
	Created           string                    `json:"created"`
	Updated           string                    `json:"updated"`
}

// EuThresholdLedgerReview says why the ledger stopped serving decisions.
type EuThresholdLedgerReview struct {
	// Code is "consumed_slice_missing", "correction_not_qualifiable" or
	// "declared_by_administrator".
	Code     string `json:"code"`
	Detail   string `json:"detail"`
	OpenedAt string `json:"openedAt"`
}

// EuThresholdLedgerOpening is what the seller declared when the year was opened.
type EuThresholdLedgerOpening struct {
	PreviousYearAmount int `json:"previousYearAmount"`
	CurrentYearOpening int `json:"currentYearOpening"`
	// The same two figures restricted to the art. 24b services perimeter.
	PreviousYearEvidenceAmount int `json:"previousYearEvidenceAmount"`
	CurrentYearEvidenceOpening int `json:"currentYearEvidenceOpening"`
	// CoverageMode is "facturino_only" or "mixed_channels".
	CoverageMode                string `json:"coverageMode"`
	ExternalCompleteThroughDate string `json:"externalCompleteThroughDate"`
	// DeclaredAt is the ISO instant the opening was actually declared.
	DeclaredAt string `json:"declaredAt"`
}

// EuThresholdLedgerEntryCorrection qualifies a correction and names what it
// rests on.
type EuThresholdLedgerEntryCorrection struct {
	// Kind is "credit_note", "cancellation" or "refund".
	Kind                string `json:"kind"`
	CorrectsEntryID     string `json:"correctsEntryId"`
	RelatedResourceType string `json:"relatedResourceType"`
	RelatedResourceID   string `json:"relatedResourceId"`
	EvidenceReference   string `json:"evidenceReference"`
}

// EuThresholdLedgerEntry is one movement. Written once, never rewritten.
type EuThresholdLedgerEntry struct {
	// ID is derived from the movement. Your own Reference is DATA, never an
	// identifier: free text must not reach a document path.
	ID       string `json:"id"`
	Sequence int    `json:"sequence"`
	// Kind is "opening", "external_adjustment", "external_correction",
	// "reservation_consumed", "reservation_released", "review_opened" or
	// "review_resolved".
	Kind string `json:"kind"`
	// AmountMin/Max are the SIGNED effect on the running total: negative for a
	// qualified correction, zero for a released slice or a review movement.
	AmountMin             int     `json:"amountMin"`
	AmountMax             int     `json:"amountMax"`
	EvidenceAmountMin     int     `json:"evidenceAmountMin"`
	EvidenceAmountMax     int     `json:"evidenceAmountMax"`
	CumulativeMin         int     `json:"cumulativeMin"`
	CumulativeMax         int     `json:"cumulativeMax"`
	EvidenceCumulativeMin int     `json:"evidenceCumulativeMin"`
	EvidenceCumulativeMax int     `json:"evidenceCumulativeMax"`
	TaxDecisionID         *string `json:"taxDecisionId"`
	EffectiveAt           *string `json:"effectiveAt"`
	// Reference is your own identifier, kept verbatim.
	Reference  *string                           `json:"reference"`
	Correction *EuThresholdLedgerEntryCorrection `json:"correction"`
	Reason     string                            `json:"reason"`
	RecordedAt string                            `json:"recordedAt"`
	// Correctable says whether this movement brought turnover in, and therefore
	// has anything to give back.
	Correctable bool `json:"correctable"`
	// CorrectedMin is what it has ALREADY given back, over every correction that
	// named it. RemainingMin is what is LEFT: a movement gives back what it
	// brought in ONCE, whatever the number of corrections, and the balance is
	// kept inside the transaction.
	CorrectedMin         int `json:"correctedMin"`
	CorrectedEvidenceMin int `json:"correctedEvidenceMin"`
	CorrectionCount      int `json:"correctionCount"`
	RemainingMin         int `json:"remainingMin"`
	RemainingEvidenceMin int `json:"remainingEvidenceMin"`
}

// EuThresholdLedgerEntryList is one page of movements, newest first.
type EuThresholdLedgerEntryList struct {
	Object     string                    `json:"object"`
	URL        string                    `json:"url"`
	Data       []*EuThresholdLedgerEntry `json:"data"`
	HasMore    bool                      `json:"has_more"`
	NextCursor *string                   `json:"next_cursor"`
}

// EuThresholdReservation is a slice held by a decision in flight. It is NOT
// acquired: it may still be given back.
type EuThresholdReservation struct {
	ID                      string `json:"id"`
	Sequence                int    `json:"sequence"`
	EffectiveAt             string `json:"effectiveAt"`
	ContributionMin         int    `json:"contributionMin"`
	ContributionMax         int    `json:"contributionMax"`
	EvidenceContributionMin int    `json:"evidenceContributionMin"`
	EvidenceContributionMax int    `json:"evidenceContributionMax"`
	ReservedAt              string `json:"reservedAt"`
	ExpiresAt               string `json:"expiresAt"`
}

// OpenEuThresholdLedgerParams opens a calendar year.
type OpenEuThresholdLedgerParams struct {
	Year string `json:"year"`
	// PreviousYearAmount is the covered turnover of the PREVIOUS calendar year,
	// VAT excluded, in integer centimes.
	PreviousYearAmount int `json:"previousYearAmount"`
	// CurrentYearOpening is what has ALREADY been made this year, VAT excluded.
	CurrentYearOpening int `json:"currentYearOpening"`
	// PreviousYearEvidenceAmount is the art. 24b perimeter for the same period:
	// every service supplied electronically to a consumer of the Union, DOMESTIC
	// ones included. INDEPENDENT of PreviousYearAmount in both directions — the
	// common threshold counts only cross-border supplies — so a seller selling
	// mostly at home legitimately declares more here than there.
	PreviousYearEvidenceAmount int `json:"previousYearEvidenceAmount"`
	// CurrentYearEvidenceOpening is the matching figure for the current year.
	CurrentYearEvidenceOpening int `json:"currentYearEvidenceOpening"`
	// CoverageMode is "facturino_only" or "mixed_channels". It is a declaration,
	// never an assumption: sales made elsewhere are never presumed absent.
	CoverageMode string `json:"coverageMode"`
	// ExternalCompleteThroughDate must belong to the ledger's own year and must
	// never be in the future.
	ExternalCompleteThroughDate string `json:"externalCompleteThroughDate"`
}

// EuThresholdAdjustmentParams records turnover made on another channel.
type EuThresholdAdjustmentParams struct {
	// Reference is YOUR identifier: it is the entry's identity, so replaying the
	// SAME body adds nothing and reusing it for a different one answers
	// eu_threshold_entry_conflict. It never becomes a document id.
	Reference string `json:"reference"`
	// Amount is VAT-excluded, in integer centimes, and NEVER negative.
	Amount int `json:"amount"`
	// EvidenceAmount is the art. 24b part of the same movement. INDEPENDENT of
	// Amount: a domestic electronic service raises this counter and not the other.
	EvidenceAmount              int    `json:"evidenceAmount"`
	ExternalCompleteThroughDate string `json:"externalCompleteThroughDate"`
	Reason                      string `json:"reason"`
}

// EuThresholdCorrectionParams takes a qualified amount back out of the total.
//
// This is NOT a negative adjustment. Directive 2006/112/EC art. 90(1) reduces
// the taxable amount of a supply on cancellation, refusal or a price reduction
// after the supply, and the thresholds count the VALUE of the supplies — so a
// correction NAMES the movement it corrects, its qualification, the resource it
// rests on and its evidence, and never gives back more than that movement
// brought in.
type EuThresholdCorrectionParams struct {
	Reference string `json:"reference"`
	// CorrectsEntryID is a movement of THIS ledger whose base is reduced. It is
	// constrained to the ids the ledger mints ("opening", or adj_/cor_/con_/rel_/
	// rev_ followed by 32 hex characters): it reaches a document path, and free
	// text must not. A movement that brought no turnover in answers
	// eu_threshold_correction_target_not_correctable.
	CorrectsEntryID string `json:"correctsEntryId"`
	// Kind is "credit_note", "cancellation" or "refund".
	Kind string `json:"kind"`
	// Amount is given back, VAT excluded, in cents. The ledger keeps the BALANCE
	// of each movement inside the transaction: the corrections of one movement
	// never add up to more than it brought in.
	Amount int `json:"amount"`
	// EvidenceAmount is its services part; independent of Amount.
	EvidenceAmount      int    `json:"evidenceAmount"`
	RelatedResourceType string `json:"relatedResourceType"`
	RelatedResourceID   string `json:"relatedResourceId"`
	EvidenceReference   string `json:"evidenceReference"`
	Reason              string `json:"reason"`
}

// EuThresholdReviewParams opens a review: a declaration, so a reason is enough.
type EuThresholdReviewParams struct {
	Reason string `json:"reason"`
}

// EuThresholdReviewResolutionParams settles a review — by RECONCILIATION, never
// by comment.
//
// A review says the running total is known to be wrong. Reopening the ledger on
// a free-text note would put that same total back in front of the next verdict
// with a sentence for only guarantee. So you state the figures you actually
// verified, and the server compares them to its own: ReconciledVersion pins the
// state that was checked (a movement recorded since answers
// eu_threshold_reconciliation_stale), and the two acquired totals must MATCH
// (eu_threshold_reconciliation_mismatch, which returns both figures).
//
// What was verified is written into the immutable "review_resolved" movement,
// with its evidence reference.
type EuThresholdReviewResolutionParams struct {
	ReconciledVersion             int    `json:"reconciledVersion"`
	ReconciledAcquiredMin         int    `json:"reconciledAcquiredMin"`
	ReconciledAcquiredEvidenceMin int    `json:"reconciledAcquiredEvidenceMin"`
	EvidenceReference             string `json:"evidenceReference"`
	Reason                        string `json:"reason"`
}

// EuThresholdEntryListParams paginates the movements with a cursor.
type EuThresholdEntryListParams struct {
	Limit         int
	StartingAfter string
}

func (p *EuThresholdEntryListParams) values() url.Values {
	values := url.Values{}
	if p == nil {
		return values
	}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.StartingAfter != "" {
		values.Set("starting_after", p.StartingAfter)
	}
	return values
}

// EuThresholdLedgerService reads and maintains the annual ledger.
type EuThresholdLedgerService struct {
	client *httpClient
}

// Open declares a calendar year. A year already open is never rewritten
// (eu_threshold_year_already_open): decisions were frozen on its figures.
func (s *EuThresholdLedgerService) Open(params *OpenEuThresholdLedgerParams) (*EuThresholdLedger, error) {
	if params == nil {
		return nil, fmt.Errorf("facturino: opening params are required")
	}
	var ledger EuThresholdLedger
	if err := s.client.post("/eu-threshold-ledgers", params, &ledger, &requestOption{}); err != nil {
		return nil, err
	}
	return &ledger, nil
}

// Get reads the ledger of a year: acquired totals, held slices, what remains,
// and the first page of movements.
func (s *EuThresholdLedgerService) Get(
	year string,
	params *EuThresholdEntryListParams,
) (*EuThresholdLedger, error) {
	var ledger EuThresholdLedger
	path := fmt.Sprintf("/eu-threshold-ledgers/%s", year)
	if err := s.client.get(path, params.values(), &ledger); err != nil {
		return nil, err
	}
	return &ledger, nil
}

// Retrieve is an alias of Get, matching the other SDKs.
func (s *EuThresholdLedgerService) Retrieve(year string) (*EuThresholdLedger, error) {
	return s.Get(year, nil)
}

// ListEntries walks the movements page by page, newest first. The ledger keeps
// every movement; a page shows some, and NextCursor names the last one returned.
func (s *EuThresholdLedgerService) ListEntries(
	year string,
	params *EuThresholdEntryListParams,
) (*EuThresholdLedgerEntryList, error) {
	var page EuThresholdLedgerEntryList
	path := fmt.Sprintf("/eu-threshold-ledgers/%s/entries", year)
	if err := s.client.get(path, params.values(), &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// Adjust records turnover made on another channel.
//
// Append-only and idempotent on Reference, compared through a canonical
// fingerprint of the WHOLE body: replaying the same adjustment adds nothing, and
// reusing the reference for any different fact answers
// eu_threshold_entry_conflict. A negative amount is impossible here — giving an
// amount back is Correct.
func (s *EuThresholdLedgerService) Adjust(
	year string,
	params *EuThresholdAdjustmentParams,
) (*EuThresholdLedger, error) {
	if params == nil {
		return nil, fmt.Errorf("facturino: adjustment params are required")
	}
	var ledger EuThresholdLedger
	path := fmt.Sprintf("/eu-threshold-ledgers/%s/adjustments", year)
	if err := s.client.post(path, params, &ledger, &requestOption{}); err != nil {
		return nil, err
	}
	return &ledger, nil
}

// Correct takes a qualified amount back out of the running total.
//
// The correction names the movement it corrects (CorrectsEntryID), its
// qualification, the resource it rests on and its evidence. An unknown movement
// answers eu_threshold_correction_target_unknown, one that brought no turnover
// in eu_threshold_correction_target_not_correctable, and a correction beyond the
// movement's REMAINING balance eu_threshold_correction_exceeds_counted — the
// balance is kept inside the transaction, so two corrections of the whole amount
// can never both pass. Each entry publishes its RemainingMin.
//
// Decisions already frozen are never rewritten: they were correct on the figures
// of their own day. Only the total the NEXT operations read changes.
func (s *EuThresholdLedgerService) Correct(
	year string,
	params *EuThresholdCorrectionParams,
) (*EuThresholdLedger, error) {
	if params == nil {
		return nil, fmt.Errorf("facturino: correction params are required")
	}
	var ledger EuThresholdLedger
	path := fmt.Sprintf("/eu-threshold-ledgers/%s/corrections", year)
	if err := s.client.post(path, params, &ledger, &requestOption{}); err != nil {
		return nil, err
	}
	return &ledger, nil
}

// Review stops deciding on this ledger: its running total is known to be wrong.
// Every reservation then answers eu_threshold_review_required. This is the
// honest exit when an amount must come out and no qualified correction can name
// the movement it corrects.
func (s *EuThresholdLedgerService) Review(
	year string,
	params *EuThresholdReviewParams,
) (*EuThresholdLedger, error) {
	if params == nil {
		return nil, fmt.Errorf("facturino: review params are required")
	}
	var ledger EuThresholdLedger
	path := fmt.Sprintf("/eu-threshold-ledgers/%s/review", year)
	if err := s.client.post(path, params, &ledger, &requestOption{}); err != nil {
		return nil, err
	}
	return &ledger, nil
}

// ResolveReview serves decisions again — on a RECONCILIATION that matches,
// never on a comment. See EuThresholdReviewResolutionParams.
func (s *EuThresholdLedgerService) ResolveReview(
	year string,
	params *EuThresholdReviewResolutionParams,
) (*EuThresholdLedger, error) {
	if params == nil {
		return nil, fmt.Errorf("facturino: review params are required")
	}
	var ledger EuThresholdLedger
	path := fmt.Sprintf("/eu-threshold-ledgers/%s/review/resolve", year)
	if err := s.client.post(path, params, &ledger, &requestOption{}); err != nil {
		return nil, err
	}
	return &ledger, nil
}
