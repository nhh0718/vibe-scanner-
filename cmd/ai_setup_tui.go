package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nhh0718/vibe-scanner-/internal/ai"
)

// MenuItem for AI setup menu
type aiMenuItem struct {
	title string
	desc  string
}

// AISetupModel represents the TUI state
type AISetupModel struct {
	state        string // "menu", "installing", "removing", "status", "select_model"
	cursor       int
	menuItems    []aiMenuItem
	models       []string
	selected     map[int]struct{}
	spinner      spinner.Model
	status       string
	message      string
	err          error
	width        int
	height       int
	modelToInstall string
}

var (
	aiTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#60A5FA")).
		MarginBottom(1)

	aiMenuStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	aiSelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#60A5FA")).
		Bold(true)

	aiNormalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8"))

	aiDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		PaddingLeft(4)

	aiStatusGood = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22C55E")).
		Bold(true)

	aiStatusBad = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true)

	aiInfoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6"))

	aiHelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		MarginTop(1)
)

func initialAISetupModel() AISetupModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))

	menuItems := []aiMenuItem{
		{title: "📊 Trạng thái hệ thống", desc: "Kiểm tra Ollama và models"},
		{title: "📦 Danh sách models", desc: "Xem models đã cài đặt"},
		{title: "⬇️  Cài đặt model", desc: "Tải và cài đặt model mới"},
		{title: "🗑️  Gỡ bỏ model", desc: "Xóa model không dùng"},
	}

	// Check if Ollama is installed
	if !ai.IsOllamaAvailable() {
		menuItems = []aiMenuItem{
			{title: "⚠️  Ollama chưa cài đặt", desc: "Ollama là công cụ chạy AI local"},
			{title: "📥 Tải Ollama tự động", desc: "Download và cài đặt Ollama cho hệ điều hành hiện tại"},
			{title: "🌐 Mở trang download", desc: "Mở https://ollama.ai/download trong browser"},
			{title: "🔙 Quay lại menu chính", desc: ""},
		}
	} else {
		menuItems = append(menuItems, aiMenuItem{title: "🔙 Quay lại menu chính", desc: ""})
	}

	return AISetupModel{
		state:     "menu",
		spinner:   s,
		menuItems: menuItems,
		selected:  make(map[int]struct{}),
	}
}

func (m AISetupModel) Init() tea.Cmd {
	if m.state == "installing" || m.state == "removing" {
		return m.spinner.Tick
	}
	return nil
}

func (m AISetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.state == "menu" {
				return m, tea.Quit
			}
			// Go back to menu from any state
			m.state = "menu"
			m.err = nil
			m.message = ""
			m = initialAISetupModel()
			return m, nil

		case "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.state == "menu" && m.cursor > 0 {
				m.cursor--
			} else if m.state == "select_model" && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.state == "menu" && m.cursor < len(m.menuItems)-1 {
				m.cursor++
			} else if m.state == "select_model" && m.cursor < len(m.models)-1 {
				m.cursor++
			}

		case "enter", " ":
			if m.state == "select_model" {
				// Install selected model
				m.modelToInstall = m.models[m.cursor]
				m.state = "installing"
				return m, installModelCmd(m.modelToInstall)
			}
			return m.handleSelection()

		case "b":
			if m.state != "menu" {
				m = initialAISetupModel()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m AISetupModel) handleSelection() (tea.Model, tea.Cmd) {
	if m.state != "menu" {
		return m, nil
	}

	// Check if Ollama is available to determine menu structure
	if !ai.IsOllamaAvailable() {
		// Menu when Ollama is NOT installed
		switch m.cursor {
		case 0: // Warning message - do nothing
			return m, nil
		case 1: // Auto download Ollama
			m.state = "downloading"
			m.message = "Đang mở trang download Ollama..."
			return m, downloadOllamaCmd()
		case 2: // Open download page
			m.message = "Mở https://ollama.ai/download trong browser..."
			openBrowserFunc("https://ollama.ai/download")
			return m, tea.Quit
		case 3: // Back to main menu
			return m, tea.Quit
		}
	} else {
		// Menu when Ollama IS installed
		switch m.cursor {
		case 0: // Status
			m.state = "status"
			m.status = "running"
			models, _ := ai.ListInstalledModels()
			m.models = models

		case 1: // List
			m.state = "list"
			models, err := ai.ListInstalledModels()
			if err != nil {
				m.err = err
			} else {
				m.models = models
			}

		case 2: // Install - show model selection
			m.state = "select_model"
			m.cursor = 0
			m.models = getRecommendedModels()
			return m, nil

		case 3: // Remove
			m.state = "remove"
			models, err := ai.ListInstalledModels()
			if err != nil {
				m.err = err
				return m, nil
			}
			if len(models) == 0 {
				m.err = fmt.Errorf("Không có model nào để gỡ")
				return m, nil
			}
			m.models = models

		case 4: // Back to main menu
			return m, tea.Quit
		}
	}

	return m, nil
}

type installMsg struct {
	model string
	err   error
}

func installModelCmd(model string) tea.Cmd {
	return func() tea.Msg {
		if err := ai.PullModel(model); err != nil {
			return installMsg{model: model, err: err}
		}
		return installMsg{model: model}
	}
}

func (m AISetupModel) View() string {
	switch m.state {
	case "status":
		return m.statusView()
	case "list":
		return m.listView()
	case "select_model":
		return m.selectModelView()
	case "installing":
		return m.installView()
	case "install":
		return m.installView()
	case "remove":
		return m.removeView()
	default:
		return m.menuView()
	}
}

func (m AISetupModel) menuView() string {
	var b strings.Builder

	b.WriteString(aiTitleStyle.Render("🤖 VibeScanner AI Setup"))
	b.WriteString("\n")

	for i, item := range m.menuItems {
		if m.cursor == i {
			b.WriteString(aiSelectedStyle.Render("▸ " + item.title))
		} else {
			b.WriteString(aiNormalStyle.Render("  " + item.title))
		}
		if item.desc != "" {
			b.WriteString("\n")
			b.WriteString(aiDescStyle.Render(item.desc))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(aiHelpStyle.Render("↑/↓: di chuyển • enter: chọn • q: thoát"))

	return b.String()
}

func (m AISetupModel) statusView() string {
	var b strings.Builder

	b.WriteString(aiTitleStyle.Render("📊 Trạng thái hệ thống"))
	b.WriteString("\n\n")

	if m.status == "running" {
		b.WriteString(aiStatusGood.Render("✅ Ollama đang chạy"))
		b.WriteString("\n\n")

		if len(m.models) > 0 {
			b.WriteString(aiInfoStyle.Render("📦 Models đã cài đặt:"))
			b.WriteString("\n")
			for _, model := range m.models {
				b.WriteString(fmt.Sprintf("  • %s\n", model))
			}
		} else {
			b.WriteString("⚠️ Chưa có model nào được cài đặt\n")
		}
	} else {
		b.WriteString(aiStatusBad.Render("❌ Ollama chưa chạy"))
		b.WriteString("\n\n")
		b.WriteString("📥 Cài đặt Ollama:\n")
		b.WriteString("   1. Truy cập: https://ollama.ai/download\n")
		b.WriteString("   2. Cài đặt theo hướng dẫn\n")
		b.WriteString("   3. Chạy 'ollama serve'\n")
	}

	b.WriteString("\n")
	b.WriteString(aiHelpStyle.Render("b: quay lại • q: thoát"))

	return b.String()
}

func (m AISetupModel) listView() string {
	var b strings.Builder

	b.WriteString(aiTitleStyle.Render("📦 Danh sách models"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(aiStatusBad.Render(fmt.Sprintf("❌ %v", m.err)))
	} else if len(m.models) == 0 {
		b.WriteString("📭 Chưa có model nào được cài đặt.\n")
		b.WriteString("\nChạy 'Cài đặt model' để tải model đầu tiên.")
	} else {
		b.WriteString(fmt.Sprintf("Tổng cộng: %d models\n\n", len(m.models)))
		for i, model := range m.models {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, model))
		}
	}

	b.WriteString("\n")
	b.WriteString(aiHelpStyle.Render("b: quay lại • q: thoát"))

	return b.String()
}

func (m AISetupModel) installView() string {
	var b strings.Builder

	b.WriteString(aiTitleStyle.Render("⬇️  Cài đặt model"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(aiStatusBad.Render(fmt.Sprintf("❌ %v", m.err)))
		b.WriteString("\n")
	} else {
		modelName := m.modelToInstall
		if modelName == "" {
			modelName = "qwen2.5-coder:3b"
		}
		b.WriteString(m.spinner.View())
		b.WriteString(fmt.Sprintf(" Đang tải model %s...\n", modelName))
		b.WriteString("\n")
		b.WriteString("Lần đầu cài có thể mất vài phút tùy vào tốc độ mạng.\n")
	}

	b.WriteString("\n")
	b.WriteString(aiHelpStyle.Render("b: quay lại • q: thoát"))

	return b.String()
}

func (m AISetupModel) removeView() string {
	var b strings.Builder

	b.WriteString(aiTitleStyle.Render("🗑️  Gỡ bỏ model"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(aiStatusBad.Render(fmt.Sprintf("❌ %v", m.err)))
	} else {
		b.WriteString("Chọn model để gỡ bỏ:\n\n")
		for i, model := range m.models {
			if m.cursor == i {
				b.WriteString(aiSelectedStyle.Render(fmt.Sprintf("▸ %d. %s", i+1, model)))
			} else {
				b.WriteString(aiNormalStyle.Render(fmt.Sprintf("  %d. %s", i+1, model)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(aiHelpStyle.Render("↑/↓: chọn • enter: gỡ • b: quay lại • q: thoát"))

	return b.String()
}

func (m AISetupModel) selectModelView() string {
	var b strings.Builder

	b.WriteString(aiTitleStyle.Render("⬇️  Chọn model để cài đặt"))
	b.WriteString("\n\n")
	b.WriteString(aiInfoStyle.Render("Chọn model phù hợp với cấu hình máy của bạn:"))
	b.WriteString("\n\n")

	modelDescriptions := map[string]string{
		"qwen2.5-coder:0.5b": "Siêu nhẹ - 0.5B params (~350MB) - Máy yếu, RAM < 4GB",
		"qwen2.5-coder:1.5b": "Nhẹ - 1.5B params (~1GB) - Máy trung bình, RAM 4-8GB",
		"qwen2.5-coder:3b":   "Cân bằng - 3B params (~2GB) - Máy tốt, RAM 8-16GB (Khuyến nghị)",
		"qwen2.5-coder:7b":   "Mạnh - 7B params (~4.7GB) - Máy mạnh, RAM 16GB+",
		"qwen2.5-coder:14b":  "Rất mạnh - 14B params (~9GB) - Máy rất mạnh, RAM 32GB+",
		"qwen2.5-coder:32b":  "Cực mạnh - 32B params (~20GB) - Workstation, RAM 64GB+",
	}

	for i, model := range m.models {
		desc := modelDescriptions[model]
		if m.cursor == i {
			b.WriteString(aiSelectedStyle.Render(fmt.Sprintf("▸ %s", model)))
			b.WriteString("\n")
			b.WriteString(aiDescStyle.Render(fmt.Sprintf("  %s", desc)))
		} else {
			b.WriteString(aiNormalStyle.Render(fmt.Sprintf("  %s", model)))
			b.WriteString("\n")
			b.WriteString(aiDescStyle.Render(fmt.Sprintf("  %s", desc)))
		}
		b.WriteString("\n\n")
	}

	b.WriteString(aiHelpStyle.Render("↑/↓: chọn • enter: cài đặt • b: quay lại • q: thoát"))

	return b.String()
}

// getRecommendedModels returns a list of recommended models based on system specs
func getRecommendedModels() []string {
	return []string{
		"qwen2.5-coder:0.5b",
		"qwen2.5-coder:1.5b",
		"qwen2.5-coder:3b",
		"qwen2.5-coder:7b",
		"qwen2.5-coder:14b",
		"qwen2.5-coder:32b",
	}
}

// runAISetupInteractive runs the AI setup TUI
func runAISetupInteractive() error {
	p := tea.NewProgram(initialAISetupModel())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("lỗi TUI: %w", err)
	}
	return nil
}

// downloadOllamaCmd opens the Ollama download page
func downloadOllamaCmd() tea.Cmd {
	return func() tea.Msg {
		url := getOllamaDownloadURL()
		openBrowserFunc(url)
		return nil
	}
}

// getOllamaDownloadURL returns the download URL for current OS
func getOllamaDownloadURL() string {
	switch runtime.GOOS {
	case "windows":
		return "https://ollama.ai/download/windows"
	case "darwin":
		return "https://ollama.ai/download/mac"
	case "linux":
		return "https://ollama.ai/download/linux"
	default:
		return "https://ollama.ai/download"
	}
}

// openBrowserFunc opens a URL in the default browser
func openBrowserFunc(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
