package cli

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"text/tabwriter"
	"time"
)

func printVersion(version, buildTime string) {
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

	fmt.Fprintf(w, " Version:\t%s\n", version)
	fmt.Fprintf(w, " Go version:\t%s\n", runtime.Version())
	fmt.Fprintf(w, " Git commit:\t%s\n", commit)
	fmt.Fprintf(w, " Built:\t%s\n", formatTime(buildTime))
	fmt.Fprintf(w, " OS/Arch:\t%s/%s\n", runtime.GOOS, runtime.GOARCH)

	w.Flush()
}

func formatTime(t string) string {
	pTime, err := time.Parse("2006-01-02T15:04:05-0700", t)
	if err != nil {
		return "unknown"
	}
	return pTime.Local().Format("Mon Jan 2 15:04:05 2006")
}
