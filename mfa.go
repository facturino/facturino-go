package facturino

// MfaSetup is the response from initiating TOTP MFA setup.
type MfaSetup struct {
	Object string `json:"object"`
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

// MfaVerification is the response from verifying a TOTP code.
type MfaVerification struct {
	Object  string `json:"object"`
	Enabled bool   `json:"enabled"`
}

// MfaDisableResponse is the response from disabling MFA.
type MfaDisableResponse struct {
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// BackupCodes is the response containing MFA backup codes.
type BackupCodes struct {
	Object string   `json:"object"`
	Codes  []string `json:"codes"`
}

// MfaService operates on MFA (Multi-Factor Authentication) settings.
type MfaService struct {
	client *httpClient
}

// Setup initiates TOTP MFA setup, returning a secret and otpauth URI.
func (s *MfaService) Setup() (*MfaSetup, error) {
	var resp MfaSetup
	err := s.client.post("/auth/mfa/setup", nil, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Verify verifies a 6-digit TOTP code and enables MFA on the account.
func (s *MfaService) Verify(code string) (*MfaVerification, error) {
	var resp MfaVerification
	body := map[string]string{"code": code}
	err := s.client.post("/auth/mfa/verify", body, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Disable disables MFA on the account. Requires a valid TOTP code for confirmation.
func (s *MfaService) Disable(code string) (*MfaDisableResponse, error) {
	var resp MfaDisableResponse
	body := map[string]string{"code": code}
	err := s.client.do("DELETE", "/auth/mfa", body, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// BackupCodes generates a new set of backup codes. MFA must be enabled.
func (s *MfaService) BackupCodes() (*BackupCodes, error) {
	var resp BackupCodes
	err := s.client.post("/auth/mfa/backup-codes", nil, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
