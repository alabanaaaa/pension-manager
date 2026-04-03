package core

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrCodeNotFound         ErrorCode = "not_found"
	ErrCodeInvalidOperation ErrorCode = "invalid_operation"
)

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

func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

func NewDomainErrorWithCause(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Cause: cause}
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
	msg := "validation errors:"
	for _, e := range ve {
		msg += fmt.Sprintf(" %s=%s;", e.Field, e.Message)
	}
	return msg
}

func IsNotFound(err error) bool {
	var de *DomainError
	if errors.As(err, &de) {
		return de.Code == ErrCodeNotFound
	}
	return false
}
