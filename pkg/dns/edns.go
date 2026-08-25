package dns

import (
	"net"
	"net/netip"

	"github.com/miekg/dns"
)

type EDNS struct {
	MaxSize  int
	ClientIP netip.Prefix
}

const ednsSize = 1232

func init() {

}

func handleEDNS(r, msg *dns.Msg) (*EDNS, bool) {
	opt := r.IsEdns0()
	if opt == nil {
		return &EDNS{MaxSize: dns.MinMsgSize}, true
	}

	msg.SetEdns0(ednsSize, false)
	msgOpt := msg.IsEdns0()

	if opt.Version() != 0 {
		*msg = *newErr(r, msgOpt, dns.RcodeBadVers)
		return nil, false
	}

	clientSize := max(512, int(opt.UDPSize()))

	ctx := &EDNS{
		MaxSize: min(clientSize, ednsSize),
	}

	for _, option := range opt.Option {
		var ok = true

		switch o := option.(type) {
		case *dns.EDNS0_SUBNET:
			ctx.ClientIP, ok = handleECS(o, msgOpt)
		}

		if !ok {
			*msg = *newErr(r, msgOpt, dns.RcodeFormatError)
			return nil, false
		}
	}

	return ctx, true
}

func handleECS(ecs *dns.EDNS0_SUBNET, opt *dns.OPT) (netip.Prefix, bool) {
	prefix, ok := toPrefix(ecs.Address, ecs.SourceNetmask)
	if !ok {
		return netip.Prefix{}, false
	}

	opt.Option = append(opt.Option, &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Family:        ecs.Family,
		SourceNetmask: ecs.SourceNetmask,
		SourceScope:   0,
		Address:       ecs.Address,
	})

	return prefix, true
}

func toPrefix(ip net.IP, mask uint8) (netip.Prefix, bool) {
	if addr, ok := netip.AddrFromSlice(ip); ok {
		return netip.PrefixFrom(addr.Unmap(), int(mask)), true
	}
	return netip.Prefix{}, false
}
