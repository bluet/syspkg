package manager

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestMockCommandRunner(t *testing.T) {
	tests := []struct {
		name           string
		commands       map[string][]byte
		errors         map[string]error
		testCommand    string
		testArgs       []string
		expectedOutput []byte
		expectedError  error
	}{
		{
			name: "successful command execution",
			commands: map[string][]byte{
				"rpm -q vim": []byte("vim-enhanced-8.0.1763-19.el8_6.4.x86_64\n"),
			},
			testCommand:    "rpm",
			testArgs:       []string{"-q", "vim"},
			expectedOutput: []byte("vim-enhanced-8.0.1763-19.el8_6.4.x86_64\n"),
			expectedError:  nil,
		},
		{
			name: "command returns error",
			errors: map[string]error{
				"rpm -q nonexistent": errors.New("package nonexistent is not installed"),
			},
			testCommand:    "rpm",
			testArgs:       []string{"-q", "nonexistent"},
			expectedOutput: nil,
			expectedError:  errors.New("package nonexistent is not installed"),
		},
		{
			name:           "command not mocked returns empty",
			commands:       map[string][]byte{},
			testCommand:    "unknown",
			testArgs:       []string{"command"},
			expectedOutput: []byte{},
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewMockCommandRunner()

			// Set up mocked commands using the proper methods
			for cmd, output := range tt.commands {
				// Parse the command string to extract name and args
				parts := strings.Fields(cmd)
				if len(parts) > 0 {
					name := parts[0]
					args := parts[1:]
					runner.AddCommand(name, args, output, nil)
				}
			}
			for cmd, err := range tt.errors {
				// Parse the command string to extract name and args
				parts := strings.Fields(cmd)
				if len(parts) > 0 {
					name := parts[0]
					args := parts[1:]
					runner.AddCommand(name, args, nil, err)
				}
			}

			// Test the command execution
			output, err := runner.Run(tt.testCommand, tt.testArgs...)

			// Verify results
			if string(output) != string(tt.expectedOutput) {
				t.Errorf("Expected output %q, got %q", string(tt.expectedOutput), string(output))
			}

			if (err == nil) != (tt.expectedError == nil) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("Expected error %q, got %q", tt.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestMockCommandRunnerAddMethods(t *testing.T) {
	runner := NewMockCommandRunner()

	// Test AddCommand
	runner.AddCommand("rpm", []string{"-q", "vim"}, []byte("vim-8.0.1763\n"), nil)
	output, err := runner.Run("rpm", "-q", "vim")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(output) != "vim-8.0.1763\n" {
		t.Errorf("Expected 'vim-8.0.1763\\n', got %q", string(output))
	}

	// Test AddError
	testErr := errors.New("test error")
	runner.AddError("rpm", []string{"-q", "missing"}, testErr)
	_, err = runner.Run("rpm", "-q", "missing")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %q", err.Error())
	}
}

func TestDefaultCommandRunner(t *testing.T) {
	runner := NewDefaultCommandRunner()

	// Test that timeout is set
	if runner.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", runner.Timeout)
	}

	// Test a simple command that should exist on most systems
	output, err := runner.Run("echo", "test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Note: With LC_ALL=C prepended, output should still be "test\n"
	if string(output) != "test\n" {
		t.Errorf("Expected 'test\\n', got %q", string(output))
	}
}

func TestDefaultCommandRunnerWithContext(t *testing.T) {
	runner := NewDefaultCommandRunner()

	// Test with a normal context
	ctx := context.Background()
	output, err := runner.RunContext(ctx, "echo", []string{"test"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(output) != "test\n" {
		t.Errorf("Expected 'test\\n', got %q", string(output))
	}

	// Test with a cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = runner.RunContext(cancelledCtx, "sleep", []string{"10"})
	if err == nil {
		t.Error("Expected error due to cancelled context")
	}
}
