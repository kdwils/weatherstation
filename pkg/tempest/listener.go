package tempest

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/kdwils/weatherstation/pkg/connection"
)

type Handler func(ctx context.Context, b []byte)

// Listener describes how to listen to weather station device events
type Listener interface {
	Listen(ctx context.Context, eventType ListenEventType, device int) error
	RegisterHandler(e EventType, hs ...Handler) error
}

// EventListener implements the listener
type EventListener struct {
	c        connection.Connection
	Handlers map[string][]Handler
}

// NewEventListener creates a new listener from a connection
func NewEventListener(c connection.Connection) *EventListener {
	return &EventListener{
		c:        c,
		Handlers: make(map[string][]Handler),
	}
}

type requestMessage struct {
	Type   string `json:"type"`
	Device int    `json:"device_id"`
	ID     string `json:"id"`
}

func newRequestMessage(eventType ListenEventType, device int) requestMessage {
	return requestMessage{
		Type:   string(eventType),
		Device: device,
		ID:     uuid.New().String(),
	}
}

// Listen listens for new events and passes them each handler of that event type. Fails silently if the event cannot be unmarshaled.
func (l EventListener) Listen(ctx context.Context, eventType ListenEventType, device int) error {
	defer l.c.Close(ctx)

	if err := l.c.Write(ctx, newRequestMessage(eventType, device)); err != nil {
		return err
	}

	for {
		b, err := l.c.Read(ctx)
		if err != nil {
			return err
		}

		var o Observation
		err = json.Unmarshal(b, &o)
		if err != nil {
			return err
		}

		hs, ok := l.Handlers[o.Type]
		if !ok {
			continue
		}

		for _, h := range hs {
			go h(ctx, b)
		}
	}
}

func (l EventListener) RegisterHandler(eventType EventType, hs ...Handler) error {
	if l.Handlers == nil {
		l.Handlers = make(map[string][]Handler)
	}

	l.Handlers[string(eventType)] = append(l.Handlers[string(eventType)], hs...)
	return nil
}
