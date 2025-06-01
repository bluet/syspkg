// Package manager provides security utilities for package manager operations
package manager

import (
	"errors"
	"regexp"
)

// packageNameRegex defines the allowed pattern for package names
// This pattern allows:
// - Letters (a-z, A-Z)
// - Numbers (0-9)
// - Dash/hyphen (-)
// - Underscore (_)
// - Period/dot (.)
// - Plus sign (+)
// - Colon (:) for architecture specifiers (e.g., package:amd64)
// - Forward slash (/) for repository specifiers (e.g., repo/package)
// The pattern requires at least one valid character
var packageNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_.+:/]+$`)

// ErrInvalidPackageName is returned when a package name contains invalid characters
var ErrInvalidPackageName = errors.New("invalid package name: contains potentially dangerous characters")

// ValidatePackageName validates that a package name only contains safe characters
// to prevent command injection attacks.
//
// Valid package names may contain:
// - Alphanumeric characters (a-z, A-Z, 0-9)
// - Dash/hyphen (-)
// - Underscore (_)
// - Period/dot (.)
// - Plus sign (+)
// - Colon (:) for architecture specifiers
// - Forward slash (/) for repository specifiers
//
// This function rejects any package names containing:
// - Shell metacharacters (;, |, &, $, `, \, ", ', <, >, (, ), {, }, [, ], *, ?, ~)
// - Whitespace characters
// - Control characters
// - Null bytes
//
// Example valid names:
// - "vim"
// - "libssl1.1"
// - "gcc-9-base"
// - "python3.8"
// - "package:amd64"
// - "repo/package"
//
// Example invalid names:
// - "package; rm -rf /"
// - "package && malicious-command"
// - "package`evil`"
// - "package$(bad)"
func ValidatePackageName(name string) error {
	// Check for empty string
	if name == "" {
		return errors.New("package name cannot be empty")
	}

	// Check length limit (most package managers have reasonable limits)
	if len(name) > 255 {
		return errors.New("package name too long (max 255 characters)")
	}

	// Validate against regex pattern
	if !packageNameRegex.MatchString(name) {
		return ErrInvalidPackageName
	}

	return nil
}

// ValidatePackageNames validates multiple package names
func ValidatePackageNames(names []string) error {
	for _, name := range names {
		if err := ValidatePackageName(name); err != nil {
			return err
		}
	}
	return nil
}
