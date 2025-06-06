// Package manager provides a package manager interface implementation
package manager

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

// CommandRunner provides an abstraction for executing system commands.
// All non-interactive commands automatically get LC_ALL=C for consistent output.
// All methods follow the Go context-first convention.
type CommandRunner interface {
	// Run executes a command with context support and LC_ALL=C, plus optional extra env.
	// Extra env vars are appended after LC_ALL=C, allowing override if needed.
	// Note: Later env values override earlier ones, so users can override LC_ALL=C
	// by passing their own LC_ALL value (e.g., "LC_ALL=zh_TW.UTF-8").
	// For commands with no args but extra env, pass nil or []string{} for args.
	// Example: Run(ctx, "apt", []string{"update"}, "DEBIAN_FRONTEND=noninteractive")
	// Example: Run(ctx, "yum", []string{"info", "vim"}, "LC_ALL=zh_TW.UTF-8") // Overrides default LC_ALL=C
	Run(ctx context.Context, name string, args []string, env ...string) ([]byte, error)

	// RunInteractive executes in interactive mode with stdin/stdout/stderr passthrough.
	// Does NOT prepend LC_ALL=C (preserves user's locale for interaction).
	// Returns only error as output is written directly to provided streams.
	RunInteractive(ctx context.Context, name string, args []string, env ...string) error
}

// DefaultCommandRunner implements CommandRunner using real system commands
type DefaultCommandRunner struct{}

// NewDefaultCommandRunner creates a new DefaultCommandRunner
func NewDefaultCommandRunner() *DefaultCommandRunner {
	return &DefaultCommandRunner{}
}

// Run executes with context support and LC_ALL=C, plus optional extra env
func (r *DefaultCommandRunner) Run(ctx context.Context, name string, args []string, env ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	// Prepend LC_ALL=C, then append any additional env vars
	// Note: Later values override earlier ones, so users can override LC_ALL=C if needed
	allEnv := append([]string{"LC_ALL=C"}, env...)
	cmd.Env = append(os.Environ(), allEnv...)

	return cmd.Output()
}

// RunInteractive executes in interactive mode with stdin/stdout/stderr passthrough
func (r *DefaultCommandRunner) RunInteractive(ctx context.Context, name string, args []string, env ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)

	// For interactive mode, preserve user's locale (no LC_ALL=C)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	// Connect stdin/stdout/stderr for interactive use
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// MockCommandRunner implements CommandRunner for testing
type MockCommandRunner struct {
	// Commands maps "command args" to expected output
	Commands map[string][]byte
	// Errors maps "command args" to expected errors
	Errors map[string]error
	// InteractiveCalls tracks interactive command calls for verification
	InteractiveCalls []string
	// EnvCalls tracks environment variables passed to Run/RunInteractive
	EnvCalls map[string][]string
}

// NewMockCommandRunner creates a new MockCommandRunner for testing
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		Commands:         make(map[string][]byte),
		Errors:           make(map[string]error),
		InteractiveCalls: []string{},
		EnvCalls:         make(map[string][]string),
	}
}

// Run returns mocked output for the given command
func (m *MockCommandRunner) Run(ctx context.Context, name string, args []string, env ...string) ([]byte, error) {
	// Build command key for lookup
	cmdKey := m.buildKey(name, args)

	// Track environment variables for testing
	m.EnvCalls[cmdKey] = env

	// Check if we have a mocked error for this command
	if err, exists := m.Errors[cmdKey]; exists {
		return nil, err
	}

	// Return mocked output if available
	if output, exists := m.Commands[cmdKey]; exists {
		return output, nil
	}

	// Default: return error when no mock is found (catches missing mocks in tests)
	return nil, errors.New("no mock found for command: " + cmdKey)
}

// RunInteractive simulates interactive command execution
func (m *MockCommandRunner) RunInteractive(ctx context.Context, name string, args []string, env ...string) error {
	// Track interactive calls for verification
	cmdKey := m.buildKey(name, args)
	m.InteractiveCalls = append(m.InteractiveCalls, cmdKey)

	// Track environment variables for testing
	m.EnvCalls[cmdKey] = env

	// Check if we have a mocked error for this command
	if err, exists := m.Errors[cmdKey]; exists {
		return err
	}

	return nil
}

// buildKey creates a consistent key for command lookup
func (m *MockCommandRunner) buildKey(name string, args []string) string {
	cmdKey := name
	if len(args) > 0 {
		for _, arg := range args {
			cmdKey += " " + arg
		}
	}
	return cmdKey
}

// AddCommand adds a mocked command response
func (m *MockCommandRunner) AddCommand(name string, args []string, output []byte, err error) {
	cmdKey := m.buildKey(name, args)
	m.Commands[cmdKey] = output
	if err != nil {
		m.Errors[cmdKey] = err
	}
}

// AddError adds a mocked command error (deprecated, use AddCommand with error)
func (m *MockCommandRunner) AddError(name string, args []string, err error) {
	cmdKey := m.buildKey(name, args)
	m.Errors[cmdKey] = err
}

// AddCommandWithEnv adds a mocked command response with environment consideration
// Note: In mock, we don't differentiate by env vars, but this method exists for API consistency
func (m *MockCommandRunner) AddCommandWithEnv(name string, args []string, env []string, output []byte, err error) {
	m.AddCommand(name, args, output, err)
}

// WasInteractiveCalled checks if an interactive command was called
func (m *MockCommandRunner) WasInteractiveCalled(name string, args []string) bool {
	cmdKey := m.buildKey(name, args)
	for _, call := range m.InteractiveCalls {
		if call == cmdKey {
			return true
		}
	}
	return false
}

// GetEnvForCommand returns the environment variables passed for a specific command
func (m *MockCommandRunner) GetEnvForCommand(name string, args []string) []string {
	cmdKey := m.buildKey(name, args)
	return m.EnvCalls[cmdKey]
}
