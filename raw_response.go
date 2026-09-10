package facturino

import "encoding/json"

// RawResponse retains the original field presence and unknown additions.
// Field returns (nil, false) for an absent key, and ("null", true) for explicit null.
type RawResponse struct { Fields map[string]json.RawMessage `json:"-"` }
func (r RawResponse) Field(name string) (json.RawMessage, bool) { value, present := r.Fields[name]; return value, present }

func (r *Invoice) UnmarshalJSON(data []byte) error {
    type wire Invoice
    var value wire
    if err := json.Unmarshal(data, &value); err != nil { return err }
    *r = Invoice(value)
    return json.Unmarshal(data, &r.Fields)
}

func (r *Payment) UnmarshalJSON(data []byte) error {
    type wire Payment
    var value wire
    if err := json.Unmarshal(data, &value); err != nil { return err }
    *r = Payment(value)
    return json.Unmarshal(data, &r.Fields)
}

func (r *Customer) UnmarshalJSON(data []byte) error {
    type wire Customer
    var value wire
    if err := json.Unmarshal(data, &value); err != nil { return err }
    *r = Customer(value)
    return json.Unmarshal(data, &r.Fields)
}

func (r *CreditNote) UnmarshalJSON(data []byte) error {
    type wire CreditNote
    var value wire
    if err := json.Unmarshal(data, &value); err != nil { return err }
    *r = CreditNote(value)
    return json.Unmarshal(data, &r.Fields)
}

func (r *TaxDecision) UnmarshalJSON(data []byte) error {
    type wire TaxDecision
    var value wire
    if err := json.Unmarshal(data, &value); err != nil { return err }
    *r = TaxDecision(value)
    return json.Unmarshal(data, &r.Fields)
}

func (r *Event) UnmarshalJSON(data []byte) error {
    type wire Event
    var value wire
    if err := json.Unmarshal(data, &value); err != nil { return err }
    *r = Event(value)
    return json.Unmarshal(data, &r.Fields)
}
