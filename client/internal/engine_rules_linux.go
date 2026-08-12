//go:build linux && !android

package internal

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	nbnet "github.com/netbirdio/netbird/client/net"
)

// installControlPlaneMarkRule installs the routing rule that sends the
// engine's control-plane traffic (ICE/STUN/WireGuard probe sockets, which are
// marked with nbnet.ControlPlaneMark) straight to the main routing table,
// bypassing any TUN-based proxy's default route.
//
// Why this is needed: when netbird is embedded into sing-box (or any other
// TUN-based proxy) with auto_route, the proxy installs a more specific
// default route (0.0.0.0/1 via its TUN interface). Control-plane packets
// without a matching rule then pick that TUN as their egress, so the kernel
// selects the TUN's interface IP as the source address. NAT on the LAN
// gateway maps that to a TUN-internal address; the return path is consumed
// by the TUN and never reaches the engine's socket — ICE hole punching fails
// even though both peers exchange correct server-reflexive candidates.
//
// Normally routemanager's SetupRouting installs the
// fwmark-0x1BD00 -> NetbirdVPNTable rule for this mark. But when client
// route management is disabled (DisableClientRoutes, which is the case when
// netbird is embedded into sing-box so the two engines don't fight over the
// default route), SetupRouting is skipped and the mark ends up unrouted.
// This function installs the rule in that mode, targeting the main table so
// marked packets egress via the physical interface with the physical source
// IP — making the NAT mapping point at the physical host, which lets the
// return path reach the engine socket normally.
func installControlPlaneMarkRule() error {
	for _, family := range []int{netlink.FAMILY_V4, netlink.FAMILY_V6} {
		rule := netlink.NewRule()
		rule.Table = unix.RT_TABLE_MAIN
		rule.Mark = nbnet.ControlPlaneMark
		rule.Family = family
		rule.Priority = 110
		if err := netlink.RuleAdd(rule); err != nil && !errors.Is(err, unix.EEXIST) {
			return fmt.Errorf("add control-plane mark rule (family %d): %w", family, err)
		}
	}
	log.Debugf("installed control-plane routing rule (fwmark %#x -> main)", nbnet.ControlPlaneMark)
	return nil
}
