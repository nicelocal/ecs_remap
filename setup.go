package ecs

import (
	"net"
	"net/netip"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("ecs_remap", setup) }

func setup(c *caddy.Controller) error {
	c.Next()

	if !c.NextBlock() {
		return plugin.Error("ecs_remap", c.SyntaxErr("{"))
	}

	lookup := make(map[netip.Addr]cidr)
	for c.Val() != "}" {
		src := c.Val()
		if !c.NextArg() {
			return plugin.Error("ecs_remap", c.ArgErr())
		}
		dst := c.Val()
		if !c.NextLine() {
			return plugin.Error("ecs_remap", c.ArgErr())
		}

		srcAddr, err := netip.ParseAddr(src)
		if err != nil {
			return plugin.Error("ecs_remap", err)
		}

		_, dstNet, err := net.ParseCIDR(dst)
		if err != nil {
			return plugin.Error("ecs_remap", err)
		}
		sz, _ := dstNet.Mask.Size()

		dstIs4 := len(dstNet.IP) == net.IPv4len
		if srcAddr.Is4() != dstIs4 {
			return plugin.Error("ecs_remap", c.Errf("Both LHS and RHS of map must be of the same family: %s cannot map to %s", src, dst))
		}

		lookup[srcAddr] = cidr{
			dstNet.IP,
			uint8(sz),
		}
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return &Ecs{Next: next, lookup: lookup}
	})

	// All OK, return a nil error.
	return nil
}
