package hlb

import "fmt"

// APIErrorResponse represents the error structure returned by the HLB API
type APIErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error represents an error from HLB library operations
type Error struct {
	// APIResponse contains the parsed API error response, or nil if the error is not from the API
	APIResponse *APIErrorResponse
	// Message is the library-level error message
	Message string
}

func (e *Error) Error() string {
	if e.APIResponse != nil && e.APIResponse.Message != "" {
		return fmt.Sprintf("API request failed with status code %d: %s", e.APIResponse.Code, e.APIResponse.Message)
	}
	return e.Message
}

// NewError creates a new library error (non-API error)
func NewError(message string) *Error {
	return &Error{
		Message: message,
	}
}

// NewErrorf creates a new library error with formatted message.
// Supports error wrapping with %w directive.
func NewErrorf(format string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Errorf(format, args...).Error(),
	}
}
