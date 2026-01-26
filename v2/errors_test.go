package aptos

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "simple error",
			err: &APIError{
				StatusCode: 404,
				Message:    "Resource not found",
			},
			expected: "aptos: API error 404: Resource not found",
		},
		{
			name: "error with code",
			err: &APIError{
				StatusCode: 400,
				Message:    "Invalid request",
				ErrorCode:  "INVALID_INPUT",
			},
			expected: "aptos: API error 400 [INVALID_INPUT]: Invalid request",
		},
		{
			name: "error with VM status",
			err: &APIError{
				StatusCode: 400,
				Message:    "Transaction failed",
				VMStatus: &VMStatus{
					Type:      "move_abort",
					Location:  "0x1::coin",
					AbortCode: 100,
				},
			},
			expected: "aptos: API error 400: Transaction failed (vm_status: move_abort at 0x1::coin, code 100)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		expected   error
	}{
		{"not found", http.StatusNotFound, ErrNotFound},
		{"bad request", http.StatusBadRequest, ErrInvalidArgument},
		{"rate limited", http.StatusTooManyRequests, ErrRateLimited},
		{"timeout", http.StatusRequestTimeout, ErrTimeout},
		{"gateway timeout", http.StatusGatewayTimeout, ErrTimeout},
		{"unavailable", http.StatusServiceUnavailable, ErrUnavailable},
		{"internal error", http.StatusInternalServerError, ErrInternal},
		{"unknown", 418, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := &APIError{StatusCode: tt.statusCode, Message: "test"}
			if tt.expected != nil {
				assert.ErrorIs(t, err, tt.expected)
			} else {
				assert.NoError(t, err.Unwrap())
			}
		})
	}
}

func TestAPIError_IsNotFound(t *testing.T) {
	t.Parallel()

	err := &APIError{StatusCode: http.StatusNotFound, Message: "not found"}
	assert.True(t, err.IsNotFound())

	err2 := &APIError{StatusCode: http.StatusOK, Message: "ok"}
	assert.False(t, err2.IsNotFound())
}

func TestAPIError_IsRateLimited(t *testing.T) {
	t.Parallel()

	err := &APIError{StatusCode: http.StatusTooManyRequests, Message: "rate limited"}
	assert.True(t, err.IsRateLimited())

	err2 := &APIError{StatusCode: http.StatusOK, Message: "ok"}
	assert.False(t, err2.IsRateLimited())
}

func TestTransactionError(t *testing.T) {
	t.Parallel()

	err := &TransactionError{
		Hash:    "0x123",
		Message: "execution failed",
		VMStatus: &VMStatus{
			Type:     "move_abort",
			Location: "0x1::coin",
		},
	}

	assert.Contains(t, err.Error(), "0x123")
	assert.Contains(t, err.Error(), "execution failed")
	assert.Contains(t, err.Error(), "move_abort")
	assert.ErrorIs(t, err, ErrTransactionFailed)
}

func TestTransactionError_WithoutVMStatus(t *testing.T) {
	t.Parallel()

	err := &TransactionError{
		Hash:    "0x456",
		Message: "unknown error",
	}

	assert.Contains(t, err.Error(), "0x456")
	assert.Contains(t, err.Error(), "unknown error")
	assert.ErrorIs(t, err, ErrTransactionFailed)
}

func TestSerializationError(t *testing.T) {
	t.Parallel()

	cause := errors.New("overflow")
	err := &SerializationError{
		Type:  "Transaction",
		Field: "MaxGasAmount",
		Cause: cause,
	}

	assert.Contains(t, err.Error(), "Transaction")
	assert.Contains(t, err.Error(), "MaxGasAmount")
	assert.Contains(t, err.Error(), "overflow")
	assert.ErrorIs(t, err, ErrSerialization)
}

func TestSerializationError_NoField(t *testing.T) {
	t.Parallel()

	cause := errors.New("invalid type")
	err := &SerializationError{
		Type:  "AccountAddress",
		Cause: cause,
	}

	assert.Contains(t, err.Error(), "AccountAddress")
	assert.NotContains(t, err.Error(), ".")
	assert.ErrorIs(t, err, ErrSerialization)
}

func TestDeserializationError(t *testing.T) {
	t.Parallel()

	cause := errors.New("unexpected EOF")
	err := &DeserializationError{
		Type:   "Transaction",
		Field:  "Payload",
		Offset: 42,
		Cause:  cause,
	}

	assert.Contains(t, err.Error(), "Transaction")
	assert.Contains(t, err.Error(), "Payload")
	assert.Contains(t, err.Error(), "42")
	assert.Contains(t, err.Error(), "unexpected EOF")
	assert.ErrorIs(t, err, ErrDeserialization)
}

func TestWrapError(t *testing.T) {
	t.Parallel()

	// Nil error returns nil
	require.NoError(t, WrapError(nil, "should not appear"))

	// Wraps error with message
	original := ErrNotFound
	wrapped := WrapError(original, "failed to get account")
	require.Error(t, wrapped)
	assert.Contains(t, wrapped.Error(), "failed to get account")
	assert.ErrorIs(t, wrapped, ErrNotFound)
}

func TestErrorHelpers(t *testing.T) {
	t.Parallel()

	t.Run("IsNotFound", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsNotFound(ErrNotFound))
		assert.True(t, IsNotFound(&APIError{StatusCode: 404, Message: "not found"}))
		assert.False(t, IsNotFound(ErrTimeout))
	})

	t.Run("IsRateLimited", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsRateLimited(ErrRateLimited))
		assert.True(t, IsRateLimited(&APIError{StatusCode: 429, Message: "rate limited"}))
		assert.False(t, IsRateLimited(ErrNotFound))
	})

	t.Run("IsTransactionFailed", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsTransactionFailed(ErrTransactionFailed))
		assert.True(t, IsTransactionFailed(&TransactionError{Hash: "0x1", Message: "failed"}))
		assert.False(t, IsTransactionFailed(ErrNotFound))
	})

	t.Run("IsTimeout", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsTimeout(ErrTimeout))
		assert.True(t, IsTimeout(&APIError{StatusCode: 408, Message: "timeout"}))
		assert.False(t, IsTimeout(ErrNotFound))
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	// Ensure all sentinel errors are distinct
	sentinels := []error{
		ErrNotFound,
		ErrInvalidAddress,
		ErrInvalidArgument,
		ErrTransactionFailed,
		ErrTransactionExpired,
		ErrSequenceNumberMismatch,
		ErrInsufficientBalance,
		ErrRateLimited,
		ErrTimeout,
		ErrUnavailable,
		ErrInternal,
		ErrSignature,
		ErrSerialization,
		ErrDeserialization,
	}

	for i, e1 := range sentinels {
		for j, e2 := range sentinels {
			if i != j {
				assert.NotErrorIs(t, e1, e2)
			}
		}
	}
}
