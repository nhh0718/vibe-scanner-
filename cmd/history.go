package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/nhh0718/vibe-scanner-/internal/output"
	"github.com/nhh0718/vibe-scanner-/internal/ui"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Xem lịch sử các lần quét",
	Long:  `Hiển thị danh sách tất cả các lần quét đã thực hiện, cho phép xem lại báo cáo chi tiết.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showScanHistory()
	},
}

// showScanHistory displays all previous scans
func showScanHistory() error {
	reports, err := output.ListScanReports()
	if err != nil {
		return fmt.Errorf("❌ %v", err)
	}

	if len(reports) == 0 {
		fmt.Println("📭 Chưa có lần quét nào được lưu.")
		fmt.Println("💡 Chạy 'vibescanner scan .' để quét project.")
		return nil
	}

	// Header
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D4AA")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("📚 LỊCH SỬ QUÉT"))
	fmt.Printf("Tìm thấy %d báo cáo\n\n", len(reports))

	// Table header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#1E3A5F")).
		Padding(0, 1)

	rowStyle := lipgloss.NewStyle().Padding(0, 1)
	altRowStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color("#1A1A2E"))

	fmt.Println(headerStyle.Render(" #  │ Thời gian          │ Project           │ Files │ Issues │ Score │ ID"))
	fmt.Println(headerStyle.Render("────┼────────────────────┼───────────────────┼───────┼────────┼───────┼────────────────"))

	// Print each report
	for i, report := range reports {
		if i >= 20 { // Show only last 20
			fmt.Printf("\n... và %d báo cáo khác\n", len(reports)-20)
			break
		}

		style := rowStyle
		if i%2 == 1 {
			style = altRowStyle
		}

		timeStr := report.Timestamp.Format("02/01 15:04")
		projectName := truncateString(report.ProjectName, 17)

		// Color code health score
		scoreColor := color.FgGreen
		if report.HealthScore < 70 {
			scoreColor = color.FgYellow
		}
		if report.HealthScore < 50 {
			scoreColor = color.FgRed
		}
		scoreStr := color.New(scoreColor).Sprintf("%3d", report.HealthScore)

		row := fmt.Sprintf(" %2d │ %s │ %-17s │ %5d │ %6d │ %s │ %s",
			i+1,
			timeStr,
			projectName,
			report.FilesScanned,
			report.FindingCount,
			scoreStr,
			truncateString(report.ScanID, 8),
		)
		fmt.Println(style.Render(row))
	}

	// Footer with instructions
	fmt.Println()
	fmt.Println(ui.Subtle("💡 Các lệnh hữu ích:"))
	fmt.Printf("   • %s - Xem chi tiết báo cáo\n", color.CyanString("vibescanner report view <ID>"))
	fmt.Printf("   • %s - Xuất báo cáo HTML\n", color.CyanString("vibescanner report export <ID> --format html"))
	fmt.Printf("   • %s - Xóa báo cáo\n", color.CyanString("vibescanner report delete <ID>"))
	fmt.Println()

	return nil
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Quản lý báo cáo quét",
	Long:  `Xem, xuất và quản lý các báo cáo quét đã lưu.`,
}

var reportViewCmd = &cobra.Command{
	Use:   "view [scan-id|filename]",
	Short: "Xem chi tiết báo cáo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return viewReport(args[0])
	},
}

var reportExportCmd = &cobra.Command{
	Use:   "export [scan-id|filename]",
	Short: "Xuất báo cáo ra file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		return exportReport(args[0], format, output)
	},
}

var reportDeleteCmd = &cobra.Command{
	Use:   "delete [scan-id|filename]",
	Short: "Xóa báo cáo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteReport(args[0])
	},
}

// viewReport displays a specific report
func viewReport(identifier string) error {
	results, err := output.LoadScanReport(identifier)
	if err != nil {
		return err
	}

	// Print the report using terminal output
	output.PrintTerminal(results)
	return nil
}

// exportReport exports a report to specified format
func exportReport(identifier, format, outputPath string) error {
	results, err := output.LoadScanReport(identifier)
	if err != nil {
		return err
	}

	switch format {
	case "html":
		return output.GenerateHTML(results, results.Project.Path, false)
	case "json":
		return output.WriteJSON(results)
	case "markdown", "md":
		return output.ExportMarkdown(results, outputPath)
	case "pdf":
		return output.ExportPDF(results, outputPath)
	case "all", "combined":
		return output.GenerateCombinedReport(results, outputPath)
	default:
		return fmt.Errorf("định dạng không hỗ trợ: %s (hỗ trợ: html, json, md, pdf, all)", format)
	}
}

// deleteReport deletes a specific report
func deleteReport(identifier string) error {
	// Get report info first for confirmation
	reports, err := output.ListScanReports()
	if err != nil {
		return err
	}

	var reportToDelete *output.ScanReportInfo
	for i, report := range reports {
		if report.ScanID == identifier || report.Filename == identifier {
			reportToDelete = &reports[i]
			break
		}
	}

	if reportToDelete == nil {
		return fmt.Errorf("không tìm thấy báo cáo: %s", identifier)
	}

	// Show confirmation
	fmt.Printf("\n🗑️  Bạn có chắc muốn xóa báo cáo này?\n\n")
	fmt.Printf("   Project: %s\n", reportToDelete.ProjectName)
	fmt.Printf("   Time: %s\n", reportToDelete.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Findings: %d\n\n", reportToDelete.FindingCount)
	fmt.Print("Nhập 'yes' để xác nhận: ")

	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println("❌ Đã hủy xóa.")
		return nil
	}

	if err := output.DeleteScanReport(identifier); err != nil {
		return err
	}

	fmt.Printf("✅ Đã xóa báo cáo: %s\n", reportToDelete.ProjectName)
	return nil
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	// Report export flags
	reportExportCmd.Flags().StringP("format", "f", "html", "Định dạng xuất (html|json|md|pdf|all)")
	reportExportCmd.Flags().StringP("output", "o", "", "Đường dẫn file đầu ra")

	// Add subcommands
	reportCmd.AddCommand(reportViewCmd)
	reportCmd.AddCommand(reportExportCmd)
	reportCmd.AddCommand(reportDeleteCmd)

	// Add to root
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(reportCmd)
}
