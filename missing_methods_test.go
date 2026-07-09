package facturino

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// Tests for:
//   - Account.RequestExport / DownloadExport
//   - Company.Create / AddMilestone
//   - Invoice.CreatePortalLink

func TestAccountRequestAndDownloadExport(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/account/export":
			fmt.Fprint(w, `{"object":"account_export","exportId":"rgpdexp_abc","status":"pending","message":"Preparing..."}`)
		case "/v1/account/exports/rgpdexp_abc/download":
			fmt.Fprint(w, `{"url":"https://signed.example/export.zip","expiresAt":"2026-05-20T12:05Z"}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	started, err := client.Account.RequestExport()
	if err != nil {
		t.Fatal(err)
	}
	if started.ExportID != "rgpdexp_abc" {
		t.Errorf("ExportID = %q, want rgpdexp_abc", started.ExportID)
	}

	dl, err := client.Account.DownloadExport(started.ExportID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(dl.URL, "https://") {
		t.Errorf("URL = %q, want https://...", dl.URL)
	}
}

func TestCompanyCreate(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/companies" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		fmt.Fprint(w, `{"id":"comp_new","name":"ACME SAS","siret":"44306184100047"}`)
	})

	c, err := client.Companies.Create(&CompanyCreateParams{
		Name:  "ACME SAS",
		SIRET: "44306184100047",
		Address: &Address{
			Line1:      "12 rue de Rivoli",
			PostalCode: "75001",
			City:       "Paris",
			Country:    "FR",
		},
		VATRegime: "normal",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "comp_new" {
		t.Errorf("ID = %q, want comp_new", c.ID)
	}
}

func TestCompanyAddMilestone(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/companies/comp_x/milestones" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		fmt.Fprint(w, `{"object":"company_milestone","milestone":"first_invoice_sent","reachedAt":"2026-05-20T12:00:00Z"}`)
	})

	resp, err := client.Companies.AddMilestone("comp_x", "first_invoice_sent")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Milestone != "first_invoice_sent" {
		t.Errorf("Milestone = %q", resp.Milestone)
	}
}

func TestInvoiceCreatePortalLink(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/invoices/inv_x/portal-link" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		fmt.Fprint(w, `{"url":"https://facturino.com/portal/inv_x?token=x","token":"plt_secret","expires_at":"2026-05-21T12:00:00Z"}`)
	})

	link, err := client.Invoices.CreatePortalLink("inv_x")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(link.URL, "/portal/") {
		t.Errorf("URL = %q, want to contain /portal/", link.URL)
	}
	if link.Token != "plt_secret" {
		t.Errorf("Token = %q", link.Token)
	}
}
