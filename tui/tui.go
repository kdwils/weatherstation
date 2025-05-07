package tui

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"

	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/connection"
	"github.com/kdwils/weatherstation/pkg/tempest"
)

type model struct {
	listener         tempest.Listener
	observation      *api.ObservationTempest
	spinner          spinner.Model
	err              error
	quitting         bool
	lightningCounter int
	updates          chan tea.Msg // Add channel for updates
	width            int
	height           int
	tempHistory      []float64
	windSpeedHistory []float64
	pressureHistory  []float64
}

func InitialModel(conn connection.Connection, device int) *model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &model{
		listener:    tempest.NewEventListener(conn, tempest.ListenGroupStart, device),
		spinner:     s,
		updates:     make(chan tea.Msg),
		tempHistory: make([]float64, 0, 30), // Keep last 30 readings
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

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case observationMsg:
		m.observation = msg.observation
		if m.lightningCounter > 0 {
			m.lightningCounter--
		}
		m.tempHistory = append(m.tempHistory, m.observation.TemperatureInFarneheit())
		if len(m.tempHistory) > 30 {
			m.tempHistory = m.tempHistory[1:]
		}
		m.windSpeedHistory = append(m.windSpeedHistory, m.observation.WindSpeedAverageMPH())
		if len(m.windSpeedHistory) > 30 {
			m.windSpeedHistory = m.windSpeedHistory[1:]
		}
		m.pressureHistory = append(m.pressureHistory, m.observation.Data.StationPressure)
		if len(m.pressureHistory) > 30 {
			m.pressureHistory = m.pressureHistory[1:]
		}
		return m, m.waitForUpdate

	case lightningStrikeMsg:
		m.lightningCounter += 2
		return m, m.waitForUpdate

	case errMsg:
		m.err = msg.err
		return m, m.waitForUpdate
	}

	return m, nil
}

func (m *model) View() string {
	if m.observation == nil {
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center,
			lipgloss.Center,
			m.spinner.View())
	}
	mainContainerStyle := lipgloss.NewStyle()

	quadrantWidth := (m.width - 8) / 2
	quadrantHeight := (m.height - 8) / 2

	dataQuadrantStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1).
		Margin(1).
		Width(quadrantWidth - 1).
		Height(quadrantHeight - 1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Width(quadrantWidth - 4).
		PaddingBottom(2)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Width(quadrantWidth - 8).
		Align(lipgloss.Left)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Width(quadrantWidth - 8).
		Align(lipgloss.Left)

	detailsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Width(quadrantWidth - 8).
		Align(lipgloss.Left)

	contentStyle := lipgloss.NewStyle().
		Width(quadrantWidth - 4).
		AlignHorizontal(lipgloss.Left).
		PaddingLeft(2).
		PaddingRight(2)

	sectionStyle := lipgloss.NewStyle().
		MarginBottom(1).
		MarginRight(2).
		Width((quadrantWidth) / 3)

	graphStyle := lipgloss.NewStyle().
		Width(quadrantWidth - 8).
		Height(quadrantHeight - 8).
		AlignVertical(lipgloss.Bottom).
		MarginBottom(0).
		PaddingBottom(0)

	temperatureSection := sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			labelStyle.Render("Temperature"),
			valueStyle.Render(fmt.Sprintf("%.1f°F", m.observation.TemperatureInFarneheit())),
			detailsStyle.Render(fmt.Sprintf("Feels Like: %.1f°F", m.observation.FeelsLikeFarenheit())),
			detailsStyle.Render(fmt.Sprintf("Wind Chill: %.1f°F", m.observation.Summary.WindChill)),
			detailsStyle.Render(fmt.Sprintf("Dew Point: %.1f°F", m.observation.DewPointFarenheit())),
		),
	)

	windSection := sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			labelStyle.Render("Wind"),
			valueStyle.Render(fmt.Sprintf("%s at %.1f mph", m.observation.WindDirection(), m.observation.WindSpeedAverageMPH())),
			detailsStyle.Render(fmt.Sprintf("Gust: %.1f mph", m.observation.WindSpeedGustMPH())),
		),
	)

	precipitationSection := sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			labelStyle.Render("Preciptiation"),
			valueStyle.Render(m.observation.PrecipitationType()),
			detailsStyle.Render(fmt.Sprintf("Today: %.2fin", m.observation.RainfallInInches())),
			detailsStyle.Render(fmt.Sprintf("Rainfall Hourly: %.2fin/hr", m.observation.Summary.PrecipTotalOneHour)),
			detailsStyle.Render(fmt.Sprintf("Total Yesterday: %.2fin", m.observation.Summary.PrecipAccumLocalYesterdayFinal)),
		),
	)

	pressureSection := sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			labelStyle.Render("Pressure"),
			valueStyle.Render(fmt.Sprintf("%.1f mb", m.observation.Data.StationPressure)),
			detailsStyle.Render(fmt.Sprintf("Trend: %s", m.observation.Summary.PressureTrend)),
		),
	)

	ultraVioletIndexSection := sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			labelStyle.Render("UV Index"),
			valueStyle.Render(fmt.Sprintf("%0.2f", m.observation.Data.UltraviolentIndex)),
		),
	)

	illuminanceSection := sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			labelStyle.Render("Illuminance"),
			valueStyle.Render(fmt.Sprintf("%d lux", m.observation.Data.Illuminance)),
		),
	)
	solarRadiationSection := sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			labelStyle.Render("Solar Radiation"),
			valueStyle.Render(fmt.Sprintf("%d W/m²", m.observation.Data.SolarRadiation)),
		),
	)

	lightningSection := sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			labelStyle.Render("Lightning Strikes"),
			valueStyle.Render(fmt.Sprintf("%d strikes/hr", m.observation.Summary.StrikeCountOneHour)),
			detailsStyle.Render(fmt.Sprintf("Last Strike: %.1f miles", m.observation.AverageLightningStrikeDistanceInMiles())),
			detailsStyle.Render(fmt.Sprintf("3hr Total: %d strikes", m.observation.Summary.StrikeCountThreeHour)),
		),
	)

	reportInterval := sectionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			labelStyle.Render("Report Interval"),
			valueStyle.Render(fmt.Sprintf("%d min", m.observation.Data.ReportInterval)),
		),
	)

	currentRows := []string{
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			temperatureSection,
			precipitationSection,
			windSection,
		),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			pressureSection,
			lightningSection,
			illuminanceSection,
		),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			solarRadiationSection,
			ultraVioletIndexSection,
			reportInterval,
		),
	}

	currentConditionsContent := dataQuadrantStyle.Render(
		contentStyle.Render(
			titleStyle.Render("Current Conditions"),
			lipgloss.JoinVertical(lipgloss.Left,
				currentRows...,
			),
		),
	)

	temperatureGraph := dataQuadrantStyle.Render(
		titleStyle.Render("Temperature Trend") +
			graphStyle.Render(
				m.renderTemperatureGraph(quadrantWidth-4, quadrantHeight-4),
			),
	)

	pressureGraph := dataQuadrantStyle.Render(
		titleStyle.Render("Pressure Trend") +
			graphStyle.Render(
				m.renderPressureGraph(quadrantWidth-4, quadrantHeight-4),
			),
	)

	windGraph := dataQuadrantStyle.Render(
		titleStyle.Render("Wind Speed Trend") +
			graphStyle.Render(
				m.renderWindGraph(quadrantWidth-4, quadrantHeight-4),
			),
	)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		currentConditionsContent,
		temperatureGraph,
	)

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
		pressureGraph,
		windGraph,
	)

	allQuadrants := lipgloss.JoinVertical(lipgloss.Center,
		topRow,
		bottomRow,
	)

	mainContainer := mainContainerStyle.Render(allQuadrants)

	fullView := lipgloss.JoinVertical(lipgloss.Center,
		mainContainer,
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center,
		lipgloss.Center,
		fullView)
}

type observationMsg struct {
	observation *api.ObservationTempest
}

type lightningStrikeMsg struct {
	strike *api.LightningStrike
}

type errMsg struct {
	err error
}

func (m model) StartListener() {
	m.listener.RegisterHandler(tempest.EventObservationTempest, m.handleObservation)
	m.listener.RegisterHandler(tempest.EventLightingStrike, m.handleLightningStrike)
	m.listener.Listen(context.Background())
	return
}

func (m *model) handleObservation(ctx context.Context, b []byte) {
	var obs api.ObservationTempest
	if err := json.Unmarshal(b, &obs); err != nil {
		return
	}

	m.updates <- observationMsg{observation: &obs}
}

func (m *model) handleLightningStrike(ctx context.Context, b []byte) {
	var obs api.LightningStrike
	err := json.Unmarshal(b, &obs)
	if err != nil {
		return
	}

	m.updates <- errMsg{err: err}
}

func (m *model) renderWindGraph(width, height int) string {

	return asciigraph.Plot(
		m.windSpeedHistory,
		asciigraph.Height(height-20),
		asciigraph.Width(width-10),
		asciigraph.Precision(1),
		asciigraph.SeriesColors(asciigraph.Blue),
		asciigraph.SeriesLegends("mph"),
	)
}

func (m *model) renderPressureGraph(width, height int) string {

	return asciigraph.Plot(
		m.pressureHistory,
		asciigraph.Height(height-20),
		asciigraph.Width(width-10),
		asciigraph.Precision(1),
		asciigraph.SeriesLegends("mb"),
		asciigraph.SeriesColors(
			asciigraph.Indigo,
		),
	)
}

func (m *model) renderTemperatureGraph(width, height int) string {
	temp := asciigraph.Plot(
		m.tempHistory,
		asciigraph.Height(height-20),
		asciigraph.Width(width-10),
		asciigraph.Precision(1),
		asciigraph.SeriesColors(asciigraph.Red),
		asciigraph.SeriesLegends("Temp °F"),
	)

	return temp
}
