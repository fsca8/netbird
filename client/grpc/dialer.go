package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/cenkalti/backoff/v4"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/netbirdio/netbird/util/embeddedroots"
)

// Backoff returns a backoff configuration for gRPC calls
func Backoff(ctx context.Context) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 10 * time.Second
	b.Clock = backoff.SystemClock
	return backoff.WithContext(b, ctx)
}

// CreateConnection creates a gRPC client connection with the appropriate transport options.
// The component parameter specifies the WebSocket proxy component path (e.g., "/management", "/signal").
func CreateConnection(ctx context.Context, addr string, tlsEnabled bool, component string, extraOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	// for js, the outer websocket layer takes care of tls
	if tlsEnabled && runtime.GOOS != "js" {
		certPool, err := x509.SystemCertPool()
		if err != nil || certPool == nil {
			log.Debugf("System cert pool not available; falling back to embedded cert, error: %v", err)
			certPool = embeddedroots.Get()
		}

		transportOption = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			RootCAs: certPool,
		}))
	}

	connCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Control-plane IPv4 pre-resolution: the grpc dns resolver's LookupHost
	// waits for BOTH A and AAAA families. On networks where AAAA queries are
	// answered slowly (or not at all — e.g. home routers stalling the query)
	// every control-plane connect stalls for 0.5-5s, and engine startup makes
	// many of them, burning the whole 60s start budget.
	// The connection target becomes the IPv4 (so the resolver is never
	// touched), but the TLS SNI / :authority MUST stay the original hostname:
	// the management certificate is only valid for the domain, and grpc
	// derives the ServerName from the dial target — dialing the IP directly
	// would fail the handshake with a hostname mismatch. grpc.WithAuthority
	// is the authoritative override for both SNI and the HTTP/2 header.
	dialAddr := addr
	authority := ""
	if host, port, err := net.SplitHostPort(addr); err == nil && net.ParseIP(host) == nil {
		if ips, err := net.DefaultResolver.LookupNetIP(connCtx, "ip4", host); err == nil && len(ips) > 0 {
			dialAddr = net.JoinHostPort(ips[0].String(), port)
			authority = host
		}
	}

	opts := []grpc.DialOption{
		transportOption,
		WithCustomDialer(tlsEnabled, component),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    30 * time.Second,
			Timeout: 10 * time.Second,
		}),
	}
	if authority != "" {
		opts = append(opts, grpc.WithAuthority(authority))
	}
	opts = append(opts, extraOpts...)

	conn, err := grpc.DialContext(connCtx, dialAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial context: %w", err)
	}

	return conn, nil
}
