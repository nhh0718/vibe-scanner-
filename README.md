# 🔍 VibeScanner

> Công cụ "khám bệnh" codebase cho vibe coders - Chạy hoàn toàn local, code không rời máy

![Version](https://img.shields.io/badge/version-0.3.2-blue)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D)

## ✨ Điểm khác biệt

| Công cụ thông thường       | VibeScanner                                                                                              |
| -------------------------- | -------------------------------------------------------------------------------------------------------- |
| `SQL injection at line 42` | `Dòng 42: Hacker có thể xóa toàn bộ database chỉ bằng cách gõ vào ô search. Sửa như này:` + code example |
| Output cho developer       | Output cho người vibe-code không biết security                                                           |
| Chạy từng tool riêng lẻ    | Một lệnh quét tất cả                                                                                     |
| Kết quả thô, kỹ thuật      | Báo cáo ngôn ngữ đời thường, có priority rõ ràng                                                         |
| Gửi code lên cloud         | **100% local, code không rời máy**                                                                       |

---

## 🚀 Quick Start

### 1. Cài đặt

**Windows (PowerShell Admin):**

```powershell
Invoke-WebRequest -Uri https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-windows-amd64.exe -OutFile vibescanner.exe
.\vibescanner.exe install
```

**macOS Intel:**

```bash
curl -L https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-darwin-amd64 -o vibescanner
chmod +x vibescanner
sudo mv vibescanner /usr/local/bin/
```

**macOS Apple Silicon (M1/M2/M3):**

```bash
curl -L https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-darwin-arm64 -o vibescanner
chmod +x vibescanner
sudo mv vibescanner /usr/local/bin/
# Nếu bị chặn: System Preferences → Security & Privacy → Open Anyway
```

**Linux:**

```bash
curl -L https://github.com/nhh0718/vibe-scanner-/releases/latest/download/vibescanner-linux-amd64 -o vibescanner
chmod +x vibescanner
sudo mv vibescanner /usr/local/bin/
```

### 2. Sử dụng cơ bản

```bash
vibescanner                          # Menu tương tác
vibescanner scan ./my-project        # Quét trực tiếp
vibescanner scan . --report html     # Tạo HTML report
vibescanner serve                    # Mở dashboard
```

---

## 📋 Tất cả lệnh

| Lệnh          | Mô tả          | Ví dụ                     |
| ------------- | -------------- | ------------------------- |
| `vibescanner` | Menu tương tác | `vibescanner`             |
| `scan [path]` | Quét codebase  | `vibescanner scan .`      |
| `serve`       | Mở dashboard   | `vibescanner serve`       |
| `ai-setup`    | Quản lý AI     | `vibescanner ai-setup`    |
| `config`      | Cấu hình       | `vibescanner config list` |
| `install`     | Cài global     | `vibescanner install`     |
| `uninstall`   | Gỡ global      | `vibescanner uninstall`   |
| `update`      | Cập nhật       | `vibescanner update`      |

---

## 🤖 AI Setup (Tùy chọn)

```bash
vibescanner ai-setup        # Menu AI tương tác
vibescanner ai-setup status # Kiểm tra trạng thái
vibescanner ai-setup install # Cài model mặc định
```

---

## 🔧 Troubleshooting

| Lỗi                                 | Giải pháp                                             |
| ----------------------------------- | ----------------------------------------------------- |
| "vibescanner không phải là lệnh..." | Chạy `vibescanner install` và restart terminal        |
| macOS: "cannot be opened"           | System Preferences → Security & Privacy → Open Anyway |
| "Không tìm thấy kết quả scan"       | Chạy `vibescanner scan .` trước khi `serve`           |
| AI không khả dụng                   | Cài Ollama tại https://ollama.ai/download             |

---

## 📄 License

MIT License

<p align="center">Made with ❤️ for vibe coders everywhere</p>
