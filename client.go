package facturino

import (
	"bytes"
	"context"
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
	sdkVersion        = "2.1.0"
	defaultMaxRetries = 3
)

// httpClient wraps net/http for authenticated requests to the Facturino API.
type httpClient struct {
	apiKey      string
	baseURL     string
	httpClient  *http.Client
	maxRetries  int
	baseContext context.Context
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
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpCl,
		maxRetries: maxRetries,
	}
}

type requestOption struct {
	idempotencyKey string
	context        context.Context
}

// do executes an HTTP request, retrying on 429/5xx, and decodes JSON into dest.
func (c *httpClient) do(method, path string, body interface{}, dest interface{}, opts *requestOption) error {
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
				return fmt.Errorf("facturino: failed to marshal request body: %w", err)
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

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		var bodyReader io.Reader
		if data != nil {
			bodyReader = bytes.NewReader(data)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return fmt.Errorf("facturino: failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "facturino-go/"+sdkVersion)
		req.Header.Set("Facturino-Version", apiDateVersion)

		if opts != nil && opts.idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", opts.idempotencyKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("facturino: request failed: %w", err)
			// Network errors are retryable
			if attempt < c.maxRetries {
				c.backoff(ctx, attempt, nil)
				continue
			}
			return lastErr
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("facturino: failed to read response body: %w", err)
		}

		// Retryable status codes
		if c.isRetryable(resp.StatusCode) && attempt < c.maxRetries {
			lastErr = c.parseError(resp.StatusCode, respBody)
			c.backoff(ctx, attempt, resp)
			continue
		}

		// Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if resp.StatusCode == 204 || dest == nil {
				return nil
			}
			if err := json.Unmarshal(respBody, dest); err != nil {
				return fmt.Errorf("facturino: failed to decode response: %w", err)
			}
			return nil
		}

		// Non-retryable error
		return c.parseError(resp.StatusCode, respBody)
	}

	return lastErr
}

// doRaw is like do but returns raw bytes instead of decoding JSON.
func (c *httpClient) doRaw(method, path string, body interface{}, opts *requestOption) ([]byte, string, error) {
	fullURL := c.baseURL + "/" + apiVersion + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, "", fmt.Errorf("facturino: failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	ctx := c.baseContext
	if ctx == nil {
		ctx = context.Background()
	}
	if opts != nil && opts.context != nil {
		ctx = opts.context
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if body != nil {
			data, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(data)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return nil, "", fmt.Errorf("facturino: failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("User-Agent", "facturino-go/"+sdkVersion)
		req.Header.Set("Facturino-Version", apiDateVersion)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("facturino: request failed: %w", err)
			if attempt < c.maxRetries {
				c.backoff(ctx, attempt, nil)
				continue
			}
			return nil, "", lastErr
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, "", fmt.Errorf("facturino: failed to read response body: %w", err)
		}

		if c.isRetryable(resp.StatusCode) && attempt < c.maxRetries {
			lastErr = c.parseError(resp.StatusCode, respBody)
			c.backoff(ctx, attempt, resp)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			contentType := resp.Header.Get("Content-Type")
			return respBody, contentType, nil
		}

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

// backoff waits with exponential delay, respecting context cancellation and Retry-After.
func (c *httpClient) backoff(ctx context.Context, attempt int, resp *http.Response) {
	var delay time.Duration

	// Check for Retry-After header
	if resp != nil {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
				if seconds > 60 {
					seconds = 60
				}
				delay = time.Duration(seconds) * time.Second
			}
		}
	}

	if delay == 0 {
		// Exponential backoff: 500ms, 1s, 2s
		delay = time.Duration(math.Pow(2, float64(attempt))) * 500 * time.Millisecond
		if delay > 8*time.Second {
			delay = 8 * time.Second
		}
	}

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return
	}
}

func (c *httpClient) parseError(statusCode int, body []byte) error {
	var envelope errorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error != nil {
		envelope.Error.HTTPStatusCode = statusCode
		return envelope.Error
	}

	// Fallback for non-standard error responses
	return &Error{
		Type:           ErrorTypeAPI,
		Code:           "unknown",
		Message:        fmt.Sprintf("unexpected status %d: %s", statusCode, string(body)),
		HTTPStatusCode: statusCode,
	}
}
