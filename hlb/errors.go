package hlb

import "fmt"

// APIErrorResponse represents an error from the HLB API
type APIErrorResponse struct {
	Code    int    `json:"code"`    // Error code from the API
	Message string `json:"message"` // Error message from the API
}

func (e *APIErrorResponse) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}
