package tempest

import (
	"context"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Connection describes a connection to a tempest weather device
type Connection interface {
	Write(context.Context, interface{}) error
	Read(context.Context)
	Close(context.Context) error
}

// Websocket implements the connection interface for a websocket connection
type Websocket struct {
	conn *websocket.Conn
	ec   chan Event
}

// NewWebsocketConnection creates a new websocket connection. Opts can be nil.
func NewWebsocketConnection(ctx context.Context, addr string, eventChan chan Event, opts *websocket.DialOptions) (Connection, error) {
	c, _, err := websocket.Dial(ctx, addr, opts)
	if err != nil {
		return nil, err
	}

	return Websocket{
		conn: c,
		ec:   eventChan,
	}, nil
}

// Write writes a new message to the websocket connection
func (w Websocket) Write(ctx context.Context, message interface{}) error {
	return wsjson.Write(ctx, w.conn, message)
}

// Read reads indefinitely from the websocket connection
func (w Websocket) Read(ctx context.Context) {
	defer w.Close(ctx)
	for {
		_, b, err := w.conn.Read(ctx)
		w.ec <- NewEvent(b, err)
	}
}

// Close closes the websocket connection with status normal closure
func (w Websocket) Close(ctx context.Context) error {
	return w.conn.Close(websocket.StatusNormalClosure, "")
}
