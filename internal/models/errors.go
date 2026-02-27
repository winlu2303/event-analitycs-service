package models

import "errors"

// Domain errors
var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUnauthorized       = errors.New("unauthorized")

	// Project errors
	ErrProjectNotFound     = errors.New("project not found")
	ErrProjectAccessDenied = errors.New("access denied to project")
	ErrProjectLimitReached = errors.New("project limit reached")
	ErrInvalidAPIKey       = errors.New("invalid API key")

	// Event errors
	ErrInvalidEventType  = errors.New("invalid event type")
	ErrInvalidEventData  = errors.New("invalid event data")
	ErrEventTooLarge     = errors.New("event too large")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// Database errors
	ErrDatabaseConnection = errors.New("database connection error")
	ErrDuplicateEntry     = errors.New("duplicate entry")
	ErrQueryFailed        = errors.New("query failed")

	// Validation errors
	ErrInvalidDateFormat = errors.New("invalid date format")
	ErrInvalidPagination = errors.New("invalid pagination parameters")
	ErrRequiredField     = errors.New("required field missing")

	// Export errors
	ErrExportFailed   = errors.New("export failed")
	ErrNoDataToExport = errors.New("no data to export")
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents validation errors response
type ValidationErrorResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors"`
}
