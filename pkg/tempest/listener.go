package tempest

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/connection"
)

type Handler func(ctx context.Context, b []byte)

// Listener describes how to listen to weather station device events
//
//go:generate mockgen -package mocks -destination mocks/mock_listeniner.go github.com/kdwils/weatherstation/pkg/connection Connection
type Listener interface {
	Listen(ctx context.Context) error
	RegisterHandler(e Event, hs ...Handler) error
}

// EventListener implements the listener
type EventListener struct {
	c           connection.Connection
	Handlers    map[string][]Handler
	ListenGroup ListenGroup
	Device      int
}

// NewEventListener creates a new listener from a connection
func NewEventListener(c connection.Connection, ListenGroup ListenGroup, device int) Listener {
	return &EventListener{
		c:           c,
		Handlers:    make(map[string][]Handler),
		ListenGroup: ListenGroup,
		Device:      device,
	}
}

type RequestMessage struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Device int    `json:"device_id"`
}

func NewRequestMessage(Event ListenGroup, device int) RequestMessage {
	return RequestMessage{
		Type:   string(Event),
		Device: device,
		ID:     uuid.New().String(),
	}
}

// Listen listens for new events and passes them each handler of that event type
func (l *EventListener) Listen(ctx context.Context) error {
	defer l.c.Close(ctx)

	if err := l.c.Write(ctx, NewRequestMessage(l.ListenGroup, l.Device)); err != nil {
		return err
	}

	for {
		b, err := l.c.Read(ctx)
		if err != nil {
			return err
		}

		var o api.Observation
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

// RegisterHandler registers new handlers for a given event type
func (l *EventListener) RegisterHandler(Event Event, hs ...Handler) error {
	if l.Handlers == nil {
		l.Handlers = make(map[string][]Handler)
	}

	l.Handlers[string(Event)] = append(l.Handlers[string(Event)], hs...)
	return nil
}
