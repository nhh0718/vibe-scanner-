package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors from iNET brand
	BlueColor = lipgloss.Color("#024799")
	RedColor  = lipgloss.Color("#cc0e0e")
)

// GetLogo returns the iNET ASCII logo with proper colors
func GetLogo() string {
	blueStyle := lipgloss.NewStyle().Foreground(BlueColor).Bold(true)
	redStyle := lipgloss.NewStyle().Foreground(RedColor).Bold(true)
	starStyle := lipgloss.NewStyle().Foreground(BlueColor).Bold(true)
	captionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8FA9CC")).Italic(true)

	lines := []string{
		"    " + starStyle.Render("★"),
		"  " + blueStyle.Render("██") + "    " + blueStyle.Render("███╗   ██╗") + "   " + redStyle.Render("███████╗") + "   " + blueStyle.Render("████████╗"),
		"  " + blueStyle.Render("██") + "    " + blueStyle.Render("████╗  ██║") + "   " + redStyle.Render("██╔════╝") + "   " + blueStyle.Render("╚══██╔══╝"),
		"  " + blueStyle.Render("██") + "    " + blueStyle.Render("██╔██╗ ██║") + "   " + redStyle.Render("█████╗  ") + "   " + blueStyle.Render("   ██║   "),
		"  " + blueStyle.Render("██") + "    " + blueStyle.Render("██║╚██╗██║") + "   " + redStyle.Render("██╔══╝  ") + "   " + blueStyle.Render("   ██║   "),
		"  " + blueStyle.Render("██") + "    " + blueStyle.Render("██║ ╚████║") + "   " + redStyle.Render("███████╗") + "   " + blueStyle.Render("   ██║   "),
		"  " + blueStyle.Render("╚═") + "    " + blueStyle.Render("╚═╝  ╚═══╝") + "   " + redStyle.Render("╚══════╝") + "   " + blueStyle.Render("   ╚═╝   "),
		captionStyle.Render("Bộ công cụ quét mã nguồn và phân tích chất lượng dự án"),
	}

	return strings.Join(lines, "\n")
}

// GetBorderedBox returns content wrapped in a bordered box
func GetBorderedBox(content string, title string) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lineColor).
		Background(surfaceSoft).
		Foreground(whiteSoftColor).
		Padding(1, 2).
		Width(AppWidth - 20)
	
	if title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(BlueColor).
			Bold(true).
			Padding(0, 1).
			Align(lipgloss.Center)
		titleText := titleStyle.Render(title)
		divider := lipgloss.NewStyle().Foreground(subtleColor).Render(strings.Repeat("─", AppWidth-28))
		content = titleText + "\n" + divider + "\n\n" + content
	}
	
	return boxStyle.Render(content)
}

// GetInfoBox returns an info box with icon
func GetInfoBox(content string) string {
	infoStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lineColor).
		Background(surfaceSoft).
		Padding(0, 2).
		Foreground(highlightColor)
	
	return infoStyle.Render("ℹ " + content)
}

// GetSuccessBox returns a success box with icon
func GetSuccessBox(content string) string {
	successStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(greenColor).
		Background(surfaceSoft).
		Padding(0, 2).
		Foreground(greenColor)
	
	return successStyle.Render("✓ " + content)
}

// GetErrorBox returns an error box with icon
func GetErrorBox(content string) string {
	errorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(RedColor).
		Background(surfaceSoft).
		Padding(0, 2).
		Foreground(RedColor)
	
	return errorStyle.Render("✗ " + content)
}
