package net

import (
	"testing"

	"github.com/netbirdio/netbird/client/iface/netstack"
)

// TestEmbeddedModeCustomRouting verifies the embedded-mode bypass switch:
// when the engine is embedded into a host app (e.g. sing-box), the control
// plane must keep custom routing enabled even in netstack mode, so sockets
// get their platform bypass (IP_UNICAST_IF / protect / fwmark). Plain
// netstack mode keeps the original behavior.
func TestEmbeddedModeCustomRouting(t *testing.T) {
	t.Setenv(netstack.EnvUseNetstackMode, "true")

	// plain netstack mode: custom routing disabled (original behavior)
	netstack.SetEmbedded(false)
	if !CustomRoutingDisabled() {
		t.Fatal("expected CustomRoutingDisabled=true in plain netstack mode")
	}

	// embedded mode: custom routing stays enabled so the bypass applies
	netstack.SetEmbedded(true)
	if CustomRoutingDisabled() {
		t.Fatal("expected CustomRoutingDisabled=false in embedded mode")
	}

	// restore
	netstack.SetEmbedded(false)
}

// TestEmbeddedFlag verifies the embedded flag defaults to false and toggles.
func TestEmbeddedFlag(t *testing.T) {
	netstack.SetEmbedded(false)
	if netstack.IsEmbedded() {
		t.Fatal("IsEmbedded should default to false")
	}
	netstack.SetEmbedded(true)
	if !netstack.IsEmbedded() {
		t.Fatal("IsEmbedded should be true after SetEmbedded(true)")
	}
	netstack.SetEmbedded(false)
}
