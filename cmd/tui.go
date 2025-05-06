package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kdwils/weatherstation/pkg/connection"
	"github.com/kdwils/weatherstation/tui"
	"github.com/spf13/cobra"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Display weather data in a terminal UI",
	Long:  `Display weather data in a terminal user interface using Bubble Tea`,
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

		m := tui.InitialModel(conn, device)

		go m.StartListener()

		p := tea.NewProgram(
			m,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running program: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
