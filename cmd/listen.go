package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/kdwils/weatherstation/logging"
	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/connection"
	"github.com/kdwils/weatherstation/pkg/tempest"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "example: listen on tempest events",
	Long:  `example: listen on tempest events`,
	Run: func(cmd *cobra.Command, args []string) {
		scheme := viper.GetString("WEATHERSTATION_TEMPEST_SCHEME")
		host := viper.GetString("WEATHERSTATION_TEMPEST_HOST")
		path := viper.GetString("WEATHERSTATION_TEMPEST_PATH")
		token := viper.GetString("WEATHERSTATION_TEMPEST_TOKEN")
		device := viper.GetInt("WEATHERSTATION_TEMPEST_DEVICE_ID")

		logger, err := zap.NewProduction()
		if err != nil {
			log.Fatal(err)
		}

		ctx := logging.WithContext(context.Background(), logger)

		conn, err := connection.NewConnection(ctx, scheme, host, path, token)
		if err != nil {
			log.Fatal(err)
		}

		listener := tempest.NewEventListener(conn, tempest.ListenGroupStart, device)

		listener.RegisterHandler(tempest.EventConnectionOpened, func(ctx context.Context, b []byte) {
			l, err := logging.FromContext(ctx)
			if err != nil {
				log.Println(err)
				return
			}

			l.Info("connection opened", zap.ByteString("event", b))
		})

		listener.RegisterHandler(tempest.EventObservationTempest, func(ctx context.Context, b []byte) {
			l, err := logging.FromContext(ctx)
			if err != nil {
				log.Println(err)
				return
			}

			var obs api.ObservationTempest
			err = json.Unmarshal(b, &obs)
			if err != nil {
				l.Error("could not parse event", zap.Error(err))
				return
			}

			l.Info("received observation", zap.Any("observation", obs))
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
		logger.Info("received signal to terminate")

	},
}

func init() {
	rootCmd.AddCommand(listenCmd)
}
