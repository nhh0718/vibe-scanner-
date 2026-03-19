package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vibescanner/vibescanner/internal/ai"
	"github.com/vibescanner/vibescanner/internal/output"
)

// aiSetupCmd represents the ai-setup command
var aiSetupCmd = &cobra.Command{
	Use:   "ai-setup",
	Short: "Quản lý AI models và Ollama",
	Long: `Quản lý Ollama runtime và AI models.

Các subcommands:
  list     - Liệt kê models đã cài
  install  - Cài đặt model mới
  remove   - Gỡ bỏ model
  status   - Kiểm tra trạng thái Ollama`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default action: show status
		return runAIStatus()
	},
}

// aiStatusCmd checks Ollama status
var aiStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Kiểm tra trạng thái Ollama",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAIStatus()
	},
}

func runAIStatus() error {
	fmt.Println("🤖 VibeScanner AI Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if ai.IsOllamaAvailable() {
		output.PrintSuccess("Ollama đang chạy")
		models, err := ai.ListInstalledModels()
		if err != nil {
			output.PrintError("Không thể liệt kê models: %v", err)
		} else if len(models) > 0 {
			fmt.Println("\n📦 Models đã cài:")
			for _, m := range models {
				fmt.Printf("  • %s\n", m)
			}
		} else {
			fmt.Println("\n⚠️ Chưa có model nào được cài đặt")
		}
	} else {
		output.PrintError("Ollama chưa được cài đặt hoặc không chạy")
		fmt.Println("\n📥 Cài đặt Ollama:")
		fmt.Println("   1. Truy cập: https://ollama.ai/download")
	}
	return nil
}

// aiListCmd lists installed models
var aiListCmd = &cobra.Command{
	Use:   "list",
	Short: "Liệt kê các models đã cài đặt",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !ai.IsOllamaAvailable() {
			return fmt.Errorf("Ollama chưa chạy")
		}
		models, err := ai.ListInstalledModels()
		if err != nil {
			return fmt.Errorf("lỗi khi liệt kê models: %w", err)
		}
		if len(models) == 0 {
			fmt.Println("📭 Chưa có model nào được cài đặt.")
			return nil
		}
		fmt.Println("📦 Models đã cài đặt:")
		for i, m := range models {
			fmt.Printf("  %d. %s\n", i+1, m)
		}
		return nil
	},
}

// aiInstallCmd installs a model
var aiInstallCmd = &cobra.Command{
	Use:   "install [model]",
	Short: "Cài đặt AI model",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		model := "qwen2.5-coder:3b"
		if len(args) > 0 {
			model = args[0]
		}
		if !ai.IsOllamaAvailable() {
			fmt.Println("📥 Cài đặt Ollama trước...")
			if err := ai.DownloadOllama(); err != nil {
				return fmt.Errorf("lỗi khi tải Ollama: %w", err)
			}
		}
		fmt.Printf("📥 Đang tải model %s...\n", model)
		if err := ai.PullModel(model); err != nil {
			return fmt.Errorf("lỗi khi tải model: %w", err)
		}
		output.PrintSuccess("Đã cài đặt model %s", model)
		return nil
	},
}

// aiRemoveCmd removes a model
var aiRemoveCmd = &cobra.Command{
	Use:   "remove [model]",
	Short: "Gỡ bỏ AI model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		model := args[0]
		if !ai.IsOllamaAvailable() {
			return fmt.Errorf("Ollama chưa chạy")
		}
		fmt.Printf("🗑️  Đang gỡ bỏ model %s...\n", model)
		if err := ai.RemoveModel(model); err != nil {
			return fmt.Errorf("lỗi khi gỡ model: %w", err)
		}
		output.PrintSuccess("Đã gỡ bỏ model %s", model)
		return nil
	},
}

func init() {
	aiSetupCmd.AddCommand(aiStatusCmd)
	aiSetupCmd.AddCommand(aiListCmd)
	aiSetupCmd.AddCommand(aiInstallCmd)
	aiSetupCmd.AddCommand(aiRemoveCmd)
	rootCmd.AddCommand(aiSetupCmd)
}
