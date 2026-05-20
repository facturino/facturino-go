package facturino

// UsageSummary is the current-period consumption snapshot for the
// authenticated account.
type UsageSummary struct {
	Object  string                `json:"object"`
	Plan    string                `json:"plan"`
	Period  *UsagePeriod          `json:"period,omitempty"`
	Metrics map[string]UsageMeter `json:"metrics,omitempty"`

	// Legacy / top-level meters mirrored on the response for convenience.
	InvoicesIssued *UsageMeter `json:"invoicesIssued,omitempty"`
	StorageBytes   *UsageMeter `json:"storageBytes,omitempty"`
	APICalls       *UsageMeter `json:"apiCalls,omitempty"`
	PASubmissions  *UsageMeter `json:"paSubmissions,omitempty"`
}

// UsagePeriod bounds the period over which the usage is reported.
type UsagePeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// UsageMeter is the consumption / limit pair for one metered
// dimension.
type UsageMeter struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

// UsageService exposes the current-period consumption metrics for the
// authenticated account (invoices issued, storage used, PA submissions,
// API calls). Useful for in-app dashboards and proactive plan-upgrade
// nudges before a quota hit triggers a 402 from the API.
type UsageService struct {
	client *httpClient
}

// Retrieve returns the current usage snapshot — plan limits and the
// consumption so far for each metered dimension.
func (s *UsageService) Retrieve() (*UsageSummary, error) {
	var out UsageSummary
	if err := s.client.get("/usage", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
