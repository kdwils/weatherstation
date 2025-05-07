package tui

import (
	_ "embed"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	//go:embed art/moon.txt
	moon string

	//go:embed art/stars.txt
	nightSky string

	//go:embed art/clouds.txt
	clouds string

	//go:embed art/partialsun.txt
	partialSun string

	//go:embed art/heavyrain.txt
	heavyRain string

	//go:embed art/lightrain.txt
	lightRain string

	//go:embed art/sunny.txt
	sunny string
)

var (
	sunStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))     // Bright yellow
	rainStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))      // Light blue
	cloudStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))     // Gray
	lightningStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))     // Yellow
	heavyRainStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#0814cb")) // Darker blue
	lightRainStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#07497e")) // Lighter blue
	lightestRainStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6a95f7")) // Lightest blue
	windStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("251"))     // Light gray

	moonStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#818181")) // Bright white
	yellowStarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d9de6a")) // Yellow
	blueStarStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#3f3da8")) // Red
)

func renderNightSky(nightSky string) string {
	var coloredSky string
	// Split into lines to preserve formatting
	lines := strings.Split(nightSky, "\n")

	for _, line := range lines {
		var coloredLine string
		for _, char := range line {
			switch string(char) {
			case "+", ">":
				coloredLine += blueStarStyle.Render(string(char))
			case "`", "*", "|", "-", "o":
				coloredLine += yellowStarStyle.Render(string(char))
			case "_", "\\", "/", "'", ".", "O":
				coloredLine += moonStyle.Render(string(char))
			default:
				coloredLine += string(char)
			}
		}
		coloredSky += coloredLine + "\n"
	}
	return coloredSky
}

func renderRainArt(art string) string {
	var coloredArt string
	lines := strings.Split(art, "\n")

	for _, line := range lines {
		var coloredLine string
		for _, char := range line {
			switch string(char) {
			case ".", "-", "(", ")", "_":
				coloredLine += cloudStyle.Render(string(char))
			case "|", "/", ";":
				coloredLine += heavyRainStyle.Render(string(char))
			case ",", "'":
				coloredLine += lightRainStyle.Render(string(char))
			case "`":
				coloredLine += lightestRainStyle.Render(string(char))
			default:
				coloredLine += string(char)
			}
		}
		coloredArt += coloredLine + "\n"
	}
	return coloredArt
}

func renderPartialSunArt(art string) string {
	sunStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))    // Bright yellow
	cloudStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))  // Gray
	treeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("28"))    // Dark green for trees
	trunkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("94"))   // Brown for tree trunks
	groundStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("235")) // Dark gray for ground
	waterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("32"))   // Blue for water (~)

	lines := strings.Split(art, "\n")
	var coloredArt string

	for _, line := range lines {
		var coloredLine string
		for _, char := range line {
			switch string(char) {
			case ".", "-", "(", ")", "_":
				coloredLine += cloudStyle.Render(string(char)) // Clouds
			case "O", ",", "!", "\\", "/", "`", "'", "*":
				coloredLine += sunStyle.Render(string(char)) // Sun and rays
			case "&":
				coloredLine += treeStyle.Render(string(char)) // Trees
			case "|":
				coloredLine += trunkStyle.Render(string(char)) // Tree trunks
			case "=":
				coloredLine += groundStyle.Render(string(char)) // Ground/terrain
			case "~":
				coloredLine += waterStyle.Render(string(char)) // Water
			default:
				coloredLine += string(char)
			}
		}
		coloredArt += coloredLine + "\n"
	}

	return coloredArt
}

// Add these new styles at the top with other styles
var (
	// ...existing styles...
	sunRaysStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Bright yellow for rays
	sunCenterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Golden yellow for center
)

func renderSunny(art string) string {
	var coloredArt string
	lines := strings.Split(art, "\n")

	for _, line := range lines {
		var coloredLine string
		for _, char := range line {
			switch string(char) {
			case "\\", "/", "|", "-":
				coloredLine += sunRaysStyle.Render(string(char))
			case "*", "⣿":
				coloredLine += sunCenterStyle.Render(string(char))
			default:
				coloredLine += string(char)
			}
		}
		coloredArt += coloredLine + "\n"
	}
	return coloredArt
}
