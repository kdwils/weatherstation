package tempest

import (
	"context"
	"encoding/json"
	"net"

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
type WebsocketProtocol struct {
	conn *websocket.Conn
	ec   chan Event
}

// NewWebsocketConnection creates a new websocket connection. Opts can be nil.
func NewWebsocketConnection(ctx context.Context, addr string, eventChan chan Event, opts *websocket.DialOptions) (Connection, error) {
	c, _, err := websocket.Dial(ctx, addr, opts)
	if err != nil {
		return nil, err
	}

	return WebsocketProtocol{
		conn: c,
		ec:   eventChan,
	}, nil
}

// Write writes a new message to the websocket connection
func (w WebsocketProtocol) Write(ctx context.Context, message interface{}) error {
	return wsjson.Write(ctx, w.conn, message)
}

// Read reads indefinitely from the websocket connection
func (w WebsocketProtocol) Read(ctx context.Context) {
	defer w.Close(ctx)
	for {
		_, b, err := w.conn.Read(ctx)
		w.ec <- NewEvent(b, err)
	}
}

// Close closes the websocket connection with status normal closure
func (w WebsocketProtocol) Close(ctx context.Context) error {
	return w.conn.Close(websocket.StatusNormalClosure, "")
}

type UserDatagramProtocol struct {
	conn *net.UDPConn
	addr *net.UDPAddr
	ec   chan Event
}

func NewUDPConnection(ctx context.Context, scheme, uri string, eventChan chan Event) (Connection, error) {
	addr, err := net.ResolveUDPAddr(scheme, uri)
	if err != nil {
		return nil, err
	}

	c, err := net.DialUDP(scheme, nil, addr)
	if err != nil {
		return nil, err
	}

	return UserDatagramProtocol{
		addr: addr,
		conn: c,
	}, nil
}

// Write writes a new message to the websocket connection
func (u UserDatagramProtocol) Write(ctx context.Context, message interface{}) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = u.conn.WriteTo(b, u.addr)
	return err
}

// Read reads indefinitely from the websocket connection
func (u UserDatagramProtocol) Read(ctx context.Context) {
	defer u.Close(ctx)
	for {
		buf := make([]byte, 2048)
		_, _, err := u.conn.ReadFrom(buf)
		u.ec <- NewEvent(buf, err)
	}
}

// Close closes the websocket connection with status normal closure
func (u UserDatagramProtocol) Close(_ context.Context) error {
	return u.conn.Close()
}
