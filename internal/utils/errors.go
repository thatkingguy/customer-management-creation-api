package utils

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Error types
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Err:     err,
	}
}

func NewNotFoundError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: message,
		Err:     err,
	}
}

func NewDuplicateError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Err:     err,
	}
}

func NewDatabaseError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
		Err:     nil,
	}
}

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	ResponseCode int   `json:"responseCode"`
}

// ToErrorResponse converts AppError to ErrorResponse
func (e *AppError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Status:       "error",
		Message:      e.Message,
		ResponseCode: e.Code,
	}
}



