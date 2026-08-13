//go:build !linux || android

package internal

// installControlPlaneMarkRule is a no-op on platforms without netlink
// (Windows, macOS, Android, etc.). The control-plane fwmark rule only
// applies to the Linux kernel-TUN embedding scenario; on other platforms
// DisableClientRoutes is never set by the embed layer, so the call site in
// engine.go is unreachable at runtime — this stub only satisfies the linker.
func installControlPlaneMarkRule() error {
	return nil
}
