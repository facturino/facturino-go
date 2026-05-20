package facturino

import (
	"fmt"
	"net/url"
	"strconv"
)

// Notification is one entry of the in-app feed.
type Notification struct {
	Object  string                 `json:"object"`
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Title   string                 `json:"title"`
	Body    string                 `json:"body,omitempty"`
	Link    string                 `json:"link,omitempty"`
	Read    bool                   `json:"read"`
	Created string                 `json:"created"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// NotificationList is the paginated response for GET /v1/notifications.
type NotificationList struct {
	Object  string         `json:"object"`
	Data    []Notification `json:"data"`
	HasMore bool           `json:"has_more"`
}

// NotificationListParams adds the Unread filter to the standard
// pagination parameters.
type NotificationListParams struct {
	Limit          int
	StartingAfter  string
	EndingBefore   string
	Unread         *bool
}

// NotificationBatch is the response for PATCH
// /v1/notifications/mark-all-read.
type NotificationBatch struct {
	Object  string `json:"object"`
	Updated int    `json:"updated"`
}

// NotificationPreferences carries the per-event channel preferences
// (email / push / in-app) for the authenticated user.
type NotificationPreferences struct {
	Object      string                                 `json:"object"`
	Preferences map[string]NotificationChannelToggles `json:"preferences"`
}

// NotificationChannelToggles enables or disables a notification type on
// each delivery channel.
type NotificationChannelToggles struct {
	Email *bool `json:"email,omitempty"`
	InApp *bool `json:"inApp,omitempty"`
	Push  *bool `json:"push,omitempty"`
}

// NotificationPreferencesUpdate is the body for PATCH
// /v1/notification-preferences. The map merges with the existing
// preferences server-side, so callers can override a single event
// without resending the full table.
type NotificationPreferencesUpdate struct {
	Preferences map[string]NotificationChannelToggles `json:"preferences"`
}

// NotificationService exposes the in-app notification feed and the
// per-event notification preferences.
//
// The in-app feed (paginated list, mark single / all read) lives under
// /v1/notifications and is scoped to the authenticated user.
// Per-event preferences live under /v1/notification-preferences and
// override the channel-matrix defaults from the API reference.
type NotificationService struct {
	client *httpClient
}

// List returns the paginated list of notifications for the
// authenticated user. Pass Unread=&true to filter on unread only.
func (s *NotificationService) List(params *NotificationListParams) (*NotificationList, error) {
	var out NotificationList
	q := url.Values{}
	if params != nil {
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.StartingAfter != "" {
			q.Set("starting_after", params.StartingAfter)
		}
		if params.EndingBefore != "" {
			q.Set("ending_before", params.EndingBefore)
		}
		if params.Unread != nil {
			q.Set("unread", strconv.FormatBool(*params.Unread))
		}
	}
	if err := s.client.get("/notifications", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MarkRead marks a single notification as read.
func (s *NotificationService) MarkRead(id string) (*Notification, error) {
	var out Notification
	body := map[string]bool{"read": true}
	if err := s.client.patch(fmt.Sprintf("/notifications/%s", id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MarkAllRead marks every unread notification as read.
func (s *NotificationService) MarkAllRead() (*NotificationBatch, error) {
	var out NotificationBatch
	if err := s.client.patch("/notifications/mark-all-read", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RetrievePreferences returns the per-event notification preferences
// for the authenticated user.
func (s *NotificationService) RetrievePreferences() (*NotificationPreferences, error) {
	var out NotificationPreferences
	if err := s.client.get("/notification-preferences", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdatePreferences updates the per-event notification preferences.
// The body merges with the existing preferences map server-side.
func (s *NotificationService) UpdatePreferences(params *NotificationPreferencesUpdate) (*NotificationPreferences, error) {
	var out NotificationPreferences
	if err := s.client.patch("/notification-preferences", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
