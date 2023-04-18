package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	// "github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/bluet/syspkg"
	"github.com/bluet/syspkg/manager"
)

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("(This command must be run with root privileges. If you got exist codes 100 or 101, please run this command with sudo.)")
	}

	pms, err := syspkg.NewPackageManager([]string{})
	if err != nil {
		fmt.Printf("Error while initializing package managers: %+v\n", err)
		os.Exit(1)
	}

	app := &cli.App{
		Name:  "syspkg",
		Usage: "A universal system package manager",
		EnableBashCompletion: true,
		UseShortOptionHandling: true,
		Suggest: true,
		// Action: func(c *cli.Context) error {
		// 	var opts = getOptions(c)
		// 	pms = filterPackageManager(pms, c)

		// 	log.Printf("Listing upgradable packages for %T...\n", pms)
		// 	listUpgradablePackages(pms, opts)
		// 	return nil
		// },
		// DefaultCommand: "show upgradable",
		Commands: []*cli.Command{
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "Install packages",
				Action: func(c *cli.Context) error {
					var opts = getOptions(c)
					pms = filterPackageManager(pms, c)

					log.Printf("Installing packages for %T...\n", pms)

					pkgNames := c.Args().Slice()
					for _, pm := range pms {
						log.Printf("Installing packages for %T...\n", pm)
						packages, err := pm.Install(pkgNames, opts)
						if err != nil {
							fmt.Printf("Error while installing packages for %T: %+v\n%+v", pm, err, packages)
							continue
						}
						log.Printf("Installed packages for %T:\n%+v\n", pm, packages)
					}
					return nil
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"remove", "uninstall", "d", "rm", "un"},
				Usage:   "Delete packages",
				Action: func(c *cli.Context) error {
					var opts = getOptions(c)
					pms = filterPackageManager(pms, c)
					pkgNames := c.Args().Slice()

					log.Printf("Deleting packages... for %T\n", pms)

					for _, pm := range pms {
						log.Printf("Deleting packages for %T...\n", pm)
						packages, err := pm.Delete(pkgNames, opts)
						if err != nil {
							fmt.Printf("Error while deleting packages for %T: %+v\n%+v\n", pm, err, packages)
							continue
						}
						log.Printf("Deleted packages for %T:\n%+v\n", pm, packages)
					}
					return nil
				},
			},
			{
				Name:    "refresh",
				Aliases: []string{"update", "r", "re", "u", "up"},
				Usage:   "Refresh package list",
				Action: func(c *cli.Context) error {
					var opts = getOptions(c)
					pms = filterPackageManager(pms, c)

					log.Printf("Refreshing package list... for %T\n", pms)
					for _, pm := range pms {
						log.Printf("Refreshing package list for %T...\n", pm)
						err := pm.Refresh(opts)
						if err != nil {
							fmt.Printf("Error while updating package list for %T: %+v\n", pm, err)
							continue
						}
						log.Printf("Refreshed package list for %T\n", pm)
					}
					return nil
				},
			},
			{
				Name:    "upgrade",
				Aliases: []string{"U", "ug"},
				Usage:   "Upgrade packages",
				Action: func(c *cli.Context) error {
					var opts = getOptions(c)
					pms = filterPackageManager(pms, c)

					log.Printf("Upgrading packages... for %T\n", pms)

					listUpgradablePackages(pms, opts)
					if !opts.AssumeYes {
						fmt.Print("\nDo you want to perform the system package upgrade? [Y/n]: ")
						input := ""
						_, _ = fmt.Scanln(&input)
						input = strings.ToLower(input)

						if input != "y" && input != "" {
							fmt.Println("Upgrade cancelled.")
							return nil
						}
						log.Println("User confirmed upgrade.")
					}

					return performUpgrade(pms, opts)
				},
			},
			{
				Name:    "find",
				Aliases: []string{"search", "f"},
				Usage:   "Find matching packages",
				Action: func(c *cli.Context) error {
					var opts = getOptions(c)
					pms = filterPackageManager(pms, c)
					keywords := c.Args().Slice()

					if len(keywords) == 0 {
						fmt.Println("Please specify keywords to search.")
						return nil
					}
					log.Printf("Finding packages for %T: %+v\n", pms, keywords)

					for _, pm := range pms {
						pkgs, err := pm.Find(keywords, opts)
						if err != nil {
							fmt.Printf("Error while searching packages for %T: %+v\n", pm, err)
							continue
						}

						fmt.Printf("Found results for %T:\n", pm)
						for _, pkg := range pkgs {
							fmt.Printf("%s: %s [%s][%s] (%s)\n", pkg.PackageManager, pkg.Name, pkg.Version, pkg.NewVersion, pkg.Status)
						}
					}
					return nil
				},
			},
			{
				Name:        "show",
				Aliases:     []string{"s"},
				Usage:       "Please specify a subcommand. " + "Use `syspkg show --help` to see the subcommands.",
				Description: `Show information. Please specify a subcommand. Use ` + "`syspkg show --help`" + ` to see the subcommands. Usage: ` + "`syspkg show [subcommand]`",
				Subcommands: []*cli.Command{
					{
						Name:    "upgradable",
						Aliases: []string{"u"},
						Usage:   "Show upgradable packages",
						Action: func(c *cli.Context) error {
							var opts = getOptions(c)
							pms = filterPackageManager(pms, c)

							log.Println("Showing upgradable packages...")

							listUpgradablePackages(pms, opts)
							return nil
						},
					},
					{
						Name:    "package",
						Aliases: []string{"p"},
						Usage:   "Show package information",
						Action: func(c *cli.Context) error {
							var opts = getOptions(c)
							pms = filterPackageManager(pms, c)
							pkgNames := c.Args().Slice()

							if len(pkgNames) != 1 {
								fmt.Println("Please specify one and only one package name.")
								return nil
							}

							log.Println("Showing package information...")

							for _, pm := range pms {
								log.Printf("Showing package information for %T...\n", pm)
								pkg, err := pm.GetPackageInfo(pkgNames[0], opts)
								if err != nil {
									fmt.Printf("Error while showing package info for %T: %+v\n", pm, err)
									continue
								}

								fmt.Printf("Search results for %T:\n", pm)
								fmt.Printf("%s: %s [%s][%s] (%s) %s:%s\n", pkg.PackageManager, pkg.Name, pkg.Version, pkg.NewVersion, pkg.Status, pkg.Category, pkg.Arch)
							}
							return nil
						},
					},
					{
						Name:    "installed",
						Aliases: []string{"i"},
						Usage:   "Show installed packages",
						Action: func(c *cli.Context) error {
							var opts = getOptions(c)
							pms = filterPackageManager(pms, c)

							log.Println("Showing installed packages...")

							for _, pm := range pms {
								log.Printf("Showing installed packages for %T...\n", pm)
								pkgs, err := pm.ListInstalled(opts)
								if err != nil {
									fmt.Printf("Error while showing installed packages for %T: %+v\n", pm, err)
									continue
								}

								fmt.Printf("Search results for %T:\n", pm)
								for _, pkg := range pkgs {
									fmt.Printf("%s: %s [%s][%s] (%s)\n", pkg.PackageManager, pkg.Name, pkg.Version, pkg.NewVersion, pkg.Status)
								}
							}
							return nil
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			// &cli.StringSliceFlag{
			// 	Name:    "package-manager",
			// 	Aliases: []string{"pm"},
			// 	Usage:   "Specify package manager to use. (e.g. apt, apk, pacman, dnf, snap, yum, zypper)",
			// },
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"dbg"},
				Usage:   "Enable debug mode",
			},
			&cli.BoolFlag{
				Name:    "assume-yes",
				Aliases: []string{"y"},
				Usage:   "Assume yes - Assume 'yes' as answer to all prompts. (if -i is not set, this is implied)",
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"dry"},
				Usage:   "Dry run - Do not actually install anything, but show what would be done.",
			},
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "Interactive - Ask questions interactively.",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Verbose - Show more information.",
			},
			&cli.BoolFlag{
				Name:    "apt",
				Usage:  "Use apt package manager",
				// Hidden: true,
			},
			&cli.BoolFlag{
				Name:    "yum",
				Usage:  "Use yum package manager",
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:    "dnf",
				Usage:  "Use dnf package manager",
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:    "pacman",
				Usage:  "Use pacman package manager",
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:    "apk",
				Usage:  "Use apk package manager",
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:    "zypper",
				Usage:  "Use zypper package manager",
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:    "flatpak",
				Usage:  "Use flatpak package manager",
				// Hidden: true,
			},
			&cli.BoolFlag{
				Name:    "snap",
				Usage:  "Use snap package manager",
				Hidden: true,
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func getOptions(c *cli.Context) *manager.Options {
	var opts manager.Options
	opts.Verbose = c.Bool("verbose")
	opts.DryRun = c.Bool("dry-run")
	opts.Interactive = c.Bool("interactive")
	opts.Debug = c.Bool("debug")

	if !opts.Interactive {
		opts.AssumeYes = true
	}

	return &opts
}

func filterPackageManager(availablePMs map[string]syspkg.PackageManager, c *cli.Context) map[string]syspkg.PackageManager {
	if len(availablePMs) == 0 {
		log.Fatal("No package managers available!")
	}

	// if no specific package manager is specified, use all available
	if !c.Bool("apt") && !c.Bool("flatpak") && !c.Bool("snap") && !c.Bool("yum") && !c.Bool("dnf") && !c.Bool("pacman") && !c.Bool("apk") && !c.Bool("zypper") {
		return availablePMs
	}

	var wantedPMs = make(map[string]syspkg.PackageManager)
	for name, pm := range availablePMs {
		if c.Bool(name) {
			wantedPMs[name] = pm
		}
	}
	return wantedPMs
}


func listUpgradablePackages(pms map[string]syspkg.PackageManager, opts *manager.Options) {
	for _, pm := range pms {
		log.Printf("Listing upgradable packages for %T...\n", pm)
		upgradablePackages, err := pm.ListUpgradable(opts)
		if err != nil {
			fmt.Printf("Error while listing upgradable packages for %T: %+v\n", pm, err)
			continue
		}

		fmt.Printf("Upgradable packages for %T:\n", pm)
		for _, pkg := range upgradablePackages {
			fmt.Printf("%s: %s %s -> %s (%s)\n", pkg.PackageManager, pkg.Name, pkg.Version, pkg.NewVersion, pkg.Status)
		}
	}
}

func performUpgrade(pms map[string]syspkg.PackageManager, opts *manager.Options) error {
	fmt.Println("Performing package upgrade...")

	for _, pm := range pms {
		packages, err := pm.Upgrade(opts)
		if err != nil {
			fmt.Printf("Error while upgrading packages for %T: %+v\n%+v", pm, err, packages)
			continue
		}
		// log.Printf("Upgraded packages for %T: %+v", pm, packages)
		log.Println("Packages upgraded:")
		for _, pkg := range packages {
			fmt.Printf("%s: %s -> %s (%s)\n", pkg.PackageManager, pkg.Name, pkg.NewVersion, pkg.Status)
		}
	}

	fmt.Println("Upgrade completed.")
	return nil
}
