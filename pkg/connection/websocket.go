package connection

import (
	"context"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Websocket satisfies the connection interface for a websocket connection
type Websocket struct {
	conn *websocket.Conn
}

// NewWebsocket dials a new websocket connection. Opts can be nil.
func NewWebsocket(ctx context.Context, addr string, opts *websocket.DialOptions) (Connection, error) {
	c, _, err := websocket.Dial(ctx, addr, opts)
	if err != nil {
		return nil, err
	}

	return &Websocket{
		conn: c,
	}, nil
}

// Write writes a new message to the websocket connection
func (w Websocket) Write(ctx context.Context, data any) error {
	return wsjson.Write(ctx, w.conn, data)
}

// Read reads from the websocket connection
func (w Websocket) Read(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		_, b, err := w.conn.Read(ctx)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
}

// Close closes the websocket connection with status normal closure
func (w Websocket) Close(ctx context.Context) error {
	return w.conn.Close(websocket.StatusNormalClosure, "")
}
