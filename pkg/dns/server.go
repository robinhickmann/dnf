package dns

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/robinhickmann/dnf/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	serial = uint32(time.Now().Unix())
	logger zerolog.Logger
)

// NewBind returns and binds all the net.PacketConn's for the dns binds.
func NewBinds(cfg *config.Config) []net.PacketConn {
	pktConns := []net.PacketConn{}

	for _, bind := range cfg.DNS.Binds {
		pktConn, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", bind, cfg.DNS.Port))
		if err != nil {
			panic(err)
		}
		pktConns = append(pktConns, pktConn)
	}

	return pktConns
}

// NewServer returns and starts a new dns.Server with the provided config options.
func NewServer(cfg *config.Config, pktConns []net.PacketConn) []*dns.Server {
	logger = log.With().Str("server", "dns").Logger()
	srv := []*dns.Server{}

	dns.HandleFunc(cfg.DNS.Zone.Name, func(w dns.ResponseWriter, r *dns.Msg) {
		dnsHandler(w, r, cfg)
	})

	for _, pktConn := range pktConns {
		s := &dns.Server{
			PacketConn:   pktConn,
			Addr:         pktConn.LocalAddr().String(),
			Net:          "udp",
			ReadTimeout:  cfg.DNS.Timeout.Read,
			WriteTimeout: cfg.DNS.Timeout.Write,
		}

		go func(s *dns.Server) {
			err := s.ActivateAndServe()
			if err != nil {
				panic(err)
			}
		}(s)

		srv = append(srv, s)
		logger.Info().Msgf("listening on udp://%s", s.Addr)
	}

	return srv
}

func dnsHandler(w dns.ResponseWriter, r *dns.Msg, cfg *config.Config) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

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
		case dns.TypeNS:
			rr, err = getNS(question, cfg)
		case dns.TypeSOA:
			if isApex(question.Name, cfg.DNS.Zone.Name) {
				rr, err = getSOA(question, cfg)
			}
		default:
			logger.Debug().
				Int("id", int(r.Id)).
				Str("type", dns.TypeToString[question.Qtype]).
				Str("name", strings.ToLower(question.Name)).
				Str("addr", w.RemoteAddr().String()).
				Msg("unknown record")
		}

		if len(rr) == 0 { // add SOA to the authority section
			if soa, err := getSOA(question, cfg); err == nil {
				msg.Ns = append(msg.Ns, soa...)
			}
			continue
		}

		if err != nil {
			logger.Err(err).Msg("unable to get record")
			continue
		}

		msg.Answer = append(msg.Answer, rr...)

		logger.Debug().
			Int("id", int(r.Id)).
			Str("type", dns.TypeToString[question.Qtype]).
			Str("name", strings.ToLower(question.Name)).
			Str("addr", w.RemoteAddr().String()).
			Msg("query")
	}

	if err := w.WriteMsg(msg); err != nil {
		fmt.Println(err)
	}
}

func getNS(question dns.Question, cfg *config.Config) ([]dns.RR, error) {
	if !isApex(question.Name, cfg.DNS.Zone.Name) {
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

func isApex(name, zone string) bool {
	return strings.EqualFold(name, zone)
}
