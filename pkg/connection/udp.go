package connection

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
)

type UDP struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

// NewUDP dials a new udp connection
func NewUDP(ctx context.Context, uri string) (Connection, error) {
	addr, err := net.ResolveUDPAddr("udp", uri)
	if err != nil {
		return nil, err
	}

	c, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	return &UDP{
		addr: addr,
		conn: c,
	}, nil
}

// Write writes a new message to the websocket connection
func (u *UDP) Write(ctx context.Context, data any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		if err := encoder.Encode(data); err != nil {
			return err
		}

		_, err := u.conn.Write(buf.Bytes())
		return err
	}
}

// Read reads from the websocket connection
func (u *UDP) Read(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		buffer := make([]byte, 1024)
		n, err := u.conn.Read(buffer)
		if err != nil {
			return nil, err
		}
		return buffer[:n], nil
	}
}

// Close closes the websocket connection with status normal closure
func (u *UDP) Close(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return u.conn.Close()
	}
}
