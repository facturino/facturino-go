package facturino

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAfterTrainAdditions(t *testing.T) {
	raw, err := os.ReadFile("testdata/contract/after-train.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Einvoicing               InvoiceEinvoicing `json:"einvoicing"`
		PaymentEvent             WebhookEvent      `json:"paymentEvent"`
		LegacyPaymentEvent       WebhookEvent      `json:"legacyPaymentEvent"`
		UnattributedPaymentEvent WebhookEvent      `json:"unattributedPaymentEvent"`
	}
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Einvoicing.SenderRoutingIdentifier == nil || *corpus.Einvoicing.SenderRoutingIdentifier != "0225:908456726_90845672600029" {
		t.Fatal("seller address lost")
	}
	if corpus.Einvoicing.SubmissionArtefact.SellerRoutingIdentifier == nil {
		t.Fatal("artifact address lost")
	}
	for i, event := range []WebhookEvent{corpus.PaymentEvent, corpus.LegacyPaymentEvent, corpus.UnattributedPaymentEvent} {
		data, err := event.TypedData()
		if err != nil {
			t.Fatal(err)
		}
		payment := data.(*PaymentReceivedWebhookData)
		if i == 0 {
			if payment.PaymentID == nil || *payment.PaymentID != "pay_example" || payment.FR212 == nil || payment.FR212.State != "pending" {
				t.Fatal("payment attribution lost")
			}
		} else if payment.PaymentID != nil || payment.FR212 != nil {
			t.Fatal("invented historical attribution")
		}
	}
}
