package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/kdwils/weatherstation/pkg/logr"
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
		scheme := viper.GetString("TEMPEST_SCHEME")
		host := viper.GetString("TEMPEST_HOST")
		path := viper.GetString("TEMPEST_URI_PATH")
		token := viper.GetString("TEMPEST_TOKEN")
		device := viper.GetInt("TEMPEST_DEVICE_ID")

		logger, err := zap.NewProduction()
		if err != nil {
			log.Fatal(err)
		}

		t := tempest.New(scheme, host, path, token)
		ctx := logr.WithContext(context.Background(), logger)

		listener, err := t.NewListener(ctx, tempest.ListenGroupStart, device)
		if err != nil {
			logger.Error("could not create listener", zap.Error(err))
			return
		}

		listener.RegisterHandler(tempest.EventConnectionOpened, func(ctx context.Context, b []byte) {
			l, err := logr.FromContext(ctx)
			if err != nil {
				log.Println(err)
				return
			}

			l.Info("connection opened", zap.ByteString("event", b))
		})

		listener.RegisterHandler(tempest.EventObservationTempest, func(ctx context.Context, b []byte) {
			l, err := logr.FromContext(ctx)
			if err != nil {
				log.Println(err)
				return
			}

			var obs tempest.ObservationTempest
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
