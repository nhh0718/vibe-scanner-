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
)

const (
	AppWidth          = 118
	MainColumnWidth   = 76
	SideColumnWidth   = 38
	CompactPanelWidth = 32
)

func GetSpinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(BlueColor).Bold(true)
}

func RenderScreen(title string, subtitle string, main string, side string, hints []string) string {
	sections := []string{Center(GetLogo()), Center(headerBlock(title, subtitle))}

	body := Panel(main, "NỘI DUNG CHÍNH", MainColumnWidth)
	if side != "" {
		body = lipgloss.JoinHorizontal(lipgloss.Top, body, "  ", Panel(side, "TỔNG HỢP NHANH", SideColumnWidth))
	}
	sections = append(sections, Center(body))

	if len(hints) > 0 {
		sections = append(sections, Center(footerBlock(hints...)))
	}

	return strings.Join(sections, "\n\n")
}

func Panel(content string, title string, width int) string {
	panel := lipgloss.NewStyle().
		Width(width).
		Padding(1, 2).
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
		content = header + "\n" + lipgloss.NewStyle().Foreground(subtleColor).Render(strings.Repeat("─", max(8, width-8))) + "\n\n" + content
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
	labelStyle := lipgloss.NewStyle().Width(18).Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(whiteSoftColor).Bold(true)
	return labelStyle.Render(label) + valueStyle.Render(value)
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
	indicator := lipgloss.NewStyle().Foreground(BlueColor).Render("  ")
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(whiteSoftColor)
	descStyle := lipgloss.NewStyle().Foreground(mutedColor)
	if active {
		indicator = lipgloss.NewStyle().Foreground(RedColor).Bold(true).Render("▸ ")
		titleStyle = titleStyle.Foreground(lipgloss.Color("#FFFFFF")).Underline(true)
		descStyle = descStyle.Foreground(highlightColor)
	}
	line1 := lipgloss.JoinHorizontal(lipgloss.Top, indicator, Badge(fmt.Sprintf("%02d", number), active), " ", titleStyle.Render(title))
	line2 := lipgloss.NewStyle().PaddingLeft(5).Render(descStyle.Render(desc))
	return line1 + "\n" + line2
}

func BulletList(items []string) string {
	var lines []string
	for _, item := range items {
		lines = append(lines, "  • "+item)
	}
	return strings.Join(lines, "\n")
}

func Center(content string) string {
	return lipgloss.NewStyle().Width(AppWidth).Align(lipgloss.Center).Render(content)
}

func headerBlock(title string, subtitle string) string {
	var lines []string
	if title != "" {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(BlueColor).Render(title))
	}
	if subtitle != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(subtleColor).Render(subtitle))
	}
	return lipgloss.NewStyle().
		Width(AppWidth - 10).
		Align(lipgloss.Center).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lineColor).
		Render(strings.Join(lines, "\n"))
}

func footerBlock(items ...string) string {
	return lipgloss.NewStyle().
		Width(AppWidth - 12).
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
