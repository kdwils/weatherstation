package connection

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/coder/websocket"
)

// Connection is a generic interface for a connection to a tempest device
type Connection interface {
	Write(context.Context, any) error
	Read(context.Context) ([]byte, error)
	Close(context.Context, ...websocket.StatusCode) error
}

const (
	wss = "wss"
	udp = "udp"
)

// NewConnection determines the connection type via the passed tempest scheme. Supports websockets or UDP client connections.
func NewConnection(ctx context.Context, scheme, host, path, token string) (Connection, error) {
	u := &url.URL{
		Host:   host,
		Scheme: scheme,
	}

	switch strings.ToLower(scheme) {
	case wss:
		qps := make(url.Values)
		qps.Set("token", token)
		u.RawQuery = qps.Encode()
		u.Path = path

		return NewWebsocket(ctx, u.String(), nil)
	case udp:
		return NewUDP(ctx, u.String())
	}

	return nil, fmt.Errorf("unsupported connection protocol: %s", scheme)
}
