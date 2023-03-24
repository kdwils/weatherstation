package tempest

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/kdwils/weatherstation/pkg/connection"
)

// Tempest describes the configuration to connect to a tempest device
type Tempest struct {
	Scheme    string
	Host      string
	Path      string
	Token     string
	eventChan chan connection.Event
}

const (
	wss   = "wss"
	token = "token"
	udp   = "udp"
)

// New creates a new tempest weatherstation configuration
func New(scheme, host, path, token string) *Tempest {
	return &Tempest{
		Scheme:    scheme,
		Host:      host,
		Path:      path,
		Token:     token,
		eventChan: make(chan connection.Event),
	}
}

// NewConnection determines the connection type via the passed tempest scheme. Supports websockets or UDP client connections.
func (t *Tempest) NewConnection(ctx context.Context) (connection.Connection, error) {
	if t.eventChan == nil {
		t.eventChan = make(chan connection.Event)
	}

	u := &url.URL{
		Host:   t.Host,
		Scheme: t.Scheme,
	}

	switch strings.ToLower(t.Scheme) {
	case wss:
		qps := make(url.Values)
		qps.Set(token, t.Token)
		u.RawQuery = qps.Encode()
		u.Scheme = t.Scheme
		u.Path = t.Path
		return connection.NewWebsocketClient(ctx, u.String(), t.eventChan, nil)
	case udp:
		return connection.NewUDPClient(ctx, t.Scheme, u.String(), t.eventChan)
	}

	return nil, fmt.Errorf("unsupported connection protocol: %s", t.Scheme)
}

// NewListener creates a new connection and listener to listen to events on
func (t *Tempest) NewListener(ctx context.Context) (Listener, error) {
	c, err := t.NewConnection(ctx)
	if err != nil {
		return nil, err
	}

	return NewEventListener(c, t.eventChan), nil
}
