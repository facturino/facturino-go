package facturino

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

const (
	defaultLimit = 25
	maxLimit     = 100
)

// ListParams holds pagination and filtering options for list endpoints.
type ListParams struct {
	Limit          int    `json:"limit,omitempty"`
	StartingAfter  string `json:"starting_after,omitempty"`
	EndingBefore   string `json:"ending_before,omitempty"`
	Status         string `json:"status,omitempty"`
	IncludeDeleted bool   `json:"include_deleted,omitempty"`
}

func (p *ListParams) toValues() url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.Limit > 0 {
		limit := p.Limit
		if limit > maxLimit {
			limit = maxLimit
		}
		v.Set("limit", strconv.Itoa(limit))
	}
	if p.StartingAfter != "" {
		v.Set("starting_after", p.StartingAfter)
	}
	if p.EndingBefore != "" {
		v.Set("ending_before", p.EndingBefore)
	}
	if p.Status != "" {
		v.Set("status", p.Status)
	}
	if p.IncludeDeleted {
		v.Set("include_deleted", "true")
	}
	return v
}

// ListResponse is the paginated list envelope from the API.
type ListResponse struct {
	Object     string            `json:"object"`
	URL        string            `json:"url"`
	Data       []json.RawMessage `json:"data"`
	HasMore    bool              `json:"has_more"`
	NextCursor *string           `json:"next_cursor"`
}

// Iterator lazily fetches paginated results from the API.
type Iterator[T any] struct {
	client      *httpClient
	path        string
	params      *ListParams
	extraParams url.Values
	decode      func(json.RawMessage) (T, error)
	current     T
	items       []json.RawMessage
	idx         int
	hasMore     bool
	cursor      string
	err         error
	started     bool
}

func newIterator[T any](client *httpClient, path string, params *ListParams, decode func(json.RawMessage) (T, error)) *Iterator[T] {
	if params == nil {
		params = &ListParams{}
	}
	if params.Limit <= 0 {
		params.Limit = defaultLimit
	}
	return &Iterator[T]{
		client:  client,
		path:    path,
		params:  params,
		decode:  decode,
		hasMore: true,
	}
}

// Next advances to the next item, returning false when done or on error.
func (it *Iterator[T]) Next() bool {
	if it.err != nil {
		return false
	}

	// If we have items remaining in the current page
	if it.started && it.idx < len(it.items) {
		item, err := it.decode(it.items[it.idx])
		if err != nil {
			it.err = fmt.Errorf("facturino: failed to decode list item: %w", err)
			return false
		}
		it.current = item
		it.idx++
		return true
	}

	// No more pages
	if it.started && !it.hasMore {
		return false
	}

	// Fetch next page
	it.started = true
	params := it.params.toValues()
	for k, vals := range it.extraParams {
		for _, v := range vals {
			params.Set(k, v)
		}
	}
	if it.cursor != "" {
		params.Set("starting_after", it.cursor)
	}

	var resp ListResponse
	if err := it.client.get(it.path, params, &resp); err != nil {
		it.err = err
		return false
	}

	it.items = resp.Data
	it.hasMore = resp.HasMore
	it.idx = 0

	if len(it.items) == 0 {
		return false
	}

	// Extract cursor from last item's ID for next page
	if resp.NextCursor != nil {
		it.cursor = *resp.NextCursor
	} else if len(it.items) > 0 {
		// Fallback: extract ID from last item
		var obj struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(it.items[len(it.items)-1], &obj); err == nil {
			it.cursor = obj.ID
		}
	}

	item, err := it.decode(it.items[it.idx])
	if err != nil {
		it.err = fmt.Errorf("facturino: failed to decode list item: %w", err)
		return false
	}
	it.current = item
	it.idx++
	return true
}

// Current returns the most recently fetched item.
func (it *Iterator[T]) Current() T {
	return it.current
}

// Err returns any error encountered during iteration.
func (it *Iterator[T]) Err() error {
	return it.err
}
