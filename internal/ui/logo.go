package ui

import (
	"fmt"
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

// DashboardInfo holds info for the dashboard banner
type DashboardInfo struct {
	URL          string
	ProjectName  string
	FindingCount int
	Timestamp    string
	HealthScore  int
}

// GetDashboardBanner returns a styled CLI panel for the running dashboard
func GetDashboardBanner(info DashboardInfo) string {
	cyan := lipgloss.Color("#22D3EE")
	dimWhite := lipgloss.Color("#CBD5E1")

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(BlueColor).
		Padding(0, 2).
		Align(lipgloss.Center)

	urlStyle := lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true).
		Underline(true)

	labelW := lipgloss.NewStyle().
		Foreground(mutedColor).
		Width(14)

	valStyle := lipgloss.NewStyle().
		Foreground(dimWhite).
		Bold(true)

	scoreStyle := lipgloss.NewStyle().Bold(true)
	if info.HealthScore >= 70 {
		scoreStyle = scoreStyle.Foreground(greenColor)
	} else if info.HealthScore >= 40 {
		scoreStyle = scoreStyle.Foreground(yellowColor)
	} else {
		scoreStyle = scoreStyle.Foreground(RedColor)
	}

	divider := lipgloss.NewStyle().Foreground(subtleColor).Render(strings.Repeat("─", 48))

	findingsLabel := "vấn đề phát hiện"
	if info.FindingCount == 0 {
		findingsLabel = "không có vấn đề"
	}

	var lines []string
	lines = append(lines, headerStyle.Render("  BẢNG ĐIỀU KHIỂN WEB  "))
	lines = append(lines, divider)
	lines = append(lines, "")
	lines = append(lines, labelW.Render("  URL")+" "+urlStyle.Render(info.URL))
	lines = append(lines, labelW.Render("  Dự án")+" "+valStyle.Render(info.ProjectName))
	lines = append(lines, labelW.Render("  Findings")+" "+valStyle.Render(fmt.Sprintf("%d %s", info.FindingCount, findingsLabel)))
	if info.HealthScore > 0 {
		lines = append(lines, labelW.Render("  Health")+" "+scoreStyle.Render(fmt.Sprintf("%d/100", info.HealthScore)))
	}
	if info.Timestamp != "" {
		lines = append(lines, labelW.Render("  Quét lúc")+" "+lipgloss.NewStyle().Foreground(mutedColor).Render(info.Timestamp))
	}
	lines = append(lines, "")
	lines = append(lines, divider)

	hintKey := lipgloss.NewStyle().Foreground(highlightColor).Bold(true)
	hintDesc := lipgloss.NewStyle().Foreground(mutedColor)
	lines = append(lines, "  "+hintKey.Render("Ctrl+C")+" "+hintDesc.Render("dừng server")+"    "+hintKey.Render("Browser")+" "+hintDesc.Render("tự động mở"))

	content := strings.Join(lines, "\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BlueColor).
		Background(surfaceSoft).
		Foreground(whiteSoftColor).
		Padding(1, 2).
		Width(56)

	return boxStyle.Render(content)
}

// GetDashboardStoppedBanner returns a styled message when dashboard stops
func GetDashboardStoppedBanner() string {
	return GetSuccessBox("Dashboard đã dừng thành công")
}
