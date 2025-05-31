package manager

import (
	"context"
	"errors"
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

			// Set up mocked commands
			for cmd, output := range tt.commands {
				runner.Commands[cmd] = output
			}
			for cmd, err := range tt.errors {
				runner.Errors[cmd] = err
			}

			// Test the command execution
			output, err := runner.Output(tt.testCommand, tt.testArgs...)

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
	runner.AddCommand("rpm", []string{"-q", "vim"}, []byte("vim-8.0.1763\n"))
	output, err := runner.Output("rpm", "-q", "vim")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(output) != "vim-8.0.1763\n" {
		t.Errorf("Expected 'vim-8.0.1763\\n', got %q", string(output))
	}

	// Test AddError
	testErr := errors.New("test error")
	runner.AddError("rpm", []string{"-q", "missing"}, testErr)
	_, err = runner.Output("rpm", "-q", "missing")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %q", err.Error())
	}
}

func TestOSCommandRunner(t *testing.T) {
	runner := NewOSCommandRunner()

	// Test that timeout is set
	if runner.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", runner.Timeout)
	}

	// Test a simple command that should exist on most systems
	output, err := runner.Output("echo", "test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(output) != "test\n" {
		t.Errorf("Expected 'test\\n', got %q", string(output))
	}
}

func TestOSCommandRunnerWithContext(t *testing.T) {
	runner := NewOSCommandRunner()

	// Test with a normal context
	ctx := context.Background()
	output, err := runner.OutputWithContext(ctx, "echo", "test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(output) != "test\n" {
		t.Errorf("Expected 'test\\n', got %q", string(output))
	}

	// Test with a cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = runner.OutputWithContext(cancelledCtx, "sleep", "10")
	if err == nil {
		t.Error("Expected error due to cancelled context")
	}
}
