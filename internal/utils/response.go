package utils

// SuccessResponse represents the success response structure
type SuccessResponse struct {
	Status       string      `json:"status"`
	Message      string      `json:"message"`
	Data         interface{} `json:"data"`
	ResponseCode int         `json:"responseCode"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string, data interface{}, code int) SuccessResponse {
	return SuccessResponse{
		Status:       "success",
		Message:      message,
		Data:         data,
		ResponseCode: code,
	}
}

