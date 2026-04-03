package domain

import (
	"errors"
	"fmt"
)

// ErrorCode represents a typed error code
type ErrorCode string

const (
	ErrCodeInvalidMember       ErrorCode = "invalid_member"
	ErrCodeInvalidBeneficiary  ErrorCode = "invalid_beneficiary"
	ErrCodeInvalidContribution ErrorCode = "invalid_contribution"
	ErrCodeInvalidClaim        ErrorCode = "invalid_claim"
	ErrCodeNotFound            ErrorCode = "not_found"
	ErrCodeDuplicate           ErrorCode = "duplicate"
	ErrCodeInsufficientFunds   ErrorCode = "insufficient_funds"
	ErrCodeInvalidStatus       ErrorCode = "invalid_status"
	ErrCodeUnauthorized        ErrorCode = "unauthorized"
	ErrCodePersistence         ErrorCode = "persistence_error"
	ErrCodeValidation          ErrorCode = "validation_error"
)

// DomainError is a typed error with a code and message
type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// NewDomainError creates a new domain error
func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

// NewDomainErrorWithCause creates a new domain error with an underlying cause
func NewDomainErrorWithCause(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Cause: cause}
}

// ValidationError represents a field-level validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

// ValidationErrors is a collection of field validation errors
type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	msg := "validation errors:"
	for _, e := range ve {
		msg += fmt.Sprintf(" %s=%s;", e.Field, e.Message)
	}
	return msg
}

// IsNotFound checks if an error is a not-found error
func IsNotFound(err error) bool {
	var de *DomainError
	if errors.As(err, &de) {
		return de.Code == ErrCodeNotFound
	}
	return false
}

// IsDuplicate checks if an error is a duplicate error
func IsDuplicate(err error) bool {
	var de *DomainError
	if errors.As(err, &de) {
		return de.Code == ErrCodeDuplicate
	}
	return false
}
