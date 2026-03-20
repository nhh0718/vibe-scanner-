# 🔍 VibeScanner

> Công cụ "khám bệnh" codebase cho vibe coders - Chạy hoàn toàn local, code không rời máy

![Version](https://img.shields.io/badge/version-0.8.0-blue)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D)
![Rules](https://img.shields.io/badge/rules-44+_AST-green)
![Gitleaks](https://img.shields.io/badge/secrets-Gitleaks-red)

## ✨ Điểm khác biệt

| Công cụ thông thường       | VibeScanner                                                                                              |
| -------------------------- | -------------------------------------------------------------------------------------------------------- |
| `SQL injection at line 42` | `Dòng 42: Hacker có thể xóa toàn bộ database chỉ bằng cách gõ vào ô search. Sửa như này:` + code example |
| Output cho developer       | Output cho người vibe-code không biết security                                                           |
| Chạy từng tool riêng lẻ    | Một lệnh quét tất cả (AST + Gitleaks)                                                                    |
| Kết quả thô, kỹ thuật      | Báo cáo ngôn ngữ đời thường, có priority rõ ràng                                                         |
| Gửi code lên cloud         | **100% local, code không rời máy**                                                                       |

---

## 🆕 What's New in v0.8.0

- **44 AST rules** — Phân tích cú pháp Tree-sitter cho JS/TS/Python/Go/Java + regex fallback
- **Gitleaks integration** — Phát hiện secrets (API keys, tokens, passwords) tự động
- **Web Dashboard cải tiến** — Report history selector, chuyển đổi báo cáo, auto-refresh
- **Auto-open browser** — Dashboard tự động mở trình duyệt khi khởi động
- **Styled CLI** — Banner dashboard đẹp với thông tin dự án, health score, findings
- **Unified storage** — Lưu trữ thống nhất, không còn hiện kết quả cũ

---

## 🚀 Quick Start

### 1. Tải binary

| Platform | Lệnh |
|----------|-------|
| **Windows** | `Invoke-WebRequest -Uri https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-windows-amd64.exe -OutFile vibescanner.exe` |
| **macOS Intel** | `curl -LO https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-darwin-amd64` |
| **macOS Apple Silicon** | `curl -LO https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-darwin-arm64` |
| **Linux** | `curl -LO https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-linux-amd64` |

### 2. Cài đặt global

```bash
# macOS/Linux: đổi tên, cấp quyền, cài vào PATH
mv vibescanner-* vibescanner && chmod +x vibescanner
./vibescanner install

# Windows (PowerShell):
.\vibescanner.exe install
```

> macOS Gatekeeper chặn? Chạy `xattr -cr ./vibescanner` hoặc vào System Settings > Privacy & Security > Allow Anyway.

### 3. Sử dụng

```bash
vibescanner                          # Menu tương tác (8 tùy chọn)
vibescanner scan ./my-project        # Quét trực tiếp
vibescanner scan . --report html     # Tạo HTML report
vibescanner serve                    # Mở dashboard (auto-open browser)
vibescanner history                  # Xem lịch sử quét
```

> Hướng dẫn chi tiết: xem [`USAGE.md`](./USAGE.md)

---

## 📋 Tất cả lệnh

| Lệnh          | Mô tả                  | Ví dụ                     |
| ------------- | ---------------------- | ------------------------- |
| `vibescanner` | Menu tương tác         | `vibescanner`             |
| `scan [path]` | Quét codebase          | `vibescanner scan .`      |
| `serve`       | Mở web dashboard       | `vibescanner serve`       |
| `history`     | Xem lịch sử & báo cáo  | `vibescanner history`     |
| `report`      | Quản lý báo cáo        | `vibescanner report`      |
| `ai-setup`    | Quản lý AI local       | `vibescanner ai-setup`    |
| `config`      | Cấu hình               | `vibescanner config list` |
| `install`     | Cài global vào PATH    | `vibescanner install`     |
| `uninstall`   | Gỡ cài đặt             | `vibescanner uninstall`   |
| `update`      | Cập nhật phiên bản mới | `vibescanner update`      |

---

## 🌐 Web Dashboard

Dashboard chạy tại `http://localhost:7420` với các tính năng:

- **Health Score** — Điểm sức khỏe tổng quát, bảo mật, chất lượng, kiến trúc
- **Report History** — Dropdown chọn và chuyển đổi giữa các lần quét
- **Filter & Sort** — Lọc theo severity, category, tìm kiếm theo keyword
- **AI Explain** — Hỏi AI giải thích lỗi (cần Ollama)
- **Auto Refresh** — Nút làm mới để tải kết quả quét mới nhất

```bash
vibescanner serve              # Mở dashboard, tự động mở browser
vibescanner serve --port 8080  # Dùng port khác
```

---

## 🛡️ Rules Coverage (44 rules)

| Category      | Rules | Ví dụ                                          |
| ------------- | ----- | ---------------------------------------------- |
| Security      | 15    | SQL Injection, XSS, CSRF, Path Traversal       |
| Quality       | 11    | Complex functions, unused vars, dead code      |
| Architecture  | 5     | Circular imports, God files, DB in controllers |
| Performance   | 5     | N+1 queries, full Lodash import, sync I/O      |
| Vibe-specific | 8     | Env load order, hardcoded ports, CORS wildcard |
| Secrets       | ∞     | Gitleaks: API keys, tokens, passwords          |

---

## 🤖 AI Setup (Tùy chọn)

```bash
vibescanner ai-setup         # Menu AI tương tác
vibescanner ai-setup status  # Kiểm tra trạng thái
vibescanner ai-setup install # Cài model mặc định
```

Cần [Ollama](https://ollama.ai/download) để sử dụng tính năng AI giải thích.

---

## 🔧 Troubleshooting

| Lỗi                                 | Giải pháp                                             |
| ----------------------------------- | ----------------------------------------------------- |
| "vibescanner không phải là lệnh..." | Chạy `vibescanner install` và restart terminal        |
| macOS: "cannot be opened"           | System Preferences → Security & Privacy → Open Anyway |
| "Không tìm thấy kết quả scan"       | Chạy `vibescanner scan .` trước khi `serve`           |
| AI không khả dụng                   | Cài Ollama tại https://ollama.ai/download             |
| Dashboard hiện kết quả cũ           | Nhấn nút "Làm mới" hoặc chạy scan lại                 |

---

## 🔨 Build from Source

Yêu cầu: **Go 1.22+**, **Node.js 18+**

```bash
git clone https://github.com/nhh0718/vibe-scanner-.git
cd vibe-scanner-

# Build web dashboard (embedded vào binary)
cd web && npm ci && npm run build && cd ..

# Build binary
go build -o vibescanner .

# (Tuỳ chọn) Cài vào PATH
./vibescanner install
```

---

## 📄 License

MIT License

<p align="center">Made with ❤️ for vibe coders everywhere</p>
