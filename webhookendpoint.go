package facturino

import (
	"encoding/json"
	"fmt"
)

// WebhookEndpointResource is a configured webhook endpoint.
type WebhookEndpointResource struct {
	ID          string   `json:"id"`
	Object      string   `json:"object,omitempty"`
	URL         string   `json:"url"`
	Secret      string   `json:"secret,omitempty"`
	Events      []string `json:"events"`
	Description string   `json:"description,omitempty"`
	Livemode    bool     `json:"livemode"`
	Active      bool     `json:"active"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
}

// WebhookEndpointParams are the parameters for creating a webhook endpoint.
type WebhookEndpointParams struct {
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Description string   `json:"description,omitempty"`

	IdempotencyKey string `json:"-"`
}

// WebhookEndpointUpdateParams are the parameters for updating a webhook endpoint.
type WebhookEndpointUpdateParams struct {
	URL         string   `json:"url,omitempty"`
	Events      []string `json:"events,omitempty"`
	Description string   `json:"description,omitempty"`
	Active      *bool    `json:"active,omitempty"`
}

// WebhookEndpointService operates on webhook endpoints.
type WebhookEndpointService struct {
	client *httpClient
}

// Create registers a webhook endpoint. The signing secret is only returned once.
func (s *WebhookEndpointService) Create(params *WebhookEndpointParams) (*WebhookEndpointResource, error) {
	var we WebhookEndpointResource
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post("/webhook-endpoints", params, &we, opts)
	if err != nil {
		return nil, err
	}
	return &we, nil
}

// Get retrieves a webhook endpoint by ID.
func (s *WebhookEndpointService) Get(id string) (*WebhookEndpointResource, error) {
	var we WebhookEndpointResource
	err := s.client.get(fmt.Sprintf("/webhook-endpoints/%s", id), nil, &we)
	if err != nil {
		return nil, err
	}
	return &we, nil
}

// Update updates a webhook endpoint.
func (s *WebhookEndpointService) Update(id string, params *WebhookEndpointUpdateParams) (*WebhookEndpointResource, error) {
	var we WebhookEndpointResource
	err := s.client.patch(fmt.Sprintf("/webhook-endpoints/%s", id), params, &we)
	if err != nil {
		return nil, err
	}
	return &we, nil
}

// Delete soft-deletes a webhook endpoint.
func (s *WebhookEndpointService) Delete(id string) error {
	return s.client.del(fmt.Sprintf("/webhook-endpoints/%s", id))
}

// Test sends a test event to a webhook endpoint.
func (s *WebhookEndpointService) Test(id string) (*WebhookEndpointResource, error) {
	var we WebhookEndpointResource
	err := s.client.post(fmt.Sprintf("/webhook-endpoints/%s/test", id), nil, &we, nil)
	if err != nil {
		return nil, err
	}
	return &we, nil
}

// List returns a paginated iterator over webhook endpoints.
func (s *WebhookEndpointService) List(params *ListParams) *WebhookEndpointIterator {
	iter := newIterator[*WebhookEndpointResource](s.client, "/webhook-endpoints", params, decodeWebhookEndpoint)
	return &WebhookEndpointIterator{iter: iter}
}

// WebhookEndpointIterator iterates over webhook endpoints.
type WebhookEndpointIterator struct {
	iter *Iterator[*WebhookEndpointResource]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *WebhookEndpointIterator) Next() bool { return it.iter.Next() }

// WebhookEndpoint returns the most recently fetched webhook endpoint.
func (it *WebhookEndpointIterator) WebhookEndpoint() *WebhookEndpointResource {
	return it.iter.Current()
}

// Err returns any error encountered during iteration.
func (it *WebhookEndpointIterator) Err() error { return it.iter.Err() }

func decodeWebhookEndpoint(raw json.RawMessage) (*WebhookEndpointResource, error) {
	var we WebhookEndpointResource
	err := json.Unmarshal(raw, &we)
	return &we, err
}
