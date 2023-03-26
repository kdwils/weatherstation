package connection

import (
	"context"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Websocket implements the connection interface for a websocket connection
type WebsocketProtocol struct {
	conn *websocket.Conn
}

// NewWebsocketClient creates a new websocket client. Opts can be nil.
func NewWebsocketClient(ctx context.Context, addr string, opts *websocket.DialOptions) (*WebsocketProtocol, error) {
	c, _, err := websocket.Dial(ctx, addr, opts)
	if err != nil {
		return nil, err
	}

	return &WebsocketProtocol{
		conn: c,
	}, nil
}

// Write writes a new message to the websocket connection
func (w WebsocketProtocol) Write(ctx context.Context, message interface{}) error {
	return wsjson.Write(ctx, w.conn, message)
}

// Read reads indefinitely from the websocket connection
func (w WebsocketProtocol) Read(ctx context.Context) ([]byte, error) {
	_, b, err := w.conn.Read(ctx)
	return b, err
}

// Close closes the websocket connection with status normal closure
func (w WebsocketProtocol) Close(ctx context.Context) error {
	return w.conn.Close(websocket.StatusNormalClosure, "")
}
