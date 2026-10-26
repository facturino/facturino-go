package facturino

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestActualServerResponses(t *testing.T) {
	for _, name := range []string{"invoice", "payment", "customer", "creditNote", "taxDecision", "legal-category", "event"} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile("testdata/contract/" + name + ".json")
			if err != nil {
				t.Fatal(err)
			}
			client := corpusClient(raw)
			var model interface{}
			switch name {
			case "invoice":
				model, err = client.Invoices.Get("inv_fixture")
			case "payment":
				model, err = client.Payments.Create("inv_fixture", &PaymentParams{Amount: 50000, Method: "transfer", PaidAt: "2026-09-09T12:00:00Z"})
			case "customer":
				model, err = client.Customers.Get("cus_fixture")
			case "creditNote":
				model, err = client.CreditNotes.Get("cn_fixture")
			case "taxDecision", "legal-category":
				model, err = client.TaxDecisions.Get("taxdec_fixture")
			case "event":
				model, err = client.Events.Get("evt_fixture")
			}
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			view := model.(interface {
				Field(string) (json.RawMessage, bool)
			})
			typ := reflect.TypeOf(model).Elem()
			typed := make(map[string]bool)
			for i := 0; i < typ.NumField(); i++ {
				tag := typ.Field(i).Tag.Get("json")
				for j, char := range tag {
					if char == ',' {
						tag = tag[:j]
						break
					}
				}
				typed[tag] = true
			}
			for key, value := range fields {
				got, present := view.Field(key)
				if !present || string(got) != string(value) {
					t.Errorf("lost field %s", key)
				}
				if !typed[key] {
					t.Errorf("no typed field %s", key)
				}
			}
			if _, exists := view.Field("not_a_field"); exists {
				t.Fatal("invented field")
			}
			if name == "invoice" && model.(*Invoice).Number != nil {
				t.Fatal("null invoice number lost")
			}
		})
	}
}

// corpusBody reads one shared corpus file as a generic JSON object, so a test
// can add the keys the corpus carries empty.
func corpusBody(t *testing.T, name string) map[string]interface{} {
	t.Helper()
	raw, err := os.ReadFile("testdata/contract/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body
}

type contractRoundTripper func(*http.Request) (*http.Response, error)

func (f contractRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// corpusClient returns stored API bytes through an in-memory HTTP transport.
func corpusClient(raw []byte) *Client {
	transport := contractRoundTripper(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(raw)),
			Request:    r,
		}, nil
	})
	return New("fac_test_contract", WithHTTPClient(&http.Client{Transport: transport}), WithMaxRetries(0))
}

// serveJSON answers every request with the given body, without network access.
func serveJSON(t *testing.T, body interface{}) *Client {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return corpusClient(raw)
}

func TestEReportingObligationResponses(t *testing.T) {
	for _, tc := range []struct {
		name    string
		fields  map[string]interface{}
		basis   string
		dueFrom string
		ack     string
	}{
		{
			name: "legal",
			fields: map[string]interface{}{
				"obligationBasis": "legal", "obligationDueFrom": "2026-09-01", "acknowledgedBy": "administration",
			},
			basis: "legal", dueFrom: "2026-09-01", ack: "administration",
		},
		{
			name: "voluntary",
			fields: map[string]interface{}{
				"obligationBasis": "voluntary_early_adoption", "obligationDueFrom": "2027-09-01", "acknowledgedBy": "platform",
			},
			basis: "voluntary_early_adoption", dueFrom: "2027-09-01", ack: "platform",
		},
		{
			name: "null",
			fields: map[string]interface{}{
				"obligationBasis": nil, "obligationDueFrom": nil, "acknowledgedBy": nil,
			},
		},
		{name: "historical"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]interface{}{"id": "erep_fixture", "object": "ereporting", "state": "acknowledged"}
			for key, value := range tc.fields {
				body[key] = value
			}
			check := func(declaration *EReporting) {
				t.Helper()
				if declaration.ID != "erep_fixture" || declaration.State != "acknowledged" {
					t.Fatalf("declaration identity or state lost: %+v", declaration)
				}
				for _, field := range []struct {
					name string
					got  *string
					want string
				}{
					{"ObligationBasis", declaration.ObligationBasis, tc.basis},
					{"ObligationDueFrom", declaration.ObligationDueFrom, tc.dueFrom},
					{"AcknowledgedBy", declaration.AcknowledgedBy, tc.ack},
				} {
					if field.want == "" {
						if field.got != nil {
							t.Errorf("%s = %q, want nil", field.name, *field.got)
						}
					} else if field.got == nil || *field.got != field.want {
						t.Errorf("%s = %v, want %q", field.name, field.got, field.want)
					}
				}
			}

			declaration, err := serveJSON(t, body).EReporting.Get("erep_fixture")
			if err != nil {
				t.Fatal(err)
			}
			check(declaration)

			iter := serveJSON(t, map[string]interface{}{
				"data": []interface{}{body}, "has_more": false,
			}).EReporting.List(nil)
			if !iter.Next() {
				t.Fatalf("declaration missing from list: %v", iter.Err())
			}
			check(iter.EReporting())
			if iter.Next() || iter.Err() != nil {
				t.Fatalf("unexpected extra declaration or pagination error: %v", iter.Err())
			}
		})
	}
}

func TestEReportingSkippedResponses(t *testing.T) {
	for _, tc := range []struct {
		name            string
		state           string
		status          string
		fields          map[string]interface{}
		reason          string
		obligationState string
	}{
		{
			name: "nothing_to_declare", state: "skipped", status: "skipped",
			fields: map[string]interface{}{"skippedReason": "nothing_to_declare"},
			reason: "nothing_to_declare", obligationState: "completed",
		},
		{
			name: "null", state: "reserved", status: "draft",
			fields:          map[string]interface{}{"skippedReason": nil},
			obligationState: "pending",
		},
		{name: "historical", state: "reserved", status: "draft"},
		{name: "historical_skipped_status", status: "skipped"},
		{
			name: "blocked", state: "blocked", status: "skipped",
			obligationState: "attention_required",
		},
		{
			name: "future_reason", state: "skipped", status: "skipped",
			fields: map[string]interface{}{"skippedReason": "future_reason"},
			reason: "future_reason", obligationState: "completed",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]interface{}{
				"id": "erep_fixture", "object": "ereporting", "status": tc.status, "volet": "payment",
			}
			if tc.state != "" {
				body["state"] = tc.state
			}
			for key, value := range tc.fields {
				body[key] = value
			}
			if tc.obligationState != "" {
				body["obligation"] = map[string]interface{}{
					"state": tc.obligationState, "reasonCode": nil, "reasonSource": nil,
					"owner": nil, "action": nil, "nextAttemptAt": nil,
				}
			}
			check := func(declaration *EReporting) {
				t.Helper()
				if declaration.ID != "erep_fixture" || declaration.State != tc.state || declaration.Status != tc.status {
					t.Fatalf("declaration identity or state lost: %+v", declaration)
				}
				if tc.reason == "" {
					if declaration.SkippedReason != nil {
						t.Errorf("SkippedReason = %q, want nil", *declaration.SkippedReason)
					}
				} else if declaration.SkippedReason == nil || *declaration.SkippedReason != tc.reason {
					t.Errorf("SkippedReason = %v, want %q", declaration.SkippedReason, tc.reason)
				}
				if tc.obligationState == "" {
					if declaration.Obligation != nil {
						t.Errorf("Obligation = %+v, want nil", declaration.Obligation)
					}
				} else if declaration.Obligation == nil || declaration.Obligation.State != tc.obligationState {
					t.Errorf("Obligation = %+v, want state %q", declaration.Obligation, tc.obligationState)
				}
				if declaration.SubmittedAt != "" || declaration.PAResponseID != "" {
					t.Errorf("unexpected transmission data: %+v", declaration)
				}
			}

			declaration, err := serveJSON(t, body).EReporting.Get("erep_fixture")
			if err != nil {
				t.Fatal(err)
			}
			check(declaration)

			iter := serveJSON(t, map[string]interface{}{
				"data": []interface{}{body}, "has_more": false,
			}).EReporting.List(nil)
			if !iter.Next() {
				t.Fatalf("declaration missing from list: %v", iter.Err())
			}
			check(iter.EReporting())
			if iter.Next() || iter.Err() != nil {
				t.Fatalf("unexpected extra declaration or pagination error: %v", iter.Err())
			}
		})
	}
}

func TestDecisionAndSnapshotObligationBasisResponses(t *testing.T) {
	for _, tc := range []struct {
		name    string
		basis   interface{}
		present bool
	}{
		{"legal", "legal", true},
		{"voluntary", "voluntary_early_adoption", true},
		{"null", nil, true},
		{"historical", nil, false},
	} {
		for _, name := range []string{"taxDecision", "invoice", "creditNote"} {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				body := corpusBody(t, name)
				fields := body
				if name != "taxDecision" {
					fields = body["taxSnapshot"].(map[string]interface{})
				}
				if tc.present {
					fields["obligationBasis"] = tc.basis
				} else {
					delete(fields, "obligationBasis")
				}

				client := serveJSON(t, body)
				var got interface{}
				var present bool
				switch name {
				case "taxDecision":
					decision, err := client.TaxDecisions.Get("taxdec_fixture")
					if err != nil {
						t.Fatal(err)
					}
					if decision.ObligationBasis != nil {
						got = *decision.ObligationBasis
					}
					raw, exists := decision.Field("obligationBasis")
					present = exists
					if tc.present {
						want, err := json.Marshal(tc.basis)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(raw, want) {
							t.Errorf("raw obligationBasis = %s, want %s", raw, want)
						}
					}
				case "invoice":
					invoice, err := client.Invoices.Get("inv_fixture")
					if err != nil {
						t.Fatal(err)
					}
					got, present = invoice.TaxSnapshot["obligationBasis"]
				case "creditNote":
					creditNote, err := client.CreditNotes.Get("cn_fixture")
					if err != nil {
						t.Fatal(err)
					}
					got, present = creditNote.TaxSnapshot["obligationBasis"]
				}
				if got != tc.basis || present != tc.present {
					t.Errorf("obligationBasis = %#v (present %v), want %#v (present %v)", got, present, tc.basis, tc.present)
				}
			})
		}
	}
}

func TestCorpusCarriesNoticesAndRetainedVatNumber(t *testing.T) {
	decision, err := serveJSON(t, corpusBody(t, "taxDecision")).TaxDecisions.Get("taxdec_c6_fixture")
	if err != nil {
		t.Fatal(err)
	}
	// Always present: an empty list, never a missing key.
	if decision.Notices == nil || len(decision.Notices) != 0 {
		t.Errorf("Notices = %#v, want an empty, non-nil list", decision.Notices)
	}
	if decision.Customer == nil {
		t.Fatal("Customer = nil")
	}
	if decision.Customer.RetainedVatNumber != nil {
		t.Errorf("RetainedVatNumber = %q, want nil", *decision.Customer.RetainedVatNumber)
	}
	if decision.ObligationBasis == nil || *decision.ObligationBasis != "legal" {
		t.Errorf("ObligationBasis = %v, want legal", decision.ObligationBasis)
	}

	invoice, err := serveJSON(t, corpusBody(t, "invoice")).Invoices.Get("inv_c6_fixture")
	if err != nil {
		t.Fatal(err)
	}
	if notices, ok := invoice.TaxSnapshot["notices"].([]interface{}); !ok || len(notices) != 0 {
		t.Errorf("invoice taxSnapshot.notices = %#v, want an empty list", invoice.TaxSnapshot["notices"])
	}
	// The shared invoice is bound to a historical decision: the basis is omitted, not null.
	if basis, present := invoice.TaxSnapshot["obligationBasis"]; present {
		t.Errorf("invoice taxSnapshot.obligationBasis = %#v, want omitted", basis)
	}

	creditNote, err := serveJSON(t, corpusBody(t, "creditNote")).CreditNotes.Get("crn_c6_fixture")
	if err != nil {
		t.Fatal(err)
	}
	if notices, ok := creditNote.TaxSnapshot["notices"].([]interface{}); !ok || len(notices) != 0 {
		t.Errorf("credit note taxSnapshot.notices = %#v, want an empty list", creditNote.TaxSnapshot["notices"])
	}
}

func TestDecisionNoticesRetainedVatNumberAndIssueDetailAreHydrated(t *testing.T) {
	withDetail := map[string]interface{}{
		"code":       "vat_number_territory_mismatch",
		"detailCode": "vat_number_ignored_national_registration",
		"message":    "The VAT number was set aside: the national registration establishes the business status.",
		"reference":  "Implementing Regulation (EU) 282/2011 art. 18",
		"lineId":     "line-1",
	}
	bare := map[string]interface{}{
		"code":    "location_evidence_contradictory",
		"message": "Every possible consumer territory gives the same complete tax result.",
	}

	// A FINAL decision carries notices, with the number set aside because the
	// national registration establishes the buyer's status.
	body := corpusBody(t, "taxDecision")
	customer := body["customer"].(map[string]interface{})
	customer["vatNumber"] = "DE811128135"
	customer["retainedVatNumber"] = nil
	body["notices"] = []interface{}{withDetail, bare}
	decision, err := serveJSON(t, body).TaxDecisions.Get("taxdec_c6_fixture")
	if err != nil {
		t.Fatal(err)
	}
	if !decision.IsFinal() {
		t.Error("IsFinal() = false on a final decision that carries notices")
	}
	if len(decision.Notices) != 2 {
		t.Fatalf("Notices = %+v, want 2", decision.Notices)
	}
	first := decision.Notices[0]
	if first.Code != "vat_number_territory_mismatch" ||
		first.DetailCode != "vat_number_ignored_national_registration" ||
		first.Reference != "Implementing Regulation (EU) 282/2011 art. 18" ||
		first.LineID != "line-1" || first.Message == "" {
		t.Errorf("first notice = %+v", first)
	}
	if second := decision.Notices[1]; second.Code != "location_evidence_contradictory" ||
		second.DetailCode != "" || second.Reference != "" || second.LineID != "" {
		t.Errorf("second notice = %+v, optional fields must stay empty", second)
	}
	if decision.Customer.RetainedVatNumber != nil {
		t.Errorf("RetainedVatNumber = %v, want nil after setting the number aside", decision.Customer.RetainedVatNumber)
	}

	// A validated number is retained when no notice sets it aside.
	customer["natureBasis"] = "declared_business_with_valid_vat_number"
	customer["retainedVatNumber"] = "DE811128135"
	body["notices"] = []interface{}{}
	decision, err = serveJSON(t, body).TaxDecisions.Get("taxdec_c6_fixture")
	if err != nil {
		t.Fatal(err)
	}
	if decision.Customer.RetainedVatNumber == nil || *decision.Customer.RetainedVatNumber != "DE811128135" {
		t.Errorf("RetainedVatNumber = %v, want DE811128135", decision.Customer.RetainedVatNumber)
	}

	// Art. 18(2) applies to electronically supplied services for an EU buyer
	// that communicated no VAT number.
	body = corpusBody(t, "taxDecision")
	customer = body["customer"].(map[string]interface{})
	customer["nature"] = "consumer"
	customer["natureBasis"] = "declared_business_not_evidenced"
	customer["vatNumber"] = nil
	customer["retainedVatNumber"] = nil
	body["lines"].([]interface{})[0].(map[string]interface{})["category"] = "electronically_supplied_services"
	body["lines"].([]interface{})[0].(map[string]interface{})["effectiveCategory"] = "electronically_supplied_services"
	body["notices"] = []interface{}{map[string]interface{}{
		"code": "business_status_not_evidenced", "message": "The buyer is treated as a non-taxable person under art. 18(2).",
		"reference": "Implementing Regulation (EU) 282/2011 art. 18(2), second subparagraph",
	}}
	decision, err = serveJSON(t, body).TaxDecisions.Get("taxdec_c6_fixture")
	if err != nil {
		t.Fatal(err)
	}
	if decision.Customer.NatureBasis != "declared_business_not_evidenced" || decision.Customer.Nature != "consumer" {
		t.Errorf("Customer = %+v", decision.Customer)
	}
	if decision.Customer.VatNumber != "" || decision.Customer.RetainedVatNumber != nil {
		t.Errorf("declared %q, retained %v: want no communicated or retained number",
			decision.Customer.VatNumber, decision.Customer.RetainedVatNumber)
	}

	// A decision that is NOT final names what to complete.
	body = corpusBody(t, "taxDecision")
	body["status"] = "pending_verification"
	body["totals"] = nil
	body["amountToCharge"] = nil
	body["issues"] = []interface{}{map[string]interface{}{
		"code":       "business_status_not_evidenced",
		"detailCode": "einvoicing_buyer_registration_required",
		"message":    "The buyer SIRET is required for the electronic invoice.",
		"reference":  "CGI art. 242 nonies A",
		"lineId":     "line-1",
	}}
	decision, err = serveJSON(t, body).TaxDecisions.Get("taxdec_c6_fixture")
	if err != nil {
		t.Fatal(err)
	}
	if decision.IsFinal() || len(decision.Issues) != 1 {
		t.Fatalf("IsFinal() = %v, Issues = %+v", decision.IsFinal(), decision.Issues)
	}
	issue := decision.Issues[0]
	if issue.Code != "business_status_not_evidenced" ||
		issue.DetailCode != "einvoicing_buyer_registration_required" ||
		issue.Reference != "CGI art. 242 nonies A" || issue.LineID != "line-1" {
		t.Errorf("issue = %+v", issue)
	}
}

func TestSnapshotNoticesKeepTheShapeOfADecisionNotice(t *testing.T) {
	notice := map[string]interface{}{
		"code":       "vat_number_territory_mismatch",
		"detailCode": "vat_number_ignored_national_registration",
		"message":    "The VAT number was set aside.",
		"lineId":     "l1",
	}
	for name, read := range map[string]func(*Client) (map[string]interface{}, error){
		"invoice": func(client *Client) (map[string]interface{}, error) {
			invoice, err := client.Invoices.Get("inv_c6_fixture")
			if err != nil {
				return nil, err
			}
			return invoice.TaxSnapshot, nil
		},
		"creditNote": func(client *Client) (map[string]interface{}, error) {
			creditNote, err := client.CreditNotes.Get("crn_c6_fixture")
			if err != nil {
				return nil, err
			}
			return creditNote.TaxSnapshot, nil
		},
	} {
		t.Run(name, func(t *testing.T) {
			body := corpusBody(t, name)
			body["taxSnapshot"].(map[string]interface{})["notices"] = []interface{}{notice}
			snapshot, err := read(serveJSON(t, body))
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(snapshot["notices"])
			if err != nil {
				t.Fatal(err)
			}
			var notices []*TaxDecisionNotice
			if err := json.Unmarshal(raw, &notices); err != nil {
				t.Fatal(err)
			}
			if len(notices) != 1 || notices[0].Code != "vat_number_territory_mismatch" ||
				notices[0].DetailCode != "vat_number_ignored_national_registration" || notices[0].LineID != "l1" {
				t.Errorf("frozen notices = %+v", notices)
			}
		})
	}
}

func TestLedgerReadsASaleTakenBackOut(t *testing.T) {
	client := serveJSON(t, map[string]interface{}{
		"object": "eu_threshold_ledger", "id": "2026_test", "year": "2026",
		"entries": []interface{}{map[string]interface{}{
			"object": "eu_threshold_entry", "id": "rev_0123456789abcdef0123456789abcdef", "sequence": 4,
			"kind": "reservation_released", "amountMin": -12000, "amountMax": -12000,
			"evidenceAmountMin": -12000, "evidenceAmountMax": -12000,
			"cumulativeMin": 0, "cumulativeMax": 0, "evidenceCumulativeMin": 0, "evidenceCumulativeMax": 0,
			"taxDecisionId": "taxdec_c6_fixture", "effectiveAt": "2026-09-15",
		}},
	})
	ledger, err := client.EuThresholdLedgers.Retrieve("2026")
	if err != nil {
		t.Fatal(err)
	}
	if len(ledger.Entries) != 1 {
		t.Fatalf("Entries = %+v", ledger.Entries)
	}
	if entry := ledger.Entries[0]; entry.Kind != "reservation_released" || entry.AmountMin >= 0 || entry.AmountMax >= 0 {
		t.Errorf("entry = %+v, want a reservation_released with negative amounts", entry)
	}
}

func TestEveryEventPayloadFamily(t *testing.T) {
	raw, err := os.ReadFile("testdata/contract/events.json")
	if err != nil {
		t.Fatal(err)
	}
	var events []WebhookEvent
	if err := json.Unmarshal(raw, &events); err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		projection, err := event.TypedData()
		if err != nil {
			t.Fatal(err)
		}
		if reflect.TypeOf(projection).Kind() != reflect.Ptr {
			t.Fatalf("no family for %s", event.Type)
		}
	}
}

func TestLocalTaxContractResponses(t *testing.T) {
	corpus := corpusBody(t, "local-tax")
	customer, err := serveJSON(t, corpus["customer"]).Customers.Get("cus_local_tax")
	if err != nil {
		t.Fatal(err)
	}
	if customer.TaxStatus != "taxable" {
		t.Errorf("TaxStatus = %q, want taxable", customer.TaxStatus)
	}

	decision, err := serveJSON(t, corpus["taxDecision"]).TaxDecisions.Get("taxdec_local_tax")
	if err != nil {
		t.Fatal(err)
	}
	if !decision.IsFinal() || decision.Totals == nil || decision.Totals.TotalLocalTax == nil ||
		*decision.Totals.TotalLocalTax != 400 || decision.AmountToCharge == nil || *decision.AmountToCharge != 20400 {
		t.Fatalf("decision = %+v, want a final decision with 400 local tax and 20400 to charge", decision)
	}
	registrations, ok := decision.SellerProfile["localTaxRegistrations"].([]interface{})
	if !ok || len(registrations) != 1 {
		t.Fatalf("localTaxRegistrations = %#v, want one registration", decision.SellerProfile["localTaxRegistrations"])
	}
	wantRegistration := map[string]interface{}{
		"territory": "MF", "reference": "MF-REG-2026", "effectiveFrom": "2026-01-01", "effectiveTo": nil,
	}
	if !reflect.DeepEqual(registrations[0], wantRegistration) {
		t.Errorf("registration = %#v, want %#v", registrations[0], wantRegistration)
	}

	invoice, err := serveJSON(t, corpus["invoice"]).Invoices.Get("inv_local_tax")
	if err != nil {
		t.Fatal(err)
	}
	if invoice.Totals == nil || invoice.Totals.TotalLocalTax == nil || *invoice.Totals.TotalLocalTax != 400 {
		t.Fatalf("invoice totals = %+v, want 400 local tax", invoice.Totals)
	}
	if len(decision.Lines) != 2 || len(invoice.Items) != 2 {
		t.Fatalf("decision lines = %d, invoice items = %d, want 2 each", len(decision.Lines), len(invoice.Items))
	}
	exemption := "cgism_253"
	wantTaxes := []*LocalTax{
		{Code: "tgca_mf", RateCentipercent: 400, Amount: 400, Reference: "CGIsm art. 252"},
		{Code: "tgca_mf", RateCentipercent: 0, Amount: 0, Reference: "CGIsm art. 253", Exemption: &exemption},
	}
	for i, want := range wantTaxes {
		for name, got := range map[string]*LocalTax{
			"decision": decision.Lines[i].LocalTax,
			"invoice":  invoice.Items[i].LocalTax,
		} {
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s line %d local tax = %+v, want %+v", name, i, got, want)
			}
		}
	}
}

func TestLocalTaxHistoricalResponsesAndZeroTotals(t *testing.T) {
	customer, err := serveJSON(t, corpusBody(t, "customer")).Customers.Get("cus_c6_fixture")
	if err != nil {
		t.Fatal(err)
	}
	if customer.TaxStatus != "" {
		t.Errorf("historical TaxStatus = %q, want empty", customer.TaxStatus)
	}
	for _, name := range []string{"taxDecision", "invoice"} {
		for _, present := range []bool{false, true} {
			state := "absent"
			if present {
				state = "zero"
			}
			t.Run(name+"/"+state, func(t *testing.T) {
				body := corpusBody(t, name)
				if present {
					body["totals"].(map[string]interface{})["totalLocalTax"] = 0
				}
				client := serveJSON(t, body)
				var total *int
				var totals interface{}
				var localTax *LocalTax
				if name == "taxDecision" {
					decision, err := client.TaxDecisions.Get("taxdec_c6_fixture")
					if err != nil {
						t.Fatal(err)
					}
					if _, exists := decision.SellerProfile["localTaxRegistrations"]; exists {
						t.Error("invented historical localTaxRegistrations")
					}
					total, totals, localTax = decision.Totals.TotalLocalTax, decision.Totals, decision.Lines[0].LocalTax
				} else {
					invoice, err := client.Invoices.Get("inv_c6_fixture")
					if err != nil {
						t.Fatal(err)
					}
					total, totals, localTax = invoice.Totals.TotalLocalTax, invoice.Totals, invoice.Items[0].LocalTax
				}
				if localTax != nil {
					t.Errorf("historical LocalTax = %+v, want nil", localTax)
				}
				if present {
					if total == nil || *total != 0 {
						t.Errorf("TotalLocalTax = %v, want a pointer to zero", total)
					}
				} else if total != nil {
					t.Errorf("historical TotalLocalTax = %v, want nil", total)
				}
				raw, err := json.Marshal(totals)
				if err != nil {
					t.Fatal(err)
				}
				var fields map[string]json.RawMessage
				if err := json.Unmarshal(raw, &fields); err != nil {
					t.Fatal(err)
				}
				if got, exists := fields["totalLocalTax"]; exists != present || (present && string(got) != "0") {
					t.Errorf("serialized totalLocalTax = %s, present = %v, want present = %v", got, exists, present)
				}
			})
		}
	}
}

func TestLocalTaxResponsesAcceptFutureStrings(t *testing.T) {
	corpus := corpusBody(t, "local-tax")
	decisionBody := corpus["taxDecision"].(map[string]interface{})
	registration := decisionBody["sellerProfile"].(map[string]interface{})["localTaxRegistrations"].([]interface{})[0].(map[string]interface{})
	registration["territory"] = "future_territory"
	registration["effectiveTo"] = "2026-12-31"
	for _, name := range []string{"taxDecision", "invoice"} {
		t.Run(name, func(t *testing.T) {
			body := corpus[name].(map[string]interface{})
			linesKey := "lines"
			if name == "invoice" {
				linesKey = "items"
			}
			tax := body[linesKey].([]interface{})[0].(map[string]interface{})["localTax"].(map[string]interface{})
			tax["code"] = "future_local_tax"
			tax["exemption"] = "future_exemption"
			client := serveJSON(t, body)
			var localTax *LocalTax
			if name == "taxDecision" {
				decision, err := client.TaxDecisions.Get("taxdec_local_tax")
				if err != nil {
					t.Fatal(err)
				}
				localTax = decision.Lines[0].LocalTax
				gotRegistration := decision.SellerProfile["localTaxRegistrations"].([]interface{})[0].(map[string]interface{})
				if gotRegistration["territory"] != "future_territory" || gotRegistration["effectiveTo"] != "2026-12-31" {
					t.Errorf("registration lost future territory or end date: %#v", gotRegistration)
				}
			} else {
				invoice, err := client.Invoices.Get("inv_local_tax")
				if err != nil {
					t.Fatal(err)
				}
				localTax = invoice.Items[0].LocalTax
			}
			if localTax == nil || localTax.Code != "future_local_tax" ||
				localTax.Exemption == nil || *localTax.Exemption != "future_exemption" {
				t.Errorf("future strings lost: %+v", localTax)
			}
		})
	}
}
