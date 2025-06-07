// syspkg - Universal package manager CLI
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bluet/syspkg/manager"

	// Import all available package managers
	_ "github.com/bluet/syspkg/manager/apk"
	_ "github.com/bluet/syspkg/manager/apt"
	_ "github.com/bluet/syspkg/manager/flatpak"
	_ "github.com/bluet/syspkg/manager/snap"
	_ "github.com/bluet/syspkg/manager/yum"
)

// Exit codes following POSIX and sysexits.h standards
const (
	// POSIX Standard
	ExitSuccess      = 0 // Success
	ExitGeneralError = 1 // General errors, network issues, unknown failures
	ExitUsageError   = 2 // Invalid arguments (POSIX shell standard)

	// sysexits.h Standard (BSD/Unix)
	ExitNoPermission = 77 // Permission denied - needs sudo (EX_NOPERM)
	ExitUnavailable  = 69 // Service unavailable - manager not found (EX_UNAVAILABLE)

	// Shell Conventions
	ExitSignalInt = 130 // SIGINT (128 + 2) - user interrupted
)

const (
	version = "2.0.0"
	usage   = `syspkg - Universal Package Manager

USAGE:
    syspkg <command> [options] [packages...]

COMMANDS:
    search <query>        Search for packages
    list [filter]         List packages (installed, upgradable, all)
    install <packages>    Install packages (use - to read from stdin)
    remove <packages>     Remove packages (use - to read from stdin)
    info <package>        Show package information
    update               Update package lists
    upgrade [packages]    Upgrade packages (use - to read from stdin, all if none specified)
    clean                Clean package cache
    autoremove           Remove orphaned packages
    verify <packages>    Verify package integrity (use - to read from stdin)
    status               Show package manager status
    managers             List available package managers

OPTIONS:
    # Multi-Manager Control
    --apt, --snap, --flatpak    Use only specific manager(s)
    --all                       Use all available managers (default, concurrent for 3x performance)
    --status                    Show real package status (installed/available/upgradable)

    # Standard Options
    -c, --category CAT   Use manager category (system, language, container, etc.)
    -m, --manager MGR    Use specific manager (apt, yum, npm, pip, etc.)
    -d, --dry-run        Show what would be done without executing
    -v, --verbose        Show detailed output
    -q, --quiet          Minimal output
    -j, --json           Output in JSON format
    -y, --yes            Assume yes to all prompts
    -h, --help           Show this help
    --version            Show version

EXAMPLES:
    # Fast concurrent search (default)
    syspkg search vim                    # Search across all managers (parallel execution)
    syspkg search vim --apt              # Search only APT (single manager)

    # Status-aware search (slower but accurate)
    syspkg search vim --status           # Show real installation status (concurrent)
    syspkg search vim --apt --status     # APT with status information

    # Package operations
    syspkg install vim curl -c system   # Install using system package manager category
    syspkg install vim curl -m apt      # Install using specific APT manager
    syspkg list installed --all          # List installed packages from ALL managers (concurrent)
    syspkg upgrade --all --dry-run       # Show what would be upgraded (all managers, parallel)
    syspkg update --all                  # Update package lists (all managers, concurrent)
    syspkg managers                      # Show available package managers

    # Pipeline support (stdin)
    cat packages.txt | syspkg install - # Install packages from file
    echo "vim curl git" | syspkg install -      # Install multiple packages
    syspkg list installed -q | cut -f1 | syspkg verify -  # Verify installed packages
`
)

type Config struct {
	Manager         string
	ManagerCategory string
	ManagerFilters  map[string]bool // apt, snap, flatpak specific filters
	UseAllManagers  bool            // Default: true (user's original superior approach)
	ShowStatus      bool            // --status flag for status-aware search
	DryRun          bool
	Verbose         bool
	Quiet           bool
	JSON            bool
	AssumeYes       bool
}

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit(true) // Error case - usage to stderr, exit 2
	}

	config := parseArgs()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup signal handling for graceful interruption
	setupSignalHandling(cancel)

	registry := manager.GetGlobalRegistry()

	// Get package managers based on config
	managers, err := selectPackageManagers(registry, config)
	if err != nil {
		printErrorAndExit(err, classifyError(err))
	}

	opts := &manager.Options{
		DryRun:     config.DryRun,
		Verbose:    config.Verbose,
		Quiet:      config.Quiet,
		AssumeYes:  config.AssumeYes,
		ShowStatus: config.ShowStatus,
	}

	// Execute command with multi-manager support
	err = executeMultiCommand(ctx, managers, config, opts)
	if err != nil {
		if !config.Quiet {
			printErrorAndExit(err, classifyError(err))
		} else {
			os.Exit(classifyError(err))
		}
	}
}

func parseArgs() *Config {
	config := &Config{
		ManagerFilters: make(map[string]bool),
		UseAllManagers: true, // User's original superior default
	}
	args := os.Args[1:]

	// Simple argument parsing
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		// Multi-manager selection (user's original superior approach)
		case "--apt":
			config.ManagerFilters["apt"] = true
			config.UseAllManagers = false
		case "--snap":
			config.ManagerFilters["snap"] = true
			config.UseAllManagers = false
		case "--flatpak":
			config.ManagerFilters["flatpak"] = true
			config.UseAllManagers = false
		case "--all":
			config.UseAllManagers = true
			config.ManagerFilters = make(map[string]bool) // Clear filters
		case "--status":
			config.ShowStatus = true
		case "-c", "--category":
			if i+1 < len(args) {
				config.ManagerCategory = args[i+1]
				config.UseAllManagers = false
				i++
			}
		case "-m", "--manager":
			if i+1 < len(args) {
				config.Manager = args[i+1]
				config.UseAllManagers = false
				i++
			}
		case "-d", "--dry-run":
			config.DryRun = true
		case "-v", "--verbose":
			config.Verbose = true
		case "-q", "--quiet":
			config.Quiet = true
		case "-j", "--json":
			config.JSON = true
		case "-y", "--yes":
			config.AssumeYes = true
		case "-h", "--help":
			printUsageAndExit(false) // Help case - usage to stdout, exit 0
		case "--version":
			fmt.Printf("syspkg version %s\n", version)
			os.Exit(0)
		default:
			// Check if this is an unrecognized flag (but allow "-" for stdin)
			if strings.HasPrefix(arg, "-") && arg != "-" {
				fmt.Fprintf(os.Stderr, "syspkg: unrecognized flag '%s'\n", arg)
				printUsageAndExit(true) // Error case - usage to stderr, exit 2
			}
		}
	}

	return config
}

// formatPackageInfo formats a single package for display
func formatPackageInfo(pkg manager.PackageInfo, config *Config, managerName string) {
	if config.Quiet {
		// Tab-separated format: package manager version status
		version := pkg.Version
		if version == "" {
			version = "-"
		}
		status := pkg.Status
		if status == "" {
			status = "-"
		}
		fmt.Printf("%s\t%s\t%s\t%s\n", pkg.Name, managerName, version, status)
		return
	}

	// Format version string for human-readable output
	versionStr := ""
	if pkg.Version != "" {
		versionStr = fmt.Sprintf("[%s]", pkg.Version)
	}
	if pkg.NewVersion != "" && pkg.NewVersion != pkg.Version {
		versionStr += fmt.Sprintf("[%s]", pkg.NewVersion)
	}

	// Add description if available (truncated)
	desc := ""
	if pkg.Description != "" {
		if len(pkg.Description) > 50 {
			desc = fmt.Sprintf(" - %s...", pkg.Description[:47])
		} else {
			desc = fmt.Sprintf(" - %s", pkg.Description)
		}
	}

	// Add manager info for human-readable output
	managerInfo := fmt.Sprintf(" [%s]", managerName)

	// Display with or without status based on config
	if config.ShowStatus {
		fmt.Printf("  %-25s%s %-18s (%s)%s\n", pkg.Name, managerInfo, versionStr, pkg.Status, desc)
	} else {
		fmt.Printf("  %-25s%s %-18s%s\n", pkg.Name, managerInfo, versionStr, desc)
	}
}

func selectPackageManagers(registry *manager.Registry, config *Config) (map[string]manager.PackageManager, error) {
	available := registry.GetAvailable()

	// Single manager specified by name
	if config.Manager != "" {
		if pm, exists := available[config.Manager]; exists {
			return map[string]manager.PackageManager{config.Manager: pm}, nil
		}
		return nil, fmt.Errorf("package manager '%s' not found", config.Manager)
	}

	// Single manager by category
	if config.ManagerCategory != "" {
		pm := registry.GetBestMatch(config.ManagerCategory)
		if pm == nil {
			return nil, fmt.Errorf("no package manager found for category '%s'", config.ManagerCategory)
		}
		return map[string]manager.PackageManager{pm.GetName(): pm}, nil
	}

	// All managers (user's original superior default)
	if config.UseAllManagers {
		return available, nil
	}

	// Filtered managers (user's original superior user control)
	result := make(map[string]manager.PackageManager)
	for name, enabled := range config.ManagerFilters {
		if enabled {
			if pm, exists := available[name]; exists {
				result[name] = pm
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no package managers available")
	}

	return result, nil
}

// getManagerEmoji returns appropriate emoji for each package manager
func getManagerEmoji(managerName string) string {
	switch managerName {
	case "apt":
		return "📦"
	case "snap":
		return "🫰"
	case "flatpak":
		return "📦"
	case "yum":
		return "🔴"
	case "dnf":
		return "🔴"
	default:
		return "📦"
	}
}

// handleSearchUnified implements unified search display for both single and multi-manager scenarios
func handleSearchUnified(ctx context.Context, managers map[string]manager.PackageManager, query []string, config *Config, opts *manager.Options) error {
	if len(managers) == 0 {
		return fmt.Errorf("no package managers available")
	}

	// Print initial search message
	if !config.Quiet && !config.JSON {
		if len(managers) == 1 {
			// Single manager
			for name, pm := range managers {
				emoji := getManagerEmoji(name)
				statusInfo := ""
				if config.ShowStatus {
					statusInfo = " (with status)"
				}
				fmt.Printf("🔍 Searching for '%s' using %s %s%s...\n\n", strings.Join(query, " "), emoji, strings.ToUpper(pm.GetName()), statusInfo)
			}
		} else {
			// Multi-manager
			managerNames := make([]string, 0, len(managers))
			for name := range managers {
				managerNames = append(managerNames, name)
			}
			fmt.Printf("🔍 Searching for '%s' across %d package managers (%s)...\n\n",
				strings.Join(query, " "), len(managers), strings.Join(managerNames, ", "))
		}
	}

	totalPackages := 0

	// Use concurrent search for multiple managers for performance
	if len(managers) > 1 {
		// Get registry for concurrent operations
		registry := manager.GetGlobalRegistry()
		searchResults := registry.SearchAllConcurrent(ctx, query, opts)

		// Process results in a consistent order (sorted by manager name)
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		sort.Strings(managerNames)

		for _, name := range managerNames {
			packages, exists := searchResults[name]
			if !exists {
				continue
			}

			if len(packages) == 0 {
				if !config.Quiet && !config.JSON {
					fmt.Printf("📦 %s: No results found\n", strings.ToUpper(name))
				}
				continue
			}

			// Display header with emoji and count
			if !config.Quiet && !config.JSON {
				emoji := getManagerEmoji(name)
				fmt.Printf("%s %s (%d packages):\n", emoji, strings.ToUpper(name), len(packages))
			}

			// Display packages using unified logic
			for _, pkg := range packages {
				formatPackageInfo(pkg, config, name)
			}

			if !config.Quiet && !config.JSON && len(packages) > 0 {
				fmt.Println()
			}
			totalPackages += len(packages)
		}
	} else {
		// Single manager - use direct call (no performance benefit from concurrency)
		for name, pm := range managers {
			packages, err := pm.Search(ctx, query, opts)
			if err != nil {
				if !config.Quiet && !config.JSON {
					fmt.Printf("❌ Error searching %s: %v\n", name, err)
				}
				continue
			}

			if len(packages) == 0 {
				if !config.Quiet && !config.JSON {
					fmt.Printf("No packages found.\n")
				}
				continue
			}

			// Display header with emoji and count
			if !config.Quiet && !config.JSON {
				emoji := getManagerEmoji(name)
				fmt.Printf("%s %s (%d packages):\n", emoji, strings.ToUpper(name), len(packages))
			}

			// Display packages using unified logic
			for _, pkg := range packages {
				formatPackageInfo(pkg, config, name)
			}

			if !config.Quiet && !config.JSON && len(packages) > 0 {
				fmt.Println()
			}
			totalPackages += len(packages)
		}
	}

	// Summary for multi-manager searches
	if !config.Quiet && !config.JSON && len(managers) > 1 && totalPackages > 0 {
		fmt.Printf("📊 Summary: %d packages found across %d managers\n", totalPackages, len(managers))
	}

	return nil
}

// handleListUnified implements unified list display for multi-manager scenarios
func handleListUnified(ctx context.Context, managers map[string]manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	if len(managers) == 0 {
		return fmt.Errorf("no package managers available")
	}

	// Parse filter from args
	filter := manager.FilterInstalled
	if len(args) > 1 {
		switch args[1] {
		case "installed":
			filter = manager.FilterInstalled
		case "upgradable":
			filter = manager.FilterUpgradable
		case "all":
			filter = manager.FilterAll
		default:
			return fmt.Errorf("invalid filter: %s", args[1])
		}
	}

	// Print initial message
	if !config.Quiet && !config.JSON {
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		fmt.Printf("📋 Listing %s packages across %d package managers (%s)...\n\n",
			filter, len(managers), strings.Join(managerNames, ", "))
	}

	totalPackages := 0

	// Use concurrent operations for installed packages when multiple managers are available
	if len(managers) > 1 && filter == manager.FilterInstalled {
		// Get registry for concurrent operations
		registry := manager.GetGlobalRegistry()
		listResults := registry.ListInstalledAllConcurrent(ctx, opts)

		// Process results in a consistent order (sorted by manager name)
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		sort.Strings(managerNames)

		for _, name := range managerNames {
			packages, exists := listResults[name]
			if !exists {
				continue
			}

			if len(packages) == 0 {
				if !config.Quiet && !config.JSON {
					fmt.Printf("📦 %s: No %s packages found\n", strings.ToUpper(name), filter)
				}
				continue
			}

			// Display header with emoji and count
			if !config.Quiet && !config.JSON {
				emoji := getManagerEmoji(name)
				fmt.Printf("%s %s (%d %s packages):\n", emoji, strings.ToUpper(name), len(packages), filter)
			}

			// Display packages using unified logic
			for _, pkg := range packages {
				formatPackageInfo(pkg, config, name)
			}

			if !config.Quiet && !config.JSON && len(packages) > 0 {
				fmt.Println()
			}
			totalPackages += len(packages)
		}
	} else {
		// Sequential processing for single manager or non-installed filters
		for name, pm := range managers {
			packages, err := pm.List(ctx, filter, opts)
			if err != nil {
				if !config.Quiet && !config.JSON {
					fmt.Printf("❌ Error listing %s packages: %v\n", name, err)
				}
				continue
			}

			if len(packages) == 0 {
				if !config.Quiet && !config.JSON {
					fmt.Printf("📦 %s: No %s packages found\n", strings.ToUpper(name), filter)
				}
				continue
			}

			// Display header with emoji and count
			if !config.Quiet && !config.JSON {
				emoji := getManagerEmoji(name)
				fmt.Printf("%s %s (%d %s packages):\n", emoji, strings.ToUpper(name), len(packages), filter)
			}

			// Display packages using unified logic
			for _, pkg := range packages {
				formatPackageInfo(pkg, config, name)
			}

			if !config.Quiet && !config.JSON && len(packages) > 0 {
				fmt.Println()
			}
			totalPackages += len(packages)
		}
	}

	// Summary for multi-manager listing
	if !config.Quiet && !config.JSON && len(managers) > 1 && totalPackages > 0 {
		fmt.Printf("📊 Summary: %d %s packages found across %d managers\n", totalPackages, filter, len(managers))
	}

	return nil
}

// handleInfoUnified implements unified package info display for multi-manager scenarios
func handleInfoUnified(ctx context.Context, managers map[string]manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	if len(args) < 2 {
		return fmt.Errorf("info requires a package name")
	}
	if len(managers) == 0 {
		return fmt.Errorf("no package managers available")
	}

	packageName := args[1]
	printInfoHeader(packageName, managers, config)

	return processConcurrentInfoResults(ctx, managers, packageName, config, opts)
}

// handleStatusUnified implements unified status display for multi-manager scenarios
func handleStatusUnified(ctx context.Context, managers map[string]manager.PackageManager, config *Config, opts *manager.Options) error {
	if len(managers) == 0 {
		return fmt.Errorf("no package managers available")
	}

	// Print initial message
	if !config.Quiet && !config.JSON {
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		fmt.Printf("📊 Getting status from %d package managers (%s)...\n\n",
			len(managers), strings.Join(managerNames, ", "))
	}

	// Use concurrent status operations for performance
	registry := manager.GetGlobalRegistry()
	statusResults := registry.StatusAllConcurrent(ctx, opts)

	// Process results in a consistent order (sorted by manager name)
	managerNames := make([]string, 0, len(managers))
	for name := range managers {
		managerNames = append(managerNames, name)
	}
	sort.Strings(managerNames)

	if config.JSON {
		// For JSON mode, collect all statuses and output as array
		var allStatuses []interface{}
		for _, name := range managerNames {
			status, exists := statusResults[name]
			if !exists {
				continue
			}
			// Add manager name to the status for JSON output
			statusWithManager := map[string]interface{}{
				"manager":         name,
				"available":       status.Available,
				"healthy":         status.Healthy,
				"version":         status.Version,
				"last_refresh":    status.LastRefresh,
				"cache_size":      status.CacheSize,
				"package_count":   status.PackageCount,
				"installed_count": status.InstalledCount,
				"issues":          status.Issues,
				"metadata":        status.Metadata,
			}
			allStatuses = append(allStatuses, statusWithManager)
		}
		outputResult(allStatuses, config, "multi")
	} else {
		// For text mode, display each manager's status
		for _, name := range managerNames {
			status, exists := statusResults[name]
			if !exists {
				if !config.Quiet {
					fmt.Printf("❌ %s: Status unavailable\n", strings.ToUpper(name))
				}
				continue
			}

			if !config.Quiet {
				emoji := getManagerEmoji(name)
				fmt.Printf("%s %s:\n", emoji, strings.ToUpper(name))
				fmt.Printf("  Available: %v\n", status.Available)
				fmt.Printf("  Healthy: %v\n", status.Healthy)
				if status.Version != "" {
					fmt.Printf("  Version: %s\n", status.Version)
				}
				if len(status.Issues) > 0 {
					fmt.Printf("  Issues: %s\n", strings.Join(status.Issues, ", "))
				}
				fmt.Println()
			} else {
				// Quiet mode: tab-separated format
				healthy := "false"
				if status.Healthy {
					healthy = "true"
				}
				available := "false"
				if status.Available {
					available = "true"
				}
				fmt.Printf("%s\t%s\t%s\t%s\n", name, available, healthy, status.Version)
			}
		}
	}

	return nil
}

// handleUpdateUnified implements unified update (refresh) for multi-manager scenarios
func handleUpdateUnified(ctx context.Context, managers map[string]manager.PackageManager, config *Config, opts *manager.Options) error {
	if len(managers) == 0 {
		return fmt.Errorf("no package managers available")
	}

	// Print initial message
	if !config.Quiet && !config.JSON {
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		fmt.Printf("🔄 Updating package lists across %d package managers (%s)...\n\n",
			len(managers), strings.Join(managerNames, ", "))
	}

	// Use Registry concurrent operations for performance
	registry := manager.GetGlobalRegistry()
	updateResults := registry.RefreshAllConcurrent(ctx, opts)

	// Process results in a consistent order (sorted by manager name)
	managerNames := make([]string, 0, len(managers))
	for name := range managers {
		managerNames = append(managerNames, name)
	}
	sort.Strings(managerNames)

	successCount := 0
	for _, name := range managerNames {
		err, exists := updateResults[name]
		if !exists {
			continue
		}

		if !config.Quiet && !config.JSON {
			emoji := getManagerEmoji(name)
			fmt.Printf("%s Updating %s package lists...\n", emoji, strings.ToUpper(name))
		}

		if err == nil {
			successCount++
			if !config.Quiet && !config.JSON {
				fmt.Printf("✅ %s: Package lists updated successfully\n", strings.ToUpper(name))
			}
		} else {
			if !config.Quiet && !config.JSON {
				fmt.Printf("❌ %s: %v\n", strings.ToUpper(name), err)
			}
		}
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("\n📊 Summary: %d/%d managers updated successfully\n", successCount, len(managers))
	}

	if successCount == 0 {
		return fmt.Errorf("failed to update package lists for any manager")
	}

	return nil
}

// handleUpgradeUnified implements unified upgrade for multi-manager scenarios
func handleUpgradeUnified(ctx context.Context, managers map[string]manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	if len(managers) == 0 {
		return fmt.Errorf("no package managers available")
	}

	// Parse packages from args (same logic as single-manager)
	packages := []string{}
	stdinFound := false
	if len(args) > 1 {
		for _, arg := range args[1:] {
			if arg == "-" {
				stdinFound = true
				stdinPackages, err := readPackagesFromStdin()
				if err != nil {
					return fmt.Errorf("failed to read packages from stdin: %w", err)
				}
				packages = stdinPackages
				break
			}
		}
		if !stdinFound {
			packages = args[1:]
		}
	}

	target := "all packages"
	if len(packages) > 0 {
		target = strings.Join(packages, ", ")
	}

	// Safety prompt for destructive operation
	if !opts.AssumeYes && !config.DryRun && !config.Quiet && !config.JSON {
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		fmt.Printf("⚠️  This will upgrade %s across %d managers (%s)\n", target, len(managers), strings.Join(managerNames, ", "))
		fmt.Print("Do you want to continue? [y/N]: ")

		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Upgrade cancelled.")
			return nil
		}
	}

	// Print initial message
	if !config.Quiet && !config.JSON {
		verb := "Upgrading"
		if config.DryRun {
			verb = "Would upgrade"
		}
		fmt.Printf("🚀 %s %s across %d package managers...\n\n", verb, target, len(managers))
	}

	// Use Registry concurrent operations for performance
	registry := manager.GetGlobalRegistry()
	upgradeResults := registry.UpgradeAllConcurrent(ctx, packages, opts)

	// Process results in a consistent order (sorted by manager name)
	managerNames := make([]string, 0, len(managers))
	for name := range managers {
		managerNames = append(managerNames, name)
	}
	sort.Strings(managerNames)

	successCount := 0
	totalPackages := 0
	for _, name := range managerNames {
		upgradePackages, exists := upgradeResults[name]
		if !exists {
			continue
		}

		if !config.Quiet && !config.JSON {
			emoji := getManagerEmoji(name)
			verb := "Upgrading"
			if config.DryRun {
				verb = "Would upgrade"
			}
			fmt.Printf("%s %s %s using %s...\n", emoji, verb, target, strings.ToUpper(name))
		}

		if len(upgradePackages) >= 0 { // Success (even if 0 packages)
			successCount++
			totalPackages += len(upgradePackages)
			if !config.Quiet && !config.JSON {
				fmt.Printf("✅ %s: Successfully processed %d packages\n", strings.ToUpper(name), len(upgradePackages))
			}

			// Display results for each manager
			if config.JSON {
				outputResult(upgradePackages, config, name)
			} else if config.Quiet {
				for _, pkg := range upgradePackages {
					formatPackageInfo(pkg, config, name)
				}
			}
		}
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("\n📊 Summary: %d packages upgraded across %d/%d managers\n", totalPackages, successCount, len(managers))
	}

	if successCount == 0 {
		return fmt.Errorf("failed to upgrade packages for any manager")
	}

	return nil
}

// handleCleanUnified implements unified clean for multi-manager scenarios
func handleCleanUnified(ctx context.Context, managers map[string]manager.PackageManager, config *Config, opts *manager.Options) error {
	if len(managers) == 0 {
		return fmt.Errorf("no package managers available")
	}

	// Safety prompt for destructive operation
	if !opts.AssumeYes && !config.DryRun && !config.Quiet && !config.JSON {
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		fmt.Printf("⚠️  This will clean package caches across %d managers (%s)\n", len(managers), strings.Join(managerNames, ", "))
		fmt.Print("Do you want to continue? [y/N]: ")

		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Clean cancelled.")
			return nil
		}
	}

	// Print initial message
	if !config.Quiet && !config.JSON {
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		verb := "Cleaning"
		if config.DryRun {
			verb = "Would clean"
		}
		fmt.Printf("🧹 %s package caches across %d package managers (%s)...\n\n",
			verb, len(managers), strings.Join(managerNames, ", "))
	}

	// Use Registry concurrent operations for performance
	registry := manager.GetGlobalRegistry()
	cleanResults := registry.CleanAllConcurrent(ctx, opts)

	// Process results in a consistent order (sorted by manager name)
	managerNames := make([]string, 0, len(managers))
	for name := range managers {
		managerNames = append(managerNames, name)
	}
	sort.Strings(managerNames)

	successCount := 0
	for _, name := range managerNames {
		err, exists := cleanResults[name]
		if !exists {
			continue
		}

		if !config.Quiet && !config.JSON {
			emoji := getManagerEmoji(name)
			verb := "Cleaning"
			if config.DryRun {
				verb = "Would clean"
			}
			fmt.Printf("%s %s %s package cache...\n", emoji, verb, strings.ToUpper(name))
		}

		if err == nil {
			successCount++
			if !config.Quiet && !config.JSON {
				fmt.Printf("✅ %s: Package cache cleaned successfully\n", strings.ToUpper(name))
			}
		} else {
			if !config.Quiet && !config.JSON {
				fmt.Printf("❌ %s: %v\n", strings.ToUpper(name), err)
			}
		}
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("\n📊 Summary: %d/%d managers cleaned successfully\n", successCount, len(managers))
	}

	if successCount == 0 {
		return fmt.Errorf("failed to clean package caches for any manager")
	}

	return nil
}

// handleAutoRemoveUnified implements unified autoremove for multi-manager scenarios
func handleAutoRemoveUnified(ctx context.Context, managers map[string]manager.PackageManager, config *Config, opts *manager.Options) error {
	if len(managers) == 0 {
		return fmt.Errorf("no package managers available")
	}

	// Safety prompt for destructive operation
	if !opts.AssumeYes && !config.DryRun && !config.Quiet && !config.JSON {
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		fmt.Printf("⚠️  This will remove orphaned packages across %d managers (%s)\n", len(managers), strings.Join(managerNames, ", "))
		fmt.Print("Do you want to continue? [y/N]: ")

		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("AutoRemove cancelled.")
			return nil
		}
	}

	// Print initial message
	if !config.Quiet && !config.JSON {
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		verb := "Removing"
		if config.DryRun {
			verb = "Would remove"
		}
		fmt.Printf("🗑️  %s orphaned packages across %d package managers (%s)...\n\n",
			verb, len(managers), strings.Join(managerNames, ", "))
	}

	// Use Registry concurrent operations for performance
	registry := manager.GetGlobalRegistry()
	autoRemoveResults := registry.AutoRemoveAllConcurrent(ctx, opts)

	// Process results in a consistent order (sorted by manager name)
	managerNames := make([]string, 0, len(managers))
	for name := range managers {
		managerNames = append(managerNames, name)
	}
	sort.Strings(managerNames)

	successCount := 0
	totalPackages := 0
	for _, name := range managerNames {
		packages, exists := autoRemoveResults[name]
		if !exists {
			continue
		}

		if !config.Quiet && !config.JSON {
			emoji := getManagerEmoji(name)
			verb := "Removing"
			if config.DryRun {
				verb = "Would remove"
			}
			fmt.Printf("%s %s orphaned packages using %s...\n", emoji, verb, strings.ToUpper(name))
		}

		if len(packages) >= 0 { // Success (even if 0 packages)
			successCount++
			totalPackages += len(packages)
			if !config.Quiet && !config.JSON {
				fmt.Printf("✅ %s: Successfully processed %d packages\n", strings.ToUpper(name), len(packages))
			}

			// Display results for each manager
			if config.JSON {
				outputResult(packages, config, name)
			} else if config.Quiet {
				for _, pkg := range packages {
					formatPackageInfo(pkg, config, name)
				}
			}
		}
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("\n📊 Summary: %d orphaned packages removed across %d/%d managers\n", totalPackages, successCount, len(managers))
	}

	if successCount == 0 {
		return fmt.Errorf("failed to remove orphaned packages for any manager")
	}

	return nil
}

// handleInstallUnified implements unified install for multi-manager scenarios
func handleInstallUnified(ctx context.Context, managers map[string]manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	packages, err := parsePackagesFromArgs(args)
	if err != nil {
		return err
	}
	if len(packages) == 0 {
		return fmt.Errorf("install requires package names")
	}

	if err := confirmDestructiveOperation(config, opts, "install", strings.Join(packages, ", "), managers); err != nil {
		return err
	}

	registry := manager.GetGlobalRegistry()
	results := registry.InstallAllConcurrent(ctx, packages, opts)

	return processPackageOperationResults(results, managers, config, "install", "installed")
}

// handleRemoveUnified implements unified remove for multi-manager scenarios
func handleRemoveUnified(ctx context.Context, managers map[string]manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	packages, err := parsePackagesFromArgs(args)
	if err != nil {
		return err
	}
	if len(packages) == 0 {
		return fmt.Errorf("remove requires package names")
	}

	if err := confirmDestructiveOperation(config, opts, "remove", strings.Join(packages, ", "), managers); err != nil {
		return err
	}

	registry := manager.GetGlobalRegistry()
	results := registry.RemoveAllConcurrent(ctx, packages, opts)

	return processPackageOperationResults(results, managers, config, "remove", "removed")
}

// handleVerifyUnified implements unified verify for multi-manager scenarios
func handleVerifyUnified(ctx context.Context, managers map[string]manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	packages, err := parsePackagesFromArgs(args)
	if err != nil {
		return err
	}
	if len(packages) == 0 {
		return fmt.Errorf("verify requires package names")
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("🔍 Verifying %s across %d package managers...\n\n", strings.Join(packages, ", "), len(managers))
	}

	registry := manager.GetGlobalRegistry()
	results := registry.VerifyAllConcurrent(ctx, packages, opts)

	return processPackageOperationResults(results, managers, config, "verify", "verified")
}

func executeMultiCommand(ctx context.Context, managers map[string]manager.PackageManager, config *Config, opts *manager.Options) error {
	args := getCommandArgs()
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]

	// For multi-manager operations (search is always multi, others when --all is specified)
	switch command {
	case "search":
		if len(args) < 2 {
			return fmt.Errorf("search requires a query")
		}
		return handleSearchUnified(ctx, managers, args[1:], config, opts)

	case "list":
		if config.UseAllManagers {
			return handleListUnified(ctx, managers, args, config, opts)
		}
		// Fall through to single-manager handling

	case "info":
		if config.UseAllManagers {
			return handleInfoUnified(ctx, managers, args, config, opts)
		}
		// Fall through to single-manager handling

	case "status":
		if config.UseAllManagers {
			return handleStatusUnified(ctx, managers, config, opts)
		}
		// Fall through to single-manager handling

	case "update":
		if config.UseAllManagers {
			return handleUpdateUnified(ctx, managers, config, opts)
		}
		// Fall through to single-manager handling

	case "upgrade":
		if config.UseAllManagers {
			return handleUpgradeUnified(ctx, managers, args, config, opts)
		}
		// Fall through to single-manager handling

	case "clean":
		if config.UseAllManagers {
			return handleCleanUnified(ctx, managers, config, opts)
		}
		// Fall through to single-manager handling

	case "autoremove":
		if config.UseAllManagers {
			return handleAutoRemoveUnified(ctx, managers, config, opts)
		}
		// Fall through to single-manager handling

	case "install":
		if config.UseAllManagers {
			return handleInstallUnified(ctx, managers, args, config, opts)
		}
		// Fall through to single-manager handling

	case "remove":
		if config.UseAllManagers {
			return handleRemoveUnified(ctx, managers, args, config, opts)
		}
		// Fall through to single-manager handling

	case "verify":
		if config.UseAllManagers {
			return handleVerifyUnified(ctx, managers, args, config, opts)
		}
		// Fall through to single-manager handling

	case "managers":
		return handleManagers(config)
	}

	// Single-manager operations (including list, info, status without --all)
	{
		// For single-manager operations, select the appropriate manager
		var pm manager.PackageManager

		if len(managers) == 1 {
			// Only one manager available, use it
			for _, mgr := range managers {
				pm = mgr
				break
			}
		} else {
			// Multiple managers available
			if config.Manager != "" {
				// User specified a manager
				var exists bool
				pm, exists = managers[config.Manager]
				if !exists {
					return fmt.Errorf("specified manager '%s' not available", config.Manager)
				}
			} else {
				// No manager specified - use the best system package manager
				registry := manager.GetGlobalRegistry()
				pm = registry.GetBestMatch(manager.CategorySystem)
				if pm == nil {
					return fmt.Errorf("no system package manager available")
				}
				if !config.Quiet && !config.JSON {
					fmt.Printf("Using %s package manager\n", pm.GetName())
				}
			}
		}
		return executeSingleCommand(ctx, pm, config, opts)
	}
}

func executeSingleCommand(ctx context.Context, pm manager.PackageManager, config *Config, opts *manager.Options) error {
	args := getCommandArgs()
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]
	return executeCommand(ctx, pm, command, args, config, opts)
}

func executeCommand(ctx context.Context, pm manager.PackageManager, command string, args []string, config *Config, opts *manager.Options) error {
	switch command {
	case "search":
		return executeSearchCommand(ctx, pm, args, config, opts)
	case "list":
		return executeListCommand(ctx, pm, args, config, opts)
	case "install":
		return executeInstallCommand(ctx, pm, args, config, opts)
	case "remove":
		return executeRemoveCommand(ctx, pm, args, config, opts)
	case "info":
		return executeInfoCommand(ctx, pm, args, config, opts)
	case "update":
		return handleUpdate(ctx, pm, config, opts)
	case "upgrade":
		return executeUpgradeCommand(ctx, pm, args, config, opts)
	case "clean":
		return handleClean(ctx, pm, config, opts)
	case "autoremove":
		return handleAutoRemove(ctx, pm, config, opts)
	case "verify":
		return executeVerifyCommand(ctx, pm, args, config, opts)
	case "status":
		return handleStatus(ctx, pm, config, opts)
	case "managers":
		return handleManagers(config)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func executeSearchCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	if len(args) < 2 {
		return fmt.Errorf("search requires a query")
	}
	// Convert single manager to map for unified handling
	managers := map[string]manager.PackageManager{pm.GetName(): pm}
	return handleSearchUnified(ctx, managers, args[1:], config, opts)
}

func executeListCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	filter := manager.FilterInstalled
	if len(args) > 1 {
		switch args[1] {
		case "installed":
			filter = manager.FilterInstalled
		case "upgradable":
			filter = manager.FilterUpgradable
		case "all":
			filter = manager.FilterAll
		default:
			return fmt.Errorf("invalid filter: %s", args[1])
		}
	}
	return handleList(ctx, pm, filter, config, opts)
}

func executeInstallCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	packages := []string{}

	// Check if user specified stdin with "-" anywhere in args
	stdinFound := false
	for _, arg := range args[1:] {
		if arg == "-" {
			stdinFound = true
			// Read from stdin
			stdinPackages, err := readPackagesFromStdin()
			if err != nil {
				return fmt.Errorf("failed to read packages from stdin: %w", err)
			}
			packages = stdinPackages
			break
		}
	}

	if !stdinFound && len(args) >= 2 {
		// Get packages from command line arguments (excluding command itself)
		packages = args[1:]
	}

	// Require at least one package
	if len(packages) == 0 {
		return fmt.Errorf("install requires package names")
	}

	return handleInstall(ctx, pm, packages, config, opts)
}

func executeRemoveCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	packages := []string{}

	// Check if user specified stdin with "-" anywhere in args
	stdinFound := false
	for _, arg := range args[1:] {
		if arg == "-" {
			stdinFound = true
			// Read from stdin
			stdinPackages, err := readPackagesFromStdin()
			if err != nil {
				return fmt.Errorf("failed to read packages from stdin: %w", err)
			}
			packages = stdinPackages
			break
		}
	}

	if !stdinFound && len(args) >= 2 {
		// Get packages from command line arguments (excluding command itself)
		packages = args[1:]
	}

	// Require at least one package
	if len(packages) == 0 {
		return fmt.Errorf("remove requires package names")
	}

	return handleRemove(ctx, pm, packages, config, opts)
}

func executeInfoCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	if len(args) < 2 {
		return fmt.Errorf("info requires a package name")
	}
	return handleInfo(ctx, pm, args[1], config, opts)
}

func executeUpgradeCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	packages := []string{}

	// Check if user specified stdin with "-" anywhere in args
	stdinFound := false
	for _, arg := range args[1:] {
		if arg == "-" {
			stdinFound = true
			// Read from stdin
			stdinPackages, err := readPackagesFromStdin()
			if err != nil {
				return fmt.Errorf("failed to read packages from stdin: %w", err)
			}
			packages = stdinPackages
			break
		}
	}

	if !stdinFound && len(args) >= 2 {
		// Get packages from command line arguments (excluding command itself)
		packages = args[1:]
	}

	// Note: if no packages specified, upgrade will upgrade all packages
	return handleUpgrade(ctx, pm, packages, config, opts)
}

func executeVerifyCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	packages := []string{}

	// Check if user specified stdin with "-" anywhere in args
	stdinFound := false
	for _, arg := range args[1:] {
		if arg == "-" {
			stdinFound = true
			// Read from stdin
			stdinPackages, err := readPackagesFromStdin()
			if err != nil {
				return fmt.Errorf("failed to read packages from stdin: %w", err)
			}
			packages = stdinPackages
			break
		}
	}

	if !stdinFound && len(args) >= 2 {
		// Get packages from command line arguments (excluding command itself)
		packages = args[1:]
	}

	// Require at least one package
	if len(packages) == 0 {
		return fmt.Errorf("verify requires package names")
	}

	return handleVerify(ctx, pm, packages, config, opts)
}

func getCommandArgs() []string {
	args := os.Args[1:]
	var result []string

	// Skip flags and extract command + args
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip flags that take values
		if (arg == "-c" || arg == "--category" || arg == "-m" || arg == "--manager") && i+1 < len(args) {
			i++ // Skip the value too
			continue
		}

		// Skip single flags, but allow "-" (stdin indicator)
		if strings.HasPrefix(arg, "-") && arg != "-" {
			continue
		}

		// This is a command or argument
		result = append(result, arg)
	}

	return result
}

func outputResult(data interface{}, config *Config, managerName string) {
	if config.JSON {
		_ = json.NewEncoder(os.Stdout).Encode(data)
	} else {
		switch v := data.(type) {
		case []manager.PackageInfo:
			for _, pkg := range v {
				formatPackageInfo(pkg, config, managerName)
			}
		case manager.PackageInfo:
			if config.Quiet {
				formatPackageInfo(v, config, managerName)
			} else {
				fmt.Printf("Name: %s\n", v.Name)
				fmt.Printf("Version: %s\n", v.Version)
				fmt.Printf("Status: %s\n", v.Status)
				if v.Description != "" {
					fmt.Printf("Description: %s\n", v.Description)
				}
			}
		case string:
			fmt.Println(v)
		default:
			fmt.Printf("%+v\n", v)
		}
	}
}

func handleList(ctx context.Context, pm manager.PackageManager, filter manager.ListFilter, config *Config, opts *manager.Options) error {
	if !config.Quiet && !config.JSON {
		fmt.Printf("Listing %s packages using %s...\n", filter, pm.GetName())
	}

	packages, err := pm.List(ctx, filter, opts)
	if err != nil {
		return err
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("Found %d packages:\n", len(packages))
	}

	outputResult(packages, config, pm.GetName())
	return nil
}

func handleInstall(ctx context.Context, pm manager.PackageManager, packages []string, config *Config, opts *manager.Options) error {
	if !config.Quiet && !config.JSON {
		verb := "Installing"
		if config.DryRun {
			verb = "Would install"
		}
		fmt.Printf("%s packages: %s\n", verb, strings.Join(packages, ", "))
	}

	results, err := pm.Install(ctx, packages, opts)
	if err != nil {
		return err
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("Successfully processed %d packages:\n", len(results))
	}

	outputResult(results, config, pm.GetName())
	return nil
}

func handleRemove(ctx context.Context, pm manager.PackageManager, packages []string, config *Config, opts *manager.Options) error {
	if !config.Quiet && !config.JSON {
		verb := "Removing"
		if config.DryRun {
			verb = "Would remove"
		}
		fmt.Printf("%s packages: %s\n", verb, strings.Join(packages, ", "))
	}

	results, err := pm.Remove(ctx, packages, opts)
	if err != nil {
		return err
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("Successfully processed %d packages:\n", len(results))
	}

	outputResult(results, config, pm.GetName())
	return nil
}

func handleInfo(ctx context.Context, pm manager.PackageManager, packageName string, config *Config, opts *manager.Options) error {
	pkg, err := pm.GetInfo(ctx, packageName, opts)
	if err != nil {
		return err
	}

	outputResult(pkg, config, pm.GetName())
	return nil
}

func handleUpdate(ctx context.Context, pm manager.PackageManager, config *Config, opts *manager.Options) error {
	if !config.Quiet && !config.JSON {
		fmt.Printf("Updating package lists using %s...\n", pm.GetName())
	}

	err := pm.Refresh(ctx, opts)
	if err != nil {
		return err
	}

	if !config.Quiet && !config.JSON {
		fmt.Println("Package lists updated successfully")
	}

	return nil
}

func handleUpgrade(ctx context.Context, pm manager.PackageManager, packages []string, config *Config, opts *manager.Options) error {
	target := "all packages"
	if len(packages) > 0 {
		target = strings.Join(packages, ", ")
	}

	if !config.Quiet && !config.JSON {
		verb := "Upgrading"
		if config.DryRun {
			verb = "Would upgrade"
		}
		fmt.Printf("%s %s...\n", verb, target)
	}

	results, err := pm.Upgrade(ctx, packages, opts)
	if err != nil {
		return err
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("Successfully processed %d packages:\n", len(results))
	}

	outputResult(results, config, pm.GetName())
	return nil
}

func handleClean(ctx context.Context, pm manager.PackageManager, config *Config, opts *manager.Options) error {
	if !config.Quiet && !config.JSON {
		fmt.Printf("Cleaning package cache using %s...\n", pm.GetName())
	}

	err := pm.Clean(ctx, opts)
	if err != nil {
		return err
	}

	if !config.Quiet && !config.JSON {
		fmt.Println("Package cache cleaned successfully")
	}

	return nil
}

func handleAutoRemove(ctx context.Context, pm manager.PackageManager, config *Config, opts *manager.Options) error {
	if !config.Quiet && !config.JSON {
		verb := "Removing"
		if config.DryRun {
			verb = "Would remove"
		}
		fmt.Printf("%s orphaned packages...\n", verb)
	}

	results, err := pm.AutoRemove(ctx, opts)
	if err != nil {
		return err
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("Successfully processed %d packages:\n", len(results))
	}

	outputResult(results, config, pm.GetName())
	return nil
}

func handleVerify(ctx context.Context, pm manager.PackageManager, packages []string, config *Config, opts *manager.Options) error {
	if !config.Quiet && !config.JSON {
		fmt.Printf("Verifying packages: %s\n", strings.Join(packages, ", "))
	}

	results, err := pm.Verify(ctx, packages, opts)
	if err != nil {
		return err
	}

	outputResult(results, config, pm.GetName())
	return nil
}

func handleStatus(ctx context.Context, pm manager.PackageManager, config *Config, opts *manager.Options) error {
	status, err := pm.Status(ctx, opts)
	if err != nil {
		return err
	}

	if config.JSON {
		outputResult(status, config, pm.GetName())
	} else {
		fmt.Printf("Package Manager: %s\n", pm.GetName())
		fmt.Printf("Type: %s\n", pm.GetCategory())
		fmt.Printf("Available: %v\n", status.Available)
		fmt.Printf("Healthy: %v\n", status.Healthy)
		if status.Version != "" {
			fmt.Printf("Version: %s\n", status.Version)
		}
		if len(status.Issues) > 0 {
			fmt.Printf("Issues: %s\n", strings.Join(status.Issues, ", "))
		}
	}

	return nil
}

func handleManagers(config *Config) error {
	registry := manager.GetGlobalRegistry()
	pluginNames := registry.List() // Get ALL registered plugins

	if config.JSON {
		type ManagerInfo struct {
			Name      string `json:"name"`
			Type      string `json:"type"`
			Available bool   `json:"available"`
		}

		var infos []ManagerInfo
		for _, name := range pluginNames {
			plugin, exists := registry.Get(name)
			if !exists {
				continue // Should not happen, but be safe
			}
			pm := plugin.CreateManager()
			infos = append(infos, ManagerInfo{
				Name:      name,
				Type:      pm.GetCategory(),
				Available: pm.IsAvailable(),
			})
		}

		outputResult(infos, config, "managers")
	} else {
		fmt.Println("Available Package Managers:")
		for _, name := range pluginNames {
			plugin, exists := registry.Get(name)
			if !exists {
				continue // Should not happen, but be safe
			}
			pm := plugin.CreateManager()
			status := "❌"
			if pm.IsAvailable() {
				status = "✅"
			}
			fmt.Printf("  %s %-10s (%s)\n", status, name, pm.GetCategory())
		}
	}

	return nil
}

// printUsageAndExit prints usage to appropriate output and exits with correct code
func printUsageAndExit(isError bool) {
	out := os.Stdout
	code := ExitSuccess
	if isError {
		out = os.Stderr       // Error cases go to stderr
		code = ExitUsageError // POSIX: 2 = usage error
	}
	fmt.Fprint(out, usage)
	os.Exit(code)
}

// printErrorAndExit prints error message and exits with appropriate code
func printErrorAndExit(err error, code int) {
	fmt.Fprintf(os.Stderr, "syspkg: %v\n", err)
	os.Exit(code)
}

// classifyError determines appropriate exit code based on error type
func classifyError(err error) int {
	if err == nil {
		return ExitSuccess
	}

	// Check for StandardStatus types first (highest priority)
	if code := classifyStandardStatus(err); code != -1 {
		return code
	}

	// Check for specific manager errors
	if code := classifyManagerErrors(err); code != -1 {
		return code
	}

	// Fallback to string-based classification for non-typed errors
	return classifyStringBasedErrors(err)
}

// classifyStandardStatus checks for StandardStatus error types
func classifyStandardStatus(err error) int {
	var standardStatus *manager.StandardStatus
	if errors.As(err, &standardStatus) {
		switch standardStatus.Status {
		case manager.StatusSuccess:
			return ExitSuccess
		case manager.StatusUsageError:
			return ExitUsageError
		case manager.StatusPermissionError:
			return ExitNoPermission
		case manager.StatusUnavailableError:
			return ExitUnavailable
		case manager.StatusGeneralError:
			return ExitGeneralError
		}
	}
	return -1 // Not a StandardStatus error
}

// classifyManagerErrors checks for specific manager error types
func classifyManagerErrors(err error) int {
	if errors.Is(err, manager.ErrOperationNotSupported) {
		return ExitUnavailable
	}
	if errors.Is(err, manager.ErrInvalidPackageName) {
		return ExitUsageError
	}
	return -1 // Not a specific manager error
}

// classifyStringBasedErrors checks error strings for classification patterns
func classifyStringBasedErrors(err error) int {
	errStr := err.Error()

	// Check for permission-related errors
	if isPermissionError(errStr) {
		return ExitNoPermission
	}

	// Check for service unavailable (package manager not found)
	if isUnavailableError(errStr) {
		return ExitUnavailable
	}

	// Check for usage errors
	if isUsageError(errStr) {
		return ExitUsageError
	}

	// Default to general error
	return ExitGeneralError
}

// isPermissionError checks if error string indicates permission issues
func isPermissionError(errStr string) bool {
	permissionKeywords := []string{
		"permission denied", "are you root", "try with sudo",
		"access denied", "operation not permitted",
	}
	for _, keyword := range permissionKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	return false
}

// isUnavailableError checks if error string indicates service unavailable
func isUnavailableError(errStr string) bool {
	unavailableKeywords := []string{"not found", "not available", "unavailable"}
	for _, keyword := range unavailableKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	return false
}

// isUsageError checks if error string indicates usage errors
func isUsageError(errStr string) bool {
	usageKeywords := []string{"requires", "invalid", "usage"}
	for _, keyword := range usageKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	return false
}

// setupSignalHandling configures graceful handling of interrupt signals
func setupSignalHandling(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\nInterrupted\n")
		cancel()               // Cancel context gracefully
		os.Exit(ExitSignalInt) // Standard SIGINT exit code (130)
	}()
}

// readPackagesFromStdin reads package names from stdin
// Supports both space-separated (same line) and newline-separated formats
// This enables pipeline support: echo "vim curl git" | syspkg install -
func readPackagesFromStdin() ([]string, error) {
	var packages []string
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			// Split by whitespace to support multiple packages per line
			fields := strings.Fields(line)
			packages = append(packages, fields...)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading from stdin: %w", err)
	}

	return packages, nil
}

// parsePackagesFromArgs extracts package names from command line arguments or stdin
func parsePackagesFromArgs(args []string) ([]string, error) {
	packages := []string{}
	stdinFound := false
	if len(args) > 1 {
		for _, arg := range args[1:] {
			if arg == "-" {
				stdinFound = true
				stdinPackages, err := readPackagesFromStdin()
				if err != nil {
					return nil, fmt.Errorf("failed to read packages from stdin: %w", err)
				}
				packages = stdinPackages
				break
			}
		}
		if !stdinFound {
			packages = args[1:]
		}
	}
	return packages, nil
}

// confirmDestructiveOperation asks for user confirmation for potentially destructive operations
func confirmDestructiveOperation(config *Config, opts *manager.Options, operation, target string, managers map[string]manager.PackageManager) error {
	if !opts.AssumeYes && !config.DryRun && !config.Quiet && !config.JSON {
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		fmt.Printf("⚠️  This will %s %s across %d managers (%s)\n", operation, target, len(managers), strings.Join(managerNames, ", "))
		fmt.Print("Do you want to continue? [y/N]: ")

		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Printf("%s cancelled.\n", strings.ToUpper(operation[:1])+operation[1:])
			return fmt.Errorf("operation cancelled by user")
		}
	}

	if !config.Quiet && !config.JSON {
		verb := getOperationVerb(operation, config.DryRun)
		fmt.Printf("📦 %s %s across %d package managers...\n\n", verb, target, len(managers))
	}
	return nil
}

// processPackageOperationResults processes and displays results from package operations
func processPackageOperationResults(results map[string][]manager.PackageInfo, managers map[string]manager.PackageManager, config *Config, operation, pastTense string) error {
	if len(managers) == 0 {
		return fmt.Errorf("no package managers available")
	}

	managerNames := make([]string, 0, len(managers))
	for name := range managers {
		managerNames = append(managerNames, name)
	}
	sort.Strings(managerNames)

	successCount := 0
	totalPackages := 0
	for _, name := range managerNames {
		packages, exists := results[name]
		if !exists {
			continue
		}

		displayOperationProgress(name, operation, config)

		if len(packages) >= 0 { // Success (even if 0 packages)
			successCount++
			totalPackages += len(packages)
			displayOperationSuccess(name, len(packages), pastTense, config)
			displayPackageResults(packages, config, name)
		}
	}

	if !config.Quiet && !config.JSON {
		fmt.Printf("\n📊 Summary: %d packages %s across %d/%d managers\n", totalPackages, pastTense, successCount, len(managers))
	}

	if successCount == 0 {
		return fmt.Errorf("failed to %s packages on any manager", operation)
	}
	return nil
}

// displayOperationProgress shows progress for individual manager operations
func displayOperationProgress(managerName, operation string, config *Config) {
	if !config.Quiet && !config.JSON {
		emoji := getManagerEmoji(managerName)
		verb := getOperationVerb(operation, config.DryRun)
		fmt.Printf("%s %s packages using %s...\n", emoji, verb, strings.ToUpper(managerName))
	}
}

// displayOperationSuccess shows success message for completed operations
func displayOperationSuccess(managerName string, count int, pastTense string, config *Config) {
	if !config.Quiet && !config.JSON {
		fmt.Printf("✅ %s: Successfully %s %d packages\n", strings.ToUpper(managerName), pastTense, count)
	}
}

// displayPackageResults shows package results based on output format
func displayPackageResults(packages []manager.PackageInfo, config *Config, managerName string) {
	if config.JSON {
		outputResult(packages, config, managerName)
	} else if config.Quiet {
		for _, pkg := range packages {
			formatPackageInfo(pkg, config, managerName)
		}
	}
}

// getOperationVerb returns the appropriate verb form for the operation
func getOperationVerb(operation string, dryRun bool) string {
	verbs := map[string]string{
		"install": "Installing",
		"remove":  "Removing",
		"verify":  "Verifying",
	}
	verb := verbs[operation]
	if dryRun && operation != "verify" {
		verb = "Would " + strings.ToLower(verb)
	}
	return verb
}

// printInfoHeader prints the initial message for info command
func printInfoHeader(packageName string, managers map[string]manager.PackageManager, config *Config) {
	if !config.Quiet && !config.JSON {
		managerNames := make([]string, 0, len(managers))
		for name := range managers {
			managerNames = append(managerNames, name)
		}
		fmt.Printf("ℹ️  Getting info for '%s' across %d package managers (%s)...\n\n",
			packageName, len(managers), strings.Join(managerNames, ", "))
	}
}

// infoResult represents the result of an info operation from a single manager
type infoResult struct {
	managerName string
	success     bool
	pkg         manager.PackageInfo
	err         error
}

// processConcurrentInfoResults handles concurrent info operations and result display
func processConcurrentInfoResults(ctx context.Context, managers map[string]manager.PackageManager, packageName string, config *Config, opts *manager.Options) error {

	results := make(chan infoResult, len(managers))
	var wg sync.WaitGroup

	// Start concurrent info operations
	for name, pm := range managers {
		wg.Add(1)
		go func(name string, pm manager.PackageManager) {
			defer wg.Done()
			pkg, err := pm.GetInfo(ctx, packageName, opts)
			results <- infoResult{name, err == nil, pkg, err}
		}(name, pm)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	return displayInfoResults(results, managers, packageName, config)
}

// displayInfoResults processes and displays info results from all managers
func displayInfoResults(results chan infoResult, managers map[string]manager.PackageManager, packageName string, config *Config) error {
	// Process results in a consistent order (sorted by manager name)
	managerNames := make([]string, 0, len(managers))
	for name := range managers {
		managerNames = append(managerNames, name)
	}
	sort.Strings(managerNames)

	// Collect results by manager name
	resultMap := make(map[string]infoResult)
	for result := range results {
		resultMap[result.managerName] = result
	}

	foundCount := 0
	for _, name := range managerNames {
		result, exists := resultMap[name]
		if !exists {
			continue
		}

		if !result.success {
			if !config.Quiet && !config.JSON {
				fmt.Printf("❌ %s: %v\n", strings.ToUpper(result.managerName), result.err)
			}
			continue
		}

		foundCount++
		displaySingleInfoResult(result, config)
	}

	if foundCount == 0 && !config.Quiet && !config.JSON {
		fmt.Printf("Package '%s' not found in any package manager\n", packageName)
	}

	return nil
}

// displaySingleInfoResult displays the result from a single manager
func displaySingleInfoResult(result infoResult, config *Config) {
	// Display header for each manager
	if !config.Quiet && !config.JSON {
		emoji := getManagerEmoji(result.managerName)
		fmt.Printf("%s %s:\n", emoji, strings.ToUpper(result.managerName))
	}

	if config.JSON {
		outputResult(result.pkg, config, result.managerName)
	} else {
		if config.Quiet {
			formatPackageInfo(result.pkg, config, result.managerName)
		} else {
			fmt.Printf("  Name: %s\n", result.pkg.Name)
			fmt.Printf("  Version: %s\n", result.pkg.Version)
			fmt.Printf("  Status: %s\n", result.pkg.Status)
			if result.pkg.Description != "" {
				fmt.Printf("  Description: %s\n", result.pkg.Description)
			}
			if result.pkg.Category != "" {
				fmt.Printf("  Category: %s\n", result.pkg.Category)
			}
			fmt.Println()
		}
	}
}
