package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bluet/syspkg"
	"github.com/bluet/syspkg/internal"
	"github.com/urfave/cli/v2"
)

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("(This command must be run with root privileges. If you got exist codes 100 or 101, please run this command with sudo.)")
	}

	pms, err := syspkg.NewPackageManager()
	if err != nil {
		fmt.Printf("Error while initializing package managers: %+v", err)
		os.Exit(1)
	}

	app := &cli.App{
		Name:  "syspkg",
		Usage: "A universal system package manager",
		Action: func(c *cli.Context) error {
			var opts = getOptions(c)

			log.Println("Listing upgradable packages...")
			listUpgradablePackages(pms, opts)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "Install packages",
				Action: func(c *cli.Context) error {
					var opts = getOptions(c)

					log.Panicln("Installing packages...")

					pkgNames := c.Args().Slice()
					for _, pm := range pms {
						log.Printf("Installing packages for %T...", pm)
						packages, err := pm.Install(pkgNames, opts)
						if err != nil {
							fmt.Printf("Error while installing packages for %T: %+v\n%+v", pm, err, packages)
							continue
						}
						log.Printf("Installed packages for %T:\n%+v", pm, packages)
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
					pkgNames := c.Args().Slice()

					log.Println("Deleting packages...")

					for _, pm := range pms {
						log.Printf("Deleting packages for %T...", pm)
						packages, err := pm.Delete(pkgNames, opts)
						if err != nil {
							fmt.Printf("Error while deleting packages for %T: %+v\n%+v", pm, err, packages)
							continue
						}
						log.Printf("Deleted packages for %T:\n%+v", pm, packages)
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
					log.Println("Refreshing package list...")
					for _, pm := range pms {
						log.Printf("Refreshing package list for %T...", pm)
						err := pm.Refresh(opts)
						if err != nil {
							fmt.Printf("Error while updating package list for %T: %+v\n", pm, err)
							continue
						}
						log.Printf("Refreshed package list for %T", pm)
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

					log.Println("Upgrading packages...")

					listUpgradablePackages(pms, opts)
					if opts.Interactive {
						fmt.Print("\nDo you want to perform the system package upgrade? [Y/n]: ")
						input := ""
						_, _ = fmt.Scanln(&input)
						input = strings.ToLower(input)

						if input != "y" && input != "" {
							fmt.Println("Upgrade cancelled.")
							return nil
						}
						log.Printf("User confirmed upgrade.")
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
					keywords := c.Args().Slice()
					if len(keywords) == 0 {
						fmt.Println("Please specify keywords to search.")
						return nil
					}
					log.Printf("Finding packages: %+v", keywords)

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
							pkgName := c.Args().Slice()[0]
							if len(pkgName) == 0 {
								fmt.Println("Please specify package name.")
								return nil
							}

							log.Println("Showing package information...")

							for _, pm := range pms {
								log.Printf("Showing package information for %T...", pm)
								pkg, err := pm.GetPackageInfo(pkgName, opts)
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

							log.Println("Showing installed packages...")

							for _, pm := range pms {
								log.Printf("Showing installed packages for %T...", pm)
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
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func getOptions(c *cli.Context) *internal.Options {
	var opts internal.Options
	opts.Verbose = c.Bool("verbose")
	opts.DryRun = c.Bool("dry-run")
	opts.Interactive = c.Bool("interactive")
	opts.Debug = c.Bool("debug")

	if !opts.Interactive {
		opts.AssumeYes = true
	}

	return &opts
}

func listUpgradablePackages(pms []syspkg.PackageManager, opts *internal.Options) {
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

func performUpgrade(pms []syspkg.PackageManager, opts *internal.Options) error {
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
