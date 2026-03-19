package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vibescanner",
	Short: "🔍 Công cụ 'khám bệnh' codebase cho vibe coders",
	Long: `VibeScanner - Công cụ phân tích code kết hợp static analysis và AI
để phát hiện lỗi bảo mật, chất lượng code, và kiến trúc issues.

Chạy hoàn toàn local, code không bao giờ rời máy.`,
	Version: "0.1.0",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand provided, show interactive menu
		return runInteractiveMenu()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
