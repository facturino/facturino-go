package facturino

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductVATEXCodeTransport(t *testing.T) {
	for _, code := range []string{
		"VATEX-FR-CGI261-4", "VATEX-FR-CGI261-7", "VATEX-FR-CGI295",
		"VATEX-FR-AE", "VATEX-EU-132-1I", "VATEX-EU-I", "VATEX-EU-J",
		"VATEX-FR-261", "VATEX-FR-AUTOLIQ", "VATEX-FUTURE-CODE",
	} {
		for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPatch} {
			t.Run(method+"/"+code, func(t *testing.T) {
				// Exercise the SDK transport without opening a network listener.
				transport := contractRoundTripper(func(r *http.Request) (*http.Response, error) {
					w := httptest.NewRecorder()
					wantPath := "/api/v1/products/prod_vatex"
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
						if string(fields["vatexCode"]) != `"`+code+`"` {
							t.Errorf("vatexCode = %s, want %q", fields["vatexCode"], code)
						}
					}
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(map[string]interface{}{
						"id": "prod_vatex", "object": "product", "vatexCode": code,
					}); err != nil {
						t.Error(err)
					}
					return w.Result(), nil
				})
				client := New("fac_test_vatex", WithHTTPClient(&http.Client{Transport: transport}), WithMaxRetries(0))
				var product *Product
				var err error
				switch method {
				case http.MethodGet:
					product, err = client.Products.Get("prod_vatex")
				case http.MethodPost:
					product, err = client.Products.Create(&ProductParams{Name: "Training", VATexCode: code})
				case http.MethodPatch:
					product, err = client.Products.Update("prod_vatex", &ProductUpdateParams{VATexCode: code})
				}
				if err != nil {
					t.Fatal(err)
				}
				if product == nil || product.VATexCode != code {
					t.Errorf("product = %+v, want vatexCode %q", product, code)
				}
			})
		}
	}
}
