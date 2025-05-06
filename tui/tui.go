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
	listener       tempest.Listener
	observation    *api.ObservationTempest
	spinner        spinner.Model
	err            error
	quitting       bool
	lightningFlash bool
	updates        chan tea.Msg // Add channel for updates
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

	case lightningStrikeMsg:
		m.lightningFlash = true
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

	w, h := m.terminalSize()

	// Single container sizing
	containerWidth := w - 8
	containerHeight := h - 6

	// Calculate widths for data and art sections
	dataWidth := (containerWidth / 2) - 2 // Half minus some padding
	artWidth := (containerWidth / 2) - 2

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Width(containerWidth - 4). // Account for container padding
		MarginBottom(1)

	// Update individual styles to maintain left alignment
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")). // Light blue
		Bold(true).
		Width(15).
		Align(lipgloss.Left)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // Bright white
		Width(20).
		Align(lipgloss.Left)

	detailsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // Gray
		Width(25).
		Align(lipgloss.Left)

	// Update styles for the main container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Margin(1, 2).
		Width(containerWidth).
		Height(containerHeight)

	innerContainerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 1).
		AlignHorizontal(lipgloss.Center). // Center horizontally
		Width(dataWidth - 4).
		Height(containerHeight - 8)

	// Style for the data section
	// dataStyle := lipgloss.NewStyle().
	// 	Width(dataWidth).
	// 	Height(containerHeight-4).
	// 	Padding(0, 1)

	// Style for the art section
	artStyle := lipgloss.NewStyle().
		Width(artWidth).
		Height(containerHeight-4).
		Align(lipgloss.Center).
		Padding(0, 1)

	// Update content style to manage the width of the content block
	contentStyle := lipgloss.NewStyle().
		Width(dataWidth - 12).           // Reduced width to allow centering
		Align(lipgloss.Left).            // Keep text left-aligned
		AlignHorizontal(lipgloss.Center) // Center the content block itself

		// Add a section style with bottom margin
	sectionStyle := lipgloss.NewStyle().
		MarginBottom(1)

	dataContent := innerContainerStyle.Render(
		contentStyle.Render(
			titleStyle.Render("Current Weather Conditions"),
			lipgloss.JoinVertical(lipgloss.Left,
				sectionStyle.Render(
					lipgloss.JoinHorizontal(lipgloss.Top,
						labelStyle.Render("Temperature"),
						valueStyle.Render(fmt.Sprintf("%.1f°F", m.observation.TemperatureInFarneheit())),
						lipgloss.JoinVertical(lipgloss.Left,
							detailsStyle.Render(fmt.Sprintf("Feels Like: %.1f°F", m.observation.FeelsLikeFarenheit())),
							detailsStyle.Render(fmt.Sprintf("Wind Chill: %.1f°F", m.observation.Summary.WindChill)),
						),
					),
				),
				// Wind section
				sectionStyle.Render(
					lipgloss.JoinHorizontal(lipgloss.Top,
						labelStyle.Render("Wind"),
						valueStyle.Render(fmt.Sprintf("%s at %.1f mph", m.observation.WindDirection(), m.observation.WindSpeedAverageMPH())),
						lipgloss.JoinVertical(lipgloss.Left,
							detailsStyle.Render(fmt.Sprintf("Gust: %.1f mph", m.observation.WindSpeedGustMPH())),
						),
					),
				),
				// Conditions section
				sectionStyle.Render(
					lipgloss.JoinHorizontal(lipgloss.Top,
						labelStyle.Render("Conditions"),
						valueStyle.Render(m.observation.PrecipitationType()),
						lipgloss.JoinVertical(lipgloss.Left,
							detailsStyle.Render(fmt.Sprintf("Rain: %.2fin/hr", m.observation.RainfallInInches())),
						),
					),
				),
				// Pressure section
				lipgloss.JoinHorizontal(lipgloss.Top,
					labelStyle.Render("Pressure"),
					valueStyle.Render(fmt.Sprintf("%.1f mb", m.observation.Data.StationPressure)),
					lipgloss.JoinVertical(lipgloss.Left,
						detailsStyle.Render(fmt.Sprintf("Trend: %s", m.observation.Summary.PressureTrend)),
					),
				),
			),
		),
	)
	// Remove the data style render since we're handling alignment in the content style
	content := lipgloss.JoinHorizontal(lipgloss.Top,
		dataContent, // Remove dataStyle.Render()
		artStyle.Render(m.getConditionsArt()),
	)
	mainContainer := containerStyle.Render(content)

	// Add title and quit message
	fullView := lipgloss.JoinVertical(lipgloss.Center,
		mainContainer,
		lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(w).
			Render("Press ESC to quit"),
	)

	return lipgloss.Place(w, h,
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

const (
	sunnyArt = `
   \     /
    \   /
 ---- O ----
    /   \
   /	 \
`
	rainyArt = `
     .--.
   .(    ).
  (___.__).)
   ' ' ' '
  ' ' ' ' '
`
	cloudyArt = `
      .--.
   .-(    ).
  (___.__)__)
`
	lightningCloudArt = `
     .--.
   .(    ).
  (___.__).)
`
	lightningArt = `
    /\
    / \
`
	nightArt = `
   *  .  *
     ☾   *
  *    .	
`

	// Sun variations
	sunnyArt1 = `
    \\|//
   -- O --
    //|\\
`
	sunnyArt2 = `
   \  |  /
  *--☀️--*
   /  |  \
`
	// Rain variations
	heavyRainCloudArt = `
     .--.
   .(    ).
  (___.__).)
`
	heavyRainArt = `
	||||||||| 
   |||||||||||`
	lightRainCloudArt = `
     .--.
   .(    ).
  (___.__).)
`
	lightRainDropsArt = `
	' . ' .
   . ' . '`
	partialSunArt = `    
	\  /
 --- O ---`

	partlyCloudyArt = `
    .--.
  .(    ).
 (__.___).)
`
	windyCloudArt = `
      .--.   ~
   .-/    \---
  (___.___)--~
     ~~~~
`
)

var (
	sunStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Bright yellow
	rainStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))  // Light blue
	cloudStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // Gray
	lightningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	nightStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))  // Purple
	heavyRainStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))  // Darker blue
	windStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("251")) // Light gray
)

func (m *model) getConditionsArt() string {
	if m.observation.Data.UltraviolentIndex < 0.1 {
		return nightStyle.Render(nightArt)
	}

	switch {
	case m.lightningFlash:
		return cloudStyle.Render(lightningCloudArt) + lightningStyle.Render(lightningArt)
	case m.observation.PrecipitationType() == "Raining":
		if m.observation.Summary.PrecipTotalOneHour > 0.5 {
			return cloudStyle.Render(heavyRainCloudArt) + heavyRainStyle.Render(heavyRainArt)
		}
		return cloudStyle.Render(lightRainCloudArt) + rainStyle.Render(lightRainDropsArt)
	case m.observation.WindSpeedAverageMPH() > 15:
		return windStyle.Render(windyCloudArt)
	case m.observation.Data.UltraviolentIndex > 5:
		if m.observation.Data.UltraviolentIndex > 8 {
			return sunStyle.Render(sunnyArt2)
		}
		return sunStyle.Render(sunnyArt1)
	default:
		if m.observation.Data.UltraviolentIndex > 0 {
			return sunStyle.Render(partialSunArt) + cloudStyle.Render(partlyCloudyArt)
		}
		return cloudStyle.Render(cloudyArt)
	}
}
