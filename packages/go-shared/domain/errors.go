package domain

import "fmt"

// Error codes
const (
	ErrNotFound    = "NOT_FOUND"
	ErrUnauthorized = "UNAUTHORIZED"
	ErrForbidden   = "FORBIDDEN"
	ErrConflict    = "CONFLICT"
	ErrValidation  = "VALIDATION"
	ErrInternal    = "INTERNAL"
)

// DomainError is a typed error carrying a code and message.
type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewNotFound(msg string) *DomainError {
	return &DomainError{Code: ErrNotFound, Message: msg}
}

func NewUnauthorized(msg string) *DomainError {
	return &DomainError{Code: ErrUnauthorized, Message: msg}
}

func NewForbidden(msg string) *DomainError {
	return &DomainError{Code: ErrForbidden, Message: msg}
}

func NewConflict(msg string) *DomainError {
	return &DomainError{Code: ErrConflict, Message: msg}
}

func NewValidation(msg string) *DomainError {
	return &DomainError{Code: ErrValidation, Message: msg}
}

func NewInternal(msg string) *DomainError {
	return &DomainError{Code: ErrInternal, Message: msg}
}
