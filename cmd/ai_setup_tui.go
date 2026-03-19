package cmd

import (
	"fmt"
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
	state     string // "menu", "installing", "removing", "status"
	cursor    int
	menuItems []aiMenuItem
	models    []string
	selected  map[int]struct{}
	spinner   spinner.Model
	status    string
	message   string
	err       error
	width     int
	height    int
}

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#60A5FA")).
		MarginBottom(1)

	menuStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#60A5FA")).
		Bold(true)

	normalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8"))

	descStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		PaddingLeft(4)

	statusGood = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22C55E")).
		Bold(true)

	statusBad = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true)

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6"))

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		MarginTop(1)
)

func initialAISetupModel() AISetupModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))

	return AISetupModel{
		state:   "menu",
		spinner: s,
		menuItems: []aiMenuItem{
			{title: "📊 Trạng thái hệ thống", desc: "Kiểm tra Ollama và models"},
			{title: "📦 Danh sách models", desc: "Xem models đã cài đặt"},
			{title: "⬇️  Cài đặt model", desc: "Tải và cài đặt model mới"},
			{title: "🗑️  Gỡ bỏ model", desc: "Xóa model không dùng"},
			{title: "❌ Thoát", desc: ""},
		},
		selected: make(map[int]struct{}),
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
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.state == "menu" && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.state == "menu" && m.cursor < len(m.menuItems)-1 {
				m.cursor++
			}

		case "enter", " ":
			return m.handleSelection()

		case "b":
			if m.state != "menu" {
				m.state = "menu"
				m.err = nil
				m.message = ""
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

	switch m.cursor {
	case 0: // Status
		m.state = "status"
		if ai.IsOllamaAvailable() {
			m.status = "running"
			models, _ := ai.ListInstalledModels()
			m.models = models
		} else {
			m.status = "stopped"
		}

	case 1: // List
		m.state = "list"
		if !ai.IsOllamaAvailable() {
			m.err = fmt.Errorf("Ollama chưa chạy")
			return m, nil
		}
		models, err := ai.ListInstalledModels()
		if err != nil {
			m.err = err
		} else {
			m.models = models
		}

	case 2: // Install
		m.state = "install"
		if !ai.IsOllamaAvailable() {
			m.err = fmt.Errorf("Ollama chưa chạy. Cài đặt tại https://ollama.ai/download")
			return m, nil
		}
		return m, installModelCmd("qwen2.5-coder:3b")

	case 3: // Remove
		m.state = "remove"
		if !ai.IsOllamaAvailable() {
			m.err = fmt.Errorf("Ollama chưa chạy")
			return m, nil
		}
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

	case 4: // Exit
		return m, tea.Quit
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

	b.WriteString(titleStyle.Render("🤖 VibeScanner AI Setup"))
	b.WriteString("\n")

	for i, item := range m.menuItems {
		if m.cursor == i {
			b.WriteString(selectedStyle.Render("▸ " + item.title))
		} else {
			b.WriteString(normalStyle.Render("  " + item.title))
		}
		if item.desc != "" {
			b.WriteString("\n")
			b.WriteString(descStyle.Render(item.desc))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: di chuyển • enter: chọn • q: thoát"))

	return b.String()
}

func (m AISetupModel) statusView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📊 Trạng thái hệ thống"))
	b.WriteString("\n\n")

	if m.status == "running" {
		b.WriteString(statusGood.Render("✅ Ollama đang chạy"))
		b.WriteString("\n\n")

		if len(m.models) > 0 {
			b.WriteString(infoStyle.Render("📦 Models đã cài đặt:"))
			b.WriteString("\n")
			for _, model := range m.models {
				b.WriteString(fmt.Sprintf("  • %s\n", model))
			}
		} else {
			b.WriteString("⚠️ Chưa có model nào được cài đặt\n")
		}
	} else {
		b.WriteString(statusBad.Render("❌ Ollama chưa chạy"))
		b.WriteString("\n\n")
		b.WriteString("📥 Cài đặt Ollama:\n")
		b.WriteString("   1. Truy cập: https://ollama.ai/download\n")
		b.WriteString("   2. Cài đặt theo hướng dẫn\n")
		b.WriteString("   3. Chạy 'ollama serve'\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("b: quay lại • q: thoát"))

	return b.String()
}

func (m AISetupModel) listView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📦 Danh sách models"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(statusBad.Render(fmt.Sprintf("❌ %v", m.err)))
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
	b.WriteString(helpStyle.Render("b: quay lại • q: thoát"))

	return b.String()
}

func (m AISetupModel) installView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("⬇️  Cài đặt model"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(statusBad.Render(fmt.Sprintf("❌ %v", m.err)))
		b.WriteString("\n")
	} else {
		b.WriteString(m.spinner.View())
		b.WriteString(" Đang tải model qwen2.5-coder:3b...\n")
		b.WriteString("\n")
		b.WriteString("Lần đầu cài có thể mất vài phút tùy vào tốc độ mạng.\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("b: quay lại • q: thoát"))

	return b.String()
}

func (m AISetupModel) removeView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🗑️  Gỡ bỏ model"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(statusBad.Render(fmt.Sprintf("❌ %v", m.err)))
	} else {
		b.WriteString("Chọn model để gỡ bỏ:\n\n")
		for i, model := range m.models {
			if m.cursor == i {
				b.WriteString(selectedStyle.Render(fmt.Sprintf("▸ %d. %s", i+1, model)))
			} else {
				b.WriteString(normalStyle.Render(fmt.Sprintf("  %d. %s", i+1, model)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: chọn • enter: gỡ • b: quay lại • q: thoát"))

	return b.String()
}

// runAISetupInteractive runs the AI setup TUI
func runAISetupInteractive() error {
	p := tea.NewProgram(initialAISetupModel())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("lỗi TUI: %w", err)
	}
	return nil
}
