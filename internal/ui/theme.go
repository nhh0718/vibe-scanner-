package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	mutedColor      = lipgloss.Color("#94A3B8")
	subtleColor     = lipgloss.Color("#64748B")
	surfaceColor    = lipgloss.Color("#0F172A")
	surfaceSoft     = lipgloss.Color("#111827")
	lineColor       = lipgloss.Color("#1D4ED8")
	highlightColor  = lipgloss.Color("#DBEAFE")
	greenColor      = lipgloss.Color("#16A34A")
	yellowColor     = lipgloss.Color("#CA8A04")
	whiteSoftColor  = lipgloss.Color("#E2E8F0")

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lineColor).
			Foreground(whiteSoftColor).
			Background(surfaceColor).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(BlueColor).
			Background(surfaceSoft).
			Padding(0, 1)

	hintStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Background(surfaceSoft).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lineColor)

	labelStyle = lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle = lipgloss.NewStyle().Foreground(whiteSoftColor).Bold(true)
)

const (
	AppWidth          = 96
	MainColumnWidth   = 60
	SideColumnWidth   = 28
	CompactPanelWidth = 32
	MaxAppWidth       = 112
	MinAppWidth       = 68
	CompactHeight     = 30
	panelFrameWidth   = 6
	panelGapWidth     = 2
	minMainWidth      = 42
	minSideWidth      = 26
)

type ScreenLayout struct {
	ScreenWidth int
	MainWidth   int
	SideWidth   int
	Stack       bool
}

func GetSpinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(BlueColor).Bold(true)
}

func RenderScreen(title string, subtitle string, main string, side string, hints []string) string {
	return renderScreenSimple(title, subtitle, main, side, hints)
}

func RenderScreenWithWidth(viewportWidth int, title string, subtitle string, main string, side string, hints []string) string {
	return renderScreenSimple(title, subtitle, main, side, hints)
}

func RenderScreenWithViewport(viewportWidth int, viewportHeight int, title string, subtitle string, main string, side string, hints []string) string {
	return renderScreenSimple(title, subtitle, main, side, hints)
}

func renderScreenSimple(title string, subtitle string, main string, side string, hints []string) string {
	var sections []string

	sections = append(sections, GetLogo())

	if title != "" {
		sections = append(sections, titleStyle.Render(title))
	}
	if subtitle != "" {
		sections = append(sections, lipgloss.NewStyle().Foreground(subtleColor).Render(subtitle))
	}

	if main != "" {
		sections = append(sections, panelStyle.Render(main))
	}

	if side != "" {
		sections = append(sections, panelStyle.Render(side))
	}

	if len(hints) > 0 {
		hintText := strings.Join(hints, lipgloss.NewStyle().Foreground(subtleColor).Render("   •   "))
		sections = append(sections, hintStyle.Render(hintText))
	}

	return strings.Join(sections, "\n\n")
}

func GetScreenLayout(viewportWidth int, hasSide bool) ScreenLayout {
	screenWidth := viewportWidth
	if screenWidth <= 0 {
		screenWidth = AppWidth
	} else {
		screenWidth -= 2
	}

	if screenWidth > MaxAppWidth {
		screenWidth = MaxAppWidth
	}
	if screenWidth < MinAppWidth {
		screenWidth = MinAppWidth
	}

	if !hasSide {
		return ScreenLayout{
			ScreenWidth: screenWidth,
			MainWidth:   max(minMainWidth, screenWidth-panelFrameWidth),
		}
	}

	usableWidth := screenWidth - (panelFrameWidth * 2) - panelGapWidth
	if usableWidth < minMainWidth+minSideWidth {
		panelWidth := max(minMainWidth, screenWidth-panelFrameWidth)
		return ScreenLayout{
			ScreenWidth: screenWidth,
			MainWidth:   panelWidth,
			SideWidth:   panelWidth,
			Stack:       true,
		}
	}

	sideWidth := usableWidth / 3
	if sideWidth < minSideWidth {
		sideWidth = minSideWidth
	}
	mainWidth := usableWidth - sideWidth
	if mainWidth < minMainWidth {
		panelWidth := max(minMainWidth, screenWidth-panelFrameWidth)
		return ScreenLayout{
			ScreenWidth: screenWidth,
			MainWidth:   panelWidth,
			SideWidth:   panelWidth,
			Stack:       true,
		}
	}

	return ScreenLayout{
		ScreenWidth: screenWidth,
		MainWidth:   mainWidth,
		SideWidth:   sideWidth,
	}
}

func Panel(content string, title string, width int) string {
	return renderPanel(content, title, width, false)
}

func renderPanel(content string, title string, width int, compact bool) string {
	verticalPadding := 1
	if compact {
		verticalPadding = 0
	}
	panel := lipgloss.NewStyle().
		Width(width).
		Padding(verticalPadding, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lineColor).
		Foreground(whiteSoftColor).
		Background(surfaceColor)

	if title != "" {
		header := lipgloss.NewStyle().
			Bold(true).
			Foreground(BlueColor).
			Background(surfaceSoft).
			Padding(0, 1).
			Render(title)
		gap := "\n\n"
		if compact {
			gap = "\n"
		}
		content = header + "\n" + lipgloss.NewStyle().Foreground(subtleColor).Render(strings.Repeat("─", max(8, width-8))) + gap + content
	}

	return panel.Render(content)
}

func HintBar(items ...string) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, lipgloss.NewStyle().Foreground(highlightColor).Render(item))
	}
	return strings.Join(parts, lipgloss.NewStyle().Foreground(subtleColor).Render("   •   "))
}

func KeyValue(label string, value string) string {
	return labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

func SectionLabel(text string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(BlueColor).Render(text)
}

func Muted(text string) string {
	return lipgloss.NewStyle().Foreground(mutedColor).Render(text)
}

func Subtle(text string) string {
	return lipgloss.NewStyle().Foreground(subtleColor).Render(text)
}

func SuccessText(text string) string {
	return lipgloss.NewStyle().Foreground(greenColor).Bold(true).Render(text)
}

func WarningText(text string) string {
	return lipgloss.NewStyle().Foreground(yellowColor).Bold(true).Render(text)
}

func ErrorText(text string) string {
	return lipgloss.NewStyle().Foreground(RedColor).Bold(true).Render(text)
}

func Badge(text string, active bool) string {
	style := lipgloss.NewStyle().
		Width(6).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(BlueColor).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lineColor).
		Background(surfaceSoft)
	if active {
		style = style.Background(BlueColor).Foreground(lipgloss.Color("#FFFFFF")).BorderForeground(BlueColor)
	}
	return style.Render(text)
}

func NumberedLine(number int, title string, desc string, active bool) string {
	numberStr := fmt.Sprintf("%02d", number)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(whiteSoftColor)
	descStyle := lipgloss.NewStyle().Foreground(mutedColor)

	if active {
		numberStr = "▸"
		titleStyle = titleStyle.Foreground(lipgloss.Color("#FFFFFF")).Underline(true)
		descStyle = descStyle.Foreground(highlightColor)
	}

	numberStyle := lipgloss.NewStyle().
		Width(4).
		Foreground(BlueColor).
		Bold(true)
	if active {
		numberStyle = numberStyle.Foreground(RedColor)
	}

	line := numberStyle.Render(numberStr) + " " + titleStyle.Render(title)
	if desc != "" {
		line = line + "\n    " + descStyle.Render(desc)
	}
	return line
}

func BulletList(items []string) string {
	var lines []string
	for _, item := range items {
		lines = append(lines, "  • "+item)
	}
	return strings.Join(lines, "\n")
}

func Center(content string) string {
	return CenterToWidth(content, AppWidth)
}

func IsCompactHeight(height int) bool {
	return height > 0 && height <= CompactHeight
}

func CenterToWidth(content string, width int) string {
	if width <= 0 {
		width = AppWidth
	}
	if lipgloss.Width(content) >= width {
		return content
	}
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(content)
}

func compactBrand() string {
	star := lipgloss.NewStyle().Foreground(BlueColor).Bold(true).Render("★")
	brand := lipgloss.NewStyle().Bold(true).Render(
		lipgloss.NewStyle().Foreground(BlueColor).Render("i") +
			lipgloss.NewStyle().Foreground(RedColor).Render("NET"),
	)
	caption := lipgloss.NewStyle().Foreground(subtleColor).Render("CLI")
	return lipgloss.JoinHorizontal(lipgloss.Center, star, " ", brand, " ", caption)
}

func headerBlock(width int, title string, subtitle string) string {
	var lines []string
	if title != "" {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(BlueColor).Render(title))
	}
	if subtitle != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(subtleColor).Render(subtitle))
	}
	if len(lines) == 0 {
		return ""
	}
	return lipgloss.NewStyle().
		Width(max(24, width-6)).
		Align(lipgloss.Center).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lineColor).
		Render(strings.Join(lines, "\n"))
}

func footerBlock(width int, items ...string) string {
	return lipgloss.NewStyle().
		Width(max(22, width-8)).
		Align(lipgloss.Center).
		Foreground(highlightColor).
		Background(surfaceSoft).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lineColor).
		Render(HintBar(items...))
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
