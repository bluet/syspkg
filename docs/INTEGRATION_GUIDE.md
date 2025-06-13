# Production Integration Guide

This guide covers advanced patterns for integrating go-syspkg into production services, long-running applications, and enterprise systems.

## Overview

While the [examples/](../examples/) directory shows basic integration patterns, this guide focuses on production concerns like monitoring, logging, performance, and reliability for services that depend on go-syspkg.


## 🔧 Long-Running Service Patterns

### Service Initialization

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/bluet/syspkg/manager"
    _ "github.com/bluet/syspkg/manager/apt"
    _ "github.com/bluet/syspkg/manager/yum"
)

type PackageService struct {
    registry *manager.Registry
    managers map[string]manager.PackageManager
    logger   *log.Logger
}

func NewPackageService(logger *log.Logger) (*PackageService, error) {
    registry := manager.GetGlobalRegistry()
    managers := registry.GetAvailable()

    if len(managers) == 0 {
        return nil, fmt.Errorf("no package managers available")
    }

    logger.Printf("Initialized with %d package managers", len(managers))
    for name, pm := range managers {
        if pm.IsAvailable() {
            logger.Printf("✅ %s available", name)
        } else {
            logger.Printf("❌ %s not available", name)
        }
    }

    return &PackageService{
        registry: registry,
        managers: managers,
        logger:   logger,
    }, nil
}

func (s *PackageService) Start(ctx context.Context) error {
    mux := http.NewServeMux()

    // Health check endpoint
    mux.HandleFunc("/health", s.healthHandler)

    // Package operations
    mux.HandleFunc("/packages/search", s.searchHandler)
    mux.HandleFunc("/packages/status", s.statusHandler)

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    // Graceful shutdown
    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        server.Shutdown(shutdownCtx)
    }()

    s.logger.Println("Package service started on :8080")
    return server.ListenAndServe()
}

func main() {
    logger := log.New(os.Stdout, "[PackageService] ", log.LstdFlags)

    service, err := NewPackageService(logger)
    if err != nil {
        logger.Fatal(err)
    }

    // Graceful shutdown handling
    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    if err := service.Start(ctx); err != nil && err != http.ErrServerClosed {
        logger.Fatal(err)
    }

    logger.Println("Service shut down gracefully")
}
```

### Health Check Implementation

```go
func (s *PackageService) healthHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    health := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
        "managers": make(map[string]map[string]interface{}),
    }

    overallHealthy := true

    for name, pm := range s.managers {
        managerHealth := map[string]interface{}{
            "available": pm.IsAvailable(),
        }

        // Test basic operation
        if pm.IsAvailable() {
            status, err := pm.Status(ctx, manager.DefaultOptions())
            if err != nil {
                managerHealth["healthy"] = false
                managerHealth["error"] = err.Error()
                overallHealthy = false
            } else {
                managerHealth["healthy"] = status.Healthy
                managerHealth["version"] = status.Version
                if !status.Healthy {
                    overallHealthy = false
                }
            }
        } else {
            managerHealth["healthy"] = false
            overallHealthy = false
        }

        health["managers"].(map[string]map[string]interface{})[name] = managerHealth
    }

    if !overallHealthy {
        health["status"] = "degraded"
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

## 📊 Production Considerations

### Structured Logging Integration

```go
import (
    "github.com/sirupsen/logrus"
    "github.com/bluet/syspkg/manager"
)

type LoggingPackageManager struct {
    manager.PackageManager
    logger *logrus.Logger
    name   string
}

func NewLoggingPackageManager(pm manager.PackageManager, logger *logrus.Logger) *LoggingPackageManager {
    return &LoggingPackageManager{
        PackageManager: pm,
        logger:        logger,
        name:          pm.GetName(),
    }
}

func (l *LoggingPackageManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    start := time.Now()

    l.logger.WithFields(logrus.Fields{
        "manager": l.name,
        "operation": "search",
        "query": query,
        "dry_run": opts != nil && opts.DryRun,
    }).Info("Starting package search")

    packages, err := l.PackageManager.Search(ctx, query, opts)

    duration := time.Since(start)
    fields := logrus.Fields{
        "manager": l.name,
        "operation": "search",
        "query": query,
        "duration_ms": duration.Milliseconds(),
        "result_count": len(packages),
    }

    if err != nil {
        l.logger.WithFields(fields).WithError(err).Error("Package search failed")
        return nil, err
    }

    l.logger.WithFields(fields).Info("Package search completed")
    return packages, nil
}
```

### Metrics Collection (Prometheus)

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    packageOperations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "syspkg_operations_total",
            Help: "Total number of package operations",
        },
        []string{"manager", "operation", "status"},
    )

    packageOperationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "syspkg_operation_duration_seconds",
            Help: "Duration of package operations",
        },
        []string{"manager", "operation"},
    )
)

type MetricsPackageManager struct {
    manager.PackageManager
    name string
}

func (m *MetricsPackageManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    timer := prometheus.NewTimer(packageOperationDuration.WithLabelValues(m.name, "search"))
    defer timer.ObserveDuration()

    packages, err := m.PackageManager.Search(ctx, query, opts)

    status := "success"
    if err != nil {
        status = "error"
    }

    packageOperations.WithLabelValues(m.name, "search", status).Inc()
    return packages, err
}
```

### Circuit Breaker Pattern

```go
import "github.com/sony/gobreaker"

type CircuitBreakerPackageManager struct {
    manager.PackageManager
    breaker *gobreaker.CircuitBreaker
}

func NewCircuitBreakerPackageManager(pm manager.PackageManager) *CircuitBreakerPackageManager {
    settings := gobreaker.Settings{
        Name:        fmt.Sprintf("syspkg-%s", pm.GetName()),
        MaxRequests: 3,
        Interval:    30 * time.Second,
        Timeout:     60 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            return counts.ConsecutiveFailures >= 3
        },
    }

    return &CircuitBreakerPackageManager{
        PackageManager: pm,
        breaker:       gobreaker.NewCircuitBreaker(settings),
    }
}

func (c *CircuitBreakerPackageManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    result, err := c.breaker.Execute(func() (interface{}, error) {
        return c.PackageManager.Search(ctx, query, opts)
    })

    if err != nil {
        return nil, err
    }

    return result.([]manager.PackageInfo), nil
}
```

## 🧪 Testing Integration Patterns

### Application Testing with go-syspkg

```go
package myapp

import (
    "context"
    "testing"

    "github.com/bluet/syspkg/manager"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockPackageManager for testing
type MockPackageManager struct {
    mock.Mock
}

func (m *MockPackageManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    args := m.Called(ctx, query, opts)
    return args.Get(0).([]manager.PackageInfo), args.Error(1)
}

func (m *MockPackageManager) GetName() string {
    return "mock"
}

func (m *MockPackageManager) GetCategory() string {
    return "mock"
}

func (m *MockPackageManager) IsAvailable() bool {
    return true
}

// Test your application logic
func TestPackageService_SearchPackages(t *testing.T) {
    mockPM := new(MockPackageManager)

    expectedPackages := []manager.PackageInfo{
        {Name: "vim", Version: "8.2", Status: "available"},
    }

    mockPM.On("Search", mock.Anything, []string{"vim"}, mock.Anything).
        Return(expectedPackages, nil)

    // Create service with mock
    service := &PackageService{
        managers: map[string]manager.PackageManager{
            "mock": mockPM,
        },
    }

    // Test your business logic
    result, err := service.SearchPackages(context.Background(), "vim")

    assert.NoError(t, err)
    assert.Len(t, result, 1)
    assert.Equal(t, "vim", result[0].Name)

    mockPM.AssertExpectations(t)
}
```

### Integration Testing with Docker

```go
// +build integration

func TestRealPackageManagerIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Use real package managers in Docker
    registry := manager.GetGlobalRegistry()
    managers := registry.GetAvailable()

    for name, pm := range managers {
        if !pm.IsAvailable() {
            t.Skipf("%s not available", name)
        }

        t.Run(name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()

            // Test real operations
            packages, err := pm.Search(ctx, []string{"curl"}, manager.DefaultOptions())
            assert.NoError(t, err)
            assert.True(t, len(packages) > 0, "Should find curl package")
        })
    }
}
```

## ⚡ Performance Optimization

### Caching Strategy

```go
import (
    "sync"
    "time"
)

type CachedPackageManager struct {
    manager.PackageManager
    cache map[string]cacheEntry
    mutex sync.RWMutex
    ttl   time.Duration
}

type cacheEntry struct {
    packages  []manager.PackageInfo
    timestamp time.Time
    err       error
}

func NewCachedPackageManager(pm manager.PackageManager, ttl time.Duration) *CachedPackageManager {
    return &CachedPackageManager{
        PackageManager: pm,
        cache:         make(map[string]cacheEntry),
        ttl:           ttl,
    }
}

func (c *CachedPackageManager) Search(ctx context.Context, query []string, opts *manager.Options) ([]manager.PackageInfo, error) {
    // Skip cache for dry-run or verbose operations
    if opts != nil && (opts.DryRun || opts.Verbose) {
        return c.PackageManager.Search(ctx, query, opts)
    }

    key := fmt.Sprintf("search:%s", strings.Join(query, ","))

    c.mutex.RLock()
    if entry, exists := c.cache[key]; exists {
        if time.Since(entry.timestamp) < c.ttl {
            c.mutex.RUnlock()
            return entry.packages, entry.err
        }
    }
    c.mutex.RUnlock()

    // Cache miss or expired
    packages, err := c.PackageManager.Search(ctx, query, opts)

    c.mutex.Lock()
    c.cache[key] = cacheEntry{
        packages:  packages,
        timestamp: time.Now(),
        err:       err,
    }
    c.mutex.Unlock()

    return packages, err
}

func (c *CachedPackageManager) ClearCache() {
    c.mutex.Lock()
    c.cache = make(map[string]cacheEntry)
    c.mutex.Unlock()
}
```

### Concurrent Operations

```go
import "golang.org/x/sync/errgroup"

func (s *PackageService) SearchAllManagers(ctx context.Context, query []string) (map[string][]manager.PackageInfo, error) {
    results := make(map[string][]manager.PackageInfo)
    var mutex sync.Mutex

    g, ctx := errgroup.WithContext(ctx)

    for name, pm := range s.managers {
        if !pm.IsAvailable() {
            continue
        }

        name, pm := name, pm // Capture for goroutine
        g.Go(func() error {
            packages, err := pm.Search(ctx, query, manager.DefaultOptions())
            if err != nil {
                s.logger.Printf("Search failed for %s: %v", name, err)
                return nil // Don't fail entire operation for one manager
            }

            mutex.Lock()
            results[name] = packages
            mutex.Unlock()

            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }

    return results, nil
}
```

### Resource Pool Management

```go
type PackageManagerPool struct {
    managers map[string]chan manager.PackageManager
    factory  func(string) manager.PackageManager
    maxSize  int
}

func NewPackageManagerPool(maxSize int, factory func(string) manager.PackageManager) *PackageManagerPool {
    return &PackageManagerPool{
        managers: make(map[string]chan manager.PackageManager),
        factory:  factory,
        maxSize:  maxSize,
    }
}

func (p *PackageManagerPool) Get(name string) manager.PackageManager {
    if pool, exists := p.managers[name]; exists {
        select {
        case pm := <-pool:
            return pm
        default:
            return p.factory(name)
        }
    }
    return p.factory(name)
}

func (p *PackageManagerPool) Put(name string, pm manager.PackageManager) {
    if _, exists := p.managers[name]; !exists {
        p.managers[name] = make(chan manager.PackageManager, p.maxSize)
    }

    select {
    case p.managers[name] <- pm:
    default:
        // Pool full, discard
    }
}
```

## 🚀 Service Integration Examples

### HTTP API Server

```go
func (s *PackageService) searchHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    if query == "" {
        http.Error(w, "Missing query parameter", http.StatusBadRequest)
        return
    }

    managerName := r.URL.Query().Get("manager")

    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()

    var results map[string][]manager.PackageInfo
    var err error

    if managerName != "" {
        pm, exists := s.managers[managerName]
        if !exists || !pm.IsAvailable() {
            http.Error(w, fmt.Sprintf("Manager %s not available", managerName), http.StatusBadRequest)
            return
        }

        packages, err := pm.Search(ctx, []string{query}, manager.DefaultOptions())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        results = map[string][]manager.PackageInfo{managerName: packages}
    } else {
        results, err = s.SearchAllManagers(ctx, []string{query})
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
}
```

### Background Worker Pattern

```go
type PackageMonitor struct {
    registry *manager.Registry
    interval time.Duration
    logger   *log.Logger
    stopCh   chan struct{}
}

func NewPackageMonitor(interval time.Duration, logger *log.Logger) *PackageMonitor {
    return &PackageMonitor{
        registry: manager.GetGlobalRegistry(),
        interval: interval,
        logger:   logger,
        stopCh:   make(chan struct{}),
    }
}

func (p *PackageMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(p.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-p.stopCh:
            return
        case <-ticker.C:
            p.checkPackageHealth(ctx)
        }
    }
}

func (p *PackageMonitor) checkPackageHealth(ctx context.Context) {
    managers := p.registry.GetAvailable()

    for name, pm := range managers {
        if !pm.IsAvailable() {
            continue
        }

        status, err := pm.Status(ctx, manager.DefaultOptions())
        if err != nil {
            p.logger.Printf("Health check failed for %s: %v", name, err)
            continue
        }

        if !status.Healthy {
            p.logger.Printf("Manager %s is unhealthy: %v", name, status.Issues)
        }
    }
}

func (p *PackageMonitor) Stop() {
    close(p.stopCh)
}
```

## 📚 Best Practices Summary

### Do's ✅
- **Use context.WithTimeout()** for all operations
- **Implement health checks** for service monitoring
- **Log structured data** with relevant fields
- **Use circuit breakers** for unreliable package managers
- **Cache expensive operations** with appropriate TTL
- **Test with mocks** in unit tests, real managers in integration tests
- **Handle graceful shutdown** in long-running services

### Don'ts ❌
- **Don't block indefinitely** - always use timeouts
- **Don't ignore errors** - package managers can fail
- **Don't cache dry-run operations** - they should always execute
- **Don't run package operations in production tests** - use Docker or mocks
- **Don't forget resource cleanup** - close contexts and connections

## 🔗 Related Documentation

- **[Basic Integration](../examples/)** - Simple integration patterns
- **[API Reference](../manager/interfaces.go)** - Complete interface documentation
- **[Testing Guide](TESTING.md)** - Testing strategies and fixtures
- **[Plugin Development](PLUGIN_DEVELOPMENT.md)** - Building custom package managers
- **[Architecture Overview](ARCHITECTURE.md)** - Technical design details

---

This guide provides production-ready patterns for integrating go-syspkg into enterprise systems and long-running services. For basic integration, start with the [examples/](../examples/) directory.
