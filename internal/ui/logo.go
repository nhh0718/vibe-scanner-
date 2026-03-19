package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors from iNET brand
	BlueColor = lipgloss.Color("#024799")
	RedColor  = lipgloss.Color("#cc0e0e")
)

// GetLogo returns the iNET ASCII logo with proper colors
func GetLogo() string {
	// Blue style for 'i' and 'N'
	blueStyle := lipgloss.NewStyle().Foreground(BlueColor).Bold(true)
	
	// Red style for 'E'
	redStyle := lipgloss.NewStyle().Foreground(RedColor).Bold(true)
	
	// Star for dot on 'i'
	starStyle := lipgloss.NewStyle().Foreground(BlueColor).Bold(true)
	
	logo := ""
	
	// Line 1: Star (dot on i) - add space before
	logo += "   " + starStyle.Render("★") + "\n"
	
	// Line 2: Top of letters
	logo += "  " + blueStyle.Render("██") + "   " + 
		blueStyle.Render("███╗   ██╗") + "  " +
		redStyle.Render("███████╗") + "  " +
		blueStyle.Render("████████╗") + "\n"
	
	// Line 3
	logo += "  " + blueStyle.Render("██") + "   " +
		blueStyle.Render("████╗  ██║") + "  " +
		redStyle.Render("██╔════╝") + "  " +
		blueStyle.Render("╚══██╔══╝") + "\n"
	
	// Line 4
	logo += "  " + blueStyle.Render("██") + "   " +
		blueStyle.Render("██╔██╗ ██║") + "  " +
		redStyle.Render("█████╗  ") + "  " +
		blueStyle.Render("   ██║   ") + "\n"
	
	// Line 5
	logo += "  " + blueStyle.Render("██") + "   " +
		blueStyle.Render("██║╚██╗██║") + "  " +
		redStyle.Render("██╔══╝  ") + "  " +
		blueStyle.Render("   ██║   ") + "\n"
	
	// Line 6
	logo += "  " + blueStyle.Render("██") + "   " +
		blueStyle.Render("██║ ╚████║") + "  " +
		redStyle.Render("███████╗") + "  " +
		blueStyle.Render("   ██║   ") + "\n"
	
	// Line 7: Bottom
	logo += "  " + blueStyle.Render("╚═") + "   " +
		blueStyle.Render("╚═╝  ╚═══╝") + "  " +
		redStyle.Render("╚══════╝") + "  " +
		blueStyle.Render("   ╚═╝   ") + "\n"
	
	return logo
}

// GetBorderedBox returns content wrapped in a bordered box
func GetBorderedBox(content string, title string) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BlueColor).
		Padding(1, 2).
		Width(70)
	
	if title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(BlueColor).
			Bold(true).
			Align(lipgloss.Center)
		titleText := titleStyle.Render(title)
		content = titleText + "\n\n" + content
	}
	
	return boxStyle.Render(content)
}

// GetInfoBox returns an info box with icon
func GetInfoBox(content string) string {
	infoStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BlueColor).
		Padding(0, 1).
		Foreground(lipgloss.Color("#64748b"))
	
	return infoStyle.Render("ℹ " + content)
}

// GetSuccessBox returns a success box with icon
func GetSuccessBox(content string) string {
	successStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#22c55e")).
		Padding(0, 1).
		Foreground(lipgloss.Color("#22c55e"))
	
	return successStyle.Render("✓ " + content)
}

// GetErrorBox returns an error box with icon
func GetErrorBox(content string) string {
	errorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(RedColor).
		Padding(0, 1).
		Foreground(RedColor)
	
	return errorStyle.Render("✗ " + content)
}
