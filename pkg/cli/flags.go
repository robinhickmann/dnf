package cli

import (
	"flag"
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
