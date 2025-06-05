# Go Library Integration Examples

This directory contains complete, runnable examples showing how to integrate go-syspkg as a Go library into your applications.

## 📚 Learning Progression

### 1. **Start Here: `demos/clean_demo.go`** (70 lines)

Perfect for first-time users. Shows the essential patterns:

```bash
go run examples/demos/clean_demo.go
```

**What you'll learn:**
- ✅ How to import and initialize the library
- ✅ Registry pattern for discovering package managers
- ✅ Basic operations (Search, Status)
- ✅ Clean error handling patterns
- ✅ Context and options usage

**Key code patterns:**
```go
// Essential imports
import "github.com/bluet/syspkg/manager"
import _ "github.com/bluet/syspkg/manager/apt"  // Auto-registers plugin

// Registry discovery
registry := manager.GetGlobalRegistry()
systemPM := registry.GetBestMatch(manager.TypeSystem)

// Basic operation
packages, err := systemPM.Search(ctx, []string{"vim"}, opts)
```

### 2. **Advanced: `complete_demo.go`** (204 lines)

Comprehensive demonstration of the full API:

```bash
go run examples/complete_demo.go
```

**What you'll learn:**
- ✅ All 13 unified operations (Search, List, Install, Remove, etc.)
- ✅ Advanced error handling and graceful degradation
- ✅ Dry-run operations and safety patterns
- ✅ Metadata handling and package information
- ✅ Status checking and health monitoring
- ✅ Architecture benefits and design highlights

**Covers these operations:**
- `Status()` - Package manager health checking
- `Search()` - Package discovery across repositories
- `List()` - Installed and upgradable package enumeration
- `GetInfo()` - Detailed package information
- `Verify()` - Package integrity checking
- `Install()` - Package installation (dry-run mode)
- `Refresh()` - Package list updates
- And more...

## 🎯 Integration Patterns

Both examples demonstrate these essential integration patterns:

### **Plugin System**
```go
// Import registers plugins automatically
import _ "github.com/bluet/syspkg/manager/apt"
import _ "github.com/bluet/syspkg/manager/yum"
```

### **Registry Discovery**
```go
// Get all available managers
registry := manager.GetGlobalRegistry()
managers := registry.GetAvailable()

// Get best system manager
systemPM := registry.GetBestMatch(manager.TypeSystem)
```

### **Context & Options**
```go
// Always use context for timeouts/cancellation
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

// Configure operation behavior
opts := manager.DefaultOptions()
opts.DryRun = true  // Safe testing
opts.Verbose = true // Detailed output
```

### **Error Handling**
```go
// Graceful error handling
packages, err := pm.Search(ctx, query, opts)
if err != nil {
    // Handle specific error types
    return fmt.Errorf("search failed: %w", err)
}
```

## 🔍 When to Use Each Example

| Use Case | Recommended Example |
|----------|-------------------|
| **Learning the API** | Start with `clean_demo.go` |
| **Quick integration** | Copy patterns from `clean_demo.go` |
| **Production reference** | Study `complete_demo.go` |
| **All operations** | Reference `complete_demo.go` |
| **Error handling** | Both demonstrate clean patterns |
| **Testing patterns** | Use dry-run mode from `complete_demo.go` |

## 🚀 Next Steps

After running these examples:

1. **Integrate into your project**: Copy the patterns that fit your use case
2. **Read the interfaces**: Check [manager/interfaces.go](../manager/interfaces.go) for complete API
3. **Build production services**: See [docs/INTEGRATION_GUIDE.md](../docs/INTEGRATION_GUIDE.md) for advanced patterns
4. **Build plugins**: Check [docs/PLUGIN_DEVELOPMENT.md](../docs/PLUGIN_DEVELOPMENT.md) for custom managers
5. **CLI reference**: Reference [cmd/syspkg/](../cmd/syspkg/) for complete implementation

## 📖 Related Documentation

- **[API Reference](../manager/interfaces.go)** - Complete interface documentation
- **[Plugin Development](../docs/PLUGIN_DEVELOPMENT.md)** - Build custom package managers
- **[Architecture Guide](../docs/ARCHITECTURE.md)** - Technical design overview
- **[CLI Implementation](../cmd/syspkg/)** - Production CLI tool as advanced reference

---

**💡 Tip**: These examples work on any Linux system with APT, YUM, Snap, or Flatpak available. They automatically detect what's available and adapt accordingly.
