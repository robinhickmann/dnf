package dns

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/robinhickmann/dnf/pkg/config"
)

// NewServer returns and starts a new dns.Server with the provided config options.
func NewServer(cfg *config.Config) []*dns.Server {
	srv := []*dns.Server{}

	dns.HandleFunc(cfg.DNS.Zone.Name, func(w dns.ResponseWriter, r *dns.Msg) {
		dnsHandler(w, r, cfg)
	})

	for _, bind := range cfg.DNS.Binds {
		s := &dns.Server{
			Addr: formatAddr(bind, cfg.DNS.Port),
			Net:  "udp",
		}

		go func(s *dns.Server) {
			err := s.ListenAndServe()
			if err != nil {
				panic(err)
			}
		}(s)

		fmt.Printf("DNS Server listening on udp://%s\n", s.Addr)
		srv = append(srv, s)
	}

	return srv
}

func dnsHandler(w dns.ResponseWriter, r *dns.Msg, cfg *config.Config) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true
	msg.Compress = true

	for _, question := range r.Question {
		var rr dns.RR
		var err error

		switch question.Qtype {
		case dns.TypeA:
			rr, err = dns.NewRR(fmt.Sprintf("%s 60 IN A %s", question.Name, cfg.DNS.Zone.IPv4))
		case dns.TypeAAAA:
			rr, err = dns.NewRR(fmt.Sprintf("%s 60 IN AAAA %s", question.Name, cfg.DNS.Zone.IPv6))
		default:
			fmt.Println(r.Id, dns.TypeToString[question.Qtype], question.Name)
			continue
		}

		if err != nil {
			continue
		}

		msg.Answer = append(msg.Answer, rr)

		fmt.Println(r.Id, dns.TypeToString[question.Qtype], question.Name)
	}

	if err := w.WriteMsg(msg); err != nil {
		fmt.Println(err)
	}
}

func formatAddr(host string, port int) string {
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]:%d", host, port)
	}
	return fmt.Sprintf("%s:%d", host, port)
}
