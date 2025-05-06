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

	totalMargin := 4  // 2 units of margin on each side
	totalPadding := 4 // 2 units of padding on each side
	totalBorder := 2  // 1 unit of border on each side
	spacing := 4      // Space between boxes

	// Calculate available width for two boxes
	availableWidth := w - (2 * (totalMargin + totalPadding + totalBorder)) - spacing
	boxWidth := availableWidth / 2

	// Style definitions
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		MarginBottom(2).
		Width(boxWidth)

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Margin(1, 2).
		Width(boxWidth).
		Height(10)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Width(boxWidth).
		MarginBottom(1).
		Align(lipgloss.Center)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Width(boxWidth).
		MarginTop(1).
		MarginBottom(1).
		Align(lipgloss.Left)

	detailsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Width(boxWidth).
		MarginTop(1).
		Align(lipgloss.Left)

	artStyle := lipgloss.NewStyle().
		Width(20).
		Height(5).
		Align(lipgloss.Center)

	temperature := containerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			// Left section - data
			lipgloss.NewStyle().Width(boxWidth-25).Render(
				labelStyle.Render("Temperature")+
					valueStyle.Render(fmt.Sprintf("%.1f°F", m.observation.TemperatureInFarneheit()))+
					detailsStyle.Render(fmt.Sprintf("Feels Like: %.1f°F", m.observation.FeelsLikeFarenheit()))+
					detailsStyle.Render(fmt.Sprintf("Wind Chill: %.1f°F", m.observation.Summary.WindChill))+
					detailsStyle.Render(fmt.Sprintf("Dew Point: %.1f°F", m.observation.DewPointFarenheit()))+
					detailsStyle.Render(fmt.Sprintf("Humidity: %d%%", m.observation.Data.RelativeHumidity))),
		),
	)

	windAverage := fmt.Sprintf("%s at %.1f mph",
		m.observation.WindDirection(),
		m.observation.WindSpeedAverageMPH())

	wind := containerStyle.Render(
		labelStyle.Render("Wind") +
			valueStyle.Render(windAverage) +
			detailsStyle.Render(fmt.Sprintf("Gust: %s at %.1f mph", m.observation.WindDirection(), m.observation.WindSpeedGustMPH())) +
			detailsStyle.Render(fmt.Sprintf("Wind Chill: %.1f°F", m.observation.Summary.WindChill)))

	conditions := containerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			// Left section - data
			lipgloss.NewStyle().Width(boxWidth-25).Render(
				labelStyle.Render("Conditions")+
					valueStyle.Render(m.observation.PrecipitationType())+
					detailsStyle.Render(fmt.Sprintf("Rainfall: %.2fin", m.observation.RainfallInInches()))+
					detailsStyle.Render(fmt.Sprintf("Rain Last Hour: %.2fin", m.observation.Summary.PrecipTotalOneHour))+
					detailsStyle.Render(fmt.Sprintf("Rainfall Yesterday: %.2fin", m.observation.RainfallYesterdayInInches()))),
			// Right section - art
			artStyle.Render(m.getConditionsArt()),
		),
	)

	pressure := containerStyle.Render(
		labelStyle.Render("Pressure") +
			valueStyle.Render(fmt.Sprintf("%.1f mb", m.observation.Data.StationPressure)) +
			detailsStyle.Render(fmt.Sprintf("Trend: %s", m.observation.Summary.PressureTrend)))

	lightning := containerStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Top,
		labelStyle.Render("Lightning")+
			valueStyle.Render(fmt.Sprintf("%d strikes/hr", m.observation.Summary.StrikeCountOneHour))+
			detailsStyle.Render(fmt.Sprintf("Last Strike: %.1f miles", m.observation.AverageLightningStrikeDistanceInMiles()))+
			detailsStyle.Render(fmt.Sprintf("3hr Total: %d strikes", m.observation.Summary.StrikeCountThreeHour))))
	// lightning := containerStyle.Render(
	// 	labelStyle.Render("Lightning") +
	// 		valueStyle.Render(fmt.Sprintf("%d strikes/hr", m.observation.Summary.StrikeCountOneHour)) +
	// 		detailsStyle.Render(fmt.Sprintf("Last Strike: %.1f miles", m.observation.AverageLightningStrikeDistanceInMiles())) +
	// 		detailsStyle.Render(fmt.Sprintf("3hr Total: %d strikes", m.observation.Summary.StrikeCountThreeHour)))

	solar := containerStyle.Render(
		labelStyle.Render("Solar & UV") +
			valueStyle.Render(fmt.Sprintf("%.1f UV", m.observation.Data.UltraviolentIndex)) +
			detailsStyle.Render(fmt.Sprintf("Solar Radiation: %d W/m²", m.observation.Data.SolarRadiation)) +
			detailsStyle.Render(fmt.Sprintf("Illuminance: %d lux", m.observation.Data.Illuminance)))

	// Join sections horizontally
	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		temperature,
		lipgloss.NewStyle().Width(spacing).Render(""), // Add explicit spacing
		wind,
	)

	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		conditions,
		lipgloss.NewStyle().Width(spacing).Render(""),
		pressure,
	)

	row3 := lipgloss.JoinHorizontal(lipgloss.Top,
		lightning,
		lipgloss.NewStyle().Width(spacing).Render(""),
		solar,
	)

	row1 = lipgloss.Place(w, lipgloss.Height(temperature), lipgloss.Center, lipgloss.Center, row1)
	row2 = lipgloss.Place(w, lipgloss.Height(conditions), lipgloss.Center, lipgloss.Center, row2)
	row3 = lipgloss.Place(w, lipgloss.Height(lightning), lipgloss.Center, lipgloss.Center, row3)

	content := titleStyle.Render("Current Weather Conditions") + "\n\n" +
		row1 + "\n" +
		row2 + "\n" +
		row3 + "\n\n" +
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
