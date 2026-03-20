package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/nhh0718/vibe-scanner-/internal/output"
	"github.com/nhh0718/vibe-scanner-/internal/ui"
	"github.com/spf13/cobra"
)

// tasksCmd quản lý tác vụ
var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Quản lý tác vụ",
	Long:  `Xem và quản lý các tác vụ scan đang chạy hoặc đã hoàn thành.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTasksInteractive()
	},
}

// runTasksInteractive hiển thị quản lý tác vụ tương tác
func runTasksInteractive() error {
	for {
		// Lấy danh sách báo cáo để hiển thị như lịch sử tác vụ
		reports, err := output.ListScanReports()
		if err != nil {
			fmt.Println(ui.GetErrorBox("Không thể tải danh sách tác vụ: " + err.Error()))
			fmt.Println(ui.GetInfoBox("Nhấn Enter để quay lại menu"))
			fmt.Scanln()
			return nil
		}

		fmt.Println(ui.GetLogo())
		fmt.Println(ui.SuccessText("QUẢN LÝ TÁC VỤ"))
		fmt.Println()

		// Tạo options cho form
		var options []huh.Option[string]
		options = append(options, huh.NewOption("🔄 Quét dự án mới", "scan"))

		if len(reports) > 0 {
			options = append(options, huh.NewOption("📊 Xem báo cáo gần nhất", "latest"))
			options = append(options, huh.NewOption("📚 Duyệt lịch sử quét", "history"))
		}

		options = append(options, huh.NewOption("🌐 Mở dashboard", "dashboard"))
		options = append(options, huh.NewOption("⚙️ Cấu hình", "config"))
		options = append(options, huh.NewOption("← Quay lại", "back"))

		var choice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Chọn tác vụ").
					Options(options...).
					Value(&choice),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		switch choice {
		case "scan":
			fmt.Print("Nhập đường dẫn dự án: ")
			var path string
			fmt.Scanln(&path)
			if path == "" {
				path = "."
			}
			fmt.Println("\n" + ui.GetInfoBox(fmt.Sprintf("Đang quét: %s", path)))
			runScanInteractive(path)
			fmt.Println("\n" + ui.GetSuccessBox("Quét hoàn tất!"))
			fmt.Println(ui.GetInfoBox("Nhấn Enter để tiếp tục"))
			fmt.Scanln()

		case "latest":
			results, err := output.LoadLastScan()
			if err != nil {
				fmt.Println(ui.GetErrorBox("Không tìm thấy báo cáo: " + err.Error()))
			} else {
				output.PrintTerminal(results)
			}
			fmt.Println(ui.GetInfoBox("Nhấn Enter để tiếp tục"))
			fmt.Scanln()

		case "history":
			return runHistoryInteractive()

		case "dashboard":
			return runServeInteractive()

		case "config":
			runConfigInteractive()
			fmt.Println(ui.GetInfoBox("Nhấn Enter để tiếp tục"))
			fmt.Scanln()

		case "back":
			return nil
		}
	}
}

// runHistoryInteractive hiển thị lịch sử quét tương tác
func runHistoryInteractive() error {
	for {
		reports, err := output.ListScanReports()
		if err != nil {
			fmt.Println(ui.GetErrorBox("Chưa có báo cáo nào: " + err.Error()))
			fmt.Println(ui.GetInfoBox("Nhấn Enter để quay lại"))
			fmt.Scanln()
			return nil
		}

		fmt.Println(ui.GetLogo())
		fmt.Println(ui.SuccessText("LỊCH SỬ & BÁO CÁO"))
		fmt.Printf("Tìm thấy %d báo cáo\n\n", len(reports))

		// Tạo options từ danh sách báo cáo
		var options []huh.Option[string]

		for i, report := range reports {
			if i >= 15 { // Giới hạn hiển thị 15 báo cáo gần nhất
				break
			}

			timeStr := report.Timestamp.Format("02/01 15:04")
			scoreStr := fmt.Sprintf("%d%%", report.HealthScore)

			// Color code health score
			if report.HealthScore >= 80 {
				scoreStr = color.GreenString(scoreStr)
			} else if report.HealthScore >= 50 {
				scoreStr = color.YellowString(scoreStr)
			} else {
				scoreStr = color.RedString(scoreStr)
			}

			label := fmt.Sprintf("%s | %s | %s | %d issues | Score: %s",
				timeStr,
				truncateString(report.ProjectName, 20),
				truncateString(report.ScanID, 8),
				report.FindingCount,
				scoreStr,
			)

			options = append(options, huh.NewOption(label, report.ScanID))
		}

		options = append(options, huh.NewOption("← Quay lại menu", "back"))

		var selectedID string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Chọn báo cáo để xem chi tiết").
					Options(options...).
					Value(&selectedID),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if selectedID == "back" {
			return nil
		}

		// Hiển thị menu con cho báo cáo được chọn
		var action string
		actionForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Hành động").
					Options(
						huh.NewOption("👁️ Xem chi tiết", "view"),
						huh.NewOption("📄 Xuất HTML", "html"),
						huh.NewOption("📝 Xuất Markdown", "md"),
						huh.NewOption("🗑️ Xóa báo cáo", "delete"),
						huh.NewOption("← Quay lại", "back"),
					).
					Value(&action),
			),
		)

		if err := actionForm.Run(); err != nil {
			return err
		}

		switch action {
		case "view":
			results, err := output.LoadScanReport(selectedID)
			if err != nil {
				fmt.Println(ui.GetErrorBox("Không thể tải báo cáo: " + err.Error()))
			} else {
				output.PrintTerminal(results)
			}
			fmt.Println(ui.GetInfoBox("Nhấn Enter để tiếp tục"))
			fmt.Scanln()

		case "html":
			results, err := output.LoadScanReport(selectedID)
			if err != nil {
				fmt.Println(ui.GetErrorBox("Không thể tải báo cáo: " + err.Error()))
			} else {
				err = output.GenerateHTML(results, results.Project.Path, true)
				if err != nil {
					fmt.Println(ui.GetErrorBox("Lỗi: " + err.Error()))
				}
			}
			fmt.Println(ui.GetInfoBox("Nhấn Enter để tiếp tục"))
			fmt.Scanln()

		case "md":
			results, err := output.LoadScanReport(selectedID)
			if err != nil {
				fmt.Println(ui.GetErrorBox("Không thể tải báo cáo: " + err.Error()))
			} else {
				err = output.ExportMarkdown(results, "")
				if err != nil {
					fmt.Println(ui.GetErrorBox("Lỗi: " + err.Error()))
				}
			}
			fmt.Println(ui.GetInfoBox("Nhấn Enter để tiếp tục"))
			fmt.Scanln()

		case "delete":
			var confirm bool
			confirmForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Bạn có chắc muốn xóa báo cáo này?").
						Value(&confirm),
				),
			)
			if err := confirmForm.Run(); err != nil {
				return err
			}

			if confirm {
				if err := output.DeleteScanReport(selectedID); err != nil {
					fmt.Println(ui.GetErrorBox("Không thể xóa: " + err.Error()))
				} else {
					fmt.Println(ui.GetSuccessBox("Đã xóa báo cáo!"))
				}
			}
			fmt.Println(ui.GetInfoBox("Nhấn Enter để tiếp tục"))
			fmt.Scanln()

		case "back":
			// Tiếp tục vòng lặp để chọn báo cáo khác
		}
	}
}

func init() {
	rootCmd.AddCommand(tasksCmd)
}
