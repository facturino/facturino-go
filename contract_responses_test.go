package facturino

import (
    "encoding/json"
    "net/http"
    "os"
    "reflect"
    "testing"
)

func TestActualServerResponses(t *testing.T) {
    for _, name := range []string{"invoice", "payment", "customer", "creditNote", "taxDecision", "event"} {
        t.Run(name, func(t *testing.T) {
            raw, err := os.ReadFile("testdata/contract/" + name + ".json")
            if err != nil { t.Fatal(err) }
            client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Content-Type", "application/json"); _, _ = w.Write(raw) })
            var model interface{}
            switch name {
            case "invoice": model, err = client.Invoices.Get("inv_fixture")
            case "payment": model, err = client.Payments.Create("inv_fixture", &PaymentParams{Amount: 50000, Method: "transfer", PaidAt: "2026-09-09T12:00:00Z"})
            case "customer": model, err = client.Customers.Get("cus_fixture")
            case "creditNote": model, err = client.CreditNotes.Get("cn_fixture")
            case "taxDecision": model, err = client.TaxDecisions.Get("taxdec_fixture")
            case "event": model, err = client.Events.Get("evt_fixture")
            }
            if err != nil { t.Fatal(err) }
            var fields map[string]json.RawMessage
            if err := json.Unmarshal(raw, &fields); err != nil { t.Fatal(err) }
            view := model.(interface{ Field(string) (json.RawMessage, bool) })
            typ := reflect.TypeOf(model).Elem()
            typed := make(map[string]bool)
            for i := 0; i < typ.NumField(); i++ {
                tag := typ.Field(i).Tag.Get("json")
                for j, char := range tag { if char == ',' { tag = tag[:j]; break } }
                typed[tag] = true
            }
            for key, value := range fields {
                got, present := view.Field(key)
                if !present || string(got) != string(value) { t.Errorf("lost field %s", key) }
                if !typed[key] { t.Errorf("no typed field %s", key) }
            }
            if _, exists := view.Field("not_a_field"); exists { t.Fatal("invented field") }
            if name == "invoice" && model.(*Invoice).Number != nil { t.Fatal("null invoice number lost") }
        })
    }
}

func TestEveryEventPayloadFamily(t *testing.T) {
    raw, err := os.ReadFile("testdata/contract/events.json")
    if err != nil { t.Fatal(err) }
    var events []WebhookEvent
    if err := json.Unmarshal(raw, &events); err != nil { t.Fatal(err) }
    for _, event := range events {
        projection, err := event.TypedData()
        if err != nil { t.Fatal(err) }
        if reflect.TypeOf(projection).Kind() != reflect.Ptr { t.Fatalf("no family for %s", event.Type) }
    }
}
