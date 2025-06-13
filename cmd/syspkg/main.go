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
    syspkg search vim -m apt             # Search only APT (single manager)

    # Status-aware search (slower but accurate)
    syspkg search vim --status           # Show real installation status (concurrent)
    syspkg search vim -m apt --status    # APT with status information

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
	UseAllManagers  bool // Default: true (user's original superior approach)
	ShowStatus      bool // --status flag for status-aware search
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

	// Execute command using new dispatcher pattern
	validator := NewCommandValidator()
	dispatcher := NewCommandDispatcher(validator)
	err = dispatcher.Dispatch(ctx, managers, config, opts)
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
		UseAllManagers: true, // User's original superior default
	}
	args := os.Args[1:]

	// Simple argument parsing
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		// Multi-manager selection (user's original superior approach)
		case "--all":
			config.UseAllManagers = true
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

	// Default behavior: use best system manager
	pm := registry.GetBestMatch(manager.CategorySystem)
	if pm == nil {
		return nil, fmt.Errorf("no system package manager available")
	}
	return map[string]manager.PackageManager{pm.GetName(): pm}, nil
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
	validator := NewCommandValidator()
	if err := validator.ValidateManagersAvailable(managers); err != nil {
		return err
	}

	formatter := NewOutputFormatter(config)
	processor := NewPackageListProcessor(formatter, config)

	// Print initial search message
	formatter.FormatSearchHeader(query, managers, config.ShowStatus)

	totalPackages := 0

	// Use concurrent search for multiple managers for performance
	if len(managers) > 1 {
		registry := manager.GetGlobalRegistry()
		searchResults := registry.SearchAllConcurrent(ctx, query, opts)
		totalPackages = processor.ProcessSearchResults(searchResults, managers)
	} else {
		// Single manager - use direct call (no performance benefit from concurrency)
		for name, pm := range managers {
			packages, err := pm.Search(ctx, query, opts)
			if err != nil {
				formatter.FormatErrorMessage(name, err)
				continue
			}

			if len(packages) == 0 {
				if !config.Quiet && !config.JSON {
					fmt.Printf("No packages found.\n")
				}
				continue
			}

			// Display header with emoji and count
			emoji := getManagerEmoji(name)
			formatter.FormatManagerHeader(name, emoji, len(packages), "packages")

			// Display packages using unified logic
			for _, pkg := range packages {
				formatter.FormatPackageInfo(pkg, name)
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
	validator := NewCommandValidator()
	if err := validator.ValidateManagersAvailable(managers); err != nil {
		return err
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

	formatter := NewOutputFormatter(config)
	processor := NewPackageListProcessor(formatter, config)

	// Print initial message
	formatter.FormatListHeader(filter, managers)

	totalPackages := 0

	// Use concurrent operations for installed packages when multiple managers are available
	if len(managers) > 1 && filter == manager.FilterInstalled {
		registry := manager.GetGlobalRegistry()
		listResults := registry.ListInstalledAllConcurrent(ctx, opts)
		totalPackages = processor.ProcessListResults(listResults, managers, filter)
	} else {
		// Sequential processing for single manager or non-installed filters
		listResults := make(map[string][]manager.PackageInfo)
		for name, pm := range managers {
			packages, err := pm.List(ctx, filter, opts)
			if err != nil {
				formatter.FormatErrorMessage(name, err)
				continue
			}
			listResults[name] = packages
		}
		totalPackages = processor.ProcessListResults(listResults, managers, filter)
	}

	// Summary for multi-manager listing
	if !config.Quiet && !config.JSON && len(managers) > 1 && totalPackages > 0 {
		fmt.Printf("📊 Summary: %d %s packages found across %d managers\n", totalPackages, filter, len(managers))
	}

	return nil
}

// handleInfoUnified implements unified package info display for multi-manager scenarios
func handleInfoUnified(ctx context.Context, managers map[string]manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	validator := NewCommandValidator()
	if err := validator.ValidateInfoCommand(args); err != nil {
		return err
	}
	if err := validator.ValidateManagersAvailable(managers); err != nil {
		return err
	}

	packageName := args[1]
	formatter := NewOutputFormatter(config)
	formatter.FormatInfoHeader(packageName, managers)

	handler := NewInfoResultHandler(formatter, config)
	return handler.ProcessConcurrentInfoResults(ctx, managers, packageName, opts)
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
	validator := NewCommandValidator()
	if err := validator.ValidateManagersAvailable(managers); err != nil {
		return err
	}

	// Parse packages from args
	packages, err := parsePackagesFromArgs(args)
	if err != nil {
		return err
	}

	target := "all packages"
	if len(packages) > 0 {
		target = strings.Join(packages, ", ")
	}

	// Safety prompt for destructive operation
	confirmHandler := NewConfirmationHandler(config, opts)
	if err := confirmHandler.ConfirmDestructiveOperation("upgrade", target, managers); err != nil {
		return err
	}

	// Print initial message
	formatter := NewOutputFormatter(config)
	formatter.FormatOperationHeader("upgrade", target, managers, config.DryRun)

	// Use Registry concurrent operations for performance with proper error handling
	registry := manager.GetGlobalRegistry()
	upgradeResults := registry.UpgradeAllConcurrentWithErrors(ctx, packages, opts)

	// Process results using new result processor with error handling
	processor := NewResultProcessor(formatter, config)
	return processor.ProcessPackageOperationResultsWithErrors(upgradeResults, managers, "upgrade", "upgraded")
}

// handleCleanUnified implements unified clean for multi-manager scenarios
func handleCleanUnified(ctx context.Context, managers map[string]manager.PackageManager, config *Config, opts *manager.Options) error {
	validator := NewCommandValidator()
	if err := validator.ValidateManagersAvailable(managers); err != nil {
		return err
	}

	// Safety prompt for destructive operation
	confirmHandler := NewConfirmationHandler(config, opts)
	if err := confirmHandler.ConfirmDestructiveOperation("clean package caches", "", managers); err != nil {
		return err
	}

	// Print initial message
	formatter := NewOutputFormatter(config)
	formatter.FormatOperationHeader("clean", "package caches", managers, config.DryRun)

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
	validator := NewCommandValidator()
	if err := validator.ValidateManagersAvailable(managers); err != nil {
		return err
	}

	// Safety prompt for destructive operation
	confirmHandler := NewConfirmationHandler(config, opts)
	if err := confirmHandler.ConfirmDestructiveOperation("remove orphaned packages", "", managers); err != nil {
		return err
	}

	// Print initial message
	formatter := NewOutputFormatter(config)
	formatter.FormatOperationHeader("autoremove", "orphaned packages", managers, config.DryRun)

	// Use Registry concurrent operations for performance
	registry := manager.GetGlobalRegistry()
	autoRemoveResults := registry.AutoRemoveAllConcurrent(ctx, opts)

	// Process results using new result processor
	processor := NewResultProcessor(formatter, config)
	return processor.ProcessPackageOperationResults(autoRemoveResults, managers, "autoremove", "removed")
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
