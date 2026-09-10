package facturino

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestRegistryDisclosure(t *testing.T) {
	raw, err := os.ReadFile("testdata/contract/registry-disclosure.json")
	if err != nil {
		t.Fatal(err)
	}
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) { w.Write(raw) })
	result, err := client.Customers.Lookup(&CustomerLookupParams{SIRET: "00012345500008"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Data.Name != "" || result.Data.Address.Line1 != "" || result.Data.Disclosure == nil || *result.Data.Disclosure.Status != "P" || len(result.Data.Disclosure.WithheldFields) != 4 {
		t.Fatal("protected lookup was changed or disclosure lost")
	}
	var legacy SireneCompany
	if err := json.Unmarshal([]byte(`{"name":"Client"}`), &legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.Disclosure != nil {
		t.Fatal("invented disclosure")
	}
	var unknown RegistryDisclosure
	if err := json.Unmarshal([]byte(`{"status":null,"withheldFields":[]}`), &unknown); err != nil {
		t.Fatal(err)
	}
	if unknown.Status != nil {
		t.Fatal("invented status")
	}
}
