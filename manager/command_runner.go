// Package manager provides a package manager interface implementation
package manager

import (
	"context"
	"os/exec"
	"time"
)

// CommandRunner abstracts command execution to enable testing and mocking
type CommandRunner interface {
	// Output executes a command and returns its stdout output
	Output(name string, args ...string) ([]byte, error)
	// OutputWithContext executes a command with context for timeout/cancellation
	OutputWithContext(ctx context.Context, name string, args ...string) ([]byte, error)
}

// OSCommandRunner implements CommandRunner using real system commands
type OSCommandRunner struct {
	Timeout time.Duration // Default timeout for commands
}

// NewOSCommandRunner creates a new OSCommandRunner with default timeout
func NewOSCommandRunner() *OSCommandRunner {
	return &OSCommandRunner{
		Timeout: 30 * time.Second, // Default 30 second timeout
	}
}

// Output executes a command and returns its stdout output
func (r *OSCommandRunner) Output(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	return r.OutputWithContext(ctx, name, args...)
}

// OutputWithContext executes a command with context for timeout/cancellation
func (r *OSCommandRunner) OutputWithContext(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

// MockCommandRunner implements CommandRunner for testing
type MockCommandRunner struct {
	// Commands maps "command args" to expected output
	Commands map[string][]byte
	// Errors maps "command args" to expected errors
	Errors map[string]error
}

// NewMockCommandRunner creates a new MockCommandRunner for testing
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		Commands: make(map[string][]byte),
		Errors:   make(map[string]error),
	}
}

// Output returns mocked output for the given command
func (m *MockCommandRunner) Output(name string, args ...string) ([]byte, error) {
	return m.OutputWithContext(context.Background(), name, args...)
}

// OutputWithContext returns mocked output for the given command
func (m *MockCommandRunner) OutputWithContext(ctx context.Context, name string, args ...string) ([]byte, error) {
	// Build command key for lookup
	cmdKey := name
	if len(args) > 0 {
		for _, arg := range args {
			cmdKey += " " + arg
		}
	}

	// Check if we have a mocked error for this command
	if err, exists := m.Errors[cmdKey]; exists {
		return nil, err
	}

	// Return mocked output if available
	if output, exists := m.Commands[cmdKey]; exists {
		return output, nil
	}

	// Default: return empty output with no error
	return []byte{}, nil
}

// AddCommand adds a mocked command response
func (m *MockCommandRunner) AddCommand(name string, args []string, output []byte) {
	cmdKey := name
	if len(args) > 0 {
		for _, arg := range args {
			cmdKey += " " + arg
		}
	}
	m.Commands[cmdKey] = output
}

// AddError adds a mocked command error
func (m *MockCommandRunner) AddError(name string, args []string, err error) {
	cmdKey := name
	if len(args) > 0 {
		for _, arg := range args {
			cmdKey += " " + arg
		}
	}
	m.Errors[cmdKey] = err
}
