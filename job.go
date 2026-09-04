package facturino

import "fmt"

// Job is an async operation (PDF generation, FEC export, etc.).
type Job struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	// Status is one of "pending", "processing", "completed",
	// "completed_with_errors", "failed" or "superseded". "superseded" is a
	// TERMINAL state without a deliverable: the render was produced for a
	// ledger revision that a collection, a cancellation or a refund has since
	// overtaken. It is not a failure — request the document again and a
	// current render is published. New values may be added: tolerate
	// unknown ones.
	Status    string `json:"status"`
	InvoiceID string `json:"invoiceId,omitempty"`
	Progress  int    `json:"progress"`
	Result    string `json:"result,omitempty"`
	Error     string `json:"error,omitempty"`
	Created   string `json:"created"`
	UpdatedAt string `json:"updatedAt"`
	ExpireAt  string `json:"expireAt,omitempty"`

	// URL is a signed, time-limited link to download the generated document,
	// populated once Status is "completed". ExpiresAt is its ISO 8601 expiry.
	URL       string `json:"url,omitempty"`
	ExpiresAt string `json:"expiresAt,omitempty"`
}

// JobService operates on async jobs.
type JobService struct {
	client *httpClient
}

// Get retrieves a job by ID for polling.
func (s *JobService) Get(id string) (*Job, error) {
	var j Job
	err := s.client.get(fmt.Sprintf("/jobs/%s", id), nil, &j)
	if err != nil {
		return nil, err
	}
	return &j, nil
}
