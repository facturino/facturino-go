package facturino

import "fmt"

// ErrorType is the category of an API error.
type ErrorType string

const (
	ErrorTypeInvalidRequest  ErrorType = "invalid_request_error"
	ErrorTypeAuthentication  ErrorType = "authentication_error"
	ErrorTypeRateLimit       ErrorType = "rate_limit_error"
	ErrorTypeAPI             ErrorType = "api_error"
	ErrorTypeValidation      ErrorType = "validation_error"
	ErrorTypePlanLimit       ErrorType = "plan_limit_error"
	ErrorTypeNotFound        ErrorType = "not_found_error"
	ErrorTypeConflict        ErrorType = "conflict_error"
)

// Error is a structured API error from Facturino.
type Error struct {
	Type           ErrorType `json:"type"`
	Code           string    `json:"code"`
	Message        string    `json:"message"`
	Param          string    `json:"param,omitempty"`
	DocURL         string    `json:"doc_url,omitempty"`
	RequestID      string    `json:"request_id"`
	Hint           string    `json:"hint,omitempty"`
	HTTPStatusCode int       `json:"-"`
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
