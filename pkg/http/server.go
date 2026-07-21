package http

import (
	"crypto/tls"
	"fmt"
	stdlog "log"
	"net"
	"net/http"

	"github.com/robinhickmann/dnf/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger

// NewBind returns and binds all the net.Listener's for the dns binds.
func NewBinds(cfg *config.Config) []net.Listener {
	listeners := []net.Listener{}

	for _, bind := range cfg.HTTP.Binds {
		ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bind, cfg.HTTP.Port))
		if err != nil {
			panic(err)
		}
		listeners = append(listeners, ln)
	}

	return listeners
}

// NewServer returns and starts a new http.Server with the provided config options.
func NewServer(cfg *config.Config, listeners []net.Listener, tlsConfig *tls.Config) []*http.Server {
	logger = log.With().Str("server", "http").Logger()
	http.HandleFunc("/", httpHandler)

	srv := []*http.Server{}

	for _, ln := range listeners {
		s := &http.Server{
			Addr:         ln.Addr().String(),
			IdleTimeout:  cfg.HTTP.Timeout.Idle,
			ReadTimeout:  cfg.HTTP.Timeout.Read,
			WriteTimeout: cfg.HTTP.Timeout.Write,
			TLSConfig:    tlsConfig,
			ErrorLog:     stdlog.New(logger, "", 0),
		}

		go func() {
			var err error

			if cfg.HTTP.TLS.Enabled {
				err = s.ServeTLS(ln, "", "")
			} else {
				err = s.Serve(ln)
			}

			if err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()

		if cfg.HTTP.TLS.Enabled {
			logger.Info().Msgf("listening on https://%s", s.Addr)
		} else {
			logger.Info().Msgf("listening on http://%s", s.Addr)
		}

		srv = append(srv, s)
	}

	return srv
}

// NewTLSConfig tries to load the key pair and returns a new tls.Config if successful.
func NewTLSConfig(certFile, keyFile string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load key pair")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("Hello World! ")); err != nil {
		logger.Err(err).Msg("failed to write response")
	}

	logger.Debug().
		Str("host", r.Host).
		Str("user_agent", r.UserAgent()).
		Str("addr", r.RemoteAddr).
		Msg("request")
}
