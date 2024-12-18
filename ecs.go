// Package example is a CoreDNS plugin that prints "example" to stdout on every packet received.
//
// It serves as an example CoreDNS plugin with numerous code comments.
package ecs

import (
	"context"
	"net"
	"net/netip"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("ecs_remap")

type cidr struct {
	ip   net.IP
	mask uint8
}

// Ecs is an example plugin to show how to write a plugin.
type Ecs struct {
	Next   plugin.Handler
	lookup map[netip.Addr]cidr
}

// setupEdns0Opt will retrieve the EDNS0 OPT or create it if it does not exist.
func setupEdns0Opt(r *dns.Msg) *dns.OPT {
	o := r.IsEdns0()
	if o == nil {
		r.SetEdns0(4096, false)
		o = r.IsEdns0()
	}
	return o
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
func (e *Ecs) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	o := setupEdns0Opt(r)

	var src netip.Addr
	ip := w.RemoteAddr()
	if i, ok := ip.(*net.UDPAddr); ok {
		src, _ = netip.AddrFromSlice(i.IP)
	}
	if i, ok := ip.(*net.TCPAddr); ok {
		src, _ = netip.AddrFromSlice(i.IP)
	}
	var entry cidr
	var ok bool
	if entry, ok = e.lookup[src]; !ok {
		return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
	}

	var family uint16
	if src.Is4() {
		family = 1
	} else {
		family = 2
	}

	var ecs *dns.EDNS0_SUBNET
	for _, s := range o.Option {
		if ecs, ok = s.(*dns.EDNS0_SUBNET); ok {
			break
		}
	}

	// add option if not found
	if ecs == nil {
		ecs = &dns.EDNS0_SUBNET{Code: dns.EDNS0SUBNET}
		o.Option = append(o.Option, ecs)
	}

	ecs.SourceNetmask = entry.mask
	ecs.Address = entry.ip
	ecs.Family = family
	ecs.SourceScope = 0

	return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
}

// Name implements the Handler interface.
func (e *Ecs) Name() string { return "ecs_remap" }
