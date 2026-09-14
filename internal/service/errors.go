package service

import "errors"

// Sentinel errors shared across the service layer. Handlers map these to
// status codes (400/401/403/404) instead of stringly error messages.
var (
	ErrNotFound     = errors.New("resource not found")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidInput = errors.New("invalid input")
)
