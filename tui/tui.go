package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/connection"
	"github.com/kdwils/weatherstation/pkg/tempest"
	"golang.org/x/term"
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

func (m *model) terminalSize() (width, height int) {
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	return w, h
}

func (m *model) View() string {
	w, h := m.terminalSize()

	if m.quitting {
		return lipgloss.Place(w, h,
			lipgloss.Center, lipgloss.Center,
			"Thanks for watching the weather!")
	}

	if m.err != nil {
		return lipgloss.Place(w, h,
			lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("Error: %v", m.err))
	}

	if m.observation == nil {
		return lipgloss.Place(w, h,
			lipgloss.Center, lipgloss.Center,
			m.spinner.View()+" Loading weather data...")
	}

	// Style definitions
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		MarginBottom(2).
		Width(w)

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Margin(1, 2).
		Width(35).
		Height(10)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Width(30).
		MarginBottom(1).
		Align(lipgloss.Center)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Width(30).
		MarginTop(1).
		MarginBottom(1).
		Align(lipgloss.Left)

	detailsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Width(28).
		MarginTop(1).
		Align(lipgloss.Left)

	temperature := containerStyle.Render(
		labelStyle.Render("Temperature") +
			valueStyle.Render(fmt.Sprintf("%.1f°F", m.observation.TemperatureInFarneheit())) +
			detailsStyle.Render(fmt.Sprintf("Feels Like: %.1f°F", m.observation.FeelsLikeFarenheit())) +
			detailsStyle.Render(fmt.Sprintf("Wind Chill: %.1f°F", m.observation.Summary.WindChill)) +
			detailsStyle.Render(fmt.Sprintf("Dew Point: %.1f°F", m.observation.DewPointFarenheit())) +
			detailsStyle.Render(fmt.Sprintf("Humidity: %d%%", m.observation.Data.RelativeHumidity)))

	wind := containerStyle.Render(
		labelStyle.Render("Wind") +
			valueStyle.Render(fmt.Sprintf("%s at %.1f mph",
				m.observation.WindDirection(),
				m.observation.WindSpeedAverageMPH())))

	conditions := containerStyle.Render(
		labelStyle.Render("Conditions") +
			valueStyle.Render(m.observation.PrecipitationType()))

	pressure := containerStyle.Render(
		labelStyle.Render("Pressure") +
			valueStyle.Render(fmt.Sprintf("%.1f mb", m.observation.Data.StationPressure)) +
			detailsStyle.Render(fmt.Sprintf("Trend: %s", m.observation.Summary.PressureTrend)))

	lightning := containerStyle.Render(
		labelStyle.Render("Lightning") +
			valueStyle.Render(fmt.Sprintf("%d strikes/hr", m.observation.Summary.StrikeCountOneHour)) +
			detailsStyle.Render(fmt.Sprintf("Last Strike: %.1f miles", m.observation.AverageLightningStrikeDistanceInMiles())) +
			detailsStyle.Render(fmt.Sprintf("3hr Total: %d strikes", m.observation.Summary.StrikeCountThreeHour)))

	solar := containerStyle.Render(
		labelStyle.Render("Solar & UV") +
			valueStyle.Render(fmt.Sprintf("%.1f UV", m.observation.Data.UltraviolentIndex)) +
			detailsStyle.Render(fmt.Sprintf("Solar Radiation: %d W/m²", m.observation.Data.SolarRadiation)) +
			detailsStyle.Render(fmt.Sprintf("Illuminance: %d lux", m.observation.Data.Illuminance)))

	// Join sections horizontally
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, temperature, wind)
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, conditions, pressure)
	row3 := lipgloss.JoinHorizontal(lipgloss.Top, lightning, solar)

	// Center each row
	row1 = lipgloss.Place(w, lipgloss.Height(temperature), lipgloss.Center, lipgloss.Center, row1)
	row2 = lipgloss.Place(w, lipgloss.Height(conditions), lipgloss.Center, lipgloss.Center, row2)
	row3 = lipgloss.Place(w, lipgloss.Height(lightning), lipgloss.Center, lipgloss.Center, row3)

	content := titleStyle.Render("Current Weather Conditions") + "" +
		row1 + "" +
		row2 + "" +
		row3 + "" +
		lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(w).
			Render("Press ESC to quit")

	// Center everything in terminal
	return lipgloss.Place(w, h,
		lipgloss.Center,
		lipgloss.Center,
		content)
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
	return
}

func (m *model) handleObservation(ctx context.Context, b []byte) {
	var obs api.ObservationTempest
	if err := json.Unmarshal(b, &obs); err != nil {
		m.updates <- errMsg{err: err}
		return
	}

	m.updates <- observationMsg{observation: &obs}
}
