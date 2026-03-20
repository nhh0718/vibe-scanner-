package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// ExportMarkdown exports scan results to a Markdown file
func ExportMarkdown(results *models.ScanResult, outputPath string) error {
	if outputPath == "" {
		timestamp := time.Now().Format("20060102-150405")
		projectName := getRepoName(results.Project.Path)
		outputPath = fmt.Sprintf("vibescanner-report-%s-%s.md", timestamp, sanitizeFilename(projectName))
	}

	markdown := generateMarkdownContent(results)

	if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("không thể ghi file Markdown: %w", err)
	}

	absPath, _ := filepath.Abs(outputPath)
	fmt.Printf("✅ Đã xuất báo cáo Markdown: %s\n", absPath)
	return nil
}

// ExportPDF exports scan results to a PDF file
func ExportPDF(results *models.ScanResult, outputPath string) error {
	if outputPath == "" {
		timestamp := time.Now().Format("20060102-150405")
		projectName := getRepoName(results.Project.Path)
		outputPath = fmt.Sprintf("vibescanner-report-%s-%s.pdf", timestamp, sanitizeFilename(projectName))
	}

	htmlContent := generateHTMLContent(results)
	htmlPath := strings.TrimSuffix(outputPath, ".pdf") + ".html"

	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("không thể ghi file HTML: %w", err)
	}

	markdownContent := generateMarkdownContent(results)
	mdPath := strings.TrimSuffix(outputPath, ".pdf") + ".md"
	os.WriteFile(mdPath, []byte(markdownContent), 0644)

	absPath, _ := filepath.Abs(htmlPath)
	fmt.Printf("✅ Đã tạo báo cáo HTML (mở bằng browser để in PDF): %s\n", absPath)
	fmt.Printf("📄 Hoặc dùng file Markdown: %s\n", mdPath)
	fmt.Println("💡 Tip: Mở file HTML trong Chrome và nhấn Ctrl+P để lưu thành PDF")

	return nil
}

// generateMarkdownContent creates the Markdown report content
func generateMarkdownContent(results *models.ScanResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# 🔍 VibeScanner Report\n\n"))

	sb.WriteString(fmt.Sprintf("## 📋 Thông tin chung\n\n"))
	sb.WriteString(fmt.Sprintf("| Thuộc tính | Giá trị |\n"))
	sb.WriteString(fmt.Sprintf("|------------|---------|\n"))
	sb.WriteString(fmt.Sprintf("| Project | %s |\n", results.Project.Path))
	sb.WriteString(fmt.Sprintf("| Scan ID | %s |\n", results.ScanID))
	sb.WriteString(fmt.Sprintf("| Thời gian | %s |\n", results.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("| Thời lượng | %s |\n", results.Duration))
	sb.WriteString(fmt.Sprintf("| Files | %d |\n", results.Project.FilesScanned))
	sb.WriteString(fmt.Sprintf("| Dòng code | %d |\n", results.Project.LinesOfCode))
	sb.WriteString(fmt.Sprintf("| Health Score | %d/100 |\n", results.HealthScore.Overall))
	sb.WriteString(fmt.Sprintf("\n"))

	sb.WriteString(fmt.Sprintf("## 📊 Tổng quan vấn đề\n\n"))
	sb.WriteString(fmt.Sprintf("| Mức độ | Số lượng |\n"))
	sb.WriteString(fmt.Sprintf("|--------|----------|\n"))
	sb.WriteString(fmt.Sprintf("| 🔴 Critical | %d |\n", results.Summary.Critical))
	sb.WriteString(fmt.Sprintf("| 🟠 High | %d |\n", results.Summary.High))
	sb.WriteString(fmt.Sprintf("| 🟡 Medium | %d |\n", results.Summary.Medium))
	sb.WriteString(fmt.Sprintf("| 🔵 Low | %d |\n", results.Summary.Low))
	sb.WriteString(fmt.Sprintf("| ⚪ Info | %d |\n", results.Summary.Info))
	sb.WriteString(fmt.Sprintf("| **Tổng** | **%d** |\n\n", results.Summary.Total))

	sb.WriteString(fmt.Sprintf("## 📂 Phân loại theo danh mục\n\n"))
	sb.WriteString(fmt.Sprintf("| Danh mục | Số lượng |\n"))
	sb.WriteString(fmt.Sprintf("|----------|----------|\n"))
	categoryCount := make(map[string]int)
	for _, f := range results.Findings {
		categoryCount[string(f.Category)]++
	}
	for category, count := range categoryCount {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", category, count))
	}
	sb.WriteString(fmt.Sprintf("\n"))

	if len(results.Findings) > 0 {
		sb.WriteString(fmt.Sprintf("## 🔍 Chi tiết phát hiện\n\n"))
		for i, finding := range results.Findings {
			severityEmoji := getSeverityEmoji(finding.Severity)
			sb.WriteString(fmt.Sprintf("### %s Phát hiện #%d: %s\n\n", severityEmoji, i+1, finding.Title))
			sb.WriteString(fmt.Sprintf("**Mã lỗi:** `%s`\n\n", finding.ID))
			sb.WriteString(fmt.Sprintf("**Mức độ:** %s\n\n", severityEmoji+" "+severityString(finding.Severity)))
			sb.WriteString(fmt.Sprintf("**Nguồn:** %s\n\n", getEngineLabel(finding.Engine)))
			sb.WriteString(fmt.Sprintf("**Danh mục:** %s\n\n", finding.Category))
			sb.WriteString(fmt.Sprintf("**Mô tả:** %s\n\n", finding.Message))
			sb.WriteString(fmt.Sprintf("**Vị trí:** `%s:%d`\n\n", finding.File, finding.Line))

			if finding.CodeSnippet != "" {
				sb.WriteString(fmt.Sprintf("**Đoạn mã:**\n```\n%s\n```\n\n", finding.CodeSnippet))
			}

			if finding.FixSuggestion != "" || finding.FixCode != "" {
				fix := finding.FixSuggestion
				if fix == "" {
					fix = finding.FixCode
				}
				sb.WriteString(fmt.Sprintf("**Cách khắc phục:**\n\n%s\n\n", fix))
			}

			if len(finding.References) > 0 {
				sb.WriteString(fmt.Sprintf("**Tham khảo:** %s\n\n", strings.Join(finding.References, ", ")))
			}

			sb.WriteString(fmt.Sprintf("---\n\n"))
		}
	}

	sb.WriteString(fmt.Sprintf("\n---\n\n"))
	sb.WriteString(fmt.Sprintf("*Báo cáo được tạo bởi [VibeScanner](https://github.com/nhh0718/vibe-scanner-)*\n"))
	sb.WriteString(fmt.Sprintf("*Thời gian tạo: %s*\n", time.Now().Format("2006-01-02 15:04:05")))

	return sb.String()
}

func getSeverityEmoji(severity models.Severity) string {
	switch severity {
	case models.Critical:
		return "🔴"
	case models.High:
		return "🟠"
	case models.Medium:
		return "🟡"
	case models.Low:
		return "🔵"
	default:
		return "⚪"
	}
}

func severityString(severity models.Severity) string {
	switch severity {
	case models.Critical:
		return "Critical"
	case models.High:
		return "High"
	case models.Medium:
		return "Medium"
	case models.Low:
		return "Low"
	default:
		return "Info"
	}
}

// GenerateCombinedReport generates a comprehensive report with all formats
func GenerateCombinedReport(results *models.ScanResult, basePath string) error {
	timestamp := time.Now().Format("20060102-150405")
	projectName := getRepoName(results.Project.Path)

	if basePath == "" {
		basePath = fmt.Sprintf("vibescanner-report-%s-%s", timestamp, sanitizeFilename(projectName))
	}

	reportDir := basePath + "-full-report"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục: %w", err)
	}

	jsonPath := filepath.Join(reportDir, "report.json")
	jsonData, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile(jsonPath, jsonData, 0644)

	htmlPath := filepath.Join(reportDir, "report.html")
	htmlContent := generateHTMLContent(results)
	os.WriteFile(htmlPath, []byte(htmlContent), 0644)

	mdPath := filepath.Join(reportDir, "report.md")
	mdContent := generateMarkdownContent(results)
	os.WriteFile(mdPath, []byte(mdContent), 0644)

	summaryPath := filepath.Join(reportDir, "SUMMARY.txt")
	summaryContent := generateSummaryText(results, reportDir)
	os.WriteFile(summaryPath, []byte(summaryContent), 0644)

	absPath, _ := filepath.Abs(reportDir)
	fmt.Printf("✅ Đã tạo báo cáo đầy đủ tại: %s\n", absPath)
	fmt.Printf("   - JSON: %s\n", jsonPath)
	fmt.Printf("   - HTML: %s\n", htmlPath)
	fmt.Printf("   - Markdown: %s\n", mdPath)
	fmt.Printf("   - Summary: %s\n", summaryPath)

	return nil
}

func generateSummaryText(results *models.ScanResult, reportDir string) string {
	var sb strings.Builder

	sb.WriteString("╔══════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║           VIBESCANNER COMPREHENSIVE REPORT                   ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════╝\n\n")

	sb.WriteString(fmt.Sprintf("Project: %s\n", results.Project.Path))
	sb.WriteString(fmt.Sprintf("Scan ID: %s\n", results.ScanID))
	sb.WriteString(fmt.Sprintf("Time: %s\n", results.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Duration: %s\n\n", results.Duration))

	sb.WriteString("──────────────────────────────────────────────────────────────\n")
	sb.WriteString("SCAN STATISTICS\n")
	sb.WriteString("──────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Files Scanned:     %d\n", results.Project.FilesScanned))
	sb.WriteString(fmt.Sprintf("Lines of Code:     %d\n", results.Project.LinesOfCode))
	sb.WriteString(fmt.Sprintf("Health Score:      %d/100\n", results.HealthScore.Overall))
	sb.WriteString(fmt.Sprintf("Total Findings:    %d\n\n", results.Summary.Total))

	sb.WriteString("──────────────────────────────────────────────────────────────\n")
	sb.WriteString("FINDINGS BY SEVERITY\n")
	sb.WriteString("──────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("🔴 Critical:       %d\n", results.Summary.Critical))
	sb.WriteString(fmt.Sprintf("🟠 High:           %d\n", results.Summary.High))
	sb.WriteString(fmt.Sprintf("🟡 Medium:         %d\n", results.Summary.Medium))
	sb.WriteString(fmt.Sprintf("🔵 Low:            %d\n", results.Summary.Low))
	sb.WriteString(fmt.Sprintf("⚪ Info:           %d\n\n", results.Summary.Info))

	sb.WriteString("──────────────────────────────────────────────────────────────\n")
	sb.WriteString("FILES IN THIS REPORT\n")
	sb.WriteString("──────────────────────────────────────────────────────────────\n")
	sb.WriteString("📄 report.json  - Full scan data in JSON format\n")
	sb.WriteString("📄 report.html  - Interactive HTML report\n")
	sb.WriteString("📄 report.md    - Markdown report for documentation\n")
	sb.WriteString("📄 SUMMARY.txt  - This summary file\n\n")

	sb.WriteString("──────────────────────────────────────────────────────────────\n")
	sb.WriteString("RECOMMENDATIONS\n")
	sb.WriteString("──────────────────────────────────────────────────────────────\n")
	if results.Summary.Critical > 0 {
		sb.WriteString("⚠️  URGENT: Address critical findings immediately!\n")
	}
	if results.Summary.High > 0 {
		sb.WriteString("⚠️  HIGH: Prioritize high severity findings\n")
	}
	if results.HealthScore.Overall < 70 {
		sb.WriteString("📉 Health score below 70 - consider code review\n")
	}
	sb.WriteString("\n")

	sb.WriteString("══════════════════════════════════════════════════════════════\n")
	sb.WriteString("Generated by VibeScanner - https://github.com/nhh0718/vibe-scanner-\n")
	sb.WriteString("══════════════════════════════════════════════════════════════\n")

	return sb.String()
}
