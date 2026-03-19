# 🔍 VibeScanner

> Công cụ "khám bệnh" codebase cho vibe coders - Chạy hoàn toàn local, code không rời máy

![Version](https://img.shields.io/badge/version-0.1.0-blue)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D)

## ✨ Điểm khác biệt

| Công cụ thông thường | VibeScanner |
|---|---|
| `SQL injection at line 42` | `Dòng 42: Hacker có thể xóa toàn bộ database chỉ bằng cách gõ vào ô search. Sửa như này:` + code example |
| Output cho developer | Output cho người vibe-code không biết security |
| Chạy từng tool riêng lẻ | Một lệnh quét tất cả |
| Kết quả thô, kỹ thuật | Báo cáo ngôn ngữ đời thường, có priority rõ ràng |
| Gửi code lên cloud | **100%% local, code không rời máy** |

## 🚀 Cài đặt

### Tải binary (khuyến nghị)

```bash
# macOS / Linux
curl -L https://github.com/vibescanner/vibescanner/releases/latest/download/vibescanner-linux-amd64 -o vibescanner
chmod +x vibescanner

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/vibescanner/vibescanner/releases/latest/download/vibescanner-windows-amd64.exe -OutFile vibescanner.exe
```

### Build từ source

```bash
git clone https://github.com/vibescanner/vibescanner
cd vibescanner
go build -o vibescanner .
```

## 📖 Sử dụng

### Quét codebase và xem kết quả trong terminal

```bash
./vibescanner scan ./my-project
```

### Tạo HTML report

```bash
./vibescanner scan ./my-project --report html --open
```

### Mở Web Dashboard

```bash
./vibescanner scan ./my-project --report html
# Sau đó mở browser tại http://localhost:7420
```

### Cài đặt AI (tùy chọn)

```bash
./vibescanner ai-setup
```

Sau khi cài đặt AI, bạn có thể click "Hỏi bác sĩ AI" trên từng finding trong dashboard để nhận giải thích chi tiết bằng tiếng Việt.

## 🏥 Điểm sức khỏe codebase

VibeScanner đánh giá codebase theo 4 hạng mục:

| Điểm | Trạng thái | Ý nghĩa |
|---|---|---|
| 80-100 | 🟢 Tốt | Có thể deploy, monitor regularly |
| 60-79 | 🟡 Trung bình | Fix Critical/High trước khi scale |
| 40-59 | 🟠 Cần cải thiện | Cần sprint fix issues |
| 0-39 | 🔴 Nguy hiểm | Không nên deploy production |

## 🔍 Các loại vấn đề được phát hiện

### 🛡️ Bảo mật (Security)
- SQL Injection, NoSQL Injection
- XSS, CSRF
- Hardcoded secrets, API keys
- JWT weak secrets
- Path traversal
- CORS misconfiguration

### ✨ Chất lượng (Quality)
- Code complexity
- Magic numbers
- Console.log trong production
- TODO/FIXME comments
- Empty catch blocks
- Long lines/functions

### 🏗️ Kiến trúc (Architecture)
- Circular dependencies
- SOLID violations
- Code duplication

### 🔑 Secrets
- API keys (AWS, Stripe, OpenAI, etc.)
- Database passwords
- Private keys
- `.env` files committed

## 🏗️ Kiến trúc

```
┌─────────────────────────────────────────────────────────┐
│                     INPUT LAYER                         │
│  Local folder  │  Git repo URL  │  File upload (ZIP)    │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                   INGESTION LAYER                       │
│  • Language detection    • File tree traversal          │
│  • Git history analysis  • Dependency extraction       │
└────────────────────────┬────────────────────────────────┘
         ┌───────────────┼───────────────┐
         │               │               │
┌────────▼───────┐ ┌─────▼─────┐ ┌─────▼──────────┐
│ SECURITY       │ │  QUALITY  │ │  ARCHITECTURE  │
│ ENGINE         │ │  ENGINE   │ │  ENGINE        │
│                │ │           │ │                │
│ • Semgrep      │ │ • Radon   │ │ • Madge        │
│ • Gitleaks     │ │ • ESLint  │ │ • Dep-cruiser  │
└────────┬───────┘ └─────┬─────┘ └─────┬──────────┘
         │               │             │
┌────────▼───────────────▼─────────────▼────────────────┐
│                  AGGREGATION LAYER                    │
│  • Deduplicate findings                               │
│  • Score severity (Critical/High/Medium/Low)          │
│  • Group by file/category/severity                    │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                    AI SYNTHESIS LAYER                 │
│  Ollama (local) — Qwen2.5-Coder                      │
│  • Giải thích lỗi bằng ngôn ngữ đơn giản            │
│  • Đề xuất code fix cụ thể                          │
│  • On-demand (không chạy ngầm)                     │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                   OUTPUT LAYER                          │
│  CLI summary  │  HTML report  │  JSON  │  Dashboard     │
└─────────────────────────────────────────────────────────┘
```

## 🛠️ Stack kỹ thuật

### Core
- **Language**: Go 1.22+
- **CLI**: Cobra
- **Web Server**: Gin
- **Concurrency**: Goroutines

### Dashboard
- **Framework**: Vue 3 + Vite
- **Styling**: Tailwind CSS
- **State**: Pinia
- **Charts**: Chart.js

### AI Layer
- **Runtime**: Ollama (local)
- **Model**: Qwen2.5-Coder (3B/7B)
- **Protocol**: REST API + SSE streaming

## 📝 Lộ trình phát triển

### ✅ Phase 0 - MVP CLI (Completed)
- [x] Core CLI commands (scan, ai-setup, serve)
- [x] Semgrep & Gitleaks integration
- [x] Complexity analysis
- [x] Terminal, JSON, HTML output

### ✅ Phase 1 - Web Dashboard (Completed)
- [x] Vue 3 + Vite setup
- [x] Dashboard components
- [x] Go embed integration
- [x] Custom Semgrep rules

### 🚧 Phase 2 - CI/CD & Extensions (In Progress)
- [ ] VS Code extension
- [ ] GitHub Action
- [ ] Pre-commit hook
- [ ] Baseline support

### 📋 Phase 3 - Monetization
- [ ] Pro tier features
- [ ] Team dashboard
- [ ] Custom rules builder

## 🤝 Đóng góp

Chúng tôi rất hoan nghênh đóng góp! Vui lòng:

1. Fork repository
2. Tạo feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Mở Pull Request

## 📄 License

MIT License - Xem [LICENSE](LICENSE) để biết thêm chi tiết.

## 🙏 Acknowledgements

- [Semgrep](https://semgrep.dev/) - Static analysis engine
- [Gitleaks](https://github.com/gitleaks/gitleaks) - Secret detection
- [Ollama](https://ollama.ai/) - Local LLM runtime
- [Vue.js](https://vuejs.org/) - Frontend framework

---

<p align="center">
  Made with ❤️ for vibe coders everywhere
</p>
