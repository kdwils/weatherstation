package cmd

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/connection"
	"github.com/kdwils/weatherstation/pkg/tempest"
	"github.com/kdwils/weatherstation/templates"
	"github.com/spf13/cobra"
)

var (
	currentObservation *api.ObservationTempest
	mu                 sync.RWMutex
	clients            = make(map[chan *api.ObservationTempest]bool)
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the weather station dashboard",
	Long:  `Serve the weather station dashboard`,
	Run: func(cmd *cobra.Command, args []string) {
		scheme := getEnvOrDefault("WEATHERSTATION_TEMPEST_SCHEME", "wss")
		host := getEnvOrDefault("WEATHERSTATION_TEMPEST_HOST", "")
		path := getEnvOrDefault("WEATHERSTATION_TEMPEST_PATH", "")
		token := getEnvOrDefault("WEATHERSTATION_TEMPEST_TOKEN", "")
		device := getEnvIntOrDefault("WEATHERSTATION_TEMPEST_DEVICE_ID", 0)

		ctx := context.Background()
		conn, err := connection.NewConnection(ctx, scheme, host, path, token)
		if err != nil {
			log.Fatal(err)
		}

		listener := tempest.NewEventListener(conn, tempest.ListenGroupStart, device)

		// Handle weather station updates
		listener.RegisterHandler(tempest.EventObservationTempest, func(ctx context.Context, b []byte) {
			var obs api.ObservationTempest
			if err := json.Unmarshal(b, &obs); err != nil {
				log.Printf("error unmarshaling observation: %v", err)
				return
			}

			log.Printf("received observation: %+v", obs)

			mu.Lock()
			currentObservation = &obs
			// Broadcast to all connected clients
			for client := range clients {
				client <- &obs
			}
			mu.Unlock()
		})

		// Start listening for weather updates
		go func() {
			if err := listener.Listen(ctx); err != nil {
				log.Fatal(err)
			}
		}()

		// HTTP handlers
		http.HandleFunc("/", handleHome)
		http.HandleFunc("/events", handleEvents)
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

		log.Fatal(http.ListenAndServe(":8080", nil))
	},
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	err := templates.Dashboard(currentObservation).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create channel for this client
	updates := make(chan *api.ObservationTempest)
	mu.Lock()
	clients[updates] = true
	mu.Unlock()

	// Clean up when client disconnects
	defer func() {
		mu.Lock()
		delete(clients, updates)
		close(updates)
		mu.Unlock()
	}()

	// Send updates to client
	for obs := range updates {
		if err := templates.Dashboard(obs).Render(r.Context(), w); err != nil {
			return
		}
		flusher.Flush()
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
