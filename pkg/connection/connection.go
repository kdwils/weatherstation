package connection

import (
	"context"
)

// Connection describes a client or server udp/websocket connection
type Connection interface {
	Write(context.Context, interface{}) error
	Read(context.Context)
	Close(context.Context) error
}
