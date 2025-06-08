package manager

import (
	"fmt"
	"strings"
	"testing"
)

// TestLogger implements Logger interface for testing
type TestLogger struct {
	messages []string
}

func (t *TestLogger) Printf(format string, args ...interface{}) {
	t.messages = append(t.messages, fmt.Sprintf(format, args...))
}

func (t *TestLogger) GetMessages() []string {
	return t.messages
}

func (t *TestLogger) Clear() {
	t.messages = nil
}

func TestDefaultLogger(t *testing.T) {
	logger := DefaultLogger{}
	// Should not panic
	logger.Printf("test message: %s", "hello")
}

func TestBaseManagerLoggingPluggable(t *testing.T) {
	// Create a test logger
	testLogger := &TestLogger{}

	// Create a base manager
	runner := NewMockCommandRunner()
	mgr := NewBaseManager("test", CategorySystem, runner)

	// Set the custom logger
	mgr.SetLogger(testLogger)

	// Test verbose logging
	opts := DefaultOptions()
	opts.Verbose = true
	mgr.LogVerbosef(opts, "test verbose message: %s", "hello")

	messages := testLogger.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	expected := "[test] test verbose message: hello"
	if messages[0] != expected {
		t.Errorf("Expected '%s', got '%s'", expected, messages[0])
	}

	// Test debug logging
	testLogger.Clear()
	opts.Debug = true
	mgr.LogDebugf(opts, "test debug message: %s", "world")

	messages = testLogger.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	expected = "[test DEBUG] test debug message: world"
	if messages[0] != expected {
		t.Errorf("Expected '%s', got '%s'", expected, messages[0])
	}

	// Test dry-run logging
	testLogger.Clear()
	opts.DryRun = true
	mgr.HandleDryRun(opts, "install", []string{"vim", "curl"})

	messages = testLogger.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	if !strings.Contains(messages[0], "[test DRY-RUN] Would install packages: [vim curl]") {
		t.Errorf("Unexpected dry-run message: %s", messages[0])
	}
}

func TestLoggingDisabledWhenFlagsOff(t *testing.T) {
	testLogger := &TestLogger{}
	runner := NewMockCommandRunner()
	mgr := NewBaseManager("test", CategorySystem, runner)
	mgr.SetLogger(testLogger)

	// Create options with all logging disabled
	opts := DefaultOptions()
	opts.Verbose = false
	opts.Debug = false
	opts.DryRun = false

	// Try to log - should not create any messages
	mgr.LogVerbosef(opts, "should not appear")
	mgr.LogDebugf(opts, "should not appear")
	mgr.HandleDryRun(opts, "install", []string{"vim"})

	messages := testLogger.GetMessages()
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d: %v", len(messages), messages)
	}
}

func TestLoggingWithNilOptions(t *testing.T) {
	testLogger := &TestLogger{}
	runner := NewMockCommandRunner()
	mgr := NewBaseManager("test", CategorySystem, runner)
	mgr.SetLogger(testLogger)

	// Try to log with nil options - should not panic or log
	mgr.LogVerbosef(nil, "should not appear")
	mgr.LogDebugf(nil, "should not appear")
	mgr.HandleDryRun(nil, "install", []string{"vim"})

	messages := testLogger.GetMessages()
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages with nil options, got %d: %v", len(messages), messages)
	}
}

func TestSetLoggerWithNil(t *testing.T) {
	runner := NewMockCommandRunner()
	mgr := NewBaseManager("test", CategorySystem, runner)

	// Should not panic when setting nil logger
	mgr.SetLogger(nil)

	// Should still work with default logger
	opts := DefaultOptions()
	opts.Verbose = true
	mgr.LogVerbosef(opts, "test message")
}
