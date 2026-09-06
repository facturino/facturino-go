package facturino

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestEventRetryToEndpointReceipt(t *testing.T) {
	var body map[string]string
	client, _ := testServerWithRetries(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"evt_1","object":"event","retryScheduled":true,"endpointId":"we_a"}`)
	}, 1)

	receipt, err := client.Events.RetryToEndpoint("evt_1", "we_a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body["endpointId"] != "we_a" {
		t.Errorf("request body endpointId = %q, want we_a", body["endpointId"])
	}
	if !receipt.RetryScheduled || receipt.EndpointID != "we_a" || receipt.ID != "evt_1" {
		t.Errorf("receipt = %+v", receipt)
	}

	legacy, err := client.Events.Retry("evt_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if legacy.ID != "evt_1" || legacy.Object != "event" {
		t.Errorf("legacy receipt = %+v", legacy)
	}
}
