package manager

import (
	"context"
	"errors"
	"strings"
	"testing"
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
			name:           "command not mocked returns error",
			commands:       map[string][]byte{},
			testCommand:    "unknown",
			testArgs:       []string{"command"},
			expectedOutput: nil,
			expectedError:  errors.New("no mock found for command: unknown command"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewMockCommandRunner()

			// Set up mocked commands using the proper methods
			for cmd, output := range tt.commands {
				// Parse the command string to extract name and args
				parts := strings.Fields(cmd)
				if len(parts) == 0 {
					t.Errorf("Invalid empty command string: %q", cmd)
					continue
				}
				name := parts[0]
				args := parts[1:]
				runner.AddCommand(name, args, output, nil)
			}
			for cmd, err := range tt.errors {
				// Parse the command string to extract name and args
				parts := strings.Fields(cmd)
				if len(parts) == 0 {
					t.Errorf("Invalid empty command string: %q", cmd)
					continue
				}
				name := parts[0]
				args := parts[1:]
				runner.AddCommand(name, args, nil, err)
			}

			// Test the command execution
			result, err := runner.Run(context.Background(), tt.testCommand, tt.testArgs)

			// Verify results
			if (err == nil) != (tt.expectedError == nil) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			if err == nil {
				// Success case: result must not be nil
				if result == nil {
					t.Fatal("Expected result, got nil")
				}
				if string(result.Output) != string(tt.expectedOutput) {
					t.Errorf("Expected output %q, got %q", string(tt.expectedOutput), string(result.Output))
				}
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
	result, err := runner.Run(context.Background(), "rpm", []string{"-q", "vim"})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if string(result.Output) != "vim-8.0.1763\n" {
		t.Errorf("Expected 'vim-8.0.1763\\n', got %q", string(result.Output))
	}

	// Test AddError
	testErr := errors.New("test error")
	runner.AddError("rpm", []string{"-q", "missing"}, testErr)
	_, err = runner.Run(context.Background(), "rpm", []string{"-q", "missing"})

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %q", err.Error())
	}
}

func TestDefaultCommandRunner(t *testing.T) {
	runner := NewDefaultCommandRunner()

	// Test a simple command that should exist on most systems
	result, err := runner.Run(context.Background(), "echo", []string{"test"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	// Note: With LC_ALL=C prepended, output should still be "test\n"
	if string(result.Output) != "test\n" {
		t.Errorf("Expected 'test\\n', got %q", string(result.Output))
	}
}

func TestDefaultCommandRunnerWithContext(t *testing.T) {
	runner := NewDefaultCommandRunner()

	// Test with a normal context
	ctx := context.Background()
	result, err := runner.Run(ctx, "echo", []string{"test"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if string(result.Output) != "test\n" {
		t.Errorf("Expected 'test\\n', got %q", string(result.Output))
	}

	// Test with a cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = runner.Run(cancelledCtx, "sleep", []string{"10"})
	if err == nil {
		t.Error("Expected error due to cancelled context")
	}
}
