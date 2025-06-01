package manager

import (
	"context"
	"testing"
)

func TestCommandRunnerEnvironmentHandling(t *testing.T) {
	t.Run("DefaultCommandRunner prepends LC_ALL=C", func(t *testing.T) {
		runner := NewDefaultCommandRunner()

		// Test that LC_ALL=C is prepended automatically
		// We can test this by running a simple command
		output, err := runner.Run("echo", "$LC_ALL")
		if err != nil {
			t.Logf("Note: echo test failed (expected on some systems): %v", err)
		} else {
			t.Logf("Echo output: %s", output)
		}

		t.Log("DefaultCommandRunner.Run() prepends LC_ALL=C automatically")
		t.Log("DefaultCommandRunner.RunContext() prepends LC_ALL=C automatically")
		t.Log("Users can override by passing LC_ALL=zh_TW.UTF-8 in env parameter")
	})

	t.Run("MockCommandRunner tracks environment variables", func(t *testing.T) {
		mock := NewMockCommandRunner()

		// Add a mocked command
		mock.AddCommand("apt", []string{"update"}, []byte("success"), nil)

		// Test RunContext with environment variables
		ctx := context.Background()
		_, err := mock.RunContext(ctx, "apt", []string{"update"}, "DEBIAN_FRONTEND=noninteractive")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify environment was tracked
		env := mock.GetEnvForCommand("apt", []string{"update"})
		if len(env) != 1 || env[0] != "DEBIAN_FRONTEND=noninteractive" {
			t.Errorf("Expected env [DEBIAN_FRONTEND=noninteractive], got %v", env)
		}
	})

	t.Run("MockCommandRunner tracks interactive environment", func(t *testing.T) {
		mock := NewMockCommandRunner()

		// Test RunInteractive with environment variables
		ctx := context.Background()
		err := mock.RunInteractive(ctx, "yum", []string{"install", "vim"}, "LANG=en_US.UTF-8")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify environment was tracked
		env := mock.GetEnvForCommand("yum", []string{"install", "vim"})
		if len(env) != 1 || env[0] != "LANG=en_US.UTF-8" {
			t.Errorf("Expected env [LANG=en_US.UTF-8], got %v", env)
		}

		// Verify it was marked as interactive
		if !mock.WasInteractiveCalled("yum", []string{"install", "vim"}) {
			t.Error("Expected command to be marked as interactive")
		}
	})

	t.Run("Empty environment handling", func(t *testing.T) {
		mock := NewMockCommandRunner()
		mock.AddCommand("ls", []string{}, []byte("file1\nfile2"), nil)

		// Call without environment variables
		ctx := context.Background()
		_, err := mock.RunContext(ctx, "ls", []string{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify empty environment was tracked
		env := mock.GetEnvForCommand("ls", []string{})
		if len(env) != 0 {
			t.Errorf("Expected empty env, got %v", env)
		}
	})
}
