package http

import (
	"fmt"
	stdlog "log"
	"net/http"

	"github.com/robinhickmann/dnf/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logger zerolog.Logger

// NewServer returns and starts a new http.Server with the provided config options.
func NewServer(cfg *config.Config) []*http.Server {
	logger = log.With().Str("server", "http").Logger()
	http.HandleFunc("/", httpHandler)

	srv := []*http.Server{}

	for _, bind := range cfg.HTTP.Binds {
		s := &http.Server{
			Addr:         fmt.Sprintf("%s:%d", bind, cfg.HTTP.Port),
			IdleTimeout:  cfg.HTTP.Timeout.Idle,
			ReadTimeout:  cfg.HTTP.Timeout.Read,
			WriteTimeout: cfg.HTTP.Timeout.Write,
			ErrorLog:     stdlog.New(logger, "", 0),
		}

		go func() {
			var err error

			if cfg.HTTP.TLS.Enabled {
				err = s.ListenAndServeTLS(cfg.HTTP.TLS.CertFile, cfg.HTTP.TLS.KeyFile)
			} else {
				err = s.ListenAndServe()
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

func httpHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("Hello World! ")); err != nil {
		logger.Error().Err(err).Msg("failed to write response")
	}

	logger.Debug().
		Str("host", r.Host).
		Str("user_agent", r.UserAgent()).
		Str("addr", r.RemoteAddr).
		Msg("request")
}
