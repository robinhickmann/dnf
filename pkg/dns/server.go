package dns

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/robinhickmann/dnf/pkg/config"
)

var serial = uint32(time.Now().Unix())

// NewServer returns and starts a new dns.Server with the provided config options.
func NewServer(cfg *config.Config) []*dns.Server {
	srv := []*dns.Server{}

	dns.HandleFunc(cfg.DNS.Zone.Name, func(w dns.ResponseWriter, r *dns.Msg) {
		dnsHandler(w, r, cfg)
	})

	for _, bind := range cfg.DNS.Binds {
		s := &dns.Server{
			Addr:         formatAddr(bind, cfg.DNS.Port),
			Net:          "udp",
			ReadTimeout:  cfg.DNS.Timeout.Read,
			WriteTimeout: cfg.DNS.Timeout.Write,
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
		var rr []dns.RR
		var err error

		switch question.Qtype {
		case dns.TypeA:
			rr, err = newRR(fmt.Sprintf("%s 60 IN A %s", question.Name, cfg.DNS.Zone.IPv4))
		case dns.TypeAAAA:
			rr, err = newRR(fmt.Sprintf("%s 60 IN AAAA %s", question.Name, cfg.DNS.Zone.IPv6))
		case dns.TypeHTTPS:
			rr, err = newRR(fmt.Sprintf("%s 60 IN HTTPS 1 . alpn=h2 ipv4hint=%s ipv6hint=%s",
				question.Name, cfg.DNS.Zone.IPv4, cfg.DNS.Zone.IPv6))
		case dns.TypeSOA:
			rr, err = getSOA(question, cfg)
		case dns.TypeNS:
			rr, err = getNS(question, cfg)
		default:
			fmt.Println(r.Id, dns.TypeToString[question.Qtype], question.Name, w.RemoteAddr())

			if soa, err := getSOA(question, cfg); err == nil {
				msg.Ns = append(msg.Ns, soa...)
			}

			continue
		}

		if err != nil {
			continue
		}

		fmt.Println(r.Id, dns.TypeToString[question.Qtype], question.Name, w.RemoteAddr())
		msg.Answer = append(msg.Answer, rr...)
	}

	if err := w.WriteMsg(msg); err != nil {
		fmt.Println(err)
	}
}

func getNS(question dns.Question, cfg *config.Config) ([]dns.RR, error) {
	if question.Name != cfg.DNS.Zone.Name { // zone apex
		return nil, nil
	}

	var rr []dns.RR
	for _, nameserver := range cfg.DNS.Zone.Nameservers {
		ns, err := dns.NewRR(fmt.Sprintf("%s 86400 IN NS %s", question.Name, nameserver))
		if err != nil {
			return nil, err
		}

		rr = append(rr, ns)
	}

	return rr, nil
}

func getSOA(question dns.Question, cfg *config.Config) ([]dns.RR, error) {
	mname := cfg.DNS.Zone.Nameservers[0]
	rname := cfg.DNS.Zone.Email

	rr, err := newRR(fmt.Sprintf("%s 3600 IN SOA %s %s %s 86400 3600 604800 1800",
		question.Name, mname, rname, fmt.Sprint(serial)))

	if err != nil {
		return nil, err
	}

	return rr, nil
}

func newRR(s string) ([]dns.RR, error) {
	rr, err := dns.NewRR(s)
	if err != nil {
		return nil, err
	}
	return []dns.RR{rr}, nil
}

func formatAddr(host string, port int) string {
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]:%d", host, port)
	}
	return fmt.Sprintf("%s:%d", host, port)
}
