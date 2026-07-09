package facturino

// UsageSummary is the current-period consumption snapshot for the
// authenticated account, as returned by GET /v1/usage. Drive in-app quota
// gauges and upgrade prompts from it, before a limit triggers a 402.
type UsageSummary struct {
	Object string `json:"object"`
	Plan   string `json:"plan"`
	// PeriodStart is the ISO 8601 start of the current monthly metering period.
	PeriodStart string `json:"periodStart"`
	// Counters maps each metered dimension to its meter, e.g.
	// "invoicesMonth", "apiRequestsMonth", "quotesMonth", "customers",
	// "products", "members", "webhookEndpoints", "companies".
	Counters map[string]UsageMeter `json:"counters"`
}

// UsageMeter is the consumption / limit pair for one metered dimension.
// Limit is nil when the dimension is unlimited on the current plan.
type UsageMeter struct {
	Used  int  `json:"used"`
	Limit *int `json:"limit"`
}

// UsageService exposes the current-period consumption metrics for the
// authenticated account. Useful for in-app dashboards and proactive
// plan-upgrade nudges before a quota hit triggers a 402 from the API.
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
