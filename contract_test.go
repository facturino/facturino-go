package facturino

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

const contractFixture = `{"invoice":{"id":"inv_00090","object":"invoice","number":"FAC2026-00090","status":"rejected","documentStatus":"finalized","transmissionStatus":"rejected","transmissionDetail":null,"paymentStatus":"unpaid","type":"standard","currency":"eur","customer":{"ref":"cus_example","snapshot":{"name":"INTEK CENTER","address":{"line1":"1 rue Exemple","postalCode":"75001","city":"Paris","country":"FR"}}},"items":[],"totals":{"totalHT":990,"totalTVA":198,"totalTTC":1188,"amountPaid":0,"amountDue":1188},"einvoicing":{"paId":"dep_current","paStatus":"rejected","paStatusCode":"fr:213","paTransactionId":null,"paErrorCode":null,"rejectionCode":"REJ_ADR","rejectionSource":"platform","rejectionReason":"REJ_ADR — Directory line matricule_plateforme (0037) does not match expected matricule (0022)","rejectionNote":null,"rejectionCategory":"addressing_error","refusalReason":null,"paIdempotencyKey":"submission_same_attempt","peppolDeliveryId":null,"ereportingId":null,"ereportingPaymentId":null,"sentAt":"2026-09-07T12:36:00.000Z","trackingId":null,"routingIdentifier":"0225:73282932000074","buyerReachableAt":null,"directoryCheckedAt":null,"submissionArtefact":{"kind":"cii","path":"invoices/inv_00090/submission.xml","generatedAt":"2026-09-07T12:35:00.000Z","correctedRules":[],"routingIdentifier":"0225:73282932000074"},"previousSubmissions":[{"paId":"dep_previous","paTransactionId":null,"paIdempotencyKey":"submission_previous","paStatus":"rejected","paStatusCode":"fr:213","paErrorCode":null,"rejectionReason":"AUTRE — Ce motif nécessite une explication en note de CDV.","rejectionCode":"AUTRE","rejectionSource":"platform","rejectionNote":"L’adresse de réception doit être confirmée.","rejectionCategory":"other","sentAt":"2026-09-07T08:00:00.000Z","closedAt":"2026-09-07T12:35:00.000Z"}]},"livemode":true,"created":"2026-09-07T12:00:00.000Z","updated":"2026-09-07T12:36:00.000Z"},"events":[{"id":"evt_invoice_rejected","object":"event","type":"invoice.rejected","apiVersion":"2026-09-01","created":"2026-09-07T12:36:00.000Z","livemode":true,"data":{"id":"inv_00090","object":"invoice","number":"FAC2026-00090","status":"rejected","previous_status":"deposited","livemode":true,"documentStatus":"finalized","transmissionStatus":"rejected","transmissionDetail":null,"paymentStatus":"unpaid","paErrorCode":null,"rejectionReason":"REJ_ADR — Directory line matricule_plateforme (0037) does not match expected matricule (0022)","rejectionCategory":"addressing_error","rejectionCode":"REJ_ADR","rejectionSource":"platform","metadata":{}},"request":null},{"id":"evt_invoice_refused","object":"event","type":"invoice.refused","apiVersion":"2026-09-01","created":"2026-09-07T12:36:00.000Z","livemode":true,"data":{"id":"inv_00090","object":"invoice","number":"FAC2026-00090","status":"refused","previous_status":"deposited","livemode":true,"documentStatus":"finalized","transmissionStatus":"rejected","transmissionDetail":"refused","paymentStatus":"unpaid","paErrorCode":null,"rejectionReason":"Montant contesté.","rejectionCategory":"refused_by_buyer","rejectionCode":null,"rejectionSource":"buyer","metadata":{}},"request":null},{"id":"evt_credit_note_credit_refused","object":"event","type":"credit_note.credit_refused","apiVersion":"2026-09-01","created":"2026-09-07T12:36:00.000Z","livemode":true,"data":{"id":"cn_example","object":"credit_note","number":"AV2026-00001","status":"credit_refused","previous_status":"deposited","livemode":true,"documentStatus":"finalized","transmissionStatus":"rejected","transmissionDetail":"refused","paymentStatus":"unpaid","paErrorCode":null,"rejectionReason":"Montant contesté.","rejectionCategory":"refused_by_buyer","rejectionCode":null,"rejectionSource":"buyer","metadata":{},"relatedInvoiceId":"inv_00090","relatedInvoiceNumber":"FAC2026-00090"},"request":null},{"id":"evt_payment_received","object":"event","type":"payment.received","apiVersion":"2026-09-01","created":"2026-09-07T12:36:00.000Z","livemode":true,"data":{"id":"inv_00090","object":"invoice","number":"FAC2026-00090","status":"paid","previous_status":"rejected","livemode":true,"documentStatus":"finalized","transmissionStatus":"rejected","transmissionDetail":null,"paymentStatus":"paid","paErrorCode":null,"rejectionReason":"REJ_ADR — Directory line matricule_plateforme (0037) does not match expected matricule (0022)","rejectionCategory":"addressing_error","rejectionCode":"REJ_ADR","rejectionSource":"platform","metadata":{},"amount":"11.88","total_paid":"11.88","total_due":"11.88","total":"11.88","amountDue":"0.00"},"request":null}],"payment":{"id":"pay_example","object":"payment","amount":1188,"method":"transfer","reference":null,"paidAt":"2026-09-07T12:40:00.000Z","recorded_by":"api","created":"2026-09-07T12:41:00.000Z","fr212":{"state":"awaiting_deposit","sentAt":null,"lastErrorCode":"awaiting_pa_deposit","updatedAt":"2026-09-07T12:41:00.000Z"}},"blockedPayment":{"id":"pay_example","object":"payment","amount":1188,"method":"transfer","reference":null,"paidAt":"2026-09-07T12:40:00.000Z","recorded_by":"api","created":"2026-09-07T12:41:00.000Z","fr212":{"state":"blocked","sentAt":null,"lastErrorCode":"pa_lifecycle_rejected","lastErrorReason":"Invoice identifier is rejected.","updatedAt":"2026-09-07T12:41:00.000Z"}},"customer":{"id":"cus_example","object":"customer","name":"SAS ALPHA MENUISERIE","warnings":[{"code":"buyer_nature_suspect","message":"The customer name suggests a business. Check the customer type and SIRET.","param":"name"}]},"taxDecision":{"id":"taxdec_example","object":"tax_decision","warnings":[{"code":"buyer_nature_suspect","message":"The customer name suggests a business. Check the customer type and SIRET.","param":"customerId"}]},"creditNote":{"id":"cn_example","object":"credit_note","relatedInvoiceId":"inv_00090","relatedInvoiceNumber":"FAC2026-00090","einvoicing":{"paId":null,"paStatus":"rejected","depositedAt":null,"paStatusCode":"fr:210","rejectionReason":"Montant contesté.","rejectionCode":null,"rejectionCategory":"refused_by_buyer","rejectionSource":"buyer","rejectionNote":"Corriger le montant convenu.","previousSubmissions":[]}}}`

func TestContract270(t *testing.T) {
	var fixture struct {
		Invoice        Invoice           `json:"invoice"`
		Payment        Payment           `json:"payment"`
		BlockedPayment Payment           `json:"blockedPayment"`
		CreditNote     CreditNote        `json:"creditNote"`
		Customer       Customer          `json:"customer"`
		TaxDecision    TaxDecision       `json:"taxDecision"`
		Events         []json.RawMessage `json:"events"`
	}
	if err := json.Unmarshal([]byte(contractFixture), &fixture); err != nil {
		t.Fatal(err)
	}
	e := fixture.Invoice.Einvoicing
	if e == nil || e.RejectionCategory == nil || *e.RejectionCategory != PaRejectionAddressingError || *e.RejectionCode != "REJ_ADR" {
		t.Fatalf("lost rejection: %+v", e)
	}
	if e.RejectionNote != nil || e.BuyerReachableAt != nil || e.DirectoryCheckedAt != nil {
		t.Fatal("null became a value")
	}
	if *e.RejectionSource != PaRejectionSourcePlatform || *e.SubmissionArtefact.RoutingIdentifier != "0225:73282932000074" {
		t.Fatal("lost source or routing")
	}
	if *e.PreviousSubmissions[0].RejectionCategory != PaRejectionOther || *e.PreviousSubmissions[0].RejectionNote != "L’adresse de réception doit être confirmée." {
		t.Fatal("lost archive")
	}
	if fixture.Payment.FR212.State != "awaiting_deposit" || fixture.Payment.FR212.SentAt != nil {
		t.Fatal("lost awaiting-deposit state")
	}
	if fixture.BlockedPayment.FR212.LastErrorReason != "Invoice identifier is rejected." {
		t.Fatal("lost raw reason")
	}
	if *fixture.CreditNote.RelatedInvoiceNumber != "FAC2026-00090" || fixture.Customer.Warnings[0].Code != BuyerNatureSuspect || fixture.TaxDecision.Warnings[0].Param != "customerId" {
		t.Fatal("lost related number or warning")
	}
	for _, payload := range fixture.Events {
		timestamp := time.Now().Unix()
		mac := hmac.New(sha256.New, []byte("whsec_contract_fixture"))
		fmt.Fprintf(mac, "%d.%s", timestamp, payload)
		header := fmt.Sprintf("t=%d,v1=%s", timestamp, hex.EncodeToString(mac.Sum(nil)))
		event, err := VerifyWebhookSignature(payload, header, "whsec_contract_fixture")
		if err != nil {
			t.Fatal(err)
		}
		if event.APIVersion != "2026-09-01" {
			t.Fatal("lost apiVersion")
		}
		data, err := event.DocumentData()
		if err != nil {
			t.Fatal(err)
		}
		if data.RejectionReason == nil {
			t.Fatal("lost raw verdict")
		}
		var record Event
		if err := json.Unmarshal(payload, &record); err != nil {
			t.Fatal(err)
		}
		projected, err := record.DocumentData()
		if err != nil || *projected.RejectionReason != *data.RejectionReason {
			t.Fatal("REST and signed projections diverge")
		}
	}
}

func TestRejectionCategoryConstants270(t *testing.T) {
	values := []PaRejectionCategory{PaRejectionBuyerNotInDirectory, PaRejectionAddressingError, PaRejectionOther, PaRejectionFormatInvalid, PaRejectionSemanticError, PaRejectionDuplicate, PaRejectionPlatformAuth, PaRejectionPlatformUnavailable, PaRejectionRefusedByBuyer, PaRejectionSuspended, PaRejectionUnknown}
	seen := map[PaRejectionCategory]bool{}
	for _, value := range values {
		if seen[value] {
			t.Fatal("duplicate category")
		}
		seen[value] = true
	}
	if len(seen) != 11 {
		t.Fatal("eleven categories required")
	}
}
