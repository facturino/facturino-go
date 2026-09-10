package facturino

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const webhookTolerance = 5 * time.Minute

// WebhookEvent is a webhook event delivered by Facturino.
type WebhookEvent struct {
	ID         string                 `json:"id"`
	Object     string                 `json:"object"`
	Type       string                 `json:"type"`
	APIVersion string                 `json:"apiVersion"`
	Livemode   bool                   `json:"livemode"`
	Created    string                 `json:"created"`
	Data       map[string]interface{} `json:"data"`
	Request    *WebhookRequest        `json:"request,omitempty"`
}

// DocumentData reads the document projection while retaining all original data.
func (e *WebhookEvent) DocumentData() (*DocumentEventData, error) {
	return decodeDocumentEventData(e.Data)
}

// WebhookRequest holds metadata about the originating API request.
type WebhookRequest struct {
	ID             string `json:"id,omitempty"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// VerifyWebhookSignature verifies the HMAC-SHA256 signature of a webhook payload
// and returns the parsed event. Signature header format: t=<timestamp>,v1=<hmac>.
func VerifyWebhookSignature(payload []byte, signatureHeader string, secret string) (*WebhookEvent, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("facturino: webhook payload is empty")
	}
	if signatureHeader == "" {
		return nil, fmt.Errorf("facturino: webhook signature header is empty")
	}
	if secret == "" {
		return nil, fmt.Errorf("facturino: webhook secret is empty")
	}

	// Parse signature header: t=<timestamp>,v1=<signature>
	parts := strings.Split(signatureHeader, ",")
	var timestamp string
	var signatures []string

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			signatures = append(signatures, kv[1])
		}
	}

	if timestamp == "" {
		return nil, fmt.Errorf("facturino: missing timestamp in webhook signature header")
	}
	if len(signatures) == 0 {
		return nil, fmt.Errorf("facturino: missing signature in webhook signature header")
	}

	// Parse timestamp
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("facturino: invalid timestamp in webhook signature: %w", err)
	}

	// Verify timestamp is within tolerance (rejects both old and future timestamps)
	eventTime := time.Unix(ts, 0)
	age := time.Since(eventTime)
	if age < -webhookTolerance || age > webhookTolerance {
		return nil, fmt.Errorf("facturino: webhook timestamp is outside tolerance (%s old)", age.Round(time.Second))
	}

	// Compute expected signature: HMAC-SHA256(secret, timestamp.payload)
	signedPayload := fmt.Sprintf("%s.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Timing-safe comparison against all provided signatures
	verified := false
	for _, sig := range signatures {
		sigBytes, err := hex.DecodeString(sig)
		if err != nil {
			continue
		}
		expectedBytes, _ := hex.DecodeString(expectedSig)
		if hmac.Equal(sigBytes, expectedBytes) {
			verified = true
			break
		}
	}

	if !verified {
		return nil, fmt.Errorf("facturino: webhook signature verification failed")
	}

	// Parse the event
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("facturino: failed to parse webhook event: %w", err)
	}

	return &event, nil
}

// GenerateWebhookSignature produces a signature header for testing.
func GenerateWebhookSignature(payload []byte, secret string, t time.Time) string {
	timestamp := strconv.FormatInt(t.Unix(), 10)
	signedPayload := fmt.Sprintf("%s.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%s,v1=%s", timestamp, sig)
}
