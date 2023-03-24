package connection

// Event describes an event read from the connection
type Event struct {
	Bytes []byte
	Err   error
}

// NewEvent creates a new event
func NewEvent(b []byte, err error) Event {
	return Event{
		Bytes: b,
		Err:   err,
	}
}
