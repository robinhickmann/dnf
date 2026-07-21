package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/robinhickmann/dnf/pkg/cli"
	"github.com/robinhickmann/dnf/pkg/config"
	"github.com/robinhickmann/dnf/pkg/dns"
	"github.com/robinhickmann/dnf/pkg/http"
	"github.com/rs/zerolog/log"
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(78) // EX_CONFIG
	}

	if flags.DryRun {
		return
	}

	setupLogger(&cfg.Log, flags.Debug)

	pktConns := dns.NewBinds(cfg)
	listeners := http.NewBinds(cfg)
	tlsConfig := http.NewTLSConfig(cfg.HTTP.TLS.CertFile, cfg.HTTP.TLS.KeyFile)

	if err := dropPrivileges(); err != nil {
		log.Fatal().Err(err).Msg("failed to drop privileges")
	}

	dns := dns.NewServer(cfg, pktConns)
	http := http.NewServer(cfg, listeners, tlsConfig)

	<-ctx.Done()
	stop()

	log.Info().Msg("shutting down")

	var wg sync.WaitGroup

	for _, s := range http {
		wg.Go(shutdownWithTimeout("http", cfg.HTTP.Timeout.Shutdown, s.Shutdown))
	}

	for _, s := range dns {
		wg.Go(shutdownWithTimeout("dns", cfg.DNS.Timeout.Shutdown, s.ShutdownContext))
	}

	wg.Wait()
	log.Info().Msg("exited")
}

func shutdownWithTimeout(field string, timeout time.Duration, fn func(context.Context) error) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := fn(ctx); err != nil {
			log.Error().Err(err).Msgf("%s shutdown error", field)
		}
	}
}
