package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bluet/syspkg"
	"github.com/urfave/cli/v2"
)

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("Error: This program must be run with root privileges.")
		// os.Exit(1)
	}

	pms, err := syspkg.NewPackageManager()
	if err != nil {
		fmt.Printf("Error while initializing package managers: %v", err)
		// fmt.Println("Error:", err)
		os.Exit(1)
	}

	app := &cli.App{
		Name:  "syspkg",
		Usage: "A universal system package manager",
		Action: func(c *cli.Context) error {
			for _, pm := range pms {
				packages, err := pm.ListUpgradable(nil)
				if err != nil {
					fmt.Printf("Error while listing upgradable packages for %T: %v", pm, err)
					continue
				}
				log.Printf("Upgradable packages for %T:\n%v", pm, packages)
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				// default command
				Name:    "show-upgradable",
				Aliases: []string{"show-upgrades", "su"},
				Usage:   "Show upgradable packages",
				Action: func(c *cli.Context) error {
					for _, pm := range pms {
						packages, err := pm.ListUpgradable(nil)
						if err != nil {
							fmt.Printf("Error while listing upgradable packages for %T: %v", pm, err)
							continue
						}
						log.Printf("Upgradable packages for %T:\n%v", pm, packages)
					}
					return nil
				},
			},
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "Install packages",
				Action: func(c *cli.Context) error {
					pkgNames := c.Args().Slice()
					for _, pm := range pms {
						packages, err := pm.Install(pkgNames, nil)
						if err != nil {
							fmt.Printf("Error while installing packages for %T: %v\n%v", pm, err, packages)
							continue
						}
						log.Printf("Installed packages for %T:\n%v", pm, packages)
					}
					return nil
				},
			},
			{
				Name:    "uninstall",
				Aliases: []string{"remove", "un", "rm"},
				Usage:   "Uninstall packages",
				Action: func(c *cli.Context) error {
					pkgNames := c.Args().Slice()
					for _, pm := range pms {
						packages, err := pm.Uninstall(pkgNames, nil)
						if err != nil {
							fmt.Printf("Error while uninstalling packages for %T: %v\n%v", pm, err, packages)
							continue
						}
						log.Printf("Uninstalled packages for %T:\n%v", pm, packages)
					}
					return nil
				},
			},
			{
				Name:    "refresh",
				Aliases: []string{"update", "r", "re", "u", "up"},
				Usage:   "Refresh package list",
				Action: func(c *cli.Context) error {
					for _, pm := range pms {
						err := pm.Refresh(nil)
						if err != nil {
							fmt.Printf("Error while updating package list for %T: %v\n", pm, err)
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
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "assume-yes",
						Aliases: []string{"y", "yes", "non-interactive"},
						Usage:   "Assume yes to all prompts",
					},
				},
				Action: func(c *cli.Context) error {
					autoYes := c.Bool("y")

					listUpgradablePackages(pms)
					if !autoYes {
						fmt.Print("\nDo you want to perform the system package upgrade? [Y/n]: ")
						input := ""
						_, _ = fmt.Scanln(&input)
						input = strings.ToLower(input)

						if input != "y" && input != "" {
							fmt.Println("Upgrade cancelled.")
							return nil
						}
					}

					return performUpgrade(pms, autoYes)
				},
			},
			{
				Name:    "search",
				Aliases: []string{"find", "s"},
				Usage:   "Search packages",
				Action: func(c *cli.Context) error {
					keywords := c.Args().Slice()
					for _, pm := range pms {
						pkgs, err := pm.Search(keywords, nil)
						if err != nil {
							fmt.Printf("Error while searching packages for %T: %v\n", pm, err)
							continue
						}

						fmt.Printf("Search results for %T:\n", pm)
						for _, pkg := range pkgs {
							fmt.Printf("%s: %s %s (%s)\n", pkg.PackageManager, pkg.Name, pkg.Version, pkg.Status)
						}
					}
					return nil
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func listUpgradablePackages(pms []syspkg.PackageManager) {
	for _, pm := range pms {
		upgradablePackages, err := pm.ListUpgradable(nil)
		if err != nil {
			fmt.Printf("Error while listing upgradable packages for %T: %v\n", pm, err)
			continue
		}

		fmt.Printf("Upgradable packages for %T:\n", pm)
		for _, pkg := range upgradablePackages {
			fmt.Printf("%s: %s %s -> %s (%s)\n", pkg.PackageManager, pkg.Name, pkg.Version, pkg.NewVersion, pkg.Status)
		}
	}
}

func performUpgrade(pms []syspkg.PackageManager, autoYes bool) error {
	fmt.Println("Performing package upgrade...")

	for _, pm := range pms {
		packages, err := pm.Upgrade(&syspkg.Options{Interactive: !autoYes})
		if err != nil {
			fmt.Printf("Error while upgrading packages for %T: %v\n%v", pm, err, packages)
			continue
		}
		log.Printf("Upgraded packages for %T: %v", pm, packages)
	}

	fmt.Println("Upgrade completed.")
	return nil
}
