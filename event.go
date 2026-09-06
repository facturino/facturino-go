package facturino

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Event is a webhook event record.
type Event struct {
	ID         string                 `json:"id"`
	Object     string                 `json:"object"`
	Livemode   bool                   `json:"livemode"`
	Type       string                 `json:"type"`
	APIVersion string                 `json:"apiVersion"`
	Data       map[string]interface{} `json:"data"`
	Request    *EventRequest          `json:"request,omitempty"`

	Delivered bool              `json:"delivered"`
	Attempts  []*WebhookAttempt `json:"attempts"`
	NextRetry string            `json:"nextRetry"`

	Created string `json:"created"`
	Updated string `json:"updated"`
}

// EventRequest holds metadata about the originating API request.
type EventRequest struct {
	ID             string `json:"id,omitempty"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// WebhookAttempt is a single delivery attempt.
type WebhookAttempt struct {
	Timestamp  string `json:"timestamp"`
	HTTPStatus int    `json:"httpStatus"`
	Error      string `json:"error,omitempty"`
	Duration   int    `json:"duration,omitempty"`
}

// EventListParams adds event-specific filters to ListParams.
type EventListParams struct {
	ListParams

	Type string // e.g. "invoice.finalized"
}

// EventService operates on events.
type EventService struct {
	client *httpClient
}

// List returns a paginated iterator over events.
func (s *EventService) List(params *EventListParams) *EventIterator {
	var lp *ListParams
	if params != nil {
		lp = &params.ListParams
	}
	iter := newIterator[*Event](s.client, "/events", lp, decodeEvent)
	if params != nil && params.Type != "" {
		iter.extraParams = url.Values{"type": {params.Type}}
	}
	return &EventIterator{iter: iter}
}

// Get retrieves a single event by ID.
func (s *EventService) Get(id string) (*Event, error) {
	var evt Event
	err := s.client.get(fmt.Sprintf("/events/%s", id), nil, &evt)
	if err != nil {
		return nil, err
	}
	return &evt, nil
}

// EventRetryResult is the answer of POST /v1/events/{id}/retry: a scheduling
// receipt, not the event itself.
type EventRetryResult struct {
	ID             string `json:"id"`
	Object         string `json:"object"`
	RetryScheduled bool   `json:"retryScheduled"`
	// EndpointID is set when the replay targets one endpoint.
	EndpointID string `json:"endpointId,omitempty"`
}

// RetryDelivery retriggers the delivery to every endpoint where it failed and
// returns the scheduling receipt.
func (s *EventService) RetryDelivery(id string) (*EventRetryResult, error) {
	var result EventRetryResult
	err := s.client.post(fmt.Sprintf("/events/%s/retry", id), nil, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Retry retriggers delivery.
//
// Deprecated: the endpoint answers a scheduling receipt, not the event; only
// ID and Object of the returned Event are populated. Use RetryDelivery.
func (s *EventService) Retry(id string) (*Event, error) {
	result, err := s.RetryDelivery(id)
	if err != nil {
		return nil, err
	}
	return &Event{ID: result.ID, Object: result.Object}, nil
}

// RetryToEndpoint replays the event to one endpoint, even if it was already
// delivered there: the endpoint must be active and subscribed to the event
// type, and the receiver must stay idempotent by event id.
func (s *EventService) RetryToEndpoint(id, endpointID string) (*EventRetryResult, error) {
	var result EventRetryResult
	body := map[string]string{"endpointId": endpointID}
	err := s.client.post(fmt.Sprintf("/events/%s/retry", id), body, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EventIterator iterates over events.
type EventIterator struct {
	iter *Iterator[*Event]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *EventIterator) Next() bool { return it.iter.Next() }

// Event returns the most recently fetched event.
func (it *EventIterator) Event() *Event { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *EventIterator) Err() error { return it.iter.Err() }

func decodeEvent(raw json.RawMessage) (*Event, error) {
	var e Event
	err := json.Unmarshal(raw, &e)
	return &e, err
}
