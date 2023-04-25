// Package manager provides utilities for managing the application.
package manager

// Options represents the various configuration options for the application.
type Options struct {
	// Interactive indicates whether the application should run in interactive mode.
	Interactive bool

	// DryRun indicates whether the application should simulate actions without actually performing them.
	DryRun bool

	// Verbose indicates whether the application should output additional information during execution.
	Verbose bool

	// AssumeYes indicates whether the application should automatically confirm any prompts without user input.
	AssumeYes bool

	// Debug indicates whether the application should run in debug mode, providing more detailed information about its internal operations.
	Debug bool

	// CustomCommandArgs is a slice of strings that can be used to pass additional custom arguments to the application.
	CustomCommandArgs []string
}
