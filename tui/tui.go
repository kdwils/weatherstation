package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kdwils/weatherstation/pkg/api"
	"github.com/kdwils/weatherstation/pkg/connection"
	"github.com/kdwils/weatherstation/pkg/tempest"
	"golang.org/x/term"
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

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case observationMsg:
		m.observation = msg.observation
		if m.lightningCounter > 0 {
			m.lightningCounter--
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

func (m *model) terminalSize() (width, height int) {
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))
	return w, h
}

func (m *model) View() string {

	if m.observation == nil {
		return lipgloss.Place(80, 24,
			lipgloss.Center,
			lipgloss.Center,
			m.spinner.View())
	}

	// Main container for entire terminal with outer border
	mainContainerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(0, 1, 1, 1).
		Width(m.width - 4).
		Height(m.height - 4)

	// Calculate quadrant sizes
	quadrantWidth := (m.width - 8) / 2
	quadrantHeight := (m.height - 8) / 2

	// Style for data quadrant (with inner border)
	dataQuadrantStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1).
		Width(quadrantWidth - 2).
		Height(quadrantHeight - 2)

	// Style for other quadrants (no border)
	plainQuadrantStyle := lipgloss.NewStyle().
		Padding(1).
		Width(quadrantWidth - 2).
		Height(quadrantHeight - 2)

	// Update text styles to be more prominent
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Width(quadrantWidth - 8).
		MarginBottom(2)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Width(20).
		Align(lipgloss.Left)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Width(25).
		Align(lipgloss.Left)

	detailsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Width(30).
		Align(lipgloss.Left)

	contentStyle := lipgloss.NewStyle().
		Width(quadrantWidth - 8).
		AlignHorizontal(lipgloss.Center)

	sectionStyle := lipgloss.NewStyle().
		MarginBottom(2)

	temperatureSection := sectionStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("Temperature"),
			valueStyle.Render(fmt.Sprintf("%.1f°F", m.observation.TemperatureInFarneheit())),
			lipgloss.JoinVertical(lipgloss.Left,
				detailsStyle.Render(fmt.Sprintf("Feels Like: %.1f°F", m.observation.FeelsLikeFarenheit())),
				detailsStyle.Render(fmt.Sprintf("Wind Chill: %.1f°F", m.observation.Summary.WindChill)),
			),
		),
	)

	windSection := sectionStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("Wind"),
			valueStyle.Render(fmt.Sprintf("%s at %.1f mph", m.observation.WindDirection(), m.observation.WindSpeedAverageMPH())),
			lipgloss.JoinVertical(lipgloss.Left,
				detailsStyle.Render(fmt.Sprintf("Gust: %.1f mph", m.observation.WindSpeedGustMPH())),
			),
		),
	)

	conditionsSection := sectionStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("Conditions"),
			valueStyle.Render(m.observation.PrecipitationType()),
			lipgloss.JoinVertical(
				lipgloss.Left,
				detailsStyle.Render(fmt.Sprintf("Rainfall Today:     %.2fin", m.observation.RainfallInInches())),
				detailsStyle.Render(fmt.Sprintf("Rainfall Hourly:    %.2fin/hr", m.observation.Summary.PrecipTotalOneHour)),
				detailsStyle.Render(fmt.Sprintf("Rainfall Yesterday: %.2fin", m.observation.Summary.PrecipAccumLocalYesterdayFinal)),
			),
		),
	)

	pressureSection := sectionStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("Pressure"),
			valueStyle.Render(fmt.Sprintf("%.1f mb", m.observation.Data.StationPressure)),
			lipgloss.JoinVertical(lipgloss.Left,
				detailsStyle.Render(fmt.Sprintf("Trend: %s", m.observation.Summary.PressureTrend)),
			),
		),
	)

	// Build data content for quadrant 1
	dataContent := dataQuadrantStyle.Render(
		contentStyle.Render(
			titleStyle.Render("Current Weather Conditions"),
			lipgloss.JoinVertical(lipgloss.Left,
				temperatureSection,
				windSection,
				conditionsSection,
				pressureSection,
			),
		),
	)

	// Style for art quadrant
	artStyle := lipgloss.NewStyle().
		Width(quadrantWidth - 8).
		Height(quadrantHeight - 4).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Build art content for quadrant 2
	artContent := plainQuadrantStyle.Render(
		artStyle.Render(
			m.getArt(),
		),
	)

	// Create the quadrants layout
	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		dataContent,
		artContent,
	)

	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
		plainQuadrantStyle.Render("Quadrant 3"),
		plainQuadrantStyle.Render("Quadrant 4"),
	)

	// Join all quadrants vertically
	allQuadrants := lipgloss.JoinVertical(lipgloss.Left,
		topRow,
		bottomRow,
	)

	// Render main container with all quadrants
	mainContainer := mainContainerStyle.Render(allQuadrants)

	// Add quit message below
	fullView := lipgloss.JoinVertical(lipgloss.Center,
		mainContainer,
		lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(m.width).
			Render("Press ESC to quit"),
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

// Modify your getArt function to use scaled art
func (m *model) getArt() string {
	// Calculate available space in the art quadrant
	artWidth := (m.width - 8) / 2   // Quadrant width
	artHeight := (m.height - 8) / 2 // Quadrant height

	currentHour := time.Now().Hour()
	isNightTime := currentHour >= 22 || currentHour < 5
	isRaining := m.observation.PrecipitationType() == "Raining"
	isLightning := m.lightningCounter > 0
	isHeavyRain := m.observation.Summary.PrecipTotalOneHour > 0.5

	var art string
	if m.observation.Data.UltraviolentIndex < 0.1 && isNightTime {
		art = scaleArt(moon, artWidth, artHeight)
		return renderNightSky(art)
	}

	switch {
	case isLightning:
		// art = scaleArt(lightRain, artWidth, artHeight)
		// return renderLightningArt(art)
	case isRaining:
		if isHeavyRain {
			art = scaleArt(heavyRain, artWidth, artHeight)
			return renderRainArt(art)
		}
		art = scaleArt(lightRain, artWidth, artHeight)
		return renderRainArt(art)
	case m.observation.Data.UltraviolentIndex > 5:
		art = scaleArt(sunny, artWidth, artHeight)
		return renderSunny(art)
	}

	if m.observation.Data.UltraviolentIndex > 0 {
		art = scaleArt(partialSun, artWidth, artHeight)
		return renderPartialSunArt(art)
	}

	art = scaleArt(clouds, artWidth, artHeight)
	return renderSunny(art)
}

func scaleArt(art string, width, height int) string {
	lines := strings.Split(art, "\n")
	if len(lines) == 0 {
		return ""
	}

	// Find the longest line for reference
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	// Scale horizontally by duplicating characters
	var scaledLines []string
	horizontalScale := float64(width) / float64(maxLen)

	for _, line := range lines {
		var scaledLine strings.Builder
		for _, char := range line {
			// Repeat each character based on scale factor
			repeats := int(horizontalScale)
			if repeats < 1 {
				repeats = 1
			}
			scaledLine.WriteString(strings.Repeat(string(char), repeats))
		}

		// Trim or pad to exact width
		lineStr := scaledLine.String()
		if len(lineStr) > width {
			lineStr = lineStr[:width]
		} else {
			lineStr += strings.Repeat(" ", width-len(lineStr))
		}
		scaledLines = append(scaledLines, lineStr)
	}

	// Scale vertically by duplicating lines
	var finalLines []string
	verticalScale := float64(height) / float64(len(lines))

	for _, line := range scaledLines {
		// Repeat each line based on scale factor
		repeats := int(verticalScale)
		if repeats < 1 {
			repeats = 1
		}
		for i := 0; i < repeats && len(finalLines) < height; i++ {
			finalLines = append(finalLines, line)
		}
	}

	// Trim to exact height if needed
	if len(finalLines) > height {
		finalLines = finalLines[:height]
	}

	return strings.Join(finalLines, "\n")
}
