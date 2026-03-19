package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nhh0718/vibe-scanner-/internal/ai"
	"github.com/nhh0718/vibe-scanner-/internal/ui"
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

func initialAISetupModel() AISetupModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ui.GetSpinnerStyle()

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
				m.modelToInstall = m.models[m.cursor]
				m.state = "installing"
				return m, installModelCmd(m.modelToInstall)
			}
			if m.state == "remove" && len(m.models) > 0 {
				m.state = "removing"
				return m, removeModelCmd(m.models[m.cursor])
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

	case installMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = "installing"
			return m, nil
		}
		m = initialAISetupModel()
		m.message = fmt.Sprintf("Đã cài đặt model %s", msg.model)
		return m, nil

	case removeMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = "remove"
			return m, nil
		}
		m = initialAISetupModel()
		m.message = fmt.Sprintf("Đã gỡ model %s", msg.model)
		return m, nil
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

type removeMsg struct {
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

func removeModelCmd(model string) tea.Cmd {
	return func() tea.Msg {
		if err := ai.RemoveModel(model); err != nil {
			return removeMsg{model: model, err: err}
		}
		return removeMsg{model: model}
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
	case "removing":
		return m.removeProgressView()
	case "install":
		return m.installView()
	case "remove":
		return m.removeView()
	default:
		return m.menuView()
	}
}

func (m AISetupModel) menuView() string {
	var main strings.Builder
	if m.message != "" {
		main.WriteString(ui.SuccessText(m.message))
		main.WriteString("\n\n")
	}
	for i, item := range m.menuItems {
		main.WriteString(ui.NumberedLine(i+1, item.title, item.desc, m.cursor == i))
		main.WriteString("\n\n")
	}

	var side strings.Builder
	side.WriteString(ui.SectionLabel("Tình trạng hiện tại"))
	side.WriteString("\n\n")
	if ai.IsOllamaAvailable() {
		side.WriteString(ui.KeyValue("Ollama", "Đang sẵn sàng"))
		models, _ := ai.ListInstalledModels()
		side.WriteString("\n")
		side.WriteString(ui.KeyValue("Số model", fmt.Sprintf("%d", len(models))))
	} else {
		side.WriteString(ui.KeyValue("Ollama", "Chưa cài đặt"))
		side.WriteString("\n")
		side.WriteString(ui.KeyValue("Khuyến nghị", "Tải tự động"))
	}
	if m.message == "" {
		side.WriteString("\n\n")
		side.WriteString(ui.SectionLabel("Gợi ý"))
		side.WriteString("\n\n")
		side.WriteString(ui.BulletList([]string{
			"Ưu tiên model nhỏ nếu RAM ít.",
			"Cài Ollama trước khi thêm model.",
			"Nhấn B để quay lại trang trước.",
		}))
	}

	return ui.RenderScreen(
		"TRUNG TÂM AI CỤC BỘ",
		"Thiết lập Ollama và các mô hình AI ngay trong dòng lệnh",
		strings.TrimSpace(main.String()),
		side.String(),
		[]string{"↑/↓ di chuyển", "Enter chọn", "B quay lại", "Q thoát"},
	)
}

func (m AISetupModel) statusView() string {
	var main strings.Builder
	if m.status == "running" {
		main.WriteString(ui.SuccessText("Ollama đang hoạt động bình thường"))
		main.WriteString("\n\n")
		if len(m.models) == 0 {
			main.WriteString(ui.WarningText("Chưa có model nào được cài đặt"))
		} else {
			main.WriteString(ui.SectionLabel("Model đã cài"))
			main.WriteString("\n\n")
			for i, model := range m.models {
				main.WriteString(fmt.Sprintf("[%02d] %s\n", i+1, model))
			}
		}
	} else {
		main.WriteString(ui.ErrorText("Ollama chưa chạy hoặc chưa cài đặt"))
		main.WriteString("\n\n")
		main.WriteString(ui.BulletList([]string{
			"Tải Ollama từ trang chính thức.",
			"Hoàn tất cài đặt theo hệ điều hành.",
			"Khởi chạy dịch vụ bằng lệnh: ollama serve",
		}))
	}

	var side strings.Builder
	side.WriteString(ui.SectionLabel("Tóm tắt"))
	side.WriteString("\n\n")
	side.WriteString(ui.KeyValue("Runtime", map[bool]string{true: "Đang chạy", false: "Chưa sẵn sàng"}[m.status == "running"]))
	side.WriteString("\n")
	side.WriteString(ui.KeyValue("Model đã cài", fmt.Sprintf("%d", len(m.models))))
	side.WriteString("\n\n")
	side.WriteString(ui.SectionLabel("Phím tắt"))
	side.WriteString("\n\n")
	side.WriteString(ui.BulletList([]string{"B quay lại menu AI", "Q thoát giao diện AI"}))

	return ui.RenderScreen(
		"TRẠNG THÁI HỆ THỐNG AI",
		"Kiểm tra runtime Ollama và các model hiện có",
		strings.TrimSpace(main.String()),
		side.String(),
		[]string{"B quay lại", "Q thoát"},
	)
}

func (m AISetupModel) listView() string {
	var main strings.Builder
	if m.err != nil {
		main.WriteString(ui.ErrorText(m.err.Error()))
	} else if len(m.models) == 0 {
		main.WriteString(ui.WarningText("Chưa có model nào được cài đặt"))
		main.WriteString("\n\n")
		main.WriteString("Hãy vào mục cài đặt model để tải model đầu tiên.")
	} else {
		main.WriteString(ui.SectionLabel("Danh sách model hiện có"))
		main.WriteString("\n\n")
		for i, model := range m.models {
			main.WriteString(fmt.Sprintf("[%02d] %-22s Trạng thái: sẵn sàng\n", i+1, model))
		}
	}

	var side strings.Builder
	side.WriteString(ui.SectionLabel("Tổng kết"))
	side.WriteString("\n\n")
	side.WriteString(ui.KeyValue("Số lượng", fmt.Sprintf("%d model", len(m.models))))
	side.WriteString("\n\n")
	side.WriteString(ui.SectionLabel("Gợi ý"))
	side.WriteString("\n\n")
	side.WriteString(ui.BulletList([]string{
		"Model 3B phù hợp đa số máy cá nhân.",
		"Model 7B+ cần RAM cao hơn.",
		"Dọn model cũ để tiết kiệm dung lượng.",
	}))

	return ui.RenderScreen(
		"DANH SÁCH MODEL AI",
		"Theo dõi các model hiện đã có trong máy",
		strings.TrimSpace(main.String()),
		side.String(),
		[]string{"B quay lại", "Q thoát"},
	)
}

func (m AISetupModel) installView() string {
	modelName := m.modelToInstall
	if modelName == "" {
		modelName = "qwen2.5-coder:3b"
	}

	var main strings.Builder
	if m.err != nil {
		main.WriteString(ui.ErrorText(m.err.Error()))
	} else {
		main.WriteString(m.spinner.View())
		main.WriteString(" ")
		main.WriteString(ui.SuccessText("Đang tải model: "+modelName))
		main.WriteString("\n\n")
		main.WriteString("Lần cài đầu tiên có thể mất vài phút tùy theo tốc độ mạng và cấu hình máy.")
	}

	var side strings.Builder
	side.WriteString(ui.SectionLabel("Lưu ý khi cài đặt"))
	side.WriteString("\n\n")
	side.WriteString(ui.BulletList([]string{
		"Không tắt Ollama trong lúc tải model.",
		"Đảm bảo còn đủ dung lượng ổ đĩa.",
		"Model lớn sẽ tải lâu hơn đáng kể.",
	}))

	return ui.RenderScreen(
		"TIẾN TRÌNH CÀI ĐẶT MODEL",
		"Theo dõi tiến trình tải và cài model AI",
		strings.TrimSpace(main.String()),
		side.String(),
		[]string{"B quay lại sau khi hoàn tất", "Q thoát"},
	)
}

func (m AISetupModel) removeView() string {
	var main strings.Builder
	if m.err != nil {
		main.WriteString(ui.ErrorText(m.err.Error()))
	} else {
		main.WriteString(ui.SectionLabel("Chọn model cần gỡ"))
		main.WriteString("\n\n")
		for i, model := range m.models {
			main.WriteString(ui.NumberedLine(i+1, model, "Nhấn Enter để gỡ model này khỏi máy", m.cursor == i))
			main.WriteString("\n\n")
		}
	}

	var side strings.Builder
	side.WriteString(ui.SectionLabel("Cảnh báo"))
	side.WriteString("\n\n")
	side.WriteString(ui.BulletList([]string{
		"Gỡ model sẽ giải phóng dung lượng ổ đĩa.",
		"Model đã gỡ sẽ cần tải lại nếu muốn dùng tiếp.",
		"Hãy giữ lại ít nhất một model thường dùng.",
	}))

	return ui.RenderScreen(
		"GỠ BỎ MODEL AI",
		"Quản lý dung lượng bằng cách xóa model không còn sử dụng",
		strings.TrimSpace(main.String()),
		side.String(),
		[]string{"↑/↓ chọn model", "Enter gỡ", "B quay lại", "Q thoát"},
	)
}

func (m AISetupModel) selectModelView() string {
	var main strings.Builder

	modelDescriptions := map[string]string{
		"qwen2.5-coder:0.5b": "Siêu nhẹ - 0.5B params (~350MB) - Máy yếu, RAM < 4GB",
		"qwen2.5-coder:1.5b": "Nhẹ - 1.5B params (~1GB) - Máy trung bình, RAM 4-8GB",
		"qwen2.5-coder:3b":   "Cân bằng - 3B params (~2GB) - Máy tốt, RAM 8-16GB (Khuyến nghị)",
		"qwen2.5-coder:7b":   "Mạnh - 7B params (~4.7GB) - Máy mạnh, RAM 16GB+",
		"qwen2.5-coder:14b":  "Rất mạnh - 14B params (~9GB) - Máy rất mạnh, RAM 32GB+",
		"qwen2.5-coder:32b":  "Cực mạnh - 32B params (~20GB) - Workstation, RAM 64GB+",
	}

	main.WriteString(ui.SectionLabel("Chọn model phù hợp với cấu hình máy"))
	main.WriteString("\n\n")
	for i, model := range m.models {
		main.WriteString(ui.NumberedLine(i+1, model, modelDescriptions[model], m.cursor == i))
		main.WriteString("\n\n")
	}

	var side strings.Builder
	side.WriteString(ui.SectionLabel("Khuyến nghị nhanh"))
	side.WriteString("\n\n")
	side.WriteString(ui.BulletList([]string{
		"0.5B-1.5B: máy yếu hoặc laptop văn phòng.",
		"3B: lựa chọn cân bằng cho đa số dự án.",
		"7B trở lên: ưu tiên máy có RAM cao.",
	}))
	side.WriteString("\n\n")
	side.WriteString(ui.KeyValue("Mặc định gợi ý", "qwen2.5-coder:3b"))

	return ui.RenderScreen(
		"CHỌN MODEL ĐỂ CÀI ĐẶT",
		"Danh sách model được đề xuất theo cấu hình máy",
		strings.TrimSpace(main.String()),
		side.String(),
		[]string{"↑/↓ chọn model", "Enter cài đặt", "B quay lại", "Q thoát"},
	)
}

func (m AISetupModel) removeProgressView() string {
	modelName := "model đã chọn"
	if len(m.models) > 0 && m.cursor < len(m.models) {
		modelName = m.models[m.cursor]
	}

	var main strings.Builder
	if m.err != nil {
		main.WriteString(ui.ErrorText(m.err.Error()))
	} else {
		main.WriteString(m.spinner.View())
		main.WriteString(" ")
		main.WriteString(ui.WarningText("Đang gỡ model: "+modelName))
		main.WriteString("\n\n")
		main.WriteString("Tiến trình gỡ đang chạy, vui lòng chờ hoàn tất.")
	}

	return ui.RenderScreen(
		"TIẾN TRÌNH GỠ MODEL",
		"Theo dõi quá trình xóa model khỏi máy",
		strings.TrimSpace(main.String()),
		ui.BulletList([]string{"Giữ kết nối với Ollama trong lúc gỡ model.", "Không đóng chương trình giữa chừng nếu chưa hoàn tất."}),
		[]string{"B quay lại sau khi hoàn tất", "Q thoát"},
	)
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
