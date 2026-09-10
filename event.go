package facturino

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Event is a webhook event record.
type Event struct {
	RawResponse `json:"-"`
	CompanyID   string                      `json:"companyId,omitempty"`
	EndpointID  string                      `json:"endpointId,omitempty"`
	Deliveries  map[string]EndpointDelivery `json:"deliveries,omitempty"`
	ExpireAt    *string                     `json:"expireAt,omitempty"`
	ID          string                      `json:"id"`
	Object      string                      `json:"object"`
	Livemode    bool                        `json:"livemode"`
	Type        string                      `json:"type"`
	APIVersion  string                      `json:"apiVersion"`
	Data        map[string]interface{}      `json:"data"`
	Request     *EventRequest               `json:"request,omitempty"`

	Delivered bool              `json:"delivered"`
	Attempts  []*WebhookAttempt `json:"attempts"`
	NextRetry *string           `json:"nextRetry"`

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
	EndpointID string `json:"endpointId,omitempty"`
	Timestamp  string `json:"timestamp"`
	HTTPStatus *int   `json:"httpStatus"`
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

// DocumentEventData is the typed document projection in invoice and credit-note
// events. Event.Data remains a map to preserve fields from other event families.
type DocumentEventData struct {
	ID                   string                 `json:"id,omitempty"`
	Object               string                 `json:"object,omitempty"`
	Number               *string                `json:"number,omitempty"`
	Status               string                 `json:"status,omitempty"`
	PreviousStatus       string                 `json:"previous_status,omitempty"`
	Livemode             bool                   `json:"livemode,omitempty"`
	DocumentStatus       string                 `json:"documentStatus,omitempty"`
	TransmissionStatus   string                 `json:"transmissionStatus,omitempty"`
	TransmissionDetail   *string                `json:"transmissionDetail,omitempty"`
	PaymentStatus        string                 `json:"paymentStatus,omitempty"`
	PAErrorCode          *string                `json:"paErrorCode,omitempty"`
	RejectionReason      *string                `json:"rejectionReason,omitempty"`
	RejectionCategory    *PaRejectionCategory   `json:"rejectionCategory,omitempty"`
	RejectionCode        *string                `json:"rejectionCode,omitempty"`
	RejectionSource      *PaRejectionSource     `json:"rejectionSource,omitempty"`
	RelatedInvoiceID     *string                `json:"relatedInvoiceId,omitempty"`
	RelatedInvoiceNumber *string                `json:"relatedInvoiceNumber,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentData reads document fields without discarding the original data map.
func (e *Event) DocumentData() (*DocumentEventData, error) {
	return decodeDocumentEventData(e.Data)
}

func decodeDocumentEventData(data map[string]interface{}) (*DocumentEventData, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var result DocumentEventData
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// EndpointDelivery is a REST delivery record; signed payloads omit it.
type EndpointDelivery struct {
	Delivered    bool             `json:"delivered"`
	Attempts     []WebhookAttempt `json:"attempts"`
	NextRetry    *string          `json:"nextRetry"`
	AttemptCount *int             `json:"attemptCount,omitempty"`
	Generation   *int             `json:"generation,omitempty"`
}
