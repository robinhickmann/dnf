package cli

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"text/tabwriter"
	"time"
)

// PrintVersion prints the version and build information if the version flag is set.
// If the version is printed it exists the program.
func (f *Flags) PrintVersion(version, buildTime string) {
	if !f.VersionCmd {
		return
	}

	commit := "unknown"

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				commit = s.Value[:10]
			case "vcs.modified":
				if s.Value == "true" {
					commit += "-dirty"
				}
			}
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)

	_, _ = fmt.Fprintf(w, " Version:\t%s\n", version)
	_, _ = fmt.Fprintf(w, " Go version:\t%s\n", runtime.Version())
	_, _ = fmt.Fprintf(w, " Git commit:\t%s\n", commit)
	_, _ = fmt.Fprintf(w, " Built:\t%s\n", formatTime(buildTime))
	_, _ = fmt.Fprintf(w, " OS/Arch:\t%s/%s\n", runtime.GOOS, runtime.GOARCH)

	if err := w.Flush(); err != nil {
		panic(err)
	}

	os.Exit(0)
}

func formatTime(t string) string {
	pTime, err := time.Parse("2006-01-02T15:04:05-0700", t)
	if err != nil {
		return "unknown"
	}
	return pTime.Local().Format("Mon Jan 2 15:04:05 2006")
}
