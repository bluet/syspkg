package main

import (
	"fmt"
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
		fmt.Errorf("Error while initializing package managers: %v", err)
		// fmt.Println("Error:", err)
		os.Exit(1)
	}

	app := &cli.App{
		Name:  "syspkg",
		Usage: "A universal system package manager",
		Commands: []*cli.Command{
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "Install packages",
				Action: func(c *cli.Context) error {
					pkgNames := c.Args().Slice()
					for _, pm := range pms {
						err := pm.Install(pkgNames)
						if err != nil {
							fmt.Printf("Error while installing packages for %T: %v\n", pm, err)
							// return err
						}
					}
					return nil
				},
			},
			{
				Name:    "uninstall",
				Aliases: []string{"u", "remove"},
				Usage:   "Uninstall packages",
				Action: func(c *cli.Context) error {
					pkgNames := c.Args().Slice()
					for _, pm := range pms {
						err := pm.Uninstall(pkgNames)
						if err != nil {
							fmt.Printf("Error while uninstalling packages for %T: %v\n", pm, err)
							// return err
						}
					}
					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"up"},
				Usage:   "Update package list",
				Action: func(c *cli.Context) error {
					for _, pm := range pms {
						err := pm.Update()
						if err != nil {
							return fmt.Errorf("Error while updating package list for %T: %v", pm, err)
						}
					}
					return nil
				},
			},
			{
				Name:    "upgrade",
				Aliases: []string{"ug"},
				Usage:   "Upgrade packages",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "y",
						Usage: "Assume yes to all prompts",
					},
				},
				Action: func(c *cli.Context) error {
					assumeYes := c.Bool("y")

					// if assumeYes {
					// 	return performUpgrade(pms, assumeYes)
					// }

					listUpgradablePackages(pms)
					if !assumeYes {
						fmt.Print("\nDo you want to perform the system package upgrade? [Y/n]: ")
						input := ""
						_, _ = fmt.Scanln(&input)
						input = strings.ToLower(input)

						if input != "y" && input != "" {
							fmt.Println("Upgrade cancelled.")
							return nil
						}
					}

					return performUpgrade(pms, assumeYes)
				},
			},
			{
				Name:    "search",
				Aliases: []string{"s"},
				Usage:   "Search packages",
				Action: func(c *cli.Context) error {
					keywords := c.Args().Slice()
					for _, pm := range pms {
						pkgs, err := pm.Search(keywords)
						if err != nil {
							return err
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
		upgradablePackages, err := pm.ListUpgradable()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		fmt.Printf("Upgradable packages for %T:\n", pm)
		for _, pkg := range upgradablePackages {
			fmt.Printf("%s: %s %s -> %s (%s)\n", pkg.PackageManager, pkg.Name, pkg.Version, pkg.NewVersion, pkg.Status)
		}
	}
}

func performUpgrade(pms []syspkg.PackageManager, assumeYes bool) error {
	// listUpgradablePackages(pms)

	// if !assumeYes {
	// 	fmt.Print("\nDo you want to perform the system package upgrade? [Y/n]: ")
	// 	input := ""
	// 	_, _ = fmt.Scanln(&input)
	// 	input = strings.ToLower(input)

	// 	if input != "y" && input != "" {
	// 		fmt.Println("Upgrade cancelled.")
	// 		return nil
	// 	}
	// }

	fmt.Println("Performing package upgrade...")

	for _, pm := range pms {
		err := pm.Upgrade()
		if err != nil {
			fmt.Printf("Error while upgrading packages for %T: %v\n", pm, err)
			os.Exit(1)
		}
	}

	fmt.Println("Upgrade completed.")
	return nil
}
