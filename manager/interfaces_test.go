package manager

import (
	"context"
	"testing"
)

func TestPackageInfoStructure(t *testing.T) {
	pkg := PackageInfo{
		Name:        "test-package",
		Version:     "1.0.0",
		NewVersion:  "1.1.0",
		Status:      StatusInstalled,
		Description: "Test package",
		Category:    "test",
		ManagerName: "test-manager",
		Metadata: map[string]interface{}{
			"arch":   "amd64",
			"source": "test",
		},
	}

	if pkg.Name != "test-package" {
		t.Errorf("Expected name 'test-package', got '%s'", pkg.Name)
	}

	if pkg.Status != StatusInstalled {
		t.Errorf("Expected status '%s', got '%s'", StatusInstalled, pkg.Status)
	}

	if pkg.ManagerName != "test-manager" {
		t.Errorf("Expected manager '%s', got '%s'", "test-manager", pkg.ManagerName)
	}

	if arch, ok := pkg.Metadata["arch"]; !ok || arch != "amd64" {
		t.Errorf("Expected arch 'amd64', got '%v'", arch)
	}
}

func TestManagerStatus(t *testing.T) {
	status := ManagerStatus{
		Available:      true,
		Healthy:        true,
		Version:        "1.0.0",
		LastRefresh:    "2024-01-01",
		CacheSize:      1024,
		PackageCount:   100,
		InstalledCount: 50,
		Issues:         []string{"warning: test"},
		Metadata: map[string]interface{}{
			"config": "/etc/test",
		},
	}

	if !status.Available {
		t.Error("Expected Available to be true")
	}

	if !status.Healthy {
		t.Error("Expected Healthy to be true")
	}

	if status.PackageCount != 100 {
		t.Errorf("Expected package count 100, got %d", status.PackageCount)
	}

	if len(status.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(status.Issues))
	}
}

func TestOptionsStructure(t *testing.T) {
	opts := DefaultOptions()

	if opts.DryRun {
		t.Error("Expected DryRun to be false by default")
	}

	if opts.Interactive {
		t.Error("Expected Interactive to be false by default")
	}

	if opts.Verbose {
		t.Error("Expected Verbose to be false by default")
	}

	if !opts.GlobalScope {
		t.Error("Expected GlobalScope to be true by default")
	}

	if opts.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}

	if opts.CustomArgs == nil {
		t.Error("Expected CustomArgs to be initialized")
	}

	if opts.Tags == nil {
		t.Error("Expected Tags to be initialized")
	}
}

func TestConstants(t *testing.T) {
	// Test status constants
	expectedStatuses := []string{
		StatusInstalled,
		StatusAvailable,
		StatusUpgradable,
		StatusUnknown,
	}

	for _, status := range expectedStatuses {
		if status == "" {
			t.Errorf("Status constant should not be empty")
		}
	}

	// Test manager type constants
	expectedTypes := []string{
		CategorySystem,
		CategoryLanguage,
		CategoryVersion,
		CategoryContainer,
		CategoryGame,
		CategoryScientific,
		CategoryBuild,
		CategoryApp,
	}

	for _, managerType := range expectedTypes {
		if managerType == "" {
			t.Errorf("Manager type constant should not be empty")
		}
	}

	// Test filter constants
	expectedFilters := []ListFilter{
		FilterInstalled,
		FilterAvailable,
		FilterUpgradable,
		FilterAll,
	}

	for _, filter := range expectedFilters {
		if filter == "" {
			t.Errorf("Filter constant should not be empty")
		}
	}
}

func TestPackageManagerInterface(t *testing.T) {
	// This test ensures the interface is properly defined
	// We can't instantiate the interface directly, but we can check
	// that a mock implementation works

	mock := &MockPackageManager{}
	var pm PackageManager = mock

	ctx := context.Background()
	opts := DefaultOptions()

	// Test basic info methods
	name := pm.GetName()
	if name == "" {
		t.Error("GetName should return non-empty string")
	}

	managerType := pm.GetCategory()
	if managerType == "" {
		t.Error("GetType should return non-empty string")
	}

	// Test that all methods are callable (even if they return errors)
	pm.IsAvailable()
	_, _ = pm.GetVersion()
	_, _ = pm.Search(ctx, []string{"test"}, opts)
	_, _ = pm.List(ctx, FilterInstalled, opts)
	_, _ = pm.Install(ctx, []string{"test"}, opts)
	_, _ = pm.Remove(ctx, []string{"test"}, opts)
	_, _ = pm.GetInfo(ctx, "test", opts)
	_ = pm.Refresh(ctx, opts)
	_, _ = pm.Upgrade(ctx, []string{"test"}, opts)
	_ = pm.Clean(ctx, opts)
	_, _ = pm.AutoRemove(ctx, opts)
	_, _ = pm.Verify(ctx, []string{"test"}, opts)
	_, _ = pm.Status(ctx, opts)
}

// MockPackageManager for interface testing
type MockPackageManager struct{}

func (m *MockPackageManager) GetName() string             { return "mock" }
func (m *MockPackageManager) GetCategory() string         { return CategorySystem }
func (m *MockPackageManager) IsAvailable() bool           { return true }
func (m *MockPackageManager) GetVersion() (string, error) { return "1.0.0", nil }
func (m *MockPackageManager) Search(ctx context.Context, query []string, opts *Options) ([]PackageInfo, error) {
	return []PackageInfo{}, ErrOperationNotSupported
}
func (m *MockPackageManager) List(ctx context.Context, filter ListFilter, opts *Options) ([]PackageInfo, error) {
	return []PackageInfo{}, ErrOperationNotSupported
}
func (m *MockPackageManager) Install(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	return []PackageInfo{}, ErrOperationNotSupported
}
func (m *MockPackageManager) Remove(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	return []PackageInfo{}, ErrOperationNotSupported
}
func (m *MockPackageManager) GetInfo(ctx context.Context, packageName string, opts *Options) (PackageInfo, error) {
	return PackageInfo{}, ErrOperationNotSupported
}
func (m *MockPackageManager) Refresh(ctx context.Context, opts *Options) error {
	return ErrOperationNotSupported
}
func (m *MockPackageManager) Upgrade(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	return []PackageInfo{}, ErrOperationNotSupported
}
func (m *MockPackageManager) Clean(ctx context.Context, opts *Options) error {
	return ErrOperationNotSupported
}
func (m *MockPackageManager) AutoRemove(ctx context.Context, opts *Options) ([]PackageInfo, error) {
	return []PackageInfo{}, ErrOperationNotSupported
}
func (m *MockPackageManager) Verify(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	return []PackageInfo{}, ErrOperationNotSupported
}
func (m *MockPackageManager) Status(ctx context.Context, opts *Options) (ManagerStatus, error) {
	return ManagerStatus{}, ErrOperationNotSupported
}
