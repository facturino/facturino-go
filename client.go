package facturino

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL    = "https://facturino.com/api"
	defaultTimeout    = 30 * time.Second
	apiVersion        = "v1"
	apiDateVersion    = "2026-09-01"
	sdkVersion        = "2.8.0"
	defaultMaxRetries = 3
)

// httpClient wraps net/http for authenticated requests to the Facturino API.
type httpClient struct {
	apiKey          string
	baseURL         string
	httpClient      *http.Client
	maxRetries      int
	baseContext     context.Context
	autoIdempotency bool
	retryBudget     time.Duration
}

// withContext returns a shallow copy bound to ctx, used as the default context
// for every request unless a per-request option overrides it.
func (c *httpClient) withContext(ctx context.Context) *httpClient {
	clone := *c
	clone.baseContext = ctx
	return &clone
}

func newHTTPClient(apiKey, baseURL string, httpCl *http.Client, maxRetries int) *httpClient {
	if httpCl == nil {
		httpCl = &http.Client{Timeout: defaultTimeout}
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	// Trim trailing slash for consistent URL building
	baseURL = strings.TrimRight(baseURL, "/")
	if maxRetries < 0 {
		maxRetries = defaultMaxRetries
	}
	return &httpClient{
		apiKey:          apiKey,
		baseURL:         baseURL,
		httpClient:      httpCl,
		maxRetries:      maxRetries,
		autoIdempotency: true,
		retryBudget:     60 * time.Second,
	}
}

type requestOption struct {
	idempotencyKey string
	context        context.Context
}

// do executes an HTTP request, retrying on 429/5xx, and decodes JSON into dest.
func (c *httpClient) do(method, path string, body interface{}, dest interface{}, opts *requestOption) error {
	data, _, err := c.doRaw(method, path, body, opts)
	if err != nil {
		return err
	}
	if len(data) == 0 || dest == nil {
		return nil
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("facturino: failed to decode response: %w", err)
	}
	return nil
}

func (c *httpClient) doRaw(method, path string, body interface{}, opts *requestOption) ([]byte, string, error) {
	fullURL := c.baseURL + "/" + apiVersion + path

	// Marshal the body ONCE (reused on each retry). A typed nil pointer (the
	// documented "no params" pattern) is a non-nil interface but marshals to
	// "null", which the backend rejects ("Expected object" → 400) — send an empty
	// object instead so no-param POSTs (Email, Remind, CreatePaymentLink…) work.
	var data []byte
	if body != nil {
		if rv := reflect.ValueOf(body); rv.Kind() == reflect.Ptr && rv.IsNil() {
			data = []byte("{}")
		} else {
			var err error
			data, err = json.Marshal(body)
			if err != nil {
				return nil, "", fmt.Errorf("facturino: failed to marshal request body: %w", err)
			}
		}
	}

	ctx := c.baseContext
	if ctx == nil {
		ctx = context.Background()
	}
	if opts != nil && opts.context != nil {
		ctx = opts.context
	}

	key := ""
	if opts != nil {
		key = opts.idempotencyKey
	}
	if method == http.MethodPost && key == "" && c.autoIdempotency {
		var token [16]byte
		if _, err := rand.Read(token[:]); err != nil {
			return nil, "", fmt.Errorf("facturino: idempotency key: %w", err)
		}
		key = hex.EncodeToString(token[:])
	}
	canRetry := method != http.MethodPost || key != ""
	remainingBudget := c.retryBudget
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		var bodyReader io.Reader
		if data != nil {
			bodyReader = bytes.NewReader(data)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return nil, "", fmt.Errorf("facturino: failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "facturino-go/"+sdkVersion)
		req.Header.Set("Facturino-Version", apiDateVersion)

		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("facturino: request failed: %w", err)
			// Network errors are retryable
			if canRetry && attempt < c.maxRetries && c.backoff(ctx, attempt, nil, &remainingBudget) {
				continue
			}
			return nil, "", lastErr
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("facturino: failed to read response body: %w", err)
			if canRetry && attempt < c.maxRetries && c.backoff(ctx, attempt, nil, &remainingBudget) {
				continue
			}
			return nil, "", lastErr
		}

		// Retryable status codes
		if canRetry && c.isRetryable(resp.StatusCode) && attempt < c.maxRetries {
			lastErr = c.parseError(resp.StatusCode, respBody)
			if c.backoff(ctx, attempt, resp, &remainingBudget) {
				continue
			}
			return nil, "", lastErr
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, resp.Header.Get("Content-Type"), nil
		}

		// Non-retryable error
		return nil, "", c.parseError(resp.StatusCode, respBody)
	}

	return nil, "", lastErr
}

func (c *httpClient) get(path string, params url.Values, dest interface{}) error {
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}
	return c.do(http.MethodGet, path, nil, dest, nil)
}

func (c *httpClient) post(path string, body interface{}, dest interface{}, opts *requestOption) error {
	return c.do(http.MethodPost, path, body, dest, opts)
}

func (c *httpClient) patch(path string, body interface{}, dest interface{}) error {
	return c.do(http.MethodPatch, path, body, dest, nil)
}

func (c *httpClient) del(path string) error {
	return c.do(http.MethodDelete, path, nil, nil, nil)
}

func (c *httpClient) isRetryable(statusCode int) bool {
	switch statusCode {
	case 429, 500, 502, 503:
		return true
	default:
		return false
	}
}

// retryDelay never caps a server Retry-After. The call's cumulative waiting
// budget determines whether to wait or return the original HTTP error.
func retryDelay(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		value := resp.Header.Get("Retry-After")
		if seconds, err := strconv.ParseFloat(value, 64); err == nil && seconds >= 0 && !math.IsInf(seconds, 0) && !math.IsNaN(seconds) {
			if seconds >= float64(math.MaxInt64)/float64(time.Second) {
				return time.Duration(math.MaxInt64)
			}
			return time.Duration(seconds * float64(time.Second))
		}
		if date, err := http.ParseTime(value); err == nil {
			delay := time.Until(date)
			if delay < 0 {
				return 0
			}
			return delay
		}
	}
	return time.Duration(math.Min(math.Pow(2, float64(attempt))*500, 8000)) * time.Millisecond
}

func (c *httpClient) backoff(ctx context.Context, attempt int, resp *http.Response, budget *time.Duration) bool {
	delay := retryDelay(attempt, resp)
	if delay > *budget || ctx.Err() != nil {
		return false
	}
	*budget -= delay
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return ctx.Err() == nil
	case <-ctx.Done():
		return false
	}
}

func (c *httpClient) parseError(statusCode int, body []byte) error {
	var envelope errorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error != nil {
		envelope.Error.HTTPStatusCode = statusCode
		// Never nil: the other SDKs publish an EMPTY list when the API sent no
		// detail, and reading Issues must not need a nil check in Go either.
		if envelope.Error.Issues == nil {
			envelope.Error.Issues = []ErrorIssue{}
		}
		return envelope.Error
	}

	// Fallback for non-standard error responses. Issues is empty, never nil:
	// reading it must not need a nil check on ANY path, including this one.
	return &Error{
		Type:           ErrorTypeAPI,
		Code:           "unknown",
		Message:        fmt.Sprintf("unexpected status %d: %s", statusCode, string(body)),
		HTTPStatusCode: statusCode,
		Issues:         []ErrorIssue{},
	}
}
