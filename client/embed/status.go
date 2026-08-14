package embed

import (
	"time"
)

// PeerStatusInfo is a wire-ready view of one peer's current connection
// state. It is exported (not peer.State) so cross-module consumers such as
// sing-box's netbird integration can serialize it without importing
// netbird's internal packages.
type PeerStatusInfo struct {
	IP             string `json:"ip"`
	FQDN           string `json:"fqdn"`
	ConnStatus     string `json:"conn_status"`
	Relayed        bool   `json:"relayed"`
	LocalEndpoint  string `json:"local_endpoint"`
	RemoteEndpoint string `json:"remote_endpoint"`
	RelayServer    string `json:"relay_server"`
	LastHandshake  string `json:"last_handshake"`
	LatencyMs      int64  `json:"latency_ms"`
}

// PeerStatuses returns the current connection state of every peer, or nil
// when the status recorder is unavailable.
func (c *Client) PeerStatuses() []PeerStatusInfo {
	full, err := c.Status()
	if err != nil {
		return nil
	}
	if len(full.Peers) == 0 {
		return []PeerStatusInfo{}
	}
	out := make([]PeerStatusInfo, 0, len(full.Peers))
	for _, p := range full.Peers {
		lastHandshake := ""
		if !p.LastWireguardHandshake.IsZero() {
			lastHandshake = p.LastWireguardHandshake.UTC().Format(time.RFC3339)
		}
		out = append(out, PeerStatusInfo{
			IP:             p.IP,
			FQDN:           p.FQDN,
			ConnStatus:     p.ConnStatus.String(),
			Relayed:        p.Relayed,
			LocalEndpoint:  p.LocalIceCandidateEndpoint,
			RemoteEndpoint: p.RemoteIceCandidateEndpoint,
			RelayServer:    p.RelayServerAddress,
			LastHandshake:  lastHandshake,
			LatencyMs:      p.Latency.Milliseconds(),
		})
	}
	return out
}
