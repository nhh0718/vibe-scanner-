package cmd

import (
	"fmt"
	"io"
	"runtime"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nhh0718/vibe-scanner-/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Styles
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#60a5fa")).
		Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
		Padding(0, 0, 0, 2)

	selectedItemStyle = lipgloss.NewStyle().
		Padding(0, 0, 0, 1).
		Foreground(lipgloss.Color("#a78bfa")).
		Bold(true)

	descStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94a3b8"))
)

// MenuItem represents a menu option
type MenuItem struct {
	Title       string
	Description string
	Command     string
}

func (i MenuItem) FilterValue() string { return i.Title }

// menuDelegate for custom list rendering
type menuDelegate struct{}

func (d menuDelegate) Height() int                             { return 2 }
func (d menuDelegate) Spacing() int                           { return 1 }
func (d menuDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d menuDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(MenuItem)
	if !ok {
		return
	}

	var str string
	if index == m.Index() {
		str = selectedItemStyle.Render("▸ " + i.Title)
		str += "\n" + descStyle.Render("  "+i.Description)
	} else {
		str = itemStyle.Render(i.Title)
		str += "\n" + descStyle.Render("  "+i.Description)
	}
	fmt.Fprint(w, str)
}

// Model for bubbletea
type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
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
		return "👋 Tạm biệt!\n"
	}

	// Header with OS and version info
	osName := getOSName()
	ver := version
	if ver == "" {
		ver = "dev"
	}

	header := titleStyle.Render("🔍 VibeScanner")
	header += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#64748b")).Render(fmt.Sprintf("v%s • %s • Công cụ khám bệnh codebase", ver, osName)) + "\n\n"

	footer := "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#64748b")).Render("q/esc: thoát • enter: chọn • ↑↓: di chuyển")

	return header + m.list.View() + footer
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
				Title:       "🔍 Scan Project",
				Description: "Quét codebase để tìm lỗi bảo mật và chất lượng",
				Command:     "scan",
			},
			MenuItem{
				Title:       "🌐 Web Dashboard",
				Description: "Mở dashboard trong browser",
				Command:     "serve",
			},
			MenuItem{
				Title:       "🤖 AI Setup",
				Description: "Cài đặt và quản lý AI models",
				Command:     "ai-setup",
			},
			MenuItem{
				Title:       "⚙️  Cấu hình",
				Description: "Xem và chỉnh sửa cấu hình",
				Command:     "config",
			},
			MenuItem{
				Title:       "📦 Cài đặt Global",
				Description: "Thêm VibeScanner vào PATH",
				Command:     "install",
			},
			MenuItem{
				Title:       "❓ Help",
				Description: "Xem hướng dẫn sử dụng",
				Command:     "help",
			},
		}

		l := list.New(items, menuDelegate{}, 60, 14)
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
			fmt.Println("\n📁 Kéo thả thư mục project vào đây hoặc nhập đường dẫn:")
			fmt.Print("> ")
			var path string
			fmt.Scanln(&path)
			if path == "" {
				path = "."
			}
			runScanInteractive(path)
			fmt.Println("\n✅ Scan hoàn tất! Nhấn Enter để quay lại menu...")
			fmt.Scanln()
		case "serve":
			runServeInteractive()
		case "ai-setup":
			runAISetupInteractive()
		case "config":
			runConfigInteractive()
			fmt.Println("\nNhấn Enter để quay lại menu...")
			fmt.Scanln()
		case "install":
			installGlobal()
			fmt.Println("\nNhấn Enter để quay lại menu...")
			fmt.Scanln()
		case "help":
			fmt.Println("\nVibeScanner - Công cụ khám bệnh codebase")
			fmt.Println("\nCác lệnh có sẵn:")
			fmt.Println("  vibescanner scan <path>    - Quét codebase")
			fmt.Println("  vibescanner serve            - Mở web dashboard")
			fmt.Println("  vibescanner ai-setup         - Cài đặt AI models")
			fmt.Println("  vibescanner config           - Quản lý cấu hình")
			fmt.Println("  vibescanner install          - Cài đặt global")
			fmt.Println("\nNhấn Enter để quay lại menu...")
			fmt.Scanln()
		}
	}
}

// runScanInteractive runs scan from interactive mode
func runScanInteractive(path string) error {
	fmt.Printf("\n🔍 Đang quét: %s\n", path)
	// Call the actual scan logic
	return performScan(path, "terminal", false)
}

// runServeInteractive runs serve from interactive mode
func runServeInteractive() error {
	fmt.Println("\n🌐 Khởi động dashboard...")
	results, err := output.LoadLastScan()
	if err != nil {
		return fmt.Errorf("không tìm thấy kết quả scan trước đó: %w", err)
	}
	fmt.Println("📂 Dashboard đang chạy tại http://localhost:7420")
	fmt.Println("⚠️  Nhấn Ctrl+C để dừng server và quay lại menu...")
	
	// Run dashboard - this will block until Ctrl+C
	if err := output.ServeDashboard(results, 7420); err != nil {
		fmt.Printf("\n⚠️  Server dừng: %v\n", err)
	}
	
	fmt.Println("\n✅ Dashboard đã dừng. Nhấn Enter để quay lại menu...")
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
