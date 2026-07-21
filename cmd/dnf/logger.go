package main

import (
	"io"
	"os"
	"strings"

	"github.com/robinhickmann/dnf/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setupLogger(cfg *config.Log, debug bool) {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = cfg.TimeFormat
	log.Logger = getLogWriter(cfg.Output, debug)
}

func getLogWriter(output string, debug bool) zerolog.Logger {
	var out io.Writer

	switch strings.ToLower(output) {
	case "stderr":
		out = os.Stderr
	default:
		out = os.Stdout
	}

	if debug {
		return log.Output(zerolog.ConsoleWriter{Out: out})
	}
	return log.Output(out)
}
