package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/robinhickmann/dnf/cmd/dns"
	"github.com/robinhickmann/dnf/cmd/http"
	"github.com/robinhickmann/dnf/pkg/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	flags := config.ParseFlags()

	if flags.VersionCmd {
		printVersion()
		return
	}

	cfg, err := config.NewConfig(flags.ConfigPath, "config.yml", "config.yaml")
	if err != nil {
		panic(err)
	}

	dns.NewServer(cfg)
	http.NewServer(cfg)

	<-ctx.Done()
	stop()

	fmt.Printf("\nexiting\n")
}
