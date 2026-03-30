package facturino

import "fmt"

// SandboxResetResponse is returned by sandbox reset/fixtures operations.
type SandboxResetResponse struct {
	Object          string `json:"object"`
	DeletedCount    int    `json:"deleted_count"`
	FixturesCreated int    `json:"fixtures_created"`
}

// SimulateStatusParams configures a PA status simulation.
type SimulateStatusParams struct {
	Status string `json:"status"`
}

// SimulateStatusResponse is returned by status simulation.
type SimulateStatusResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Status    string `json:"status"`
	Simulated bool   `json:"simulated"`
}

// SandboxService provides test-only operations (fac_test_* keys only).
type SandboxService struct {
	client *httpClient
}

// ResetData wipes test data and reloads fixtures.
func (s *SandboxService) ResetData() (*SandboxResetResponse, error) {
	var resp SandboxResetResponse
	err := s.client.post("/sandbox/reset", nil, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SimulateStatus simulates a PA status change on a test invoice.
func (s *SandboxService) SimulateStatus(invoiceID string, params *SimulateStatusParams) (*SimulateStatusResponse, error) {
	var resp SimulateStatusResponse
	err := s.client.post(
		fmt.Sprintf("/sandbox/simulate-status/%s", invoiceID),
		params, &resp, nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateFixtures loads sandbox test fixtures by resetting and reloading data.
func (s *SandboxService) CreateFixtures() (*SandboxResetResponse, error) {
	var resp SandboxResetResponse
	err := s.client.post("/sandbox/reset", nil, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
