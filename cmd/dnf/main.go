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
		fmt.Fprint(os.Stderr, err)
		os.Exit(2)
	}

	if flags.DryRun {
		return
	}

	pktConns := dns.NewBinds(cfg)
	listeners := http.NewBinds(cfg)
	tlsConfig := http.NewTLSConfig(cfg.HTTP.TLS.CertFile, cfg.HTTP.TLS.KeyFile)

	if err := dropPrivileges(); err != nil {
		fmt.Fprint(os.Stderr, "failed to drop privileges: %w", err)
		os.Exit(1)
	}

	dns := dns.NewServer(cfg, pktConns)
	http := http.NewServer(cfg, listeners, tlsConfig)

	<-ctx.Done()
	stop()

	fmt.Printf("\nexiting")

	httpCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.Timeout.Shutdown)
	defer cancel()

	dnsCtx, cancel := context.WithTimeout(context.Background(), cfg.DNS.Timeout.Shutdown)
	defer cancel()

	for _, s := range http {
		if err := s.Shutdown(httpCtx); err != nil {
			panic(err)
		}
	}

	for _, s := range dns {
		if err := s.ShutdownContext(dnsCtx); err != nil {
			panic(err)
		}
	}
}
