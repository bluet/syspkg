package manager

import (
	"context"
	"testing"
	"time"
)

func TestRegistry_SearchAllConcurrent(t *testing.T) {
	registry := NewRegistry()

	// Register a mock plugin for testing
	mockPlugin := &MockPlugin{
		manager: &MockPackageManager{},
	}

	err := registry.Register("test", mockPlugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	ctx := context.Background()
	results := registry.SearchAllConcurrent(ctx, []string{"vim"}, nil)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if _, exists := results["test"]; !exists {
		t.Errorf("Expected result for 'test' manager")
	}
}

func TestRegistry_ListInstalledAllConcurrent(t *testing.T) {
	registry := NewRegistry()

	// Register a mock plugin for testing
	mockPlugin := &MockPlugin{
		manager: &MockPackageManager{},
	}

	err := registry.Register("test", mockPlugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	ctx := context.Background()
	results := registry.ListInstalledAllConcurrent(ctx, nil)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if _, exists := results["test"]; !exists {
		t.Errorf("Expected result for 'test' manager")
	}
}

func TestRegistry_StatusAllConcurrent(t *testing.T) {
	registry := NewRegistry()

	// Register a mock plugin for testing
	mockPlugin := &MockPlugin{
		manager: &MockPackageManager{},
	}

	err := registry.Register("test", mockPlugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	ctx := context.Background()
	results := registry.StatusAllConcurrent(ctx, nil)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if _, exists := results["test"]; !exists {
		t.Errorf("Expected result for 'test' manager")
	}
}

func TestConcurrentPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	registry := NewRegistry()

	// Register multiple slow mock managers
	for i := 0; i < 3; i++ {
		mockPlugin := &MockPlugin{
			manager: &SlowMockPackageManager{
				MockPackageManager: MockPackageManager{},
				delay:              100 * time.Millisecond, // 100ms delay per operation
			},
		}

		err := registry.Register(string(rune('a'+i)), mockPlugin)
		if err != nil {
			t.Fatalf("Failed to register plugin: %v", err)
		}
	}

	ctx := context.Background()

	// Test concurrent vs sequential performance
	start := time.Now()
	_ = registry.SearchAllConcurrent(ctx, []string{"vim"}, nil)
	concurrentTime := time.Since(start)

	// Sequential would take ~300ms (3 managers * 100ms each)
	// Concurrent should take ~100ms (all run in parallel)
	if concurrentTime > 200*time.Millisecond {
		t.Errorf("Concurrent search took too long: %v (expected < 200ms)", concurrentTime)
	}

	t.Logf("Concurrent search completed in: %v", concurrentTime)
}

// MockPlugin for testing
type MockPlugin struct {
	manager PackageManager
}

func (p *MockPlugin) CreateManager() PackageManager {
	return p.manager
}

func (p *MockPlugin) GetPriority() int {
	return 50
}

// SlowMockPackageManager simulates slow operations for performance testing
type SlowMockPackageManager struct {
	MockPackageManager
	delay time.Duration
}

func (m *SlowMockPackageManager) Search(ctx context.Context, query []string, opts *Options) ([]PackageInfo, error) {
	time.Sleep(m.delay)
	return []PackageInfo{}, nil
}

func (m *SlowMockPackageManager) List(ctx context.Context, filter ListFilter, opts *Options) ([]PackageInfo, error) {
	time.Sleep(m.delay)
	return []PackageInfo{}, nil
}

func (m *SlowMockPackageManager) Status(ctx context.Context, opts *Options) (ManagerStatus, error) {
	time.Sleep(m.delay)
	return ManagerStatus{Available: true, Healthy: true}, nil
}

func (m *SlowMockPackageManager) Install(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	time.Sleep(m.delay)
	return []PackageInfo{{Name: packages[0], Version: "1.0", Status: "installed"}}, nil
}

func (m *SlowMockPackageManager) Remove(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	time.Sleep(m.delay)
	return []PackageInfo{{Name: packages[0], Version: "1.0", Status: "removed"}}, nil
}

func (m *SlowMockPackageManager) Verify(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	time.Sleep(m.delay)
	return []PackageInfo{{Name: packages[0], Version: "1.0", Status: "verified"}}, nil
}

func (m *SlowMockPackageManager) Refresh(ctx context.Context, opts *Options) error {
	time.Sleep(m.delay)
	return nil
}

func (m *SlowMockPackageManager) Upgrade(ctx context.Context, packages []string, opts *Options) ([]PackageInfo, error) {
	time.Sleep(m.delay)
	return []PackageInfo{{Name: "upgraded-pkg", Version: "2.0", Status: "upgraded"}}, nil
}

func (m *SlowMockPackageManager) Clean(ctx context.Context, opts *Options) error {
	time.Sleep(m.delay)
	return nil
}

func (m *SlowMockPackageManager) AutoRemove(ctx context.Context, opts *Options) ([]PackageInfo, error) {
	time.Sleep(m.delay)
	return []PackageInfo{{Name: "orphaned-pkg", Version: "1.0", Status: "removed"}}, nil
}

func TestRegistry_InstallAllConcurrent(t *testing.T) {
	registry := NewRegistry()

	// Register a mock plugin for testing
	mockPlugin := &MockPlugin{
		manager: &MockPackageManager{},
	}

	err := registry.Register("test", mockPlugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	ctx := context.Background()
	results := registry.InstallAllConcurrent(ctx, []string{"vim"}, nil)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if _, exists := results["test"]; !exists {
		t.Errorf("Expected result for 'test' manager")
	}
}

func TestRegistry_RemoveAllConcurrent(t *testing.T) {
	registry := NewRegistry()

	// Register a mock plugin for testing
	mockPlugin := &MockPlugin{
		manager: &MockPackageManager{},
	}

	err := registry.Register("test", mockPlugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	ctx := context.Background()
	results := registry.RemoveAllConcurrent(ctx, []string{"vim"}, nil)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if _, exists := results["test"]; !exists {
		t.Errorf("Expected result for 'test' manager")
	}
}

func TestRegistry_VerifyAllConcurrent(t *testing.T) {
	registry := NewRegistry()

	// Register a mock plugin for testing
	mockPlugin := &MockPlugin{
		manager: &MockPackageManager{},
	}

	err := registry.Register("test", mockPlugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	ctx := context.Background()
	results := registry.VerifyAllConcurrent(ctx, []string{"vim"}, nil)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if _, exists := results["test"]; !exists {
		t.Errorf("Expected result for 'test' manager")
	}
}

func TestAllConcurrentOperationsPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	registry := NewRegistry()

	// Register multiple slow mock managers
	for i := 0; i < 3; i++ {
		mockPlugin := &MockPlugin{
			manager: &SlowMockPackageManager{
				MockPackageManager: MockPackageManager{},
				delay:              100 * time.Millisecond, // 100ms delay per operation
			},
		}

		err := registry.Register(string(rune('a'+i)), mockPlugin)
		if err != nil {
			t.Fatalf("Failed to register plugin: %v", err)
		}
	}

	ctx := context.Background()

	// Test all concurrent operations
	operations := []struct {
		name string
		test func() time.Duration
	}{
		{"Install", func() time.Duration {
			start := time.Now()
			_ = registry.InstallAllConcurrent(ctx, []string{"vim"}, nil)
			return time.Since(start)
		}},
		{"Remove", func() time.Duration {
			start := time.Now()
			_ = registry.RemoveAllConcurrent(ctx, []string{"vim"}, nil)
			return time.Since(start)
		}},
		{"Verify", func() time.Duration {
			start := time.Now()
			_ = registry.VerifyAllConcurrent(ctx, []string{"vim"}, nil)
			return time.Since(start)
		}},
		{"Refresh", func() time.Duration {
			start := time.Now()
			_ = registry.RefreshAllConcurrent(ctx, nil)
			return time.Since(start)
		}},
		{"Upgrade", func() time.Duration {
			start := time.Now()
			_ = registry.UpgradeAllConcurrent(ctx, []string{}, nil)
			return time.Since(start)
		}},
		{"Clean", func() time.Duration {
			start := time.Now()
			_ = registry.CleanAllConcurrent(ctx, nil)
			return time.Since(start)
		}},
		{"AutoRemove", func() time.Duration {
			start := time.Now()
			_ = registry.AutoRemoveAllConcurrent(ctx, nil)
			return time.Since(start)
		}},
	}

	for _, op := range operations {
		duration := op.test()
		// Sequential would take ~300ms (3 managers * 100ms each)
		// Concurrent should take ~100ms (all run in parallel)
		if duration > 200*time.Millisecond {
			t.Errorf("%s concurrent operation took too long: %v (expected < 200ms)", op.name, duration)
		}
		t.Logf("%s concurrent operation completed in: %v", op.name, duration)
	}
}
