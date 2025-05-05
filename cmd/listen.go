package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/connection"
	"github.com/kdwils/weatherstation/pkg/tempest"
	"github.com/spf13/cobra"
)

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "example: listen on tempest events",
	Long:  `example: listen on tempest events`,
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

		listener.RegisterHandler(tempest.EventConnectionOpened, func(ctx context.Context, b []byte) {
			log.Printf("connection opened: %s", b)
		})

		listener.RegisterHandler(tempest.EventObservationTempest, func(ctx context.Context, b []byte) {
			var obs api.ObservationTempest
			err = json.Unmarshal(b, &obs)
			if err != nil {
				log.Fatal(err)
				return
			}

			log.Printf("received observation: %+v", obs)
		})

		go func(ctx context.Context, device int) {
			err := listener.Listen(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}(ctx, device)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		<-c
		log.Println("received signal to terminate")
	},
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}
	return value
}

func init() {
	rootCmd.AddCommand(listenCmd)
}
