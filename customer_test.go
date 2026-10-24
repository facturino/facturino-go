package facturino

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCustomerTaxStatusOnCreateAndUpdate(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPatch} {
		for _, status := range []string{"", "taxable", "non_taxable"} {
			name := status
			if name == "" {
				name = "omitted"
			}
			t.Run(method+"/"+name, func(t *testing.T) {
				response := corpusBody(t, "customer")
				if status != "" {
					response["taxStatus"] = status
				}
				var body map[string]interface{}
				client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
					path := "/v1/customers"
					if method == http.MethodPatch {
						path += "/cus_c6_fixture"
					}
					if r.Method != method || r.URL.Path != path {
						t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, method, path)
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("decode request: %v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(response); err != nil {
						t.Errorf("encode response: %v", err)
					}
				})
				var customer *Customer
				var err error
				if method == http.MethodPost {
					customer, err = client.Customers.Create(&CustomerParams{
						Name: "Local tax customer", Type: "company", TaxStatus: status,
					})
				} else {
					customer, err = client.Customers.Update("cus_c6_fixture", &CustomerUpdateParams{TaxStatus: status})
				}
				if err != nil {
					t.Fatal(err)
				}
				if got, present := body["taxStatus"]; status == "" {
					if present {
						t.Errorf("taxStatus = %#v, want omitted", got)
					}
				} else if !present || got != status {
					t.Errorf("taxStatus = %#v, want %q", got, status)
				}
				if customer.TaxStatus != status {
					t.Errorf("TaxStatus = %q, want %q", customer.TaxStatus, status)
				}
			})
		}
	}
}
