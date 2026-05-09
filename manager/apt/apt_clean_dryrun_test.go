package apt_test

import (
	"testing"

	"github.com/bluet/syspkg/manager"
	"github.com/bluet/syspkg/manager/apt"
)

// TestCleanRespectsDryRun is the regression test for the security-relevant bug
// where Clean() executed `apt autoclean` even when opts.DryRun was true.
// Behavior contract: Clean(DryRun=true) MUST NOT execute any underlying command.
func TestCleanRespectsDryRun(t *testing.T) {
	mockRunner := manager.NewMockCommandRunner()
	pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

	if err := pm.Clean(&manager.Options{DryRun: true}); err != nil {
		t.Fatalf("Clean(DryRun=true) returned error: %v", err)
	}

	if got := len(mockRunner.EnvCalls); got != 0 {
		t.Errorf("Clean(DryRun=true) executed %d non-interactive command(s); expected 0. Calls: %v",
			got, mockRunner.EnvCalls)
	}
	if got := len(mockRunner.InteractiveCalls); got != 0 {
		t.Errorf("Clean(DryRun=true) executed %d interactive command(s); expected 0. Calls: %v",
			got, mockRunner.InteractiveCalls)
	}
}

// TestCleanRunsWithoutDryRun guards against the Clean(DryRun) fix being
// implemented as a blanket no-op. Without DryRun, Clean MUST invoke
// `apt autoclean`.
func TestCleanRunsWithoutDryRun(t *testing.T) {
	mockRunner := manager.NewMockCommandRunner()
	mockRunner.AddCommand("apt", []string{"autoclean"}, []byte("Reading package lists...\n"), nil)
	pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

	if err := pm.Clean(&manager.Options{DryRun: false}); err != nil {
		t.Fatalf("Clean(DryRun=false) returned error: %v", err)
	}

	if _, ok := mockRunner.EnvCalls["apt autoclean"]; !ok {
		t.Errorf("Clean(DryRun=false) didn't invoke 'apt autoclean'. Recorded calls: %v",
			mockRunner.EnvCalls)
	}
}

// TestCleanRespectsDryRunWithNilOptsDefault verifies the nil-opts branch:
// when opts == nil, the code path defaults DryRun to false, so Clean
// SHOULD execute (proving the nil-opts default is preserved by the fix).
func TestCleanRunsWithNilOpts(t *testing.T) {
	mockRunner := manager.NewMockCommandRunner()
	mockRunner.AddCommand("apt", []string{"autoclean"}, []byte("Reading package lists...\n"), nil)
	pm := apt.NewPackageManagerWithCustomRunner(mockRunner)

	if err := pm.Clean(nil); err != nil {
		t.Fatalf("Clean(nil) returned error: %v", err)
	}

	if _, ok := mockRunner.EnvCalls["apt autoclean"]; !ok {
		t.Errorf("Clean(nil) didn't invoke 'apt autoclean'. Recorded calls: %v",
			mockRunner.EnvCalls)
	}
}
