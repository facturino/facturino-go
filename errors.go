package facturino

import "fmt"

// ErrorType is the category of an API error.
type ErrorType string

const (
	ErrorTypeInvalidRequest ErrorType = "invalid_request_error"
	ErrorTypeAuthentication ErrorType = "authentication_error"
	ErrorTypeRateLimit      ErrorType = "rate_limit_error"
	ErrorTypeAPI            ErrorType = "api_error"
	ErrorTypeValidation     ErrorType = "validation_error"
	ErrorTypePlanLimit      ErrorType = "plan_limit_error"
	ErrorTypeNotFound       ErrorType = "not_found_error"
	ErrorTypeConflict       ErrorType = "conflict_error"
)

// ErrorIssue is one detailed reason behind a refusal.
//
// Code is stable and safe to branch on, and is more precise than the refusal's
// own Code when that one covers several distinct facts. Param names the field
// in cause as a dotted path, and is empty when the reason names no field: an
// approximate pointer would send the caller to correct a field that was right.
type ErrorIssue struct {
	Code    string `json:"code"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message"`
}

// Error is a structured API error from Facturino.
type Error struct {
	Type      ErrorType `json:"type"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Param     string    `json:"param,omitempty"`
	DocURL    string    `json:"doc_url,omitempty"`
	RequestID string    `json:"request_id"`
	Hint      string    `json:"hint,omitempty"`
	// Issues carries the detailed reasons when one refusal has several. It is
	// additive: Type, Code, Message, Param and Hint keep their exact meaning,
	// and Param still points at the first field in cause. An EMPTY, non-nil
	// slice when the refusal had nothing more to say — never nil.
	Issues         []ErrorIssue `json:"issues,omitempty"`
	HTTPStatusCode int          `json:"-"`
}

func (e *Error) Error() string {
	if e.Param != "" {
		return fmt.Sprintf("facturino: %s (type=%s, code=%s, param=%s, request_id=%s)",
			e.Message, e.Type, e.Code, e.Param, e.RequestID)
	}
	return fmt.Sprintf("facturino: %s (type=%s, code=%s, request_id=%s)",
		e.Message, e.Type, e.Code, e.RequestID)
}

type errorEnvelope struct {
	Error *Error `json:"error"`
}

// IsErrorType reports whether err is a *Error with the given type.
func IsErrorType(err error, t ErrorType) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.Type == t
}

// IsNotFound reports whether err is a not_found_error.
func IsNotFound(err error) bool {
	return IsErrorType(err, ErrorTypeNotFound)
}

// IsRateLimit reports whether err is a rate_limit_error.
func IsRateLimit(err error) bool {
	return IsErrorType(err, ErrorTypeRateLimit)
}

// IsAuthentication reports whether err is an authentication_error.
func IsAuthentication(err error) bool {
	return IsErrorType(err, ErrorTypeAuthentication)
}
