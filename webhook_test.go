package facturino

import (
	"encoding/json"
	"testing"
	"time"
)

func TestVerifyWebhookSignature(t *testing.T) {
	secret := "whsec_test_secret_123"
	payload := []byte(`{"id":"evt_123","object":"event","type":"invoice.finalized","data":{"id":"inv_abc"}}`)
	now := time.Now()

	sig := GenerateWebhookSignature(payload, secret, now)

	event, err := VerifyWebhookSignature(payload, sig, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.ID != "evt_123" {
		t.Errorf("ID = %q, want %q", event.ID, "evt_123")
	}
	if event.Type != "invoice.finalized" {
		t.Errorf("Type = %q, want %q", event.Type, "invoice.finalized")
	}
}

func TestVerifyWebhookSignatureInvalidSecret(t *testing.T) {
	secret := "whsec_correct"
	wrongSecret := "whsec_wrong"
	payload := []byte(`{"id":"evt_123","object":"event","type":"invoice.paid"}`)
	now := time.Now()

	sig := GenerateWebhookSignature(payload, wrongSecret, now)

	_, err := VerifyWebhookSignature(payload, sig, secret)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestVerifyWebhookSignatureExpired(t *testing.T) {
	secret := "whsec_test_secret"
	payload := []byte(`{"id":"evt_123","object":"event","type":"invoice.paid"}`)
	// 10 minutes ago (exceeds 5-minute tolerance)
	oldTime := time.Now().Add(-10 * time.Minute)

	sig := GenerateWebhookSignature(payload, secret, oldTime)

	_, err := VerifyWebhookSignature(payload, sig, secret)
	if err == nil {
		t.Fatal("expected error for expired signature")
	}
}

func TestVerifyWebhookSignatureEmptyPayload(t *testing.T) {
	_, err := VerifyWebhookSignature(nil, "t=123,v1=abc", "secret")
	if err == nil {
		t.Fatal("expected error for empty payload")
	}
}

func TestVerifyWebhookSignatureEmptyHeader(t *testing.T) {
	_, err := VerifyWebhookSignature([]byte("payload"), "", "secret")
	if err == nil {
		t.Fatal("expected error for empty header")
	}
}

func TestVerifyWebhookSignatureEmptySecret(t *testing.T) {
	_, err := VerifyWebhookSignature([]byte("payload"), "t=123,v1=abc", "")
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestVerifyWebhookSignatureMissingTimestamp(t *testing.T) {
	_, err := VerifyWebhookSignature([]byte("payload"), "v1=abc123", "secret")
	if err == nil {
		t.Fatal("expected error for missing timestamp")
	}
}

func TestVerifyWebhookSignatureMissingSignature(t *testing.T) {
	_, err := VerifyWebhookSignature([]byte("payload"), "t=12345", "secret")
	if err == nil {
		t.Fatal("expected error for missing signature")
	}
}

func TestVerifyWebhookSignatureInvalidTimestamp(t *testing.T) {
	_, err := VerifyWebhookSignature([]byte("payload"), "t=notanumber,v1=abc", "secret")
	if err == nil {
		t.Fatal("expected error for invalid timestamp")
	}
}

func TestGenerateWebhookSignature(t *testing.T) {
	secret := "whsec_test"
	payload := []byte(`{"test":true}`)
	ts := time.Unix(1700000000, 0)

	sig := GenerateWebhookSignature(payload, secret, ts)

	if sig == "" {
		t.Fatal("expected non-empty signature")
	}
	// Should contain t= and v1=
	if len(sig) < 10 {
		t.Error("signature seems too short")
	}
}

func TestVerifyWebhookSignatureWithEventData(t *testing.T) {
	secret := "whsec_production_key"
	event := WebhookEvent{
		ID:     "evt_full_test",
		Object: "event",
		Type:   "invoice.approved",
		Data: map[string]interface{}{
			"id":              "inv_999",
			"object":          "invoice",
			"status":          "approved",
			"previous_status": "received",
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	now := time.Now()
	sig := GenerateWebhookSignature(payload, secret, now)

	parsed, err := VerifyWebhookSignature(payload, sig, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.ID != "evt_full_test" {
		t.Errorf("ID = %q, want %q", parsed.ID, "evt_full_test")
	}
	if parsed.Type != "invoice.approved" {
		t.Errorf("Type = %q, want %q", parsed.Type, "invoice.approved")
	}
	invoiceID, ok := parsed.Data["id"].(string)
	if !ok || invoiceID != "inv_999" {
		t.Errorf("Data.id = %v, want %q", parsed.Data["id"], "inv_999")
	}
}

func TestVerifyWebhookSignatureTimingWindow(t *testing.T) {
	secret := "whsec_timing_test"
	payload := []byte(`{"id":"evt_timing"}`)

	// Just within the 5-minute window (4 minutes ago)
	recentTime := time.Now().Add(-4 * time.Minute)
	sig := GenerateWebhookSignature(payload, secret, recentTime)

	_, err := VerifyWebhookSignature(payload, sig, secret)
	if err != nil {
		t.Fatalf("signature from 4 minutes ago should be valid: %v", err)
	}

	// Just outside the window (6 minutes ago)
	oldTime := time.Now().Add(-6 * time.Minute)
	sig2 := GenerateWebhookSignature(payload, secret, oldTime)

	_, err = VerifyWebhookSignature(payload, sig2, secret)
	if err == nil {
		t.Fatal("signature from 6 minutes ago should be rejected")
	}
}
