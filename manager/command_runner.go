// Package manager provides a package manager interface implementation
package manager

import (
	"context"
	"os"
	"os/exec"
	"time"
)

// CommandRunner provides an abstraction for executing system commands.
// All non-interactive commands automatically get LC_ALL=C for consistent output.
type CommandRunner interface {
	// Run executes a command with LC_ALL=C for consistent English output.
	// This is the primary method for simple non-interactive commands.
	Run(name string, args ...string) ([]byte, error)

	// RunContext executes with context support and LC_ALL=C, plus optional extra env.
	// Extra env vars are appended after LC_ALL=C, allowing override if needed.
	// Note: Later env values override earlier ones, so users can override LC_ALL=C
	// by passing their own LC_ALL value (e.g., "LC_ALL=zh_TW.UTF-8").
	// For commands with no args but extra env, pass nil or []string{} for args.
	// Example: RunContext(ctx, "apt", []string{"update"}, "DEBIAN_FRONTEND=noninteractive")
	// Example: RunContext(ctx, "yum", []string{"info", "vim"}, "LC_ALL=zh_TW.UTF-8") // Overrides default LC_ALL=C
	RunContext(ctx context.Context, name string, args []string, env ...string) ([]byte, error)

	// RunInteractive executes in interactive mode with stdin/stdout/stderr passthrough.
	// Does NOT prepend LC_ALL=C (preserves user's locale for interaction).
	// Returns only error as output is written directly to provided streams.
	RunInteractive(ctx context.Context, name string, args []string, env ...string) error
}

// DefaultCommandRunner implements CommandRunner using real system commands
type DefaultCommandRunner struct {
	Timeout time.Duration // Default timeout for commands
}

// NewDefaultCommandRunner creates a new DefaultCommandRunner with default timeout
func NewDefaultCommandRunner() *DefaultCommandRunner {
	return &DefaultCommandRunner{
		Timeout: 30 * time.Second, // Default 30 second timeout
	}
}

// Run executes a command with LC_ALL=C for consistent English output
func (r *DefaultCommandRunner) Run(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	return r.RunContext(ctx, name, args)
}

// RunContext executes with context support and LC_ALL=C, plus optional extra env
func (r *DefaultCommandRunner) RunContext(ctx context.Context, name string, args []string, env ...string) ([]byte, error) {
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
	// EnvCalls tracks environment variables passed to RunContext/RunInteractive
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

// Run returns mocked output for the given command with LC_ALL=C
func (m *MockCommandRunner) Run(name string, args ...string) ([]byte, error) {
	return m.RunContext(context.Background(), name, args)
}

// RunContext returns mocked output for the given command
func (m *MockCommandRunner) RunContext(ctx context.Context, name string, args []string, env ...string) ([]byte, error) {
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

	// Default: return empty output with no error
	return []byte{}, nil
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
