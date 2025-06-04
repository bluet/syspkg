// syspkg - Universal package manager CLI
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bluet/syspkg/manager"

	// Import all available package managers
	_ "github.com/bluet/syspkg/manager/apt"
	_ "github.com/bluet/syspkg/manager/flatpak"
	_ "github.com/bluet/syspkg/manager/snap"
	_ "github.com/bluet/syspkg/manager/yum"
)

const (
	version = "2.0.0"
	usage   = `syspkg - Universal Package Manager

USAGE:
    syspkg <command> [options] [packages...]

COMMANDS:
    search <query>        Search for packages
    list [filter]         List packages (installed, upgradable, all)
    install <packages>    Install packages
    remove <packages>     Remove packages
    info <package>        Show package information
    update               Update package lists
    upgrade [packages]    Upgrade packages (all if none specified)
    clean                Clean package cache
    autoremove           Remove orphaned packages
    verify <packages>    Verify package integrity
    status               Show package manager status
    managers             List available package managers

OPTIONS:
    # Multi-Manager Control
    --apt, --snap, --flatpak    Use only specific manager(s)
    --all                       Use all available managers (default)
    --status                    Show real package status (installed/available/upgradable)

    # Standard Options
    -m, --manager TYPE   Use specific manager type (system, language, etc.)
    -n, --name NAME      Use specific manager by name (apt, npm, etc.)
    -d, --dry-run        Show what would be done without executing
    -v, --verbose        Show detailed output
    -q, --quiet          Minimal output
    -j, --json           Output in JSON format
    -y, --yes            Assume yes to all prompts
    -h, --help           Show this help
    --version            Show version

EXAMPLES:
    # Fast repository search (default)
    syspkg search vim                    # Search across all managers (fast)
    syspkg search vim --apt              # Search only APT (fast)

    # Status-aware search (slower but accurate)
    syspkg search vim --status           # Show real installation status
    syspkg search vim --apt --status     # APT with status information

    # Other operations
    syspkg install vim curl -m system   # Install using system package manager
    syspkg list installed                # List all installed packages
    syspkg upgrade --dry-run             # Show what would be upgraded
    syspkg managers                      # Show available package managers
`
)

type Config struct {
	Manager        string
	ManagerType    string
	ManagerFilters map[string]bool // apt, snap, flatpak specific filters
	UseAllManagers bool            // Default: true (user's original superior approach)
	ShowStatus     bool            // --status flag for status-aware search
	DryRun         bool
	Verbose        bool
	Quiet          bool
	JSON           bool
	AssumeYes      bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	config := parseArgs()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	registry := manager.GetGlobalRegistry()

	// Get package managers based on config
	managers, err := selectPackageManagers(registry, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
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
		case "-m", "--manager":
			if i+1 < len(args) {
				config.ManagerType = args[i+1]
				i++
			}
		case "-n", "--name":
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
			fmt.Print(usage)
			os.Exit(0)
		case "--version":
			fmt.Printf("syspkg version %s\n", version)
			os.Exit(0)
		}
	}

	return config
}

// formatPackageInfo formats a single package for display
func formatPackageInfo(pkg manager.PackageInfo, config *Config) {
	if config.Quiet {
		fmt.Println(pkg.Name)
		return
	}

	// Format version string
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

	// Display with or without status based on config
	if config.ShowStatus {
		fmt.Printf("  %-25s %-18s (%s)%s\n", pkg.Name, versionStr, pkg.Status, desc)
	} else {
		fmt.Printf("  %-25s %-18s%s\n", pkg.Name, versionStr, desc)
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

	// Single manager by type
	if config.ManagerType != "" {
		pm := registry.GetBestMatch(config.ManagerType)
		if pm == nil {
			return nil, fmt.Errorf("no package manager found for type '%s'", config.ManagerType)
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
	if !config.Quiet {
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
	for name, pm := range managers {
		packages, err := pm.Search(ctx, query, opts)
		if err != nil {
			if !config.Quiet {
				fmt.Printf("❌ Error searching %s: %v\n", name, err)
			}
			continue
		}

		if len(packages) == 0 {
			if !config.Quiet {
				if len(managers) == 1 {
					fmt.Printf("No packages found.\n")
				} else {
					fmt.Printf("📦 %s: No results found\n", strings.ToUpper(name))
				}
			}
			continue
		}

		// Display header with emoji and count
		if !config.Quiet {
			emoji := getManagerEmoji(name)
			fmt.Printf("%s %s (%d packages):\n", emoji, strings.ToUpper(name), len(packages))
		}

		// Display packages using unified logic
		for _, pkg := range packages {
			formatPackageInfo(pkg, config)
		}

		if !config.Quiet && len(packages) > 0 {
			fmt.Println()
		}
		totalPackages += len(packages)
	}

	// Summary for multi-manager searches
	if !config.Quiet && len(managers) > 1 && totalPackages > 0 {
		fmt.Printf("📊 Summary: %d packages found across %d managers\n", totalPackages, len(managers))
	}

	return nil
}

func executeMultiCommand(ctx context.Context, managers map[string]manager.PackageManager, config *Config, opts *manager.Options) error {
	args := getCommandArgs()
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]

	// For multi-manager operations (search is the main one)
	switch command {
	case "search":
		if len(args) < 2 {
			return fmt.Errorf("search requires a query")
		}
		return handleSearchUnified(ctx, managers, args[1:], config, opts)

	case "managers":
		return handleManagers(config)

	default:
		// For single-manager operations, use the first available manager
		// or let user specify with -n flag
		var pm manager.PackageManager
		if len(managers) == 1 {
			for _, mgr := range managers {
				pm = mgr
				break
			}
		} else {
			// Multiple managers available, use the first one or require user to specify
			if config.Manager == "" {
				return fmt.Errorf("multiple package managers available (%d), please specify one with -n flag", len(managers))
			}
			var exists bool
			pm, exists = managers[config.Manager]
			if !exists {
				return fmt.Errorf("specified manager '%s' not available", config.Manager)
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
	if len(args) < 2 {
		return fmt.Errorf("install requires package names")
	}
	return handleInstall(ctx, pm, args[1:], config, opts)
}

func executeRemoveCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	if len(args) < 2 {
		return fmt.Errorf("remove requires package names")
	}
	return handleRemove(ctx, pm, args[1:], config, opts)
}

func executeInfoCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	if len(args) < 2 {
		return fmt.Errorf("info requires a package name")
	}
	return handleInfo(ctx, pm, args[1], config, opts)
}

func executeUpgradeCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	packages := []string{}
	if len(args) > 1 {
		packages = args[1:]
	}
	return handleUpgrade(ctx, pm, packages, config, opts)
}

func executeVerifyCommand(ctx context.Context, pm manager.PackageManager, args []string, config *Config, opts *manager.Options) error {
	if len(args) < 2 {
		return fmt.Errorf("verify requires package names")
	}
	return handleVerify(ctx, pm, args[1:], config, opts)
}

func getCommandArgs() []string {
	args := os.Args[1:]
	var result []string

	// Skip flags and extract command + args
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip flags that take values
		if (arg == "-m" || arg == "--manager" || arg == "-n" || arg == "--name") && i+1 < len(args) {
			i++ // Skip the value too
			continue
		}

		// Skip single flags
		if strings.HasPrefix(arg, "-") {
			continue
		}

		// This is a command or argument
		result = append(result, arg)
	}

	return result
}

func outputResult(data interface{}, config *Config) {
	if config.JSON {
		_ = json.NewEncoder(os.Stdout).Encode(data)
	} else {
		switch v := data.(type) {
		case []manager.PackageInfo:
			for _, pkg := range v {
				formatPackageInfo(pkg, config)
			}
		case manager.PackageInfo:
			if config.Quiet {
				fmt.Println(v.Name)
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
	if !config.Quiet {
		fmt.Printf("Listing %s packages using %s...\n", filter, pm.GetName())
	}

	packages, err := pm.List(ctx, filter, opts)
	if err != nil {
		return err
	}

	if !config.Quiet {
		fmt.Printf("Found %d packages:\n", len(packages))
	}

	outputResult(packages, config)
	return nil
}

func handleInstall(ctx context.Context, pm manager.PackageManager, packages []string, config *Config, opts *manager.Options) error {
	if !config.Quiet {
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

	if !config.Quiet {
		fmt.Printf("Successfully processed %d packages:\n", len(results))
	}

	outputResult(results, config)
	return nil
}

func handleRemove(ctx context.Context, pm manager.PackageManager, packages []string, config *Config, opts *manager.Options) error {
	if !config.Quiet {
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

	if !config.Quiet {
		fmt.Printf("Successfully processed %d packages:\n", len(results))
	}

	outputResult(results, config)
	return nil
}

func handleInfo(ctx context.Context, pm manager.PackageManager, packageName string, config *Config, opts *manager.Options) error {
	pkg, err := pm.GetInfo(ctx, packageName, opts)
	if err != nil {
		return err
	}

	outputResult(pkg, config)
	return nil
}

func handleUpdate(ctx context.Context, pm manager.PackageManager, config *Config, opts *manager.Options) error {
	if !config.Quiet {
		fmt.Printf("Updating package lists using %s...\n", pm.GetName())
	}

	err := pm.Refresh(ctx, opts)
	if err != nil {
		return err
	}

	if !config.Quiet {
		fmt.Println("Package lists updated successfully")
	}

	return nil
}

func handleUpgrade(ctx context.Context, pm manager.PackageManager, packages []string, config *Config, opts *manager.Options) error {
	target := "all packages"
	if len(packages) > 0 {
		target = strings.Join(packages, ", ")
	}

	if !config.Quiet {
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

	if !config.Quiet {
		fmt.Printf("Successfully processed %d packages:\n", len(results))
	}

	outputResult(results, config)
	return nil
}

func handleClean(ctx context.Context, pm manager.PackageManager, config *Config, opts *manager.Options) error {
	if !config.Quiet {
		fmt.Printf("Cleaning package cache using %s...\n", pm.GetName())
	}

	err := pm.Clean(ctx, opts)
	if err != nil {
		return err
	}

	if !config.Quiet {
		fmt.Println("Package cache cleaned successfully")
	}

	return nil
}

func handleAutoRemove(ctx context.Context, pm manager.PackageManager, config *Config, opts *manager.Options) error {
	if !config.Quiet {
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

	if !config.Quiet {
		fmt.Printf("Successfully processed %d packages:\n", len(results))
	}

	outputResult(results, config)
	return nil
}

func handleVerify(ctx context.Context, pm manager.PackageManager, packages []string, config *Config, opts *manager.Options) error {
	if !config.Quiet {
		fmt.Printf("Verifying packages: %s\n", strings.Join(packages, ", "))
	}

	results, err := pm.Verify(ctx, packages, opts)
	if err != nil {
		return err
	}

	outputResult(results, config)
	return nil
}

func handleStatus(ctx context.Context, pm manager.PackageManager, config *Config, opts *manager.Options) error {
	status, err := pm.Status(ctx, opts)
	if err != nil {
		return err
	}

	if config.JSON {
		outputResult(status, config)
	} else {
		fmt.Printf("Package Manager: %s\n", pm.GetName())
		fmt.Printf("Type: %s\n", pm.GetType())
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
	managers := registry.GetAvailable()

	if config.JSON {
		type ManagerInfo struct {
			Name      string `json:"name"`
			Type      string `json:"type"`
			Available bool   `json:"available"`
		}

		var infos []ManagerInfo
		for name, pm := range managers {
			infos = append(infos, ManagerInfo{
				Name:      name,
				Type:      pm.GetType(),
				Available: pm.IsAvailable(),
			})
		}

		outputResult(infos, config)
	} else {
		fmt.Println("Available Package Managers:")
		for name, pm := range managers {
			status := "❌"
			if pm.IsAvailable() {
				status = "✅"
			}
			fmt.Printf("  %s %-10s (%s)\n", status, name, pm.GetType())
		}
	}

	return nil
}
