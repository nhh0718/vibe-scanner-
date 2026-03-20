package cmd

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nhh0718/vibe-scanner-/internal/output"
	"github.com/nhh0718/vibe-scanner-/internal/ui"
	"github.com/spf13/cobra"
)

// MenuItem represents a menu option
type MenuItem struct {
	Number      int
	Title       string
	Description string
	Command     string
}

func (i MenuItem) FilterValue() string { return i.Title }

// menuDelegate for custom list rendering
type menuDelegate struct{}

func (d menuDelegate) Height() int                             { return 2 }
func (d menuDelegate) Spacing() int                           { return 0 }
func (d menuDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d menuDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(MenuItem)
	if !ok {
		return
	}

	fmt.Fprint(w, ui.NumberedLine(i.Number, i.Title, i.Description, index == m.Index()))
}

// Model for bubbletea
type model struct {
	list     list.Model
	choice   string
	quitting bool
	width    int
	height   int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		layout := ui.GetScreenLayout(msg.Width, true)
		if layout.MainWidth > 2 {
			m.list.SetWidth(layout.MainWidth - 2)
		}
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(MenuItem)
			if ok {
				m.choice = i.Command
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		if m.choice != "" {
			return ""
		}
		return ui.GetSuccessBox("Tạm biệt! Hẹn gặp lại.") + "\n"
	}

	osName := getOSName()
	ver := version
	if ver == "" {
		ver = "dev"
	}

	var sb strings.Builder
	sb.WriteString(ui.GetLogo())
	sb.WriteString("\n")
	sb.WriteString(ui.SuccessText("TRUNG TÂM ĐIỀU KHIỂN"))
	sb.WriteString(" | ")
	sb.WriteString(ui.Muted(fmt.Sprintf("v%s | %s", ver, osName)))
	sb.WriteString("\n")
	sb.WriteString(m.list.View())
	sb.WriteString("\n")
	sb.WriteString(ui.Muted("↑/↓ chọn  •  Enter  •  Q thoát"))

	return sb.String()
}

// interactiveCmd opens the TUI menu
var interactiveCmd = &cobra.Command{
	Use:    "menu",
	Short:  "Mở interactive menu",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractiveMenu()
	},
}

func runInteractiveMenu() error {
	for {
		items := []list.Item{
			MenuItem{
				Number:      1,
				Title:       "Quét dự án",
				Description: "Phân tích codebase tìm lỗi bảo mật và chất lượng",
				Command:     "scan",
			},
			MenuItem{
				Number:      2,
				Title:       "Lịch sử & Báo cáo",
				Description: "Xem và quản lý các lần quét đã lưu",
				Command:     "history",
			},
			MenuItem{
				Number:      3,
				Title:       "Xem bảng điều khiển",
				Description: "Mở giao diện web hiển thị kết quả quét",
				Command:     "serve",
			},
			MenuItem{
				Number:      4,
				Title:       "Quản lý tác vụ",
				Description: "Theo dõi và quản lý các tác vụ đang chạy",
				Command:     "tasks",
			},
			MenuItem{
				Number:      5,
				Title:       "Cài đặt AI",
				Description: "Thiết lập và quản lý các mô hình AI",
				Command:     "ai-setup",
			},
			MenuItem{
				Number:      6,
				Title:       "Cấu hình hệ thống",
				Description: "Xem và chỉnh sửa các thiết lập",
				Command:     "config",
			},
			MenuItem{
				Number:      7,
				Title:       "Cài đặt toàn cục",
				Description: "Thêm công cụ vào PATH hệ thống",
				Command:     "install",
			},
			MenuItem{
				Number:      8,
				Title:       "Trợ giúp",
				Description: "Hiển thị hướng dẫn sử dụng",
				Command:     "help",
			},
		}

		layout := ui.GetScreenLayout(0, true)
		l := list.New(items, menuDelegate{}, layout.MainWidth-2, 18)
		l.Title = ""
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.SetShowHelp(false)
		l.SetShowPagination(false)

		m := model{list: l}
		p := tea.NewProgram(m)
		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		m = finalModel.(model)
		if m.choice == "" {
			// User pressed q/esc - exit
			return nil
		}

		// Execute the chosen command
		switch m.choice {
		case "scan":
			fmt.Println(ui.GetInfoBox("Nhập hoặc kéo thả đường dẫn dự án cần quét"))
			fmt.Print("> ")
			var path string
			fmt.Scanln(&path)
			if path == "" {
				path = "."
			}
			runScanInteractive(path)
			fmt.Println("\n" + ui.GetSuccessBox("Quét hoàn tất. Nhấn Enter để quay lại menu."))
			fmt.Scanln()
		case "history":
			runHistoryInteractive()
		case "serve":
			runServeInteractive()
		case "tasks":
			runTasksInteractive()
		case "ai-setup":
			runAISetupInteractive()
		case "config":
			runConfigInteractive()
			fmt.Println("\n" + ui.GetInfoBox("Nhấn Enter để quay lại menu."))
			fmt.Scanln()
		case "install":
			installGlobal()
			fmt.Println("\n" + ui.GetInfoBox("Nhấn Enter để quay lại menu."))
			fmt.Scanln()
		case "help":
			fmt.Println(ui.GetBorderedBox(strings.Join([]string{
				ui.SectionLabel("Hướng dẫn nhanh"),
				"",
				"vibescanner scan <đường_dẫn>    Quét mã nguồn dự án",
				"vibescanner history             Xem lịch sử và báo cáo",
				"vibescanner serve               Mở bảng điều khiển web",
				"vibescanner tasks               Quản lý tác vụ",
				"vibescanner ai-setup            Thiết lập AI cục bộ",
				"vibescanner config              Quản lý cấu hình",
				"vibescanner install             Cài đặt toàn cục",
			}, "\n"), "TRỢ GIÚP"))
			fmt.Println("\n" + ui.GetInfoBox("Nhấn Enter để quay lại menu."))
			fmt.Scanln()
		}
	}
}

// runScanInteractive runs scan from interactive mode
func runScanInteractive(path string) error {
	fmt.Println("\n" + ui.GetInfoBox(fmt.Sprintf("Đang quét dự án: %s", path)))
	// Call the actual scan logic
	return performScan(path, "terminal", false)
}

// runServeInteractive runs serve from interactive mode
func runServeInteractive() error {
	results, err := output.LoadLastScan()
	if err != nil {
		fmt.Println(ui.GetErrorBox("Không tìm thấy kết quả scan. Hãy chạy scan trước."))
		fmt.Println("\n" + ui.GetInfoBox("Nhấn Enter để quay lại menu."))
		fmt.Scanln()
		return nil
	}

	// Run dashboard - this will block until Ctrl+C
	if err := output.ServeDashboard(results, 7420); err != nil {
		fmt.Println(ui.GetErrorBox(fmt.Sprintf("Lỗi server: %v", err)))
	} else {
		fmt.Println(ui.GetDashboardStoppedBanner())
	}

	fmt.Println(ui.GetInfoBox("Nhấn Enter để quay lại menu."))
	fmt.Scanln()
	return nil
}

// runConfigInteractive runs config from interactive mode
func runConfigInteractive() error {
	return runConfigList()
}

// getOSName returns a friendly OS name
func getOSName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return runtime.GOOS
	}
}
