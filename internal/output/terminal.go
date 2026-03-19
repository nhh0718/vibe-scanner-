package output

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/nhh0718/vibe-scanner-/internal/aggregation"
	"github.com/nhh0718/vibe-scanner-/internal/models"
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
	cyan := color.New(color.FgCyan, color.Bold)
	_ = cyan
	fmt.Println()
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	color.Cyan("  🔍 VibeScanner — Kết quả khám bệnh")
	fmt.Printf("  Dự án: %s    Thời gian: %s\n", results.Project.Name, results.Duration.Round(time.Second))
	fmt.Println()
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// printHealthScore in điểm sức khỏe
func printHealthScore(score models.HealthScore) {
	status, desc := aggregation.GetHealthStatus(score.Overall)

	fmt.Printf("  🏥 ĐIỂM SỨC KHỎE TỔNG QUÁT: %d/100 %s\n", score.Overall, status)
	fmt.Printf("     %s\n", desc)
	fmt.Println()

	fmt.Println("  ┌─────────────┬─────────┬──────────────────────────┐")
	fmt.Println("  │ Hạng mục   │  Điểm  │ Tình trạng               │")
	fmt.Println("  ├─────────────┼─────────┼──────────────────────────┤")

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

	for _, cat := range categories {
		status := getScoreStatus(cat.score)
		fmt.Printf("  │ %s %-9s │ %3d/100 │ %-24s │\n", cat.emoji, cat.name, cat.score, status)
	}

	fmt.Println("  └─────────────┴─────────┴──────────────────────────┘")
	fmt.Println()
}

// printSummary in tóm tắt findings
func printSummary(summary models.ScanSummary) {
	fmt.Printf("  PHÁT HIỆN: %s\n", formatSummary(summary))
	fmt.Println()
}

// printCriticalFindings in các findings critical/high
func printCriticalFindings(results *models.ScanResult) {
	critical := results.FindingsBySeverity(models.Critical)
	high := results.FindingsBySeverity(models.High)

	if len(critical) > 0 {
		red := color.New(color.FgRed, color.Bold)
		fmt.Println()
		red.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		red.Println("  🚨 CRITICAL — Cần xử lý NGAY trước khi deploy")
		red.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
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
		yellow := color.New(color.FgYellow, color.Bold)
		fmt.Println()
		yellow.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		yellow.Println("  ⚠️  HIGH — Nên xử lý trong tuần này")
		yellow.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
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
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	fmt.Printf("  [%s] %s\n", f.ID, aggregation.GetSeverityLabel(f.Severity))
	fmt.Printf("  File: %s\n", cyan.Sprintf("%s:%d", f.File, f.Line))
	fmt.Println()
	fmt.Printf("  %s\n\n", f.Message)

	if f.CodeSnippet != "" {
		fmt.Println("  Code:")
		lines := splitLines(f.CodeSnippet, 5) // Max 5 lines
		for _, line := range lines {
			fmt.Printf("  %s %s\n", yellow.Sprintf("│"), line)
		}
		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// printFooter in footer
func printFooter(results *models.ScanResult) {
	green := color.New(color.FgGreen)
	fmt.Println()
	green.Println("✅ Quét hoàn tất!")
	fmt.Printf("📊 Tổng cộng: %d files được quét, %d issues phát hiện\n", 
		results.Project.FilesScanned, results.Summary.Total)
	fmt.Println()
	fmt.Println("💡 Chạy với --report html để xem báo cáo chi tiết trong browser")
	fmt.Println("💡 Chạy 'vibescanner serve' để mở dashboard")
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

// LoadLastScan load kết quả scan gần nhất từ file
func LoadLastScan() (*models.ScanResult, error) {
	// TODO: Implement loading from ~/.vibescanner/cache/
	return nil, fmt.Errorf("chưa có kết quả scan nào được lưu")
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
	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✅ ")
	fmt.Printf(format+"\n", args...)
}

// PrintError prints an error message
func PrintError(format string, args ...interface{}) {
	red := color.New(color.FgRed, color.Bold)
	red.Printf("❌ ")
	fmt.Printf(format+"\n", args...)
}

// PrintInfo prints an info message
func PrintInfo(format string, args ...interface{}) {
	blue := color.New(color.FgBlue)
	blue.Printf("ℹ️  ")
	fmt.Printf(format+"\n", args...)
}
