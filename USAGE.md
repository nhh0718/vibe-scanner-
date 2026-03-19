# 📖 Hướng Dẫn Sử Dụng VibeScanner

> Hướng dẫn từ A-Z để sử dụng VibeScanner - Công cụ "khám bệnh" codebase cho vibe coders

> 📚 **Xem tổng quan dự án tại:** [`README.md`](./README.md)

---

## 📋 Mục lục

1. [Cài đặt](#1-cài-đặt)
2. [Chạy lần đầu](#2-chạy-lần-đầu)
3. [Các lệnh cơ bản](#3-các-lệnh-cơ-bản)
4. [Sử dụng Interactive Menu](#4-sử-dụng-interactive-menu)
5. [Phân tích một dự án](#5-phân-tích-một-dự-án)
6. [Sử dụng Web Dashboard](#6-sử-dụng-web-dashboard)
7. [Cài đặt và quản lý AI](#7-cài-đặt-và-quản-lý-ai)
8. [Quản lý cấu hình](#8-quản-lý-cấu-hình)
9. [Troubleshooting](#9-troubleshooting)

---

## 1. Cài đặt

### Cách 1: Tải Binary (Khuyến nghị)

#### Windows (PowerShell):
```powershell
# Tải về
Invoke-WebRequest -Uri https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-windows-amd64.exe -OutFile vibescanner.exe

# Hoặc dùng curl
curl -L -o vibescanner.exe https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-windows-amd64.exe
```

#### macOS (Intel - older Macs):
```bash
# Download for Intel Macs
curl -L https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-darwin-amd64 -o vibescanner
chmod +x vibescanner
```

#### macOS (Apple Silicon - M1/M2/M3):
```bash
# Download for Apple Silicon (M1/M2/M3)
curl -L https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-darwin-arm64 -o vibescanner
chmod +x vibescanner
```

> ⚠️ **Lưu ý về Gatekeeper trên macOS:**
> Khi chạy lần đầu, macOS có thể báo:
> > "vibescanner cannot be opened because the developer cannot be verified"
>
> **Cách khắc phục:**
> 1. Vào **System Settings** → **Privacy & Security**
> 2. Kéo xuống phần **Security**, tìm thông báo về vibescanner
> 3. Click **"Allow Anyway"**
> 4. Chạy lại: `./vibescanner --version`
>
> Hoặc dùng terminal:
> ```bash
> xattr -cr ./vibescanner
> ./vibescanner --version
> ```

#### Linux:
```bash
curl -L https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-linux-amd64 -o vibescanner
chmod +x vibescanner
```

### Cách 2: Build từ Source

Yêu cầu: Go 1.22+, Node.js 18+

```bash
# Clone repo
git clone https://github.com/nhh0718/vibe-scanner-.git
cd vibescanner

# Build web dashboard
cd web
npm install
npm run build
cd ..

# Build binary
go build -o vibescanner .
```

### Cách 3: Cài đặt Global (Khuyến nghị sau khi tải)

Sau khi tải binary, chạy lệnh sau để cài đặt vào PATH:

```bash
# Windows (PowerShell - chạy tại thư mục chứa vibescanner.exe)
.\vibescanner.exe install

# macOS/Linux
./vibescanner install
```

**Lưu ý:** Sau khi cài đặt global, bạn cần **mở terminal mới** hoặc restart terminal để lệnh `vibescanner` có hiệu lực.

Kiểm tra cài đặt:
```bash
vibescanner --version
# Output: vibescanner version 0.1.0
```

---

## 2. Chạy lần đầu

### Không có gì cả? Mở Menu tương tác!

```bash
vibescanner
```

Sẽ hiển thị menu đẹp mắt với các tùy chọn:
- 🔍 Scan Project
- 🌐 Web Dashboard
- 🤖 AI Setup
- ⚙️ Cấu hình
- 📦 Cài đặt Global
- ❓ Help

**Điều hướng:** Dùng phím ↑ ↓ để di chuyển, Enter để chọn, q hoặc Esc để thoát.

---

## 3. Các lệnh cơ bản

### 3.1 Quét codebase

```bash
# Quét thư mục hiện tại
vibescanner scan .

# Quét thư mục cụ thể
vibescanner scan ./my-project

# Quét và xuất HTML report
vibescanner scan ./my-project --report html --open

# Quét và xuất JSON
vibescanner scan ./my-project --report json

# Quét nhưng tắt AI
vibescanner scan ./my-project --no-ai
```

### 3.2 Mở Web Dashboard

```bash
# Mở dashboard với kết quả scan gần nhất
vibescanner serve

# Mở ở port khác
vibescanner serve --port 8080
```

Dashboard sẽ mở tại: `http://localhost:7420`

### 3.3 Quản lý AI

```bash
# Xem trạng thái AI
vibescanner ai-setup

# Liệt kê models đã cài
vibescanner ai-setup list

# Cài đặt model mặc định (qwen2.5-coder:3b)
vibescanner ai-setup install

# Cài đặt model cụ thể
vibescanner ai-setup install qwen2.5-coder:7b

# Gỡ bỏ model
vibescanner ai-setup remove qwen2.5-coder:3b
```

### 3.4 Quản lý cấu hình

```bash
# Tạo file config mặc định
vibescanner config init

# Xem tất cả cấu hình
vibescanner config list

# Lấy giá trị cụ thể
vibescanner config get default_model

# Thiết lập giá trị
vibescanner config set default_model qwen2.5-coder:7b
vibescanner config set auto_open false
```

### 3.5 Cài đặt / Gỡ cài đặt

```bash
# Cài đặt global (thêm vào PATH)
vibescanner install

# Gỡ cài đặt global
vibescanner uninstall
```

---

## 4. Sử dụng Interactive Menu

Khi chạy `vibescanner` không có arguments, bạn sẽ thấy menu tương tác:

```
🔍 VibeScanner - Chào mừng!
Công cụ khám bệnh codebase cho vibe coders

  ▸ 🔍 Scan Project
    🌐 Web Dashboard
    🤖 AI Setup
    ⚙️  Cấu hình
    📦 Cài đặt Global
    ❓ Help

q/esc: thoát • enter: chọn • ↑↓: di chuyển
```

### Chọn "🔍 Scan Project"
1. Menu sẽ hỏi đường dẫn project
2. Bạn có thể:
   - Nhập đường dẫn thủ công (vd: `./my-project`)
   - Nhấn Enter để quét thư mục hiện tại (`.`)
   - Kéo thả thư mục vào terminal (Windows)

### Chọn "🌐 Web Dashboard"
- Tự động khởi động server tại `http://localhost:7420`
- Mở browser và hiển thị kết quả scan gần nhất

### Chọn "🤖 AI Setup"
- Hiển thị trạng thái Ollama
- Cho phép cài đặt models

---

## 5. Phân tích một dự án

### Bước 1: Chuẩn bị
```bash
cd /path/to/your/project
```

### Bước 2: Chạy quét
```bash
vibescanner scan .
```

### Bước 3: Đọc kết quả terminal

VibeScanner sẽ hiển thị:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  🔍 VibeScanner — Kết quả khám bệnh
  Dự án: my-project    Thời gian: 2s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  🏥 ĐIỂM SỨC KHỎE TỔNG QUÁT: 65/100 🟡 Trung bình

  ┌─────────────┬─────────┬──────────────────────────┐
  │ Hạng mục   │  Điểm  │ Tình trạng               │
  ├─────────────┼─────────┼──────────────────────────┤
  │ 🛡️ Bảo mật │   45   │ 🔴 Cần cải thiện ngay    │
  │ ✨ Chất lượng│   70   │ 🟡 Trung bình            │
  │ 🏗️ Kiến trúc│   60   │ 🟡 Trung bình            │
  └─────────────┴─────────┴──────────────────────────┘

  📊 TỔNG KẾT PHÁT HIỆN
  🔴 Critical: 2  🟠 High: 5  🟡 Medium: 12  🔵 Low: 20

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚨 PHÁT HIỆN NGUY HIỂM CẦN XỬ LÝ NGAY:

[CRI-001] 🔴 SQL Injection vulnerability
    📁 src/db/user.js:45
    📝 User input is directly concatenated into SQL query

    → Lỗi này cho phép hacker xóa toàn bộ database!
      Sửa ngay bằng cách dùng parameterized queries.
```

### Bước 4: Xem chi tiết từng lỗi

Mỗi finding hiển thị:
- **ID**: Mã định danh (vd: CRI-001, HIGH-005)
- **Severity**: Critical / High / Medium / Low / Info
- **Category**: Bảo mật / Chất lượng / Kiến trúc / Secrets
- **File**: Vị trí chính xác (file:line)
- **Mô tả**: Giải thích vấn đề
- **Code snippet**: Đoạn code có vấn đề

---

## 6. Sử dụng Web Dashboard

### Khởi động Dashboard

```bash
# Sau khi đã scan
vibescanner serve
```

Dashboard sẽ mở tại `http://localhost:7420`

### Tính năng Dashboard

#### 1. Health Score Cards
- Hiển thị điểm tổng quát và từng hạng mục
- Màu sắc: 🟢 Xanh (>80) / 🟡 Vàng (60-79) / 🔴 Đỏ (<60)

#### 2. Summary Bar
- Tổng số findings theo severity
- Critical / High / Medium / Low / Info

#### 3. Filters
- **Severity Filter**: Chọn xem findings theo mức độ nghiêm trọng
- **Category Filter**: Lọc theo Bảo mật / Chất lượng / Kiến trúc / Secrets
- **Search**: Tìm theo tên, file, message

#### 4. Finding Cards
- Mỗi card hiển thị một finding
- Click để xem code snippet
- Click "🤖 Hỏi bác sĩ AI" để nhận giải thích

#### 5. AI Explanation
Khi click "Hỏi bác sĩ AI", bạn sẽ nhận được:
- Giải thích lỗi bằng tiếng Việt dễ hiểu
- Hậu quả nếu không sửa
- Code fix cụ thể
- Mức độ ưu tiên

---

## 7. Cài đặt và quản lý AI

### Yêu cầu
- Ít nhất 4GB RAM (cho model 3B)
- 8GB RAM (cho model 7B)

### Cài đặt Ollama và Model

```bash
# Kiểm tra trạng thái
vibescanner ai-setup status

# Cài đặt model nhẹ (nhanh, ít RAM)
vibescanner ai-setup install qwen2.5-coder:1.5b

# Cài đặt model cân bằng (khuyến nghị)
vibescanner ai-setup install qwen2.5-coder:3b

# Cài đặt model chính xác cao (cần nhiều RAM)
vibescanner ai-setup install qwen2.5-coder:7b
```

### Models khuyến nghị

| Model | Dung lượng | RAM cần | Tốc độ | Độ chính xác |
|-------|-----------|---------|--------|--------------|
| qwen2.5-coder:1.5b | ~1GB | 4GB | ⚡ Nhanh | ⭐⭐ |
| qwen2.5-coder:3b | ~2GB | 6GB | 🚀 Tốt | ⭐⭐⭐ |
| qwen2.5-coder:7b | ~4GB | 8GB | 🐢 Chậm | ⭐⭐⭐⭐ |

### Quản lý models

```bash
# Liệt kê đã cài
vibescanner ai-setup list

# Xóa model không dùng
vibescanner ai-setup remove qwen2.5-coder:1.5b
```

---

## 8. Quản lý cấu hình

### File config lưu ở đâu?

- **Windows**: `%APPDATA%\vibescanner\config.json`
  - Thường là: `C:\Users\<username>\AppData\Roaming\vibescanner\config.json`

- **macOS/Linux**: `~/.config/vibescanner/config.json`

### Các tùy chọn cấu hình

```json
{
  "ollama_url": "http://localhost:11434",
  "default_model": "qwen2.5-coder:3b",
  "theme": "dark",
  "auto_open": true,
  "ignore_paths": ["node_modules", ".git", "vendor", "dist", "build"],
  "custom_rules": [],
  "installed_models": []
}
```

### Thiết lập bằng CLI

```bash
# Đổi model mặc định
vibescanner config set default_model qwen2.5-coder:7b

# Tắt tự động mở browser
vibescanner config set auto_open false

# Đổi theme
vibescanner config set theme light
```

---

## 9. Troubleshooting

### Lỗi: "vibescanner không phải là lệnh nội bộ..."

**Nguyên nhân**: Chưa cài đặt global hoặc chưa restart terminal

**Giải pháp**:
```bash
# Cài đặt global
.\vibescanner.exe install  # Windows
./vibescanner install      # macOS/Linux

# Mở terminal mới và thử lại
vibescanner --version
```

### Lỗi: "Không tìm thấy kết quả scan trước đó"

**Nguyên nhân**: Chưa chạy scan lần nào, hoặc scan cũ đã bị xóa

**Giải pháp**:
```bash
# Chạy scan trước
vibescanner scan ./my-project

# Sau đó mới mở dashboard
vibescanner serve
```

### Lỗi: "AI không khả dụng"

**Nguyên nhân**: Ollama chưa chạy hoặc chưa cài model

**Giải pháp**:
```bash
# Kiểm tra trạng thái
vibescanner ai-setup status

# Cài đặt Ollama thủ công nếu cần
# Truy cập: https://ollama.ai/download

# Cài model
vibescanner ai-setup install qwen2.5-coder:3b
```

### Lỗi: Semgrep/Gitleaks không tìm thấy

**Nguyên nhân**: Các công cụ này chưa được cài đặt

**Giải pháp**:
```bash
# Cài Semgrep
pip install semgrep

# Hoặc
curl -L https://github.com/returntocorp/semgrep/releases/latest/download/semgrep-linux-amd64 -o semgrep
chmod +x semgrep
sudo mv semgrep /usr/local/bin/

# Cài Gitleaks
# Truy cập: https://github.com/gitleaks/gitleaks/releases
```

### Lỗi: "Không thể bind port 7420"

**Nguyên nhân**: Port đang được sử dụng

**Giải pháp**:
```bash
# Dùng port khác
vibescanner serve --port 8080
```

### Cần xóa cấu hình và bắt đầu lại?

```bash
# Windows
Remove-Item -Path "$env:APPDATA\vibescanner" -Recurse

# macOS/Linux
rm -rf ~/.config/vibescanner

# Tạo lại config
vibescanner config init
```

---

## 💡 Mẹo sử dụng

1. **Luôn quét trước khi commit**: `vibescanner scan .`
2. **Dùng HTML report cho review**: `vibescanner scan . --report html --open`
3. **Chạy dashboard trong background**: `vibescanner serve &`
4. **Thêm vào git pre-commit hook**: Tự động quét trước mỗi commit
5. **Chọn model phù hợp với RAM**: 3B cho đa số, 7B nếu máy mạnh

---

## 🆘 Cần trợ giúp?

- **Issues**: https://github.com/vibescanner/vibescanner/issues
- **Documentation**: Xem thêm `README.md`
- **Help**: Chạy `vibescanner --help` hoặc `vibescanner [command] --help`

---

<p align="center">
  <strong>Chúc bạn code sạch, không bug! 🎉</strong>
</p>
