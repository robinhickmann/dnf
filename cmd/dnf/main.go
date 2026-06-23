package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/robinhickmann/dnf/pkg/cli"
	"github.com/robinhickmann/dnf/pkg/config"
	"github.com/robinhickmann/dnf/pkg/dns"
	"github.com/robinhickmann/dnf/pkg/http"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	flags := cli.ParseFlags()
	flags.PrintVersion(version, buildTime)

	cfg, err := config.NewConfig(flags.ConfigPath)
	if err != nil {
		panic(err)
	}

	if flags.DryRun {
		return
	}

	dns := dns.NewServer(cfg)
	http := http.NewServer(cfg)

	<-ctx.Done()
	stop()

	fmt.Printf("\nexiting")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.Timeout.Shutdown)
	defer cancel()

	for _, s := range http {
		if err := s.Shutdown(shutdownCtx); err != nil {
			panic(err)
		}
	}

	for _, s := range dns {
		if err := s.Shutdown(); err != nil {
			panic(err)
		}
	}
}
