package cli

import (
	"flag"
	"os"
)

type Flags struct {
	ConfigPath string
	VersionCmd bool
	DryRun     bool
	Debug      bool
}

// ParseFlags returns a new Flags struct with all the parsed cli flags.
func ParseFlags() *Flags {
	var flags Flags

	flag.StringVar(&flags.ConfigPath, "config", "", "Path to the config file")
	flag.BoolVar(&flags.VersionCmd, "version", false, "Print the version and exit")
	flag.BoolVar(&flags.DryRun, "dry-run", false, "Validate config file and exit")
	flag.BoolVar(&flags.Debug, "debug", false, "Enable debug logging")
	flag.Parse()

	return &flags
}

// PrintVersion prints the version and build information if the version flag is set.
// If the version is printed it exists the program.
func (f *Flags) PrintVersion(version, buildTime string) {
	if f.VersionCmd {
		printVersion(version, buildTime)
		os.Exit(0)
	}
}
