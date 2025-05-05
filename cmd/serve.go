package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/connection"
	"github.com/kdwils/weatherstation/pkg/tempest"
	"github.com/kdwils/weatherstation/server"
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
		serverPort := getEnvIntOrDefault("WEATHERSTATION_SERVER_PORT", 8080)

		ctx := context.Background()
		conn, err := connection.NewConnection(ctx, scheme, host, path, token)
		if err != nil {
			log.Fatal(err)
		}

		listener := tempest.NewEventListener(conn, tempest.ListenGroupStart, device)

		srv := server.New(listener)

		http.HandleFunc("/", server.CORSMiddleware(srv.HandleHome()))
		http.HandleFunc("/events", server.CORSMiddleware(srv.HandleEvents()))
		fs := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))

		http.Handle("/static/", server.CORSMiddleware(fs))

		log.Printf("Serving on port %d", serverPort)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil))
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
