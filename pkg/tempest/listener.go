package tempest

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/kdwils/weatherstation/pkg/connection"
)

type Handler func(ctx context.Context, e connection.Event)

// Listener describes how to listen to weather station device events
type Listener interface {
	Listen(ctx context.Context, eventType ListenEventType, device int) error
	RegisterHandler(e EventType, hs ...Handler) error
	Stop()
}

// EventListener implements the listener
type EventListener struct {
	c         connection.Connection
	Handlers  map[string][]Handler
	EventChan chan connection.Event
	stopChan  chan bool
}

// NewEventListener creates a new listener from a connection
func NewEventListener(c connection.Connection, eventChan chan connection.Event) *EventListener {
	return &EventListener{
		c:         c,
		Handlers:  make(map[string][]Handler),
		EventChan: eventChan,
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

	go l.c.Read(ctx)

	for {
		select {
		case <-l.stopChan:
			return nil
		case e := <-l.EventChan:
			var o Observation
			err := json.Unmarshal(e.Bytes, &o)
			if err != nil {
				continue
			}

			hs, ok := l.Handlers[o.Type]
			if !ok {
				continue
			}

			for _, h := range hs {
				go h(ctx, e)
			}
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

func (l EventListener) Stop() {
	l.stopChan <- true
}
