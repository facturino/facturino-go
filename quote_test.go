package facturino

import (
	"fmt"
	"net/http"
	"testing"
)

func TestQuoteClone(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/v1/quotes/quo_123/clone" {
			t.Errorf("Path = %q, want /v1/quotes/quo_123/clone", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"id":"quo_456","object":"quote","status":"draft"}`)
	})

	q, err := client.Quotes.Clone("quo_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.ID != "quo_456" {
		t.Errorf("ID = %q, want %q", q.ID, "quo_456")
	}
	if q.Status != "draft" {
		t.Errorf("Status = %q, want %q", q.Status, "draft")
	}
}
