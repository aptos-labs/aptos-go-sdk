package aptos

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel errors for common error conditions.
// These can be used with errors.Is() to check for specific error types.
var (
	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("aptos: resource not found")

	// ErrInvalidAddress indicates an invalid account address format.
	ErrInvalidAddress = errors.New("aptos: invalid address format")

	// ErrInvalidArgument indicates an invalid argument was provided.
	ErrInvalidArgument = errors.New("aptos: invalid argument")

	// ErrTransactionFailed indicates transaction execution failed.
	ErrTransactionFailed = errors.New("aptos: transaction execution failed")

	// ErrTransactionExpired indicates the transaction expired before execution.
	ErrTransactionExpired = errors.New("aptos: transaction expired")

	// ErrSequenceNumberMismatch indicates a sequence number conflict.
	ErrSequenceNumberMismatch = errors.New("aptos: sequence number mismatch")

	// ErrInsufficientBalance indicates insufficient funds for the operation.
	ErrInsufficientBalance = errors.New("aptos: insufficient balance")

	// ErrRateLimited indicates the request was rate limited.
	ErrRateLimited = errors.New("aptos: rate limit exceeded")

	// ErrTimeout indicates the request timed out.
	ErrTimeout = errors.New("aptos: request timeout")

	// ErrUnavailable indicates the service is temporarily unavailable.
	ErrUnavailable = errors.New("aptos: service unavailable")

	// ErrInternal indicates an internal error occurred.
	ErrInternal = errors.New("aptos: internal error")

	// ErrSignature indicates a signature verification failure.
	ErrSignature = errors.New("aptos: signature verification failed")

	// ErrSerialization indicates a BCS serialization error.
	ErrSerialization = errors.New("aptos: serialization failed")

	// ErrDeserialization indicates a BCS deserialization error.
	ErrDeserialization = errors.New("aptos: deserialization failed")
)

// APIError represents an error response from the Aptos API.
// It provides rich context about the error including HTTP status,
// VM status for transaction failures, and the original request.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int

	// Message is the error message from the API.
	Message string

	// ErrorCode is the Aptos-specific error code, if any.
	ErrorCode string

	// VMStatus contains VM execution status for transaction errors.
	VMStatus *VMStatus

	// RequestMethod is the HTTP method of the failed request.
	RequestMethod string

	// RequestURL is the URL of the failed request.
	RequestURL string
}

// VMStatus contains details about VM execution failure.
type VMStatus struct {
	// Type is the VM status type (e.g., "move_abort", "execution_failure").
	Type string `json:"type"`

	// Location is the Move location where the error occurred.
	Location string `json:"location,omitempty"`

	// AbortCode is the abort code for Move aborts.
	AbortCode uint64 `json:"abort_code,omitempty"`

	// FunctionIndex is the function index for execution failures.
	FunctionIndex uint16 `json:"function_index,omitempty"`

	// CodeOffset is the code offset for execution failures.
	CodeOffset uint16 `json:"code_offset,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.VMStatus != nil {
		return fmt.Sprintf("aptos: API error %d: %s (vm_status: %s at %s, code %d)",
			e.StatusCode, e.Message, e.VMStatus.Type, e.VMStatus.Location, e.VMStatus.AbortCode)
	}
	if e.ErrorCode != "" {
		return fmt.Sprintf("aptos: API error %d [%s]: %s", e.StatusCode, e.ErrorCode, e.Message)
	}
	return fmt.Sprintf("aptos: API error %d: %s", e.StatusCode, e.Message)
}

// Unwrap returns the underlying sentinel error based on the status code.
// This allows using errors.Is() to check for specific error categories.
func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusBadRequest:
		return ErrInvalidArgument
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return ErrTimeout
	case http.StatusServiceUnavailable:
		return ErrUnavailable
	case http.StatusInternalServerError:
		return ErrInternal
	default:
		return nil
	}
}

// IsNotFound returns true if this is a not found error.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsRateLimited returns true if this is a rate limit error.
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// TransactionError represents an error that occurred during transaction execution.
type TransactionError struct {
	// Hash is the transaction hash.
	Hash string

	// VMStatus is the VM execution status.
	VMStatus *VMStatus

	// Message is a human-readable error message.
	Message string
}

// Error implements the error interface.
func (e *TransactionError) Error() string {
	if e.VMStatus != nil {
		return fmt.Sprintf("aptos: transaction %s failed: %s (vm_status: %s at %s)",
			e.Hash, e.Message, e.VMStatus.Type, e.VMStatus.Location)
	}
	return fmt.Sprintf("aptos: transaction %s failed: %s", e.Hash, e.Message)
}

// Unwrap returns ErrTransactionFailed.
func (e *TransactionError) Unwrap() error {
	return ErrTransactionFailed
}

// SerializationError represents an error during BCS serialization.
type SerializationError struct {
	// Type is the name of the type being serialized.
	Type string

	// Field is the field name, if applicable.
	Field string

	// Cause is the underlying error.
	Cause error
}

// Error implements the error interface.
func (e *SerializationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("aptos: failed to serialize %s.%s: %v", e.Type, e.Field, e.Cause)
	}
	return fmt.Sprintf("aptos: failed to serialize %s: %v", e.Type, e.Cause)
}

// Unwrap returns the underlying cause.
func (e *SerializationError) Unwrap() error {
	return errors.Join(ErrSerialization, e.Cause)
}

// DeserializationError represents an error during BCS deserialization.
type DeserializationError struct {
	// Type is the name of the type being deserialized.
	Type string

	// Field is the field name, if applicable.
	Field string

	// Offset is the byte offset where the error occurred.
	Offset int

	// Cause is the underlying error.
	Cause error
}

// Error implements the error interface.
func (e *DeserializationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("aptos: failed to deserialize %s.%s at offset %d: %v",
			e.Type, e.Field, e.Offset, e.Cause)
	}
	return fmt.Sprintf("aptos: failed to deserialize %s at offset %d: %v",
		e.Type, e.Offset, e.Cause)
}

// Unwrap returns the underlying cause.
func (e *DeserializationError) Unwrap() error {
	return errors.Join(ErrDeserialization, e.Cause)
}

// WrapError wraps an error with additional context while preserving the error chain.
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsNotFound returns true if the error indicates a resource was not found.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsRateLimited returns true if the error indicates rate limiting.
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsTransactionFailed returns true if the error indicates transaction failure.
func IsTransactionFailed(err error) bool {
	return errors.Is(err, ErrTransactionFailed)
}

// IsTimeout returns true if the error indicates a timeout.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}
