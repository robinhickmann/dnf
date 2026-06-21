package main

import (
	"fmt"
	"runtime/debug"
)

func printVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("no build info available")
		return
	}

	for _, s := range info.Settings {
		fmt.Println(s.Key, s.Value)
	}
}
