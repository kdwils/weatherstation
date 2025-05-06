package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/connection"
	"github.com/kdwils/weatherstation/pkg/tempest"
)

type model struct {
	listener    tempest.Listener
	observation *api.ObservationTempest
	spinner     spinner.Model
	err         error
	quitting    bool
	updates     chan tea.Msg // Add channel for updates
}

func InitialModel(conn connection.Connection, device int) *model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &model{
		listener: tempest.NewEventListener(conn, tempest.ListenGroupStart, device),
		spinner:  s,
		updates:  make(chan tea.Msg),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.waitForUpdate,
	)
}

// Add a new command to wait for updates
func (m model) waitForUpdate() tea.Msg {
	log.Println("wwaiting for updates...")
	return <-m.updates
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			m.quitting = true
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case observationMsg:
		m.observation = msg.observation
		return m, m.waitForUpdate

	case errMsg:
		m.err = msg.err
		return m, m.waitForUpdate
	}

	return m, nil
}

func (m *model) View() string {
	if m.quitting {
		return "Thanks for watching the weather!\n"
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.observation == nil {
		return fmt.Sprintf("\n\n   %s Loading weather data...\n\n", m.spinner.View())
	}

	// Style definitions
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Margin(0, 1).
		Background(lipgloss.Color("0")).
		Width(30)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))

	detailsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginLeft(2)

	// Build sections
	temperature := containerStyle.Render(
		labelStyle.Render("Temperature\n") +
			valueStyle.Render(fmt.Sprintf("%.1f°F\n", m.observation.TemperatureInFarneheit())) +
			detailsStyle.Render(fmt.Sprintf("Feels Like: %.1f°F\n", m.observation.FeelsLikeFarenheit())) +
			detailsStyle.Render(fmt.Sprintf("Wind Chill: %.1f°F\n", m.observation.Summary.WindChill)) +
			detailsStyle.Render(fmt.Sprintf("Dew Point: %.1f°F", m.observation.DewPointFarenheit())))

	wind := containerStyle.Render(
		labelStyle.Render("Wind\n") +
			valueStyle.Render(fmt.Sprintf("%s at %.1f mph\n",
				m.observation.WindDirection(),
				m.observation.WindSpeedAverageMPH())))

	humidity := containerStyle.Render(
		labelStyle.Render("Humidity\n") +
			valueStyle.Render(fmt.Sprintf("%d%%", m.observation.Data.RelativeHumidity)))

	conditions := containerStyle.Render(
		labelStyle.Render("Conditions\n") +
			valueStyle.Render(m.observation.PrecipitationType()))

	pressure := containerStyle.Render(
		labelStyle.Render("Pressure\n") +
			valueStyle.Render(fmt.Sprintf("%.1f mb\n", m.observation.Data.StationPressure)) +
			detailsStyle.Render(fmt.Sprintf("Trend: %s", m.observation.Summary.PressureTrend)))

	lightning := containerStyle.Render(
		labelStyle.Render("Lightning\n") +
			valueStyle.Render(fmt.Sprintf("%d strikes/hr\n", m.observation.Summary.StrikeCountOneHour)) +
			detailsStyle.Render(fmt.Sprintf("Last Strike: %.1f miles\n", m.observation.AverageLightningStrikeDistanceInMiles())) +
			detailsStyle.Render(fmt.Sprintf("3hr Total: %d strikes", m.observation.Summary.StrikeCountThreeHour)))

	solar := containerStyle.Render(
		labelStyle.Render("Solar & UV\n") +
			valueStyle.Render(fmt.Sprintf("%.1f UV\n", m.observation.Data.UltraviolentIndex)) +
			detailsStyle.Render(fmt.Sprintf("Solar Radiation: %d W/m²\n", m.observation.Data.SolarRadiation)) +
			detailsStyle.Render(fmt.Sprintf("Illuminance: %d lux", m.observation.Data.Illuminance)))

	// Layout sections in a grid
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, temperature, wind, humidity)
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, conditions, pressure)
	row3 := lipgloss.JoinHorizontal(lipgloss.Top, lightning, solar)

	return titleStyle.Render("Current Weather Conditions") + "\n\n" +
		row1 + "\n" +
		row2 + "\n" +
		row3 + "\n\n" +
		labelStyle.Render("Press ESC to quit")
}

type observationMsg struct {
	observation *api.ObservationTempest
}

type errMsg struct {
	err error
}

func (m model) StartListener() {
	m.listener.RegisterHandler(tempest.EventObservationTempest, m.handleObservation)
	m.listener.Listen(context.Background())
}

func (m *model) handleObservation(ctx context.Context, b []byte) {
	var obs api.ObservationTempest
	if err := json.Unmarshal(b, &obs); err != nil {
		m.updates <- errMsg{err: err}
		return
	}

	m.updates <- observationMsg{observation: &obs}
}
