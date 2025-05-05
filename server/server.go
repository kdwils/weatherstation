package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/tempest"
	"github.com/kdwils/weatherstation/templates"
)

type Server struct {
	listener          tempest.Listener
	latestObservation *api.ObservationTempest
	mu                sync.RWMutex
	clients           map[chan api.ObservationTempest]bool
	events            chan api.ObservationTempest
	clientsMu         sync.RWMutex
	port              int
}

// New creates a new dashboard expecting a configured tempest listener
func New(listener tempest.Listener, port int) *Server {
	s := &Server{
		listener:          listener,
		mu:                sync.RWMutex{},
		latestObservation: &api.ObservationTempest{},
		events:            make(chan api.ObservationTempest),
		clients:           make(map[chan api.ObservationTempest]bool),
		port:              port,
	}

	// Start global event handler
	go s.handleEvents()

	// Register global observation handler
	s.listener.RegisterHandler(tempest.EventObservationTempest, s.handleObservation)

	// Start listener in background
	go func() {
		if err := s.listener.Listen(context.Background()); err != nil {
			log.Printf("global listener error: %v", err)
		}
	}()

	return s
}

func (s *Server) handleObservation(ctx context.Context, b []byte) {
	var obs api.ObservationTempest
	if err := json.Unmarshal(b, &obs); err != nil {
		log.Printf("error unmarshaling observation: %v", err)
		return
	}

	s.mu.Lock()
	s.latestObservation = &obs
	s.mu.Unlock()

	s.events <- obs
}

func (s *Server) handleEvents() {
	for obs := range s.events {
		s.clientsMu.RLock()
		for clientChan := range s.clients {
			select {
			case clientChan <- obs:
			default:
			}
		}
		s.clientsMu.RUnlock()
	}
}

func (s *Server) HandleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := templates.Dashboard(s.latestObservation, s.port).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) HandleEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		clientChan := make(chan api.ObservationTempest, 1)

		s.clientsMu.Lock()
		s.clients[clientChan] = true
		s.clientsMu.Unlock()

		defer func() {
			s.clientsMu.Lock()
			delete(s.clients, clientChan)
			close(clientChan)
			s.clientsMu.Unlock()
		}()

		for {
			select {
			case <-r.Context().Done():
				return
			case obs := <-clientChan:
				var buf bytes.Buffer
				if err := templates.Dashboard(&obs, s.port).Render(r.Context(), &buf); err != nil {
					log.Printf("error rendering template: %v", err)
					continue
				}

				_, err := fmt.Fprintf(w, "data: %s\n\n", buf.String())
				if err != nil {
					log.Printf("error writing SSE data: %v", err)
					return
				}

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}
	}
}
