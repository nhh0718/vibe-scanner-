package output

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nhh0718/vibe-scanner-/internal/aggregation"
	"github.com/nhh0718/vibe-scanner-/internal/models"
	"github.com/nhh0718/vibe-scanner-/internal/ui"
)

// PrintTerminal in kết quả ra terminal
func PrintTerminal(results *models.ScanResult) {
	printHeader(results)
	printHealthScore(results.HealthScore)
	printSummary(results.Summary)
	printCriticalFindings(results)
	printFooter(results)
}

// printHeader in header
func printHeader(results *models.ScanResult) {
	fmt.Println()
	var main strings.Builder
	main.WriteString(ui.SectionLabel("Kết quả quét dự án"))
	main.WriteString("\n\n")
	main.WriteString(ui.KeyValue("Dự án", results.Project.Name))
	main.WriteString("\n")
	main.WriteString(ui.KeyValue("Đường dẫn", results.Project.Path))
	main.WriteString("\n")
	main.WriteString(ui.KeyValue("Thời gian quét", results.Duration.Round(time.Second).String()))

	var side strings.Builder
	side.WriteString(ui.SectionLabel("Tổng quan nhanh"))
	side.WriteString("\n\n")
	side.WriteString(ui.KeyValue("Tệp đã quét", fmt.Sprintf("%d", results.Project.FilesScanned)))
	side.WriteString("\n")
	side.WriteString(ui.KeyValue("Dòng mã", fmt.Sprintf("%d", results.Project.LinesOfCode)))
	side.WriteString("\n")
	side.WriteString(ui.KeyValue("Tổng lỗi", fmt.Sprintf("%d", results.Summary.Total)))

	fmt.Println(ui.RenderScreen(
		"BÁO CÁO QUÉT MÃ NGUỒN",
		"Tổng hợp sức khỏe mã nguồn và các điểm cần ưu tiên xử lý",
		main.String(),
		side.String(),
		[]string{"Chạy vibescanner serve để mở bảng điều khiển", "Dùng --report html để xuất báo cáo web"},
	))
	fmt.Println()
}

// printHealthScore in điểm sức khỏe
func printHealthScore(score models.HealthScore) {
	status, desc := aggregation.GetHealthStatus(score.Overall)
	var main strings.Builder
	main.WriteString(ui.SectionLabel(fmt.Sprintf("Điểm sức khỏe tổng quát: %d/100", score.Overall)))
	main.WriteString("\n")
	main.WriteString(ui.Muted(desc))
	main.WriteString("\n\n")
	main.WriteString(ui.KeyValue("Trạng thái", status))
	main.WriteString("\n")
	main.WriteString(ui.KeyValue("Bảo mật", fmt.Sprintf("%d/100", score.Security)))
	main.WriteString("\n")
	main.WriteString(ui.KeyValue("Chất lượng", fmt.Sprintf("%d/100", score.Quality)))
	main.WriteString("\n")
	main.WriteString(ui.KeyValue("Kiến trúc", fmt.Sprintf("%d/100", score.Architecture)))
	main.WriteString("\n")
	main.WriteString(ui.KeyValue("Hiệu năng", fmt.Sprintf("%d/100", score.Performance)))

	categories := []struct {
		name  string
		emoji string
		score int
	}{
		{"Bảo mật", "🔴", score.Security},
		{"Chất lượng", "🟡", score.Quality},
		{"Kiến trúc", "🟡", score.Architecture},
		{"Hiệu năng", "🟢", score.Performance},
	}

	var side strings.Builder
	side.WriteString(ui.SectionLabel("Phân loại theo hạng mục"))
	side.WriteString("\n\n")
	for _, cat := range categories {
		side.WriteString(fmt.Sprintf("%s %-10s %3d/100  %s\n", cat.emoji, cat.name, cat.score, getScoreStatus(cat.score)))
	}

	fmt.Println(ui.RenderScreen(
		"ĐIỂM SỨC KHỎE MÃ NGUỒN",
		"Đánh giá nhanh theo từng nhóm chất lượng quan trọng",
		main.String(),
		strings.TrimSpace(side.String()),
		[]string{"Ưu tiên xử lý mục có điểm thấp trước"},
	))
	fmt.Println()
}

// printSummary in tóm tắt findings
func printSummary(summary models.ScanSummary) {
	fmt.Println(ui.GetInfoBox("Tóm tắt phát hiện: " + formatSummary(summary)))
	fmt.Println()
}

// printCriticalFindings in các findings critical/high
func printCriticalFindings(results *models.ScanResult) {
	critical := results.FindingsBySeverity(models.Critical)
	high := results.FindingsBySeverity(models.High)

	if len(critical) > 0 {
		fmt.Println(ui.GetErrorBox("Mức CRITICAL: cần xử lý ngay trước khi triển khai"))
		fmt.Println()

		for i, f := range critical {
			if i >= 5 { // Show max 5
				fmt.Printf("  ... và %d vấn đề critical khác\n", len(critical)-5)
				break
			}
			printFinding(f)
		}
	}

	if len(high) > 0 {
		fmt.Println(ui.GetInfoBox("Mức HIGH: nên lên kế hoạch xử lý trong tuần này"))
		fmt.Println()

		for i, f := range high {
			if i >= 5 { // Show max 5
				fmt.Printf("  ... và %d vấn đề high khác\n\n", len(high)-5)
				break
			}
			printFinding(f)
		}
	}
}

// printFinding in một finding
func printFinding(f models.Finding) {
	var body strings.Builder
	body.WriteString(ui.KeyValue("Mã lỗi", f.ID))
	body.WriteString("\n")
	body.WriteString(ui.KeyValue("Mức độ", aggregation.GetSeverityLabel(f.Severity)))
	body.WriteString("\n")
	body.WriteString(ui.KeyValue("Vị trí", fmt.Sprintf("%s:%d", f.File, f.Line)))
	body.WriteString("\n\n")
	body.WriteString(f.Message)

	if f.CodeSnippet != "" {
		body.WriteString("\n\n")
		body.WriteString(ui.SectionLabel("Đoạn mã liên quan"))
		body.WriteString("\n")
		lines := splitLines(f.CodeSnippet, 5) // Max 5 lines
		for _, line := range lines {
			body.WriteString("  │ ")
			body.WriteString(line)
			body.WriteString("\n")
		}
	}

	fmt.Println(ui.GetBorderedBox(strings.TrimSpace(body.String()), fmt.Sprintf("CHI TIẾT PHÁT HIỆN %s", f.ID)))
	fmt.Println()
}

// printFooter in footer
func printFooter(results *models.ScanResult) {
	fmt.Println()
	fmt.Println(ui.GetSuccessBox(fmt.Sprintf("Quét hoàn tất: %d tệp, %d phát hiện", results.Project.FilesScanned, results.Summary.Total)))
	fmt.Println()
	fmt.Println(ui.GetInfoBox("Dùng --report html để mở báo cáo dạng web"))
	fmt.Println(ui.GetInfoBox("Dùng vibescanner serve để xem bảng điều khiển"))
	fmt.Println()
}

// formatSummary format summary thành string
func formatSummary(s models.ScanSummary) string {
	parts := []string{}
	if s.Critical > 0 {
		parts = append(parts, color.RedString("%d Critical", s.Critical))
	}
	if s.High > 0 {
		parts = append(parts, color.YellowString("%d High", s.High))
	}
	if s.Medium > 0 {
		parts = append(parts, color.BlueString("%d Medium", s.Medium))
	}
	if s.Low > 0 {
		parts = append(parts, fmt.Sprintf("%d Low", s.Low))
	}
	if s.Info > 0 {
		parts = append(parts, fmt.Sprintf("%d Info", s.Info))
	}

	if len(parts) == 0 {
		return "Không phát hiện vấn đề nào 🎉"
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " · "
		}
		result += part
	}
	return result
}

// getScoreStatus trả về status text cho score
func getScoreStatus(score int) string {
	switch {
	case score >= 80:
		return "Tốt"
	case score >= 60:
		return "Trung bình"
	case score >= 40:
		return "Cần cải thiện"
	default:
		return "Nguy hiểm"
	}
}

// splitLines chia string thành lines, giới hạn số lượng
func splitLines(s string, maxLines int) []string {
	var lines []string
	start := 0
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
			count++
			if count >= maxLines {
				break
			}
		}
	}
	if start < len(s) && count < maxLines {
		lines = append(lines, s[start:])
	}
	return lines
}

// WriteJSON ghi kết quả ra file JSON
func WriteJSON(results *models.ScanResult) error {
	filename := fmt.Sprintf("vibescanner-report-%s.json", time.Now().Format("20060102-150405"))

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("lỗi marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("lỗi ghi file: %w", err)
	}

	fmt.Printf("✅ Đã lưu báo cáo JSON: %s\n", filename)
	return nil
}

// openBrowser mở browser với URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

// PrintSuccess prints a success message with checkmark
func PrintSuccess(format string, args ...interface{}) {
	fmt.Println(ui.GetSuccessBox(fmt.Sprintf(format, args...)))
}

// PrintError prints an error message
func PrintError(format string, args ...interface{}) {
	fmt.Println(ui.GetErrorBox(fmt.Sprintf(format, args...)))
}

// PrintInfo prints an info message
func PrintInfo(format string, args ...interface{}) {
	fmt.Println(ui.GetInfoBox(fmt.Sprintf(format, args...)))
}
