package facturino

import "fmt"

// Job is an async operation (PDF generation, FEC export, etc.).
type Job struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	InvoiceID string `json:"invoiceId,omitempty"`
	Progress  int    `json:"progress"`
	Result    string `json:"result,omitempty"`
	Error     string `json:"error,omitempty"`
	Created   string `json:"created"`
	UpdatedAt string `json:"updatedAt"`
	ExpireAt  string `json:"expireAt,omitempty"`

	DownloadURL string `json:"download_url,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
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
