package connection

import (
	"context"
	"encoding/json"
	"net"
)

type UserDatagramProtocol struct {
	conn *net.UDPConn
	addr *net.UDPAddr
	ec   chan Event
}

func NewUDPClient(ctx context.Context, scheme, uri string, eventChan chan Event) (Connection, error) {
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
