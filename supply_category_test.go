package facturino

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// supplyCategoryFixture reads SDK-only examples, independent of the shared corpus.
func supplyCategoryFixture(t *testing.T, name string) json.RawMessage {
	t.Helper()
	raw, err := os.ReadFile("testdata/sdk/supply-category.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus map[string]json.RawMessage
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatal(err)
	}
	body, ok := corpus[name]
	if !ok {
		t.Fatalf("missing supply-category fixture %q", name)
	}
	return body
}

func assertSupplyFields(t *testing.T, value interface{}, kind, category string) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"kind": kind, "supplyCategory": category} {
		got, present := fields[key]
		if want == "" {
			if present {
				t.Errorf("%s = %s, want omitted", key, got)
			}
		} else if string(got) != `"`+want+`"` {
			t.Errorf("%s = %s, want %q", key, got, want)
		}
	}
}

func TestProductSupplyFieldsTransport(t *testing.T) {
	for _, category := range []string{"goods", "services", "electronically_supplied_services", "deposit", "ancillary_costs", "future_supply"} {
		for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPatch} {
			t.Run(method+"/"+category, func(t *testing.T) {
				kind := "service"
				if category == "goods" {
					kind = "goods"
				}
				var response map[string]interface{}
				if err := json.Unmarshal(supplyCategoryFixture(t, "product"), &response); err != nil {
					t.Fatal(err)
				}
				response["kind"], response["supplyCategory"] = kind, category
				if method == http.MethodPatch {
					response["active"] = false
				}
				transport := contractRoundTripper(func(r *http.Request) (*http.Response, error) {
					wantPath := "/api/v1/products/prod_supply"
					if method == http.MethodPost {
						wantPath = "/api/v1/products"
					}
					if r.Method != method || r.URL.Path != wantPath {
						t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, method, wantPath)
					}
					if method != http.MethodGet {
						var fields map[string]json.RawMessage
						if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
							t.Error(err)
						}
						assertSupplyFields(t, fields, kind, category)
						if string(fields["reference"]) != `"SUB-ONLINE"` || string(fields["category"]) != `"subscriptions"` {
							t.Error("product reference or catalog category lost")
						}
						if method == http.MethodPatch && string(fields["active"]) != "false" {
							t.Errorf("active = %s, want false", fields["active"])
						}
					}
					w := httptest.NewRecorder()
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(response); err != nil {
						t.Error(err)
					}
					return w.Result(), nil
				})
				client := New("fac_test_supply", WithHTTPClient(&http.Client{Transport: transport}), WithMaxRetries(0))
				var product *Product
				var err error
				switch method {
				case http.MethodGet:
					product, err = client.Products.Get("prod_supply")
				case http.MethodPost:
					product, err = client.Products.Create(&ProductParams{
						Name: "Online subscription", Reference: "SUB-ONLINE", Category: "subscriptions",
						Kind: kind, SupplyCategory: category,
					})
				case http.MethodPatch:
					inactive := false
					product, err = client.Products.Update("prod_supply", &ProductUpdateParams{
						Reference: "SUB-ONLINE", Category: "subscriptions", Active: &inactive,
						Kind: kind, SupplyCategory: category,
					})
				}
				if err != nil {
					t.Fatal(err)
				}
				if product.Kind != kind || product.SupplyCategory != category {
					t.Fatalf("product = %+v, want kind %q and supply category %q", product, kind, category)
				}
				if method == http.MethodPatch && product.Active {
					t.Error("updated product must remain inactive")
				}
				assertSupplyFields(t, product, kind, category)
			})
		}
	}
}

func TestQuoteSupplyFieldsTransport(t *testing.T) {
	for _, category := range []string{"electronically_supplied_services", "future_supply", ""} {
		for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPatch} {
			name := category
			if name == "" {
				name = "undeclared"
			}
			t.Run(method+"/"+name, func(t *testing.T) {
				var response map[string]interface{}
				if err := json.Unmarshal(supplyCategoryFixture(t, "quote"), &response); err != nil {
					t.Fatal(err)
				}
				line := response["items"].([]interface{})[0].(map[string]interface{})
				if category == "" {
					delete(line, "supplyCategory")
				} else {
					line["supplyCategory"] = category
				}
				transport := contractRoundTripper(func(r *http.Request) (*http.Response, error) {
					wantPath := "/api/v1/quotes/quo_supply"
					if method == http.MethodPost {
						wantPath = "/api/v1/quotes"
					}
					if r.Method != method || r.URL.Path != wantPath {
						t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, method, wantPath)
					}
					if method != http.MethodGet {
						var body struct {
							Lines []map[string]json.RawMessage `json:"lines"`
						}
						if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
							t.Error(err)
						}
						if len(body.Lines) != 1 {
							t.Errorf("lines = %+v, want one line", body.Lines)
						} else {
							assertSupplyFields(t, body.Lines[0], "service", category)
						}
					}
					w := httptest.NewRecorder()
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(response); err != nil {
						t.Error(err)
					}
					return w.Result(), nil
				})
				client := New("fac_test_supply", WithHTTPClient(&http.Client{Transport: transport}), WithMaxRetries(0))
				items := []*ItemParams{{
					Description: "Online subscription", Quantity: "1", Unit: "month", UnitPrice: 2900,
					VATRate: 2000, VATCode: "S", Kind: "service", SupplyCategory: category,
				}}
				var quote *Quote
				var err error
				switch method {
				case http.MethodGet:
					quote, err = client.Quotes.Get("quo_supply")
				case http.MethodPost:
					quote, err = client.Quotes.Create(&QuoteParams{Customer: "cus_supply", Items: items})
				case http.MethodPatch:
					quote, err = client.Quotes.Update("quo_supply", &QuoteUpdateParams{Items: items})
				}
				if err != nil {
					t.Fatal(err)
				}
				if len(quote.Items) != 1 || quote.Items[0].Kind != "service" || quote.Items[0].SupplyCategory != category {
					t.Fatalf("quote items = %+v, want one service with supply category %q", quote.Items, category)
				}
				assertSupplyFields(t, quote.Items[0], "service", category)
			})
		}
	}
}

func TestQuoteConversionPreservesSupplyCategory(t *testing.T) {
	for _, category := range []string{"electronically_supplied_services", "future_supply", ""} {
		name := category
		if name == "" {
			name = "undeclared"
		}
		t.Run(name, func(t *testing.T) {
			var body map[string]interface{}
			if err := json.Unmarshal(supplyCategoryFixture(t, "convertedInvoice"), &body); err != nil {
				t.Fatal(err)
			}
			item := body["items"].([]interface{})[0].(map[string]interface{})
			draftLine := body["commercialDraft"].(map[string]interface{})["lines"].([]interface{})[0].(map[string]interface{})
			for _, line := range []map[string]interface{}{item, draftLine} {
				if category == "" {
					delete(line, "supplyCategory")
				} else {
					line["supplyCategory"] = category
				}
			}
			invoice, err := serveJSON(t, body).Quotes.Convert("quo_supply")
			if err != nil {
				t.Fatal(err)
			}
			if invoice.TaxSource != "" || invoice.CommercialDraft == nil || len(invoice.CommercialDraft.Lines) != 1 || len(invoice.Items) != 1 {
				t.Fatalf("converted invoice lost its commercial draft: %+v", invoice)
			}
			if invoice.Items[0].SupplyCategory != category || invoice.CommercialDraft.Lines[0].SupplyCategory != category {
				t.Fatalf("converted category = %q / %q, want %q", invoice.Items[0].SupplyCategory, invoice.CommercialDraft.Lines[0].SupplyCategory, category)
			}
			assertSupplyFields(t, invoice.CommercialDraft.Lines[0], "", category)
		})
	}
}

func TestSupplyFieldsHistoricalOmissions(t *testing.T) {
	product, err := corpusClient(supplyCategoryFixture(t, "historicalProduct")).Products.Get("prod_historical")
	if err != nil {
		t.Fatal(err)
	}
	quote, err := corpusClient(supplyCategoryFixture(t, "historicalQuote")).Quotes.Get("quo_historical")
	if err != nil {
		t.Fatal(err)
	}
	invoice, err := corpusClient(supplyCategoryFixture(t, "historicalInvoice")).Quotes.Convert("quo_historical")
	if err != nil {
		t.Fatal(err)
	}
	if product.Kind != "" || product.SupplyCategory != "" || len(quote.Items) != 1 ||
		len(invoice.Items) != 1 || invoice.CommercialDraft == nil || len(invoice.CommercialDraft.Lines) != 1 {
		t.Fatal("historical resources lost their structure or acquired a product supply category")
	}
	if quote.Items[0].Kind != "" || quote.Items[0].SupplyCategory != "" ||
		invoice.Items[0].Kind != "" || invoice.Items[0].SupplyCategory != "" || invoice.CommercialDraft.Lines[0].SupplyCategory != "" {
		t.Fatal("invented a historical line's kind or supply category")
	}
	for name, value := range map[string]interface{}{
		"product": product, "product create": &ProductParams{Name: "Historical service"},
		"product update": &ProductUpdateParams{}, "quote line": quote.Items[0],
		"quote input": &ItemParams{}, "invoice line": invoice.Items[0],
		"commercial draft line": invoice.CommercialDraft.Lines[0],
	} {
		t.Run(name, func(t *testing.T) { assertSupplyFields(t, value, "", "") })
	}
}

func TestSupplyCategoryDecisionResponses(t *testing.T) {
	for _, name := range []string{"pendingDecision", "finalDecision", "historicalDecision"} {
		t.Run(name, func(t *testing.T) {
			decision, err := corpusClient(supplyCategoryFixture(t, name)).TaxDecisions.Get("taxdec_supply")
			if err != nil {
				t.Fatal(err)
			}
			if len(decision.Lines) != 1 {
				t.Fatalf("Lines = %+v, want one line", decision.Lines)
			}
			line := decision.Lines[0]
			if name == "pendingDecision" {
				if decision.Status != "pending_verification" || line.Category != nil || decision.Totals != nil || decision.AmountToCharge != nil ||
					line.AmountHT != nil || line.AmountVAT != nil || line.AmountTTC != nil || line.RateCentipercent != nil {
					t.Fatalf("pending decision invented a category or amounts: %+v / %+v", decision, line)
				}
				if len(decision.Issues) != 1 || decision.Issues[0].Code != "supply_category_not_declared" ||
					decision.Issues[0].LineID != line.Reference || len(decision.Notices) != 0 {
					t.Fatalf("pending issue = %+v, notices = %+v", decision.Issues, decision.Notices)
				}
			} else {
				if !decision.IsFinal() || line.Category == nil || *line.Category != "services" || decision.Totals == nil ||
					decision.AmountToCharge == nil || *decision.AmountToCharge != 12000 || line.AmountTTC == nil || *line.AmountTTC != 12000 {
					t.Fatalf("final decision lost its category or amounts: %+v / %+v", decision, line)
				}
				if len(decision.Issues) != 0 {
					t.Errorf("final Issues = %+v, want none", decision.Issues)
				}
				if name == "finalDecision" {
					if len(decision.Notices) != 1 || decision.Notices[0].Code != "supply_category_not_declared" || decision.Notices[0].LineID != line.Reference {
						t.Fatalf("final notices = %+v, want a line-specific supply_category_not_declared", decision.Notices)
					}
				} else if len(decision.Notices) != 0 {
					t.Errorf("invented historical notices: %+v", decision.Notices)
				}
			}
			raw, err := json.Marshal(line)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			want := `"services"`
			if name == "pendingDecision" {
				want = "null"
			}
			if string(fields["category"]) != want {
				t.Errorf("serialized category = %s, want %s", fields["category"], want)
			}
		})
	}
}

func TestTaxDecisionInputCategoryRemainsRequired(t *testing.T) {
	for _, source := range []string{"facturino", "integration"} {
		for _, category := range []string{"services", ""} {
			name := category
			if name == "" {
				name = "empty"
			}
			t.Run(source+"/"+name, func(t *testing.T) {
				params := decisionParams()
				params.TaxSource = source
				params.Lines[0].Category = category
				raw, err := json.Marshal(params)
				if err != nil {
					t.Fatal(err)
				}
				var body struct {
					Lines []map[string]json.RawMessage `json:"lines"`
				}
				if err := json.Unmarshal(raw, &body); err != nil {
					t.Fatal(err)
				}
				if len(body.Lines) != 1 {
					t.Fatalf("Lines = %+v, want one line", body.Lines)
				}
				if got, present := body.Lines[0]["category"]; !present || string(got) != `"`+category+`"` {
					t.Errorf("input category = %s, present = %v; want the required string %q", got, present, category)
				}
			})
		}
	}
}
