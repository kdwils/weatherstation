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

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	var s string
	s += titleStyle.Render("Current Weather Conditions") + "\n\n"

	s += labelStyle.Render("Temperature: ") +
		valueStyle.Render(fmt.Sprintf("%.1f°F", m.observation.TemperatureInFarneheit())) + "\n"

	s += labelStyle.Render("Feels Like: ") +
		valueStyle.Render(fmt.Sprintf("%.1f°F", m.observation.FeelsLikeFarenheit())) + "\n"

	s += labelStyle.Render("Wind: ") +
		valueStyle.Render(fmt.Sprintf("%s at %.1f mph",
			m.observation.WindDirection(),
			m.observation.WindSpeedAverageMPH())) + "\n"

	s += labelStyle.Render("Humidity: ") +
		valueStyle.Render(fmt.Sprintf("%d%%", m.observation.Data.RelativeHumidity)) + "\n"

	s += labelStyle.Render("Pressure: ") +
		valueStyle.Render(fmt.Sprintf("%.1f mb", m.observation.Data.StationPressure)) + "\n"

	s += "\n" + labelStyle.Render("Press ESC to quit")

	return s
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
