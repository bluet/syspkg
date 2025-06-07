package manager

import (
	"errors"
	"fmt"
	"testing"
)

func TestStandardStatusTypes(t *testing.T) {
	tests := []struct {
		name           string
		statusFunc     func() error
		expectedStatus ReturnStatus
	}{
		{
			name: "explicit status with WrapReturn",
			statusFunc: func() error {
				return WrapReturn(StatusPermissionError, "requires sudo", nil)
			},
			expectedStatus: StatusPermissionError,
		},
		{
			name: "usage error status",
			statusFunc: func() error {
				return WrapReturn(StatusUsageError, "invalid package name", nil)
			},
			expectedStatus: StatusUsageError,
		},
		{
			name: "unavailable error status",
			statusFunc: func() error {
				return WrapReturn(StatusUnavailableError, "package not found", nil)
			},
			expectedStatus: StatusUnavailableError,
		},
		{
			name: "general error status",
			statusFunc: func() error {
				return WrapReturn(StatusGeneralError, "something went wrong", nil)
			},
			expectedStatus: StatusGeneralError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.statusFunc()

			var standardStatus *StandardStatus
			if !errors.As(err, &standardStatus) {
				t.Fatalf("Expected StandardStatus, got %T", err)
			}

			if standardStatus.Status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, standardStatus.Status)
			}
		})
	}
}

func TestWrapReturnUsage(t *testing.T) {
	// Test the main WrapReturn function
	tests := []struct {
		name     string
		status   ReturnStatus
		message  string
		wrapErr  error
		expected string
	}{
		{
			name:     "simple status",
			status:   StatusUsageError,
			message:  "invalid input",
			wrapErr:  nil,
			expected: "invalid input",
		},
		{
			name:     "wrapped error",
			status:   StatusPermissionError,
			message:  "access denied",
			wrapErr:  fmt.Errorf("underlying error"),
			expected: "access denied: underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapReturn(tt.status, tt.message, tt.wrapErr)

			if result.Error() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.Error())
			}

			var standardStatus *StandardStatus
			if !errors.As(result, &standardStatus) {
				t.Fatalf("Expected StandardStatus, got %T", result)
			}

			if standardStatus.Status != tt.status {
				t.Errorf("Expected status %v, got %v", tt.status, standardStatus.Status)
			}
		})
	}
}

func TestStandardStatusWrapping(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedStatus := WrapReturn(StatusUsageError, "wrapper message", originalErr)

	// Test Error() method
	expectedMessage := "wrapper message: original error"
	if wrappedStatus.Error() != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, wrappedStatus.Error())
	}

	// Test Unwrap() method
	if !errors.Is(wrappedStatus, originalErr) {
		t.Error("Expected wrapped status to contain original error")
	}
}

func TestStandardStatusWithoutWrappedError(t *testing.T) {
	status := WrapReturn(StatusUsageError, "standalone message", nil)

	expectedMessage := "standalone message"
	if status.Error() != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, status.Error())
	}

	var standardStatus *StandardStatus
	if !errors.As(status, &standardStatus) {
		t.Fatalf("Expected StandardStatus, got %T", status)
	}

	if standardStatus.Unwrap() != nil {
		t.Error("Expected Unwrap() to return nil when no wrapped error")
	}
}
