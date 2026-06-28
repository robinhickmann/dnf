package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/robinhickmann/dnf/pkg/config"
)

// NewServer returns and starts a new http.Server with the provided config options.
func NewServer(cfg *config.Config) []*http.Server {
	http.HandleFunc("/", indexHandler)
	srv := []*http.Server{}

	for _, bind := range cfg.HTTP.Binds {
		s := &http.Server{
			Addr:         formatAddr(bind, cfg.HTTP.Port),
			IdleTimeout:  cfg.HTTP.Timeout.Idle,
			ReadTimeout:  cfg.HTTP.Timeout.Read,
			WriteTimeout: cfg.HTTP.Timeout.Write,
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
			fmt.Printf("HTTP Server listening on https://%s\n", s.Addr)
		} else {
			fmt.Printf("HTTP Server listening on http://%s\n", s.Addr)
		}

		srv = append(srv, s)
	}

	return srv
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("Hello World! ")); err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.RemoteAddr, r.UserAgent(), r.Host)
}

func formatAddr(host string, port int) string {
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]:%d", host, port)
	}
	return fmt.Sprintf("%s:%d", host, port)
}
