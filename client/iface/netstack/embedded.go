//go:build !js

package netstack

import "sync/atomic"

// embedded 模式: 引擎被宿主(如 sing-box)嵌入运行。此时引擎的控制面 socket
// (STUN/ICE/WireGuard, 由 client/net 创建)必须绕过宿主的 TUN, 否则探测流量
// 会被宿主 TUN 劫持(源 IP 变 TUN 内部地址, NAT 回程被 TUN 消费, 打洞失败)。
//
// 独立 netstack 模式(NB_USE_NETSTACK_MODE=true, 引擎自管网络/自带 SOCKS5)
// 行为不受影响——旁路机制仅在嵌入模式下启用。
var embedded atomic.Bool

// SetEmbedded 设置嵌入模式。宿主(如 sing-box 的 netbird_integration)在启动引擎前调用。
func SetEmbedded(b bool) {
	embedded.Store(b)
}

// IsEmbedded 返回当前是否处于嵌入模式。
func IsEmbedded() bool {
	return embedded.Load()
}
