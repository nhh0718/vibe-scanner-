package cmd

import (
	"fmt"
	"strings"

	"github.com/nhh0718/vibe-scanner-/internal/ai"
	"github.com/nhh0718/vibe-scanner-/internal/ui"
	"github.com/spf13/cobra"
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
		return runAISetupInteractive()
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
	var main strings.Builder
	var side strings.Builder

	if ai.IsOllamaAvailable() {
		main.WriteString(ui.SuccessText("Ollama đang chạy và sẵn sàng xử lý yêu cầu"))
		main.WriteString("\n\n")
		models, err := ai.ListInstalledModels()
		if err != nil {
			main.WriteString(ui.ErrorText(fmt.Sprintf("Không thể liệt kê model: %v", err)))
		} else if len(models) > 0 {
			main.WriteString(ui.SectionLabel("Model đã cài"))
			main.WriteString("\n\n")
			for i, m := range models {
				main.WriteString(fmt.Sprintf("[%02d] %s\n", i+1, m))
			}
		} else {
			main.WriteString(ui.WarningText("Chưa có model nào được cài đặt"))
		}
		side.WriteString(ui.SectionLabel("Tóm tắt"))
		side.WriteString("\n\n")
		side.WriteString(ui.KeyValue("Runtime", "Đang hoạt động"))
	} else {
		main.WriteString(ui.ErrorText("Ollama chưa được cài đặt hoặc chưa chạy"))
		main.WriteString("\n\n")
		main.WriteString(ui.BulletList([]string{
			"Truy cập https://ollama.ai/download để cài đặt.",
			"Hoàn tất cài đặt đúng theo hệ điều hành.",
			"Khởi động bằng lệnh: ollama serve",
		}))
		side.WriteString(ui.SectionLabel("Khuyến nghị"))
		side.WriteString("\n\n")
		side.WriteString(ui.BulletList([]string{"Dùng vibescanner ai-setup để mở giao diện đầy đủ.", "Ưu tiên cài model 3B nếu máy tầm trung."}))
	}

	fmt.Println(ui.RenderScreen(
		"TRẠNG THÁI AI CỤC BỘ",
		"Kiểm tra nhanh tình trạng Ollama và model đã cài",
		main.String(),
		side.String(),
		[]string{"vibescanner ai-setup mở giao diện đầy đủ", "vibescanner ai-setup list xem danh sách model"},
	))
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
			fmt.Println(ui.GetInfoBox("Chưa có model nào được cài đặt."))
			return nil
		}
		fmt.Println(ui.GetBorderedBox(strings.Join(func() []string {
			lines := []string{ui.SectionLabel("Danh sách model đã cài"), ""}
			for i, m := range models {
				lines = append(lines, fmt.Sprintf("[%02d] %s", i+1, m))
			}
			return lines
		}(), "\n"), "MODEL ĐÃ CÀI"))
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
			fmt.Println(ui.GetInfoBox("Đang cài Ollama trước khi tải model..."))
			if err := ai.DownloadOllama(); err != nil {
				return fmt.Errorf("lỗi khi tải Ollama: %w", err)
			}
		}
		fmt.Println(ui.GetInfoBox(fmt.Sprintf("Đang tải model %s", model)))
		if err := ai.PullModel(model); err != nil {
			return fmt.Errorf("lỗi khi tải model: %w", err)
		}
		fmt.Println(ui.GetSuccessBox(fmt.Sprintf("Đã cài đặt model %s", model)))
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
		fmt.Println(ui.GetInfoBox(fmt.Sprintf("Đang gỡ bỏ model %s", model)))
		if err := ai.RemoveModel(model); err != nil {
			return fmt.Errorf("lỗi khi gỡ model: %w", err)
		}
		fmt.Println(ui.GetSuccessBox(fmt.Sprintf("Đã gỡ bỏ model %s", model)))
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
