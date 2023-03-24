package tempest

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// Tempest describes the configuration to connect to a tempest device
type Tempest struct {
	Scheme    string
	Host      string
	Path      string
	Token     string
	eventChan chan Event
}

const (
	ws    = "ws"
	wss   = "wss"
	token = "token"
)

// New creates a new tempest weatherstation configuration
func New(scheme, host, path, token string) *Tempest {
	return &Tempest{
		Scheme:    scheme,
		Host:      host,
		Path:      path,
		Token:     token,
		eventChan: make(chan Event),
	}
}

// NewConnection determines the connection type via the passed tempest scheme. Supports websockets or TODO: UDP connections.
func (t *Tempest) NewConnection(ctx context.Context) (Connection, error) {
	if t.eventChan == nil {
		t.eventChan = make(chan Event)
	}

	u := &url.URL{
		Host:   t.Host,
		Scheme: t.Scheme,
	}

	// TODO: support udp for local network data aggregation
	switch strings.ToLower(t.Scheme) {
	case ws, wss:
		qps := make(url.Values)
		qps.Set(token, t.Token)
		u.RawQuery = qps.Encode()
		u.Scheme = t.Scheme
		u.Path = t.Path
		return NewWebsocketConnection(ctx, u.String(), t.eventChan, nil)
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
