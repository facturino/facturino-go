package facturino

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func testServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := New("fac_test_testkey123", WithBaseURL(server.URL), WithMaxRetries(0))
	return client, server
}

func testServerWithRetries(t *testing.T, handler http.HandlerFunc, maxRetries int) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := New("fac_test_testkey123", WithBaseURL(server.URL), WithMaxRetries(maxRetries))
	return client, server
}

func TestNew(t *testing.T) {
	client := New("fac_test_xxx")
	if client == nil {
		t.Fatal("New() returned nil")
	}
	if client.Invoices == nil {
		t.Error("Invoices service is nil")
	}
	if client.Customers == nil {
		t.Error("Customers service is nil")
	}
	if client.Products == nil {
		t.Error("Products service is nil")
	}
	if client.Quotes == nil {
		t.Error("Quotes service is nil")
	}
	if client.CreditNotes == nil {
		t.Error("CreditNotes service is nil")
	}
	if client.Events == nil {
		t.Error("Events service is nil")
	}
	if client.WebhookEndpoints == nil {
		t.Error("WebhookEndpoints service is nil")
	}
	if client.RecurringInvoices == nil {
		t.Error("RecurringInvoices service is nil")
	}
	if client.Companies == nil {
		t.Error("Companies service is nil")
	}
	if client.Exports == nil {
		t.Error("Exports service is nil")
	}
	if client.EReporting == nil {
		t.Error("EReporting service is nil")
	}
	if client.Jobs == nil {
		t.Error("Jobs service is nil")
	}
	if client.Sandbox == nil {
		t.Error("Sandbox service is nil")
	}
	if client.Payments == nil {
		t.Error("Payments service is nil")
	}
}

func TestNewWithOptions(t *testing.T) {
	httpCl := &http.Client{}
	client := New("fac_live_xxx",
		WithBaseURL("https://custom.example.com/api"),
		WithHTTPClient(httpCl),
		WithMaxRetries(5),
	)
	if client == nil {
		t.Fatal("New() with options returned nil")
	}
	if client.client.baseURL != "https://custom.example.com/api" {
		t.Errorf("baseURL = %q, want %q", client.client.baseURL, "https://custom.example.com/api")
	}
	if client.client.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want %d", client.client.maxRetries, 5)
	}
}

func TestAuthorizationHeader(t *testing.T) {
	var receivedAuth string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice"}`)
	})

	_, err := client.Invoices.Get("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedAuth != "Bearer fac_test_testkey123" {
		t.Errorf("Authorization header = %q, want %q", receivedAuth, "Bearer fac_test_testkey123")
	}
}

func TestUserAgent(t *testing.T) {
	var receivedUA string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice"}`)
	})

	_, err := client.Invoices.Get("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "facturino-go/" + sdkVersion
	if receivedUA != expected {
		t.Errorf("User-Agent = %q, want %q", receivedUA, expected)
	}
}

func TestIdempotencyKeyHeader(t *testing.T) {
	var receivedKey string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedKey = r.Header.Get("Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice","status":"draft"}`)
	})

	_, err := client.Invoices.Create(&InvoiceParams{
		Customer:       "cus_123",
		TaxDecisionID:  "taxdec_9c1f",
		DecisionLines:  []*DecisionLineParams{{TaxLineRef: "l1", Unit: "unit"}},
		IdempotencyKey: "idem_test_123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedKey != "idem_test_123" {
		t.Errorf("Idempotency-Key = %q, want %q", receivedKey, "idem_test_123")
	}
}

func TestErrorParsing(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":{"type":"not_found_error","code":"resource_missing","message":"Invoice not found","request_id":"req_abc123"}}`)
	})

	_, err := client.Invoices.Get("inv_nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Type != ErrorTypeNotFound {
		t.Errorf("Type = %q, want %q", apiErr.Type, ErrorTypeNotFound)
	}
	if apiErr.Code != "resource_missing" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "resource_missing")
	}
	if apiErr.RequestID != "req_abc123" {
		t.Errorf("RequestID = %q, want %q", apiErr.RequestID, "req_abc123")
	}
	if apiErr.HTTPStatusCode != 404 {
		t.Errorf("HTTPStatusCode = %d, want %d", apiErr.HTTPStatusCode, 404)
	}
}

// A refusal keeps ONE main code — that is what a caller branches on. When it
// rests on a more precise reason, `issues` publishes it as data, with the field
// in cause, instead of joining it into the message.
func TestErrorIssues(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		fmt.Fprint(w, `{"error":{"type":"validation_error","code":"validation_error",`+
			`"message":"Buyer territory could not be resolved.","param":"customerId",`+
			`"request_id":"req_abc","issues":[{"code":"invalid_postal_code",`+
			`"param":"customer.address.postalCode","message":"Buyer territory could not be resolved."}]}}`)
	})

	_, err := client.Invoices.Get("inv_1")
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Code != "validation_error" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "validation_error")
	}
	if apiErr.Param != "customerId" {
		t.Errorf("Param = %q, want %q", apiErr.Param, "customerId")
	}
	if len(apiErr.Issues) != 1 {
		t.Fatalf("len(Issues) = %d, want 1", len(apiErr.Issues))
	}
	if apiErr.Issues[0].Code != "invalid_postal_code" {
		t.Errorf("Issues[0].Code = %q, want %q", apiErr.Issues[0].Code, "invalid_postal_code")
	}
	if apiErr.Issues[0].Param != "customer.address.postalCode" {
		t.Errorf("Issues[0].Param = %q, want %q", apiErr.Issues[0].Param, "customer.address.postalCode")
	}
}

// A refusal with nothing more to say carries no issues at all.
func TestErrorWithoutIssues(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		fmt.Fprint(w, `{"error":{"type":"not_found_error","code":"resource_missing","message":"Invoice not found","request_id":"req_abc"}}`)
	})

	_, err := client.Invoices.Get("inv_1")
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if len(apiErr.Issues) != 0 {
		t.Errorf("len(Issues) = %d, want 0", len(apiErr.Issues))
	}
	if apiErr.Issues == nil {
		t.Errorf("Issues is nil, want an empty non-nil slice")
	}
}

// A response that is not the JSON envelope at all — a gateway HTML page — still
// yields the same contract: reading Issues never needs a nil check.
func TestErrorIssuesOnNonJSONResponse(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(502)
		fmt.Fprint(w, `<html>bad gateway</html>`)
	})

	_, err := client.Invoices.Get("inv_1")
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Code != "unknown" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "unknown")
	}
	if apiErr.Issues == nil {
		t.Fatalf("Issues is nil, want an empty non-nil slice")
	}
	if len(apiErr.Issues) != 0 {
		t.Errorf("len(Issues) = %d, want 0", len(apiErr.Issues))
	}
}

func TestErrorHelpers(t *testing.T) {
	notFoundErr := &Error{Type: ErrorTypeNotFound}
	rateLimitErr := &Error{Type: ErrorTypeRateLimit}
	authErr := &Error{Type: ErrorTypeAuthentication}

	if !IsNotFound(notFoundErr) {
		t.Error("IsNotFound should return true")
	}
	if IsNotFound(rateLimitErr) {
		t.Error("IsNotFound should return false for rate limit error")
	}
	if !IsRateLimit(rateLimitErr) {
		t.Error("IsRateLimit should return true")
	}
	if !IsAuthentication(authErr) {
		t.Error("IsAuthentication should return true")
	}
	if IsNotFound(fmt.Errorf("random error")) {
		t.Error("IsNotFound should return false for non-API error")
	}
}

func TestRetryOn429(t *testing.T) {
	var callCount int32
	client, _ := testServerWithRetries(t, func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count <= 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			fmt.Fprint(w, `{"error":{"type":"rate_limit_error","code":"rate_limit","message":"Too many requests","request_id":"req_xyz"}}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice","status":"draft"}`)
	}, 3)

	inv, err := client.Invoices.Get("inv_123")
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if inv.ID != "inv_123" {
		t.Errorf("ID = %q, want %q", inv.ID, "inv_123")
	}
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("call count = %d, want 3", atomic.LoadInt32(&callCount))
	}
}

func TestRetryOn500(t *testing.T) {
	var callCount int32
	client, _ := testServerWithRetries(t, func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count == 1 {
			w.WriteHeader(500)
			fmt.Fprint(w, `{"error":{"type":"api_error","code":"internal","message":"Internal error","request_id":"req_err"}}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"cus_abc","object":"customer","name":"ACME"}`)
	}, 2)

	cus, err := client.Customers.Get("cus_abc")
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if cus.Name != "ACME" {
		t.Errorf("Name = %q, want %q", cus.Name, "ACME")
	}
}

func TestNoRetryOn400(t *testing.T) {
	var callCount int32
	client, _ := testServerWithRetries(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(400)
		fmt.Fprint(w, `{"error":{"type":"invalid_request_error","code":"invalid_field","message":"Bad request","request_id":"req_bad"}}`)
	}, 3)

	_, err := client.Invoices.Get("inv_123")
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("call count = %d, want 1 (no retries on 400)", atomic.LoadInt32(&callCount))
	}
}

func TestDelete204(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %q, want DELETE", r.Method)
		}
		w.WriteHeader(204)
	})

	err := client.Invoices.Delete("inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPaginationIterator(t *testing.T) {
	page := 0
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case 1:
			fmt.Fprint(w, `{
				"object":"list",
				"url":"/v1/invoices",
				"data":[
					{"id":"inv_1","object":"invoice","status":"draft"},
					{"id":"inv_2","object":"invoice","status":"finalized"}
				],
				"has_more":true,
				"next_cursor":"inv_2"
			}`)
		case 2:
			fmt.Fprint(w, `{
				"object":"list",
				"url":"/v1/invoices",
				"data":[
					{"id":"inv_3","object":"invoice","status":"paid"}
				],
				"has_more":false,
				"next_cursor":null
			}`)
		default:
			t.Error("unexpected extra page request")
			w.WriteHeader(500)
		}
	})

	iter := client.Invoices.List(&InvoiceListParams{ListParams: ListParams{Limit: 2}})
	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Invoice().ID)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("got %d invoices, want 3", len(ids))
	}
	expected := []string{"inv_1", "inv_2", "inv_3"}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("ids[%d] = %q, want %q", i, id, expected[i])
		}
	}
}

func TestPaginationEmpty(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"object":"list","url":"/v1/invoices","data":[],"has_more":false,"next_cursor":null}`)
	})

	iter := client.Invoices.List(nil)
	count := 0
	for iter.Next() {
		count++
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if count != 0 {
		t.Errorf("got %d items, want 0", count)
	}
}

func TestListParamsToValues(t *testing.T) {
	tests := []struct {
		name   string
		params *ListParams
		want   map[string]string
	}{
		{
			name:   "nil params",
			params: nil,
			want:   map[string]string{},
		},
		{
			name:   "empty params",
			params: &ListParams{},
			want:   map[string]string{},
		},
		{
			name: "all params",
			params: &ListParams{
				Limit:          50,
				StartingAfter:  "inv_abc",
				Status:         "paid",
				IncludeDeleted: true,
			},
			want: map[string]string{
				"limit":           "50",
				"starting_after":  "inv_abc",
				"status":          "paid",
				"include_deleted": "true",
			},
		},
		{
			name:   "limit capped at 100",
			params: &ListParams{Limit: 200},
			want:   map[string]string{"limit": "100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals := tt.params.toValues()
			for k, v := range tt.want {
				if got := vals.Get(k); got != v {
					t.Errorf("toValues()[%q] = %q, want %q", k, got, v)
				}
			}
		})
	}
}

func TestErrorString(t *testing.T) {
	err := &Error{
		Type:      ErrorTypeNotFound,
		Code:      "resource_missing",
		Message:   "Invoice not found",
		RequestID: "req_123",
	}
	s := err.Error()
	if s == "" {
		t.Error("Error() returned empty string")
	}

	errWithParam := &Error{
		Type:      ErrorTypeValidation,
		Code:      "invalid_field",
		Message:   "Invalid email",
		Param:     "email",
		RequestID: "req_456",
	}
	s2 := errWithParam.Error()
	if s2 == "" {
		t.Error("Error() with param returned empty string")
	}
}

func TestBaseURLTrailingSlash(t *testing.T) {
	client := New("fac_test_key", WithBaseURL("https://example.com/api/"))
	if client.client.baseURL != "https://example.com/api" {
		t.Errorf("baseURL = %q, trailing slash should be stripped", client.client.baseURL)
	}
}

func TestContentTypeHeader(t *testing.T) {
	var receivedCT string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		receivedCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"cus_123","object":"customer"}`)
	})

	_, err := client.Customers.Create(&CustomerParams{
		Name: "Test",
		Type: "company",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedCT != "application/json" {
		t.Errorf("Content-Type = %q, want %q", receivedCT, "application/json")
	}
}

func TestMalformedJSON(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{invalid json}`)
	})

	_, err := client.Invoices.Get("inv_123")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestMalformedErrorResponse(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		fmt.Fprint(w, `not json at all`)
	})

	_, err := client.Invoices.Get("inv_123")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Code != "unknown" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "unknown")
	}
}

func TestRequestBodySerialization(t *testing.T) {
	var receivedBody map[string]interface{}
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"inv_123","object":"invoice","status":"draft"}`)
	})

	_, err := client.Invoices.Create(&InvoiceParams{
		Customer:      "cus_abc",
		TaxDecisionID: "taxdec_9c1f",
		DecisionLines: []*DecisionLineParams{{TaxLineRef: "service", Unit: "unit"}},
		Notes:         "test note",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody["customerId"] != "cus_abc" {
		t.Errorf("customerId = %v, want %q", receivedBody["customerId"], "cus_abc")
	}
	if receivedBody["notes"] != "test note" {
		t.Errorf("notes = %v, want %q", receivedBody["notes"], "test note")
	}
	lines, ok := receivedBody["decisionLines"].([]interface{})
	if !ok || len(lines) != 1 {
		t.Fatalf("lines should have 1 item, got %v", receivedBody["lines"])
	}
}
