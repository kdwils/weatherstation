package tempest

import (
	"context"

	"github.com/kdwils/weatherstation/pkg/connection"
)

// Config describes the configuration to connect to a tempest device
type Config struct {
	Scheme string
	Host   string
	Path   string
	Token  string
}

// New creates a new tempest weatherstation configuration
func NewConfig(scheme, host, path, token string) Config {
	return Config{
		Scheme: scheme,
		Host:   host,
		Path:   path,
		Token:  token,
	}
}

// NewListener creates a new connection and listener to listen to events on
func (c Config) NewEventListener(ctx context.Context, ListenGroup ListenGroup, device int) (Listener, error) {
	conn, err := connection.NewConnection(ctx, c.Scheme, c.Host, c.Path, c.Token)
	if err != nil {
		return nil, err
	}

	return NewEventListener(conn, ListenGroup, device), nil
}
