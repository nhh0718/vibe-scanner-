package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/nhh0718/vibe-scanner-/internal/engines"
	"github.com/nhh0718/vibe-scanner-/internal/output"
	"github.com/spf13/cobra"
)

var (
	reportFormat string
	openBrowser  bool
	disableAI    bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Quét và phân tích codebase",
	Long:  `Quét toàn bộ codebase tại đường dẫn được chỉ định và xuất báo cáo phân tích.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return performScan(args[0], reportFormat, openBrowser)
	},
}

// performScan thực hiện quét và output kết quả
func performScan(path, format string, autoOpen bool) error {
	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("không thể resolve path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path không tồn tại: %s", absPath)
	}

	// Print scan header
	printScanHeader(absPath)

	// Run scan
	results, err := engines.ScanProject(absPath)
	if err != nil {
		return fmt.Errorf("lỗi khi quét: %w", err)
	}

	// Output results
	switch format {
	case "html":
		return output.GenerateHTML(results, absPath, autoOpen)
	case "json":
		return output.WriteJSON(results)
	case "pdf":
		return output.GeneratePDF(results, absPath)
	default:
		output.PrintTerminal(results)
	}

	// Save to reports directory with timestamp and project name
	reportInfo, err := output.SaveScanReport(results)
	if err != nil {
		fmt.Printf("⚠️ Không thể lưu báo cáo: %v\n", err)
	} else {
		fmt.Printf("✅ Đã lưu báo cáo: %s\n", reportInfo.Filename)
	}

	return nil
}

func init() {
	scanCmd.Flags().StringVar(&reportFormat, "report", "terminal", "Định dạng báo cáo: terminal|html|json|pdf")
	scanCmd.Flags().BoolVar(&openBrowser, "open", true, "Tự động mở browser sau khi tạo HTML report")
	scanCmd.Flags().BoolVar(&disableAI, "no-ai", false, "Tắt tính năng AI explanation")
	rootCmd.AddCommand(scanCmd)
}

func printScanHeader(path string) {
	cyan := color.New(color.FgCyan, color.Bold)
	fmt.Println()
	cyan.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	cyan.Println("  🔍 VibeScanner — Đang quét codebase...")
	fmt.Printf("  📁 %s\n", path)
	fmt.Println()
	cyan.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}
