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
    syspkg search vim                    # Search for vim across all managers
    syspkg install vim curl -m system   # Install using system package manager
    syspkg list installed                # List all installed packages
    syspkg upgrade --dry-run             # Show what would be upgraded
    syspkg managers                      # Show available package managers
`
)

type Config struct {
	Manager     string
	ManagerType string
	DryRun      bool
	Verbose     bool
	Quiet       bool
	JSON        bool
	AssumeYes   bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	config := parseArgs()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	registry := manager.GetGlobalRegistry()

	// Get package manager based on config
	pm, err := selectPackageManager(registry, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	opts := &manager.Options{
		DryRun:    config.DryRun,
		Verbose:   config.Verbose,
		Quiet:     config.Quiet,
		AssumeYes: config.AssumeYes,
	}

	// Execute command
	err = executeCommand(ctx, pm, config, opts)
	if err != nil {
		if !config.Quiet {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func parseArgs() *Config {
	config := &Config{}
	args := os.Args[1:]

	// Simple argument parsing
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-m", "--manager":
			if i+1 < len(args) {
				config.ManagerType = args[i+1]
				i++
			}
		case "-n", "--name":
			if i+1 < len(args) {
				config.Manager = args[i+1]
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
			fmt.Println(usage)
			os.Exit(0)
		case "--version":
			fmt.Printf("syspkg version %s\n", version)
			os.Exit(0)
		}
	}

	return config
}

func selectPackageManager(registry *manager.Registry, config *Config) (manager.PackageManager, error) {
	// Get by specific name if provided
	if config.Manager != "" {
		managers := registry.GetAvailable()
		if pm, exists := managers[config.Manager]; exists {
			return pm, nil
		}
		return nil, fmt.Errorf("package manager '%s' not found", config.Manager)
	}

	// Get by type if provided
	if config.ManagerType != "" {
		pm := registry.GetBestMatch(config.ManagerType)
		if pm == nil {
			return nil, fmt.Errorf("no package manager found for type '%s'", config.ManagerType)
		}
		return pm, nil
	}

	// Default: get best system package manager
	pm := registry.GetBestMatch(manager.TypeSystem)
	if pm == nil {
		return nil, fmt.Errorf("no package managers available")
	}

	return pm, nil
}

func executeCommand(ctx context.Context, pm manager.PackageManager, config *Config, opts *manager.Options) error {
	args := getCommandArgs()
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	command := args[0]

	switch command {
	case "search":
		if len(args) < 2 {
			return fmt.Errorf("search requires a query")
		}
		return handleSearch(ctx, pm, args[1:], config, opts)

	case "list":
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

	case "install":
		if len(args) < 2 {
			return fmt.Errorf("install requires package names")
		}
		return handleInstall(ctx, pm, args[1:], config, opts)

	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("remove requires package names")
		}
		return handleRemove(ctx, pm, args[1:], config, opts)

	case "info":
		if len(args) < 2 {
			return fmt.Errorf("info requires a package name")
		}
		return handleInfo(ctx, pm, args[1], config, opts)

	case "update":
		return handleUpdate(ctx, pm, config, opts)

	case "upgrade":
		packages := []string{}
		if len(args) > 1 {
			packages = args[1:]
		}
		return handleUpgrade(ctx, pm, packages, config, opts)

	case "clean":
		return handleClean(ctx, pm, config, opts)

	case "autoremove":
		return handleAutoRemove(ctx, pm, config, opts)

	case "verify":
		if len(args) < 2 {
			return fmt.Errorf("verify requires package names")
		}
		return handleVerify(ctx, pm, args[1:], config, opts)

	case "status":
		return handleStatus(ctx, pm, config, opts)

	case "managers":
		return handleManagers(config)

	default:
		return fmt.Errorf("unknown command: %s", command)
	}
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
		json.NewEncoder(os.Stdout).Encode(data)
	} else {
		switch v := data.(type) {
		case []manager.PackageInfo:
			for _, pkg := range v {
				if config.Quiet {
					fmt.Println(pkg.Name)
				} else {
					fmt.Printf("%-30s %-15s %s\n", pkg.Name, pkg.Version, pkg.Status)
				}
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

func handleSearch(ctx context.Context, pm manager.PackageManager, query []string, config *Config, opts *manager.Options) error {
	if !config.Quiet {
		fmt.Printf("Searching for '%s' using %s...\n", strings.Join(query, " "), pm.GetName())
	}

	packages, err := pm.Search(ctx, query, opts)
	if err != nil {
		return err
	}

	if !config.Quiet {
		fmt.Printf("Found %d packages:\n", len(packages))
	}

	outputResult(packages, config)
	return nil
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
