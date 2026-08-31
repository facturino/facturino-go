package facturino

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

const finalDecisionJSON = `{
  "id": "taxdec_9c1f",
  "object": "tax_decision",
  "status": "final",
  "customerId": "cus_8f2k4m9n",
  "currency": "eur",
  "priceMode": "tax_exclusive",
  "effectiveAt": "2026-09-15",
  "expired": false,
  "rulesVersion": "fr-vat-2021-07-01",
  "operationFingerprint": "sha256:op",
  "totals": {"totalHT": 2900, "totalVAT": 580, "totalTTC": 3480},
  "amountToCharge": 3480,
  "invoiceChannel": "einvoicing",
  "transactionReporting": "none",
  "paymentReporting": "fr212",
  "foreignTaxReviewRequired": false,
  "vies": null,
  "issues": [],
  "obligationReasons": [{"axis": "paymentReporting", "code": "fr212_b2b", "reference": "flux 212"}]
}`

func decisionParams() *TaxDecisionParams {
	return &TaxDecisionParams{
		TaxSource:   "facturino",
		Customer:    "cus_8f2k4m9n",
		EffectiveAt: "2026-09-15",
		Currency:    "eur",
		PriceMode:   "tax_exclusive",
		Lines: []*TaxDecisionLineParams{{
			Reference:    "abo-pro",
			Description:  "Abonnement Pro",
			Category:     "electronically_supplied_services",
			RateCategory: "standard",
			UnitAmount:   2900,
			Quantity:     "1",
		}},
		IdempotencyKey: "order-default",
	}
}

func TestTaxDecisionCreateRequiresIdempotencyKey(t *testing.T) {
	client := New("fac_test_xxx")
	params := decisionParams()
	params.IdempotencyKey = ""
	if _, err := client.TaxDecisions.Create(params); err == nil || !strings.Contains(err.Error(), "Idempotency-Key") {
		t.Fatalf("Create error = %v, want missing Idempotency-Key", err)
	}
}

func TestTaxDecisionCreateRefusesAnOverLongIdempotencyKey(t *testing.T) {
	// Checked locally: an over-long key fails without a round trip.
	client := New("fac_test_xxx")
	params := decisionParams()
	params.IdempotencyKey = strings.Repeat("k", MaxIdempotencyKeyLength+1)
	if _, err := client.TaxDecisions.Create(params); err == nil ||
		!strings.Contains(err.Error(), "at most 255 characters") {
		t.Fatalf("Create error = %v, want an over-long Idempotency-Key refusal", err)
	}
}

func TestTaxDecisionCreateAcceptsAKeyOfExactlyTheMaximumLength(t *testing.T) {
	var seen string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, finalDecisionJSON)
	})
	params := decisionParams()
	params.IdempotencyKey = strings.Repeat("k", MaxIdempotencyKeyLength)
	if _, err := client.TaxDecisions.Create(params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(seen) != MaxIdempotencyKeyLength {
		t.Errorf("Idempotency-Key length = %d, want %d", len(seen), MaxIdempotencyKeyLength)
	}
}

func TestTaxDecisionCreateSendsTheCurrentDatedContractVersion(t *testing.T) {
	var version string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		version = r.Header.Get("Facturino-Version")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, finalDecisionJSON)
	})
	if _, err := client.TaxDecisions.Create(decisionParams()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "2026-09-01" {
		t.Errorf("Facturino-Version = %q, want 2026-09-01", version)
	}
}

func TestTaxDecisionCreateReturnsTheDecisionOnA200Replay(t *testing.T) {
	// 201 on creation, 200 when the same key already produced the decision.
	// Both carry the decision itself: the caller reads one shape either way.
	for _, status := range []int{201, 200} {
		client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			fmt.Fprint(w, finalDecisionJSON)
		})
		d, err := client.TaxDecisions.Create(decisionParams())
		if err != nil {
			t.Fatalf("status %d: unexpected error: %v", status, err)
		}
		if d.AmountToCharge == nil || *d.AmountToCharge != 3480 {
			t.Errorf("status %d: AmountToCharge = %v, want 3480", status, d.AmountToCharge)
		}
	}
}

func TestTaxDecisionCreateSurfacesA409AsAConflict(t *testing.T) {
	calls := 0
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(409)
		fmt.Fprint(w, `{"error":{"type":"invalid_request_error","code":"conflict",`+
			`"message":"This Idempotency-Key was already used with a different request body."}}`)
	})

	_, err := client.TaxDecisions.Create(decisionParams())
	if err == nil {
		t.Fatal("Create error = nil, want a conflict")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T, want *facturino.Error", err)
	}
	if apiErr.HTTPStatusCode != 409 {
		t.Errorf("HTTPStatusCode = %d, want 409", apiErr.HTTPStatusCode)
	}
	if apiErr.Code != "conflict" {
		t.Errorf("Code = %q, want conflict", apiErr.Code)
	}
	// A conflict is never retried: the key identifies one request.
	if calls != 1 {
		t.Errorf("server calls = %d, want 1", calls)
	}
}

func TestTaxDecisionsExposedOnClient(t *testing.T) {
	client := New("fac_test_xxx")
	if client.TaxDecisions == nil {
		t.Fatal("TaxDecisions service is nil")
	}
}

func TestTaxDecisionCreate(t *testing.T) {
	var body map[string]interface{}
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/v1/tax-decisions" {
			t.Errorf("Path = %q, want /v1/tax-decisions", r.URL.Path)
		}
		if got := r.Header.Get("Idempotency-Key"); got != "order-4711" {
			t.Errorf("Idempotency-Key = %q, want order-4711", got)
		}
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("body is not JSON: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, finalDecisionJSON)
	})

	params := decisionParams()
	params.IdempotencyKey = "order-4711"
	d, err := client.TaxDecisions.Create(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.AmountToCharge == nil || *d.AmountToCharge != 3480 {
		t.Errorf("AmountToCharge = %v, want 3480", d.AmountToCharge)
	}
	if !d.IsFinal() {
		t.Error("IsFinal() = false, want true")
	}
	if d.InvoiceChannel == nil || *d.InvoiceChannel != "einvoicing" {
		t.Errorf("InvoiceChannel = %v, want einvoicing", d.InvoiceChannel)
	}

	// The idempotency key travels in the header, never in the body.
	if _, present := body["idempotencyKey"]; present {
		t.Error("idempotencyKey leaked into the request body")
	}
	lines, _ := body["lines"].([]interface{})
	if len(lines) != 1 {
		t.Fatalf("lines length = %d, want 1", len(lines))
	}
	line := lines[0].(map[string]interface{})
	// A quantity is a decimal STRING: a float would not survive the round trip.
	if quantity, ok := line["quantity"].(string); !ok || quantity != "1" {
		t.Errorf("quantity = %v, want the string \"1\"", line["quantity"])
	}
	// Amounts are integer centimes.
	if amount, ok := line["unitAmount"].(float64); !ok || amount != 2900 {
		t.Errorf("unitAmount = %v, want 2900", line["unitAmount"])
	}
	if body["priceMode"] != "tax_exclusive" {
		t.Errorf("priceMode = %v, want tax_exclusive", body["priceMode"])
	}
}

func TestTaxDecisionCreateKeepsIdempotencyKeyAcrossRetries(t *testing.T) {
	var keys []string
	client, _ := testServerWithRetries(t, func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		if len(keys) == 1 {
			w.WriteHeader(503)
			fmt.Fprint(w, `{"error":{"message":"unavailable"}}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, finalDecisionJSON)
	}, 2)

	params := decisionParams()
	params.IdempotencyKey = "order-4711"
	if _, err := client.TaxDecisions.Create(params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(keys) < 2 {
		t.Fatalf("expected a retry, got %d attempt(s)", len(keys))
	}
	// A retry that changed the key would take a SECOND decision.
	for i, key := range keys {
		if key != "order-4711" {
			t.Errorf("attempt %d sent Idempotency-Key %q, want order-4711", i, key)
		}
	}
}

func TestTaxDecisionCreateWithEvidence(t *testing.T) {
	var raw string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, finalDecisionJSON)
	})

	params := decisionParams()
	params.LocationEvidence = []*LocationEvidenceParams{{
		Kind:        "ip_geolocation",
		Country:     "FR",
		PostalCode:  "75002",
		ThirdParty:  true,
		Source:      "psp",
		CollectedAt: "2026-09-15",
		Reference:   "ch_3Kj9aLZ",
	}}
	params.NonEuBusinessEvidence = &NonEuBusinessEvidenceParams{
		Kind:                            "vat_or_similar_number",
		Reference:                       "CH-123456",
		IssuedByCountry:                 "CH",
		ReasonableVerificationPerformed: true,
		CollectedAt:                     "2026-09-15",
	}
	if _, err := client.TaxDecisions.Create(params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{`"thirdParty":true`, `"reasonableVerificationPerformed":true`, `"locationEvidence"`} {
		if !strings.Contains(raw, want) {
			t.Errorf("body missing %s: %s", want, raw)
		}
	}
	// The territorial signal travels, never the raw one.
	for _, forbidden := range []string{`"ip"`, `"payload"`, `"iban"`} {
		if strings.Contains(raw, forbidden) {
			t.Errorf("body leaked a raw signal field %s", forbidden)
		}
	}
}

func TestTaxDecisionGet(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/v1/tax-decisions/taxdec_9c1f" {
			t.Errorf("Path = %q, want /v1/tax-decisions/taxdec_9c1f", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, finalDecisionJSON)
	})

	d, err := client.TaxDecisions.Get("taxdec_9c1f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID != "taxdec_9c1f" {
		t.Errorf("ID = %q, want taxdec_9c1f", d.ID)
	}
	if len(d.ObligationReasons) != 1 || d.ObligationReasons[0].Axis != "paymentReporting" {
		t.Errorf("ObligationReasons = %+v", d.ObligationReasons)
	}
}

func TestTaxDecisionRetrieveIsAliasOfGet(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tax-decisions/taxdec_9c1f" {
			t.Errorf("Path = %q, want /v1/tax-decisions/taxdec_9c1f", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, finalDecisionJSON)
	})

	if _, err := client.TaxDecisions.Retrieve("taxdec_9c1f"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaxDecisionNonFinalHasNilAmounts(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{
		  "id": "taxdec_pending", "object": "tax_decision",
		  "status": "pending_verification",
		  "totals": null, "amountToCharge": null,
		  "invoiceChannel": null, "transactionReporting": null, "paymentReporting": null,
		  "issues": [{"code": "vies_unavailable", "message": "VIES is unreachable."}]
		}`)
	})

	d, err := client.TaxDecisions.Create(decisionParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.IsFinal() {
		t.Error("IsFinal() = true on a pending decision")
	}
	// nil, never 0: absent is not "nothing to charge".
	if d.AmountToCharge != nil {
		t.Errorf("AmountToCharge = %v, want nil", *d.AmountToCharge)
	}
	if d.Totals != nil {
		t.Errorf("Totals = %+v, want nil", d.Totals)
	}
	if len(d.Issues) != 1 || d.Issues[0].Code != "vies_unavailable" {
		t.Errorf("Issues = %+v", d.Issues)
	}
}

func TestTaxDecisionRetryCarriesLineage(t *testing.T) {
	var body map[string]interface{}
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"taxdec_new","object":"tax_decision","status":"final","retryOfTaxDecisionId":"taxdec_previous"}`)
	})

	params := decisionParams()
	params.RetryOfTaxDecisionID = "taxdec_previous"
	params.LocationEvidence = []*LocationEvidenceParams{{
		Kind: "billing_address", Country: "FR", ThirdParty: false,
		Source: "declared", CollectedAt: "2026-09-15",
	}}
	d, err := client.TaxDecisions.Create(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body["retryOfTaxDecisionId"] != "taxdec_previous" {
		t.Errorf("retryOfTaxDecisionId = %v", body["retryOfTaxDecisionId"])
	}
	if d.RetryOfTaxDecisionID != "taxdec_previous" {
		t.Errorf("RetryOfTaxDecisionID = %q", d.RetryOfTaxDecisionID)
	}
}

func TestTaxDecisionServiceHasNoMutation(t *testing.T) {
	service := reflect.TypeOf(&TaxDecisionService{})
	for _, forbidden := range []string{"Update", "Patch", "Delete", "Del", "Cancel"} {
		if _, found := service.MethodByName(forbidden); found {
			t.Errorf("TaxDecisionService exposes %s: a decision is immutable", forbidden)
		}
	}
}

func TestInvoiceBackedByDecision(t *testing.T) {
	var body map[string]interface{}
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{
		  "id": "inv_1", "object": "invoice", "status": "draft",
		  "documentStatus": "draft", "transmissionStatus": "not_applicable",
		  "paymentStatus": "unpaid", "taxSource": "facturino",
		  "taxDecisionId": "taxdec_9c1f"
		}`)
	})

	inv, err := client.Invoices.Create(&InvoiceParams{
		Customer:      "cus_8f2k4m9n",
		TaxDecisionID: "taxdec_9c1f",
		DecisionLines: []*DecisionLineParams{{TaxLineRef: "abo-pro", Unit: "month"}},
		Buyer:         &BuyerParams{CompanyName: "ACME SAS"},
		Dates:         &InvoiceDatesParams{Issued: "2026-09-15", Due: "2026-10-15"},
		Payment:       &PaymentTermsParams{Terms: "30 jours", TermsDays: 30, Method: "transfer"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body["taxDecisionId"] != "taxdec_9c1f" {
		t.Errorf("taxDecisionId = %v", body["taxDecisionId"])
	}
	// The historical `lines` field must not travel alongside a decision.
	if _, present := body["lines"]; present {
		t.Error("lines was sent together with taxDecisionId")
	}
	if inv.TaxSource != "facturino" || inv.DocumentStatus != "draft" || inv.PaymentStatus != "unpaid" {
		t.Errorf("axes = %q/%q/%q", inv.TaxSource, inv.DocumentStatus, inv.PaymentStatus)
	}
}

func TestDepositsAndScheduleTravelWithTheDecision(t *testing.T) {
	var body map[string]interface{}
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"inv_2","object":"invoice","status":"draft"}`)
	})

	_, err := client.Invoices.Create(&InvoiceParams{
		Customer:      "cus_8f2k4m9n",
		TaxDecisionID: "taxdec_9c1f",
		DecisionLines: []*DecisionLineParams{{TaxLineRef: "abo-pro", Unit: "month"}},
		Buyer:         &BuyerParams{CompanyName: "ACME SAS"},
		Dates:         &InvoiceDatesParams{Issued: "2026-09-15", Due: "2026-10-15"},
		Payment:       &PaymentTermsParams{Terms: "30 jours", TermsDays: 30, Method: "transfer"},
		Deposits:      []*DepositParam{{InvoiceID: "inv_dep"}},
		Schedule: []*ScheduleParam{
			{Amount: 1740, DueDate: "2026-10-15"},
			{Amount: 1740, DueDate: "2026-11-15"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both settle SERVER-SIDE against the decided amount; the SDK only
	// forwards them.
	deposits, ok := body["deposits"].([]interface{})
	if !ok || len(deposits) != 1 {
		t.Errorf("deposits = %v, want one entry", body["deposits"])
	}
	schedule, ok := body["schedule"].([]interface{})
	if !ok || len(schedule) != 2 {
		t.Errorf("schedule = %v, want two instalments", body["schedule"])
	}
}

func TestIntegrationSourceLinesCarryTheirSuppliedVat(t *testing.T) {
	var body map[string]interface{}
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"taxdec_int1","object":"tax_decision","status":"final","taxSource":"integration"}`)
	})

	standardRate := 2000
	zeroRate := 0
	decision, err := client.TaxDecisions.Create(&TaxDecisionParams{
		TaxSource:   "integration",
		Customer:    "cus_8f2k4m9n",
		EffectiveAt: "2026-09-15",
		Currency:    "eur",
		PriceMode:   "tax_exclusive",
		Lines: []*TaxDecisionLineParams{
			{
				Reference: "conseil", Description: "Conseil", Category: "services",
				UnitAmount: 2900, Quantity: "1",
				VatRate: &standardRate, VatCode: "S",
			},
			{
				Reference: "formation", Description: "Formation exportée", Category: "services",
				UnitAmount: 5000, Quantity: "1",
				VatRate: &zeroRate, VatCode: "G",
				VatexCode: "VATEX-EU-G", PlaceOfSupply: "GB",
			},
		},
		IdempotencyKey: "order-integration-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body["taxSource"] != "integration" {
		t.Errorf("taxSource = %v, want integration", body["taxSource"])
	}
	lines := body["lines"].([]interface{})
	first := lines[0].(map[string]interface{})
	second := lines[1].(map[string]interface{})
	if first["vatRate"] != float64(2000) || first["vatCode"] != "S" {
		t.Errorf("line 1 supplied VAT = %v/%v", first["vatRate"], first["vatCode"])
	}
	// Zero is a REAL rate on the wire, not an omitted field.
	if second["vatRate"] != float64(0) || second["vatexCode"] != "VATEX-EU-G" || second["placeOfSupply"] != "GB" {
		t.Errorf("line 2 supplied VAT = %v/%v/%v", second["vatRate"], second["vatexCode"], second["placeOfSupply"])
	}
	if decision.TaxSource != "integration" {
		t.Errorf("TaxSource = %q, want integration", decision.TaxSource)
	}
}

func TestInvoiceReadsThreeAxes(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
		  "id": "inv_1", "object": "invoice", "status": "deposited",
		  "documentStatus": "finalized", "transmissionStatus": "deposited",
		  "transmissionDetail": "", "paymentStatus": "partially_paid",
		  "taxSource": "facturino"
		}`)
	})

	inv, err := client.Invoices.Get("inv_1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A collection never moves the transmission axis, and `status` stays
	// populated as their projection.
	if inv.Status != "deposited" || inv.TransmissionStatus != "deposited" {
		t.Errorf("transmission = %q / %q", inv.Status, inv.TransmissionStatus)
	}
	if inv.PaymentStatus != "partially_paid" {
		t.Errorf("PaymentStatus = %q", inv.PaymentStatus)
	}
}

func TestCreditNoteFromCreditedLines(t *testing.T) {
	var body map[string]interface{}
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{
		  "id": "crn_1", "object": "credit_note", "status": "draft",
		  "taxSource": "facturino", "originalInvoiceId": "inv_1",
		  "originalTaxDecisionId": "taxdec_9c1f"
		}`)
	})

	cn, err := client.CreditNotes.Create(&CreditNoteParams{
		RelatedInvoiceID: "inv_1",
		CreditNoteType:   "partial",
		ReasonCode:       "quality",
		CreditedLines:    []*CreditedLineParams{{TaxLineRef: "abo-pro", AmountTTC: 1200}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	credited, _ := body["creditedLines"].([]interface{})
	if len(credited) != 1 {
		t.Fatalf("creditedLines length = %d, want 1", len(credited))
	}
	// The VAT is inherited from the invoice snapshot, never restated.
	if _, present := body["items"]; present {
		t.Error("items was sent together with creditedLines")
	}
	if cn.OriginalTaxDecisionID != "taxdec_9c1f" {
		t.Errorf("OriginalTaxDecisionID = %q", cn.OriginalTaxDecisionID)
	}
}

func TestRecurringInvoiceSendsTaxInputs(t *testing.T) {
	var raw string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		raw = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"rin_1","object":"recurring_invoice"}`)
	})

	_, err := client.RecurringInvoices.Create(&RecurringInvoiceParams{
		CustomerID:         "cus_1",
		Frequency:          "monthly",
		StartDate:          "2026-09-01",
		NextGenerationDate: "2026-09-01",
		TaxInputs: &RecurringTaxInputsParams{
			PriceMode: "tax_exclusive",
			Lines: []*RecurringTaxLineParams{{
				TaxDecisionLineParams: TaxDecisionLineParams{
					Reference: "abo-pro", Description: "Abonnement Pro",
					Category:     "electronically_supplied_services",
					RateCategory: "standard", UnitAmount: 2900, Quantity: "1",
				},
				Unit: "month",
			}},
		},
		TemplateInvoice: &RecurringTemplateParams{PaymentTermsDays: 30},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(raw, `"taxInputs"`) {
		t.Errorf("taxInputs missing: %s", raw)
	}
	// No decision id travels: each occurrence is decided on its own date.
	if strings.Contains(raw, "taxDecisionId") {
		t.Errorf("a decision id leaked into a recurrence: %s", raw)
	}
	if strings.Contains(raw, `"items"`) {
		t.Errorf("templateInvoice.items sent with taxInputs: %s", raw)
	}
}

// TestFiscalSurfaceParity fails if the fiscal contract disappears from the SDK.
func TestFiscalSurfaceParity(t *testing.T) {
	client := New("fac_test_xxx")
	if client.TaxDecisions == nil {
		t.Fatal("TaxDecisions dropped from the client")
	}

	service := reflect.TypeOf(client.TaxDecisions)
	for _, method := range []string{"Create", "Get", "Retrieve"} {
		if _, found := service.MethodByName(method); !found {
			t.Errorf("TaxDecisionService.%s is missing", method)
		}
	}

	decision := reflect.TypeOf(TaxDecision{})
	for _, field := range []string{
		"Status", "TaxSource", "AmountToCharge", "Totals", "InvoiceChannel",
		"TransactionReporting", "PaymentReporting", "ForeignTaxReviewRequired",
		"RetryOfTaxDecisionID", "Expired", "RulesVersion", "OperationFingerprint",
		"ObligationReasons", "Vies", "Issues", "LocationEvidence",
	} {
		if _, found := decision.FieldByName(field); !found {
			t.Errorf("TaxDecision.%s is missing", field)
		}
	}

	invoice := reflect.TypeOf(Invoice{})
	for _, field := range []string{
		"DocumentStatus", "TransmissionStatus", "TransmissionDetail",
		"PaymentStatus", "TaxSource", "TaxDecisionID", "TaxSnapshot",
	} {
		if _, found := invoice.FieldByName(field); !found {
			t.Errorf("Invoice.%s is missing", field)
		}
	}

	creditNote := reflect.TypeOf(CreditNote{})
	for _, field := range []string{
		"DocumentStatus", "PaymentStatus", "TaxSource",
		"OriginalInvoiceID", "OriginalTaxDecisionID", "TaxSnapshot",
	} {
		if _, found := creditNote.FieldByName(field); !found {
			t.Errorf("CreditNote.%s is missing", field)
		}
	}
}

func TestInvoiceCreateRequiresTheDecisionLocally(t *testing.T) {
	// Checked locally: the refusal happens without a round trip — no server is
	// configured, so any HTTP attempt would fail loudly.
	client := New("fac_test_xxx")

	withoutDecision := &InvoiceParams{
		Customer:      "cus_8f2k4m9n",
		DecisionLines: []*DecisionLineParams{{TaxLineRef: "abo-pro", Unit: "month"}},
		Buyer:         &BuyerParams{CompanyName: "ACME SAS"},
		Dates:         &InvoiceDatesParams{Issued: "2026-09-15", Due: "2026-10-15"},
		Payment:       &PaymentTermsParams{Terms: "30 jours", TermsDays: 30, Method: "transfer"},
	}
	if _, err := client.Invoices.Create(withoutDecision); err == nil || !strings.Contains(err.Error(), "TaxDecisionID is required") {
		t.Fatalf("Create error = %v, want a local TaxDecisionID refusal", err)
	}

	withoutLines := &InvoiceParams{
		Customer:      "cus_8f2k4m9n",
		TaxDecisionID: "taxdec_9c1f",
		Buyer:         &BuyerParams{CompanyName: "ACME SAS"},
		Dates:         &InvoiceDatesParams{Issued: "2026-09-15", Due: "2026-10-15"},
		Payment:       &PaymentTermsParams{Terms: "30 jours", TermsDays: 30, Method: "transfer"},
	}
	if _, err := client.Invoices.Create(withoutLines); err == nil || !strings.Contains(err.Error(), "DecisionLines is required") {
		t.Fatalf("Create error = %v, want a local DecisionLines refusal", err)
	}

	if _, err := client.Invoices.Create(nil); err == nil {
		t.Fatal("Create(nil) must refuse locally")
	}
}
