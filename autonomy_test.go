package facturino

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAutonomousObligationContract(t *testing.T) {
	raw, err := os.ReadFile("testdata/contract/autonomy.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Obligation ObligationFollowUp      `json:"obligation"`
		Completed  ObligationFollowUp      `json:"completed"`
		Einvoicing InvoiceEinvoicing       `json:"einvoicing"`
		FR212      PaymentCollectionStatus `json:"fr212"`
		Events     []WebhookEvent          `json:"events"`
	}
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Obligation.Owner == nil || *corpus.Obligation.Owner != "facturino" || corpus.Completed.Owner != nil {
		t.Fatal("nullable owner lost")
	}
	if corpus.Einvoicing.Obligation == nil || corpus.FR212.Obligation == nil {
		t.Fatal("nested follow-up lost")
	}
	for _, event := range corpus.Events {
		data, err := event.TypedData()
		if err != nil {
			t.Fatal(err)
		}
		typed, ok := data.(*ObligationWebhookData)
		if !ok || typed.Obligation == nil || typed.Obligation.Action == nil || *typed.Obligation.Action != "retry" {
			t.Fatal("untyped follow-up event")
		}
	}
	var historical InvoiceEinvoicing
	if err := json.Unmarshal([]byte(`{"obligation":null}`), &historical); err != nil || historical.Obligation != nil {
		t.Fatal("invented historical follow-up")
	}
}
