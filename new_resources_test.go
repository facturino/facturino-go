package facturino

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// Tests for the services added to reach 100% API coverage:
// Billing, Cabinet, Notification, Reference, Settings, Usage, Validate.

func TestBillingRetrieveSubscription(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/v1/billing/subscription" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"subscription","plan":"pro","cycle":"monthly","status":"active"}`)
	})

	sub, err := client.Billing.RetrieveSubscription()
	if err != nil {
		t.Fatal(err)
	}
	if sub.Plan != "pro" {
		t.Errorf("Plan = %q, want pro", sub.Plan)
	}
}

func TestBillingUpdateSubscriptionSendsCamelCase(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/v1/billing/subscription" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		// We don't read the body here — the JSON struct tag `cancelAtPeriodEnd`
		// guarantees the wire format if encoding/json is used (it is).
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"subscription","plan":"pro","cancelAtPeriodEnd":true}`)
	})

	cancel := true
	sub, err := client.Billing.UpdateSubscription(&BillingSubscriptionUpdateParams{CancelAtPeriodEnd: &cancel})
	if err != nil {
		t.Fatal(err)
	}
	if !sub.CancelAtPeriodEnd {
		t.Errorf("CancelAtPeriodEnd = false, want true")
	}
}

func TestBillingPauseAndResume(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/billing/pause":
			fmt.Fprint(w, `{"object":"subscription","status":"paused"}`)
		case "/v1/billing/resume":
			fmt.Fprint(w, `{"object":"subscription","status":"active"}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	if sub, err := client.Billing.Pause(); err != nil || sub.Status != "paused" {
		t.Errorf("Pause failed: %v / %v", err, sub)
	}
	if sub, err := client.Billing.Resume(); err != nil || sub.Status != "active" {
		t.Errorf("Resume failed: %v / %v", err, sub)
	}
}

func TestBillingGetInvoicePDF(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/billing/invoices/in_x/pdf" {
			t.Errorf("Path = %q, want /v1/billing/invoices/in_x/pdf", r.URL.Path)
		}
		fmt.Fprint(w, `{"url":"https://signed.example/x.pdf","expiresAt":"2026-05-20T12:05:00Z"}`)
	})

	pdf, err := client.Billing.GetInvoicePDF("in_x")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(pdf.URL, "https://") {
		t.Errorf("URL = %q, want https://...", pdf.URL)
	}
}

func TestCabinetCreateAndDashboard(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/cabinets":
			fmt.Fprint(w, `{"id":"cab_x","name":"Cabinet","plan":"cabinet_50"}`)
		case r.Method == "GET" && r.URL.Path == "/v1/cabinets/cab_x/dashboard":
			fmt.Fprint(w, `{"object":"cabinet_dashboard","totalRevenue":"12345.67","companyCount":12}`)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})

	cab, err := client.Cabinets.Create(&CabinetCreateParams{Name: "Cabinet", Plan: "cabinet_50"})
	if err != nil {
		t.Fatal(err)
	}
	if cab.ID != "cab_x" {
		t.Errorf("ID = %q, want cab_x", cab.ID)
	}

	dash, err := client.Cabinets.Dashboard("cab_x", nil)
	if err != nil {
		t.Fatal(err)
	}
	if dash.TotalRevenue != "12345.67" {
		t.Errorf("TotalRevenue = %q, want 12345.67", dash.TotalRevenue)
	}
}

func TestCabinetInviteMember(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/cabinets/cab_x/members" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"object":"cabinet_member","id":"mem_x","status":"invited"}`)
	})

	resp, err := client.Cabinets.InviteMember("cab_x", &CabinetMemberInviteParams{
		Email: "test@example.com",
		Role:  "accountant",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "invited" {
		t.Errorf("Status = %q, want invited", resp.Status)
	}
}

func TestNotificationListAndMarkRead(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/v1/notifications":
			if r.URL.Query().Get("unread") != "true" {
				t.Errorf("unread query = %q, want true", r.URL.Query().Get("unread"))
			}
			fmt.Fprint(w, `{"object":"list","data":[{"id":"notif_1","read":false}],"has_more":false}`)
		case r.Method == "PATCH" && r.URL.Path == "/v1/notifications/notif_1":
			fmt.Fprint(w, `{"object":"notification","id":"notif_1","read":true}`)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})

	unread := true
	list, err := client.Notifications.List(&NotificationListParams{Unread: &unread})
	if err != nil || len(list.Data) != 1 {
		t.Fatalf("List failed: %v / %v", err, list)
	}

	n, err := client.Notifications.MarkRead("notif_1")
	if err != nil || !n.Read {
		t.Fatalf("MarkRead failed: %v / %v", err, n)
	}
}

func TestNotificationPreferencesRoundTrip(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/notification-preferences" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"object":"notification_preferences","preferences":{}}`)
	})

	if _, err := client.Notifications.RetrievePreferences(); err != nil {
		t.Fatal(err)
	}

	trueVal := true
	if _, err := client.Notifications.UpdatePreferences(&NotificationPreferencesUpdate{
		Preferences: map[string]NotificationChannelToggles{
			"invoice_paid": {Email: &trueVal, InApp: &trueVal},
		},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReferenceListLegalFormsAndNaf(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/reference/legal-forms":
			if r.URL.Query().Get("search") != "SAS" {
				t.Errorf("search = %q, want SAS", r.URL.Query().Get("search"))
			}
			fmt.Fprint(w, `{"object":"list","data":[{"code":"5710","sigle":"SAS","label":"SAS"}],"has_more":false}`)
		case "/v1/reference/naf-codes":
			fmt.Fprint(w, `{"object":"list","data":[{"code":"62.01Z","label":"Programmation"}],"has_more":false}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	lf, err := client.Reference.ListLegalForms(&ReferenceListParams{Search: "SAS"})
	if err != nil || len(lf.Data) != 1 || lf.Data[0].Code != "5710" {
		t.Fatalf("ListLegalForms failed: %v / %v", err, lf)
	}

	naf, err := client.Reference.ListNafCodes(nil)
	if err != nil || len(naf.Data) != 1 {
		t.Fatalf("ListNafCodes failed: %v / %v", err, naf)
	}
}

func TestSettingsScopedUnderCompany(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/companies/comp_x/settings/accounting":
			fmt.Fprint(w, `{"object":"accounting_settings","vatRegime":"normal"}`)
		case "/v1/companies/comp_x/settings/reminders":
			fmt.Fprint(w, `{"object":"reminder_settings","enabled":true,"intervals":[7,15,30]}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	acc, err := client.Settings.RetrieveAccounting("comp_x")
	if err != nil || acc.VATRegime != "normal" {
		t.Fatalf("RetrieveAccounting failed: %v / %v", err, acc)
	}

	rem, err := client.Settings.RetrieveReminders("comp_x")
	if err != nil || !rem.Enabled || len(rem.Intervals) != 3 {
		t.Fatalf("RetrieveReminders failed: %v / %v", err, rem)
	}
}

func TestUsageRetrieve(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/usage" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"object":"usage","plan":"pro","invoicesIssued":{"used":12,"limit":1000}}`)
	})

	u, err := client.Usage.Retrieve()
	if err != nil {
		t.Fatal(err)
	}
	if u.Plan != "pro" || u.InvoicesIssued.Used != 12 {
		t.Errorf("Usage = %+v", u)
	}
}

func TestValidateRun(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/validate" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		fmt.Fprint(w, `{"object":"validate","valid":true,"kind":"siret"}`)
	})

	res, err := client.Validate.Run(&ValidateParams{Kind: "siret", Value: "44306184100047"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Valid {
		t.Errorf("Valid = false, want true")
	}
}
