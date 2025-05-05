package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/tempest"
	"github.com/kdwils/weatherstation/templates"
)

type Server struct {
	listener tempest.Listener
}

// New creates a new dashboard expecting a configured tempest listener
func New(listener tempest.Listener) *Server {
	return &Server{
		listener: listener,
	}
}

// HandleHome handles the home page /
func (s *Server) HandleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := templates.Dashboard(&api.ObservationTempest{}).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// HandleEvents handles the SSE events stream /events
func (s *Server) HandleEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		events := make(chan api.ObservationTempest)

		s.listener.RegisterHandler(tempest.EventObservationTempest, func(ctx context.Context, b []byte) {
			var obs api.ObservationTempest
			if err := json.Unmarshal(b, &obs); err != nil {
				log.Printf("error unmarshaling observation: %v", err)
				return
			}

			log.Printf("received observation: %+v", obs)
			events <- obs
		})

		go func() {
			if err := s.listener.Listen(r.Context()); err != nil {
				log.Printf("listener error: %v", err)
			}
		}()

		for {
			select {
			case <-r.Context().Done():
				return
			case obs := <-events:
				var buf bytes.Buffer
				if err := templates.Dashboard(&obs).Render(r.Context(), &buf); err != nil {
					log.Printf("error rendering template: %v", err)
					continue
				}

				_, err := fmt.Fprintf(w, "data: %s\n\n", buf.String())
				if err != nil {
					log.Printf("error writing SSE data: %v", err)
					continue
				}

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}
	}
}
