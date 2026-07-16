package facturino

// HealthStatus is the service snapshot returned by GET /v1/health. It exposes
// no account data — only the service status and the current API version — so it
// is safe to poll for connectivity checks, readiness gates and uptime monitors.
type HealthStatus struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Region    string `json:"region"`
	Timestamp string `json:"timestamp"`
}

// HealthService is a lightweight liveness probe for the Facturino API.
type HealthService struct {
	client *httpClient
}

// Check returns the current service health snapshot. Useful to confirm an API
// key reaches the platform before running real traffic.
func (s *HealthService) Check() (*HealthStatus, error) {
	var out HealthStatus
	if err := s.client.get("/health", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
