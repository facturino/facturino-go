package facturino

import (
	"fmt"
	"net/http"
	"testing"
)

func TestAccountRetrieve(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/v1/account" {
			t.Errorf("Path = %q, want /v1/account", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprint(w, `{
			"object":"account",
			"userId":"usr_abc",
			"companyId":"comp_xyz",
			"plan":"pro",
			"livemode":false,
			"apiKeyPrefix":"fac_test_",
			"permissions":[],
			"company":{"id":"comp_xyz","name":"ACME","siret":"44306184100047","vatRegime":"normal"},
			"user":{"emailVerified":true}
		}`)
	})

	acc, err := client.Account.Retrieve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.UserID != "usr_abc" {
		t.Errorf("UserID = %q, want usr_abc", acc.UserID)
	}
	if acc.Plan != "pro" {
		t.Errorf("Plan = %q, want pro", acc.Plan)
	}
	if acc.Livemode {
		t.Errorf("Livemode = true, want false for fac_test_ key")
	}
	if acc.APIKeyPrefix != "fac_test_" {
		t.Errorf("APIKeyPrefix = %q, want fac_test_", acc.APIKeyPrefix)
	}
	if acc.Company == nil || acc.Company.Name != "ACME" {
		t.Errorf("Company.Name = %v, want ACME", acc.Company)
	}
	if acc.User == nil || !acc.User.EmailVerified {
		t.Errorf("User.EmailVerified = %v, want true", acc.User)
	}
}

func TestAccountRetrieveNoCompany(t *testing.T) {
	// A freshly provisioned user has no company yet — the endpoint must
	// gracefully return company:null without breaking the JSON decode.
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprint(w, `{
			"object":"account",
			"userId":"usr_new",
			"companyId":"",
			"plan":"free",
			"livemode":true,
			"apiKeyPrefix":"fac_live_",
			"permissions":[],
			"company":null,
			"user":{"emailVerified":false}
		}`)
	})

	acc, err := client.Account.Retrieve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Company != nil {
		t.Errorf("Company = %v, want nil for an unconnected account", acc.Company)
	}
	if !acc.Livemode {
		t.Errorf("Livemode = false, want true for fac_live_ key")
	}
}
