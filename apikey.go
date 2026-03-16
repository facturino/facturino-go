package facturino

import (
	"encoding/json"
	"fmt"
)

// APIKey is an API key. The full value is only returned at creation.
type APIKey struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Prefix      string   `json:"prefix"`
	Livemode    bool     `json:"livemode"`
	Permissions []string `json:"permissions"`
	Revoked     bool     `json:"revoked"`
	RevokedAt   string   `json:"revokedAt,omitempty"`
	LastUsedAt  string   `json:"lastUsedAt,omitempty"`
	Created     string   `json:"created"`

	// Only populated on creation; cannot be retrieved afterward.
	Key string `json:"key,omitempty"`
}

// APIKeyParams are the parameters for creating an API key.
type APIKeyParams struct {
	Name        string   `json:"name"`
	Livemode    bool     `json:"livemode"`
	Permissions []string `json:"permissions,omitempty"`

	IdempotencyKey string `json:"-"`
}

// APIKeyService operates on API keys.
type APIKeyService struct {
	client *httpClient
}

// Create creates a new API key. The full value cannot be retrieved again.
func (s *APIKeyService) Create(params *APIKeyParams) (*APIKey, error) {
	var key APIKey
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	err := s.client.post("/api-keys", params, &key, opts)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// Get retrieves an API key by ID (without the secret value).
func (s *APIKeyService) Get(id string) (*APIKey, error) {
	var key APIKey
	err := s.client.get(fmt.Sprintf("/api-keys/%s", id), nil, &key)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// List returns a paginated iterator over API keys.
func (s *APIKeyService) List(params *ListParams) *APIKeyIterator {
	iter := newIterator[*APIKey](s.client, "/api-keys", params, decodeAPIKey)
	return &APIKeyIterator{iter: iter}
}

// Revoke permanently revokes an API key.
func (s *APIKeyService) Revoke(id string) (*APIKey, error) {
	var key APIKey
	err := s.client.post(fmt.Sprintf("/api-keys/%s/revoke", id), nil, &key, nil)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// APIKeyIterator iterates over API keys.
type APIKeyIterator struct {
	iter *Iterator[*APIKey]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *APIKeyIterator) Next() bool { return it.iter.Next() }

// APIKey returns the most recently fetched API key.
func (it *APIKeyIterator) APIKey() *APIKey { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *APIKeyIterator) Err() error { return it.iter.Err() }

func decodeAPIKey(raw json.RawMessage) (*APIKey, error) {
	var k APIKey
	err := json.Unmarshal(raw, &k)
	return &k, err
}
