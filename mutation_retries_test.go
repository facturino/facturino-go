package facturino

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestMutationCommittedBeforeResponseLoss(t *testing.T) {
	var mu sync.Mutex
	movements := map[string]int{}
	var keys []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		mu.Lock()
		defer mu.Unlock()
		key := r.Header.Get("Idempotency-Key")
		keys = append(keys, key)
		if _, exists := movements[key]; !exists {
			movements[key] = len(movements) + 1
		}
		if len(keys) == 1 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			conn.Close()
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]int{"id": movements[key]})
	}))
	defer server.Close()
	client := New("fac_test_local", WithBaseURL(server.URL))
	var result map[string]int
	if err := client.client.post("/payments", map[string]int{"amount": 1188}, &result, nil); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	if result["id"] != 1 || len(movements) != 1 || len(keys) != 2 || keys[0] == "" || keys[0] != keys[1] {
		t.Fatalf("unexpected replay: %v %v", result, keys)
	}
	mu.Unlock()
	if err := client.client.post("/payments", map[string]int{"amount": 1188}, &result, nil); err != nil {
		t.Fatal(err)
	}
	if result["id"] != 2 {
		t.Fatalf("new call reused key: %v", result)
	}
}

func TestUnkeyedPostNeverRetries(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Idempotency-Key") != "" {
			t.Error("unexpected key")
		}
		w.WriteHeader(503)
	}))
	defer server.Close()
	client := New("fac_test_local", WithBaseURL(server.URL), WithAutoIdempotency(false))
	if err := client.client.post("/payments", nil, nil, nil); err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("got %d attempts", calls)
	}
}

func TestRetryAfterIsNeverShortened(t *testing.T) {
	response := &http.Response{Header: http.Header{"Retry-After": []string{"90"}}}
	if got := retryDelay(0, response); got != 90*time.Second {
		t.Fatalf("shortened Retry-After: %s", got)
	}
	client := New("fac_test_local")
	budget := 100 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if client.client.backoff(ctx, 0, response, &budget) {
		t.Fatal("premature retry before 90s")
	}
	if budget != 10*time.Second {
		t.Fatalf("wrong cumulative budget: %s", budget)
	}
}

func TestRetryAfterOverBudgetReturns429(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "90")
		w.WriteHeader(429)
		_, _ = io.WriteString(w, `{"error":{"type":"rate_limit_error","code":"rate_limit_exceeded","message":"Slow down"}}`)
	}))
	defer server.Close()
	client := New("fac_test_local", WithBaseURL(server.URL))
	err := client.client.post("/payments", nil, nil, nil)
	apiErr, ok := err.(*Error)
	if !ok || apiErr.HTTPStatusCode != 429 || calls != 1 {
		t.Fatalf("unexpected result: %v, calls %d", err, calls)
	}
}

func TestExplicitKeyAndRetryAfterWaiting(t *testing.T) {
	calls := 0
	start := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Idempotency-Key") != "caller-key" {
			t.Error("caller key lost")
		}
		if calls == 1 {
			w.Header().Set("Retry-After", "0.08")
			w.WriteHeader(429)
			return
		}
		if time.Since(start) < 80*time.Millisecond {
			t.Error("premature retry")
		}
		_, _ = io.WriteString(w, `{}`)
	}))
	defer server.Close()
	client := New("fac_test_local", WithBaseURL(server.URL), WithAutoIdempotency(false))
	if err := client.client.post("/payments", nil, nil, &requestOption{idempotencyKey: "caller-key"}); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("attempts %d", calls)
	}
}
