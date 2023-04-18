package manager

type Options struct {
	Interactive bool
	DryRun      bool
	Verbose     bool
	AssumeYes   bool
	Debug       bool
	CustomCommandArgs []string
}
