# VibeScanner — Blueprint xây dựng công cụ "khám bệnh" code

> Tài liệu phân tích toàn diện: kiến trúc, triển khai, lộ trình và chiến lược phát triển.

---

## Mục lục

1. [Tổng quan & bối cảnh](#1-tổng-quan--bối-cảnh)
2. [Phân tích thị trường & cơ hội](#2-phân-tích-thị-trường--cơ-hội)
3. [Kiến trúc hệ thống](#3-kiến-trúc-hệ-thống)
4. [Các module phân tích chi tiết](#4-các-module-phân-tích-chi-tiết)
5. [Stack kỹ thuật](#5-stack-kỹ-thuật)
6. [Cấu trúc dự án](#6-cấu-trúc-dự-án)
7. [Lộ trình phát triển](#7-lộ-trình-phát-triển)
8. [Thiết kế Report Output](#8-thiết-kế-report-output)
9. [AI Layer — Cách tích hợp LLM](#9-ai-layer--cách-tích-hợp-llm)
10. [Hướng dẫn triển khai MVP](#10-hướng-dẫn-triển-khai-mvp)
11. [Chiến lược mở rộng & kinh doanh](#11-chiến-lược-mở-rộng--kinh-doanh)
12. [Rủi ro & cách xử lý](#12-rủi-ro--cách-xử-lý)

---

## 1. Tổng quan & bối cảnh

### Vấn đề

Làn sóng "vibe coding" — lập trình dựa hoàn toàn vào AI (Cursor, Lovable, v0, Bolt...) — đang tạo ra hàng triệu dòng code được deploy lên production mỗi ngày bởi những người không có nền tảng kỹ thuật. Hệ quả:

- **Bảo mật:** SQL injection, hardcoded API keys, CORS mở wildcard, không có rate limiting, JWT secret là `"secret"`.
- **Chất lượng:** God functions dài 500 dòng, không có error handling, magic numbers, không có test.
- **Kiến trúc:** Circular dependencies, business logic nằm trong view layer, không có separation of concerns.
- **Hiệu năng:** N+1 queries, thiếu index database, memory leaks, blocking I/O trong event loop.

### Giải pháp

**VibeScanner** là một công cụ CLI + Web Dashboard chạy hoàn toàn local, kết hợp static analysis engines với một AI layer (LLM local qua Ollama) để "khám bệnh" toàn bộ codebase và xuất ra "hồ sơ bệnh án" dễ hiểu — kể cả với người không có kiến thức bảo mật hay software engineering.

### Điểm khác biệt cốt lõi

| Công cụ thông thường | VibeScanner |
|---|---|
| `SQL injection at line 42` | `Dòng 42: Hacker có thể xóa toàn bộ database chỉ bằng cách gõ vào ô search. Sửa như này:` + code example |
| Output cho developer | Output cho người vibe-code không biết security |
| Chạy từng tool riêng lẻ | Một lệnh quét tất cả |
| Kết quả thô, kỹ thuật | Báo cáo ngôn ngữ đời thường, có priority rõ ràng |
| Gửi code lên cloud | **100% local, code không rời máy** |

---

## 2. Phân tích thị trường & cơ hội

### Đối tượng người dùng chính

**Tier 1 — Indie hacker / Vibe coder (target chính)**
- Không có background lập trình hoặc có ít kinh nghiệm
- Dùng Cursor/Lovable/v0 để build SaaS, tool nội bộ
- Sắp deploy hoặc vừa deploy product
- Câu hỏi của họ: "Code AI tạo ra có an toàn không?"

**Tier 2 — Developer dùng AI assistant**
- Biết lập trình nhưng dùng AI để code nhanh hơn
- Muốn catch các lỗi AI hallucinate trước khi review
- Tích hợp vào pre-commit hook hoặc CI/CD

**Tier 3 — Tech lead / CTO startup**
- Muốn audit codebase của team trước khi scale
- Cần report để trình bày với investor hoặc khách hàng enterprise

### Phân tích đối thủ

| Tool | Mạnh | Yếu | Cơ hội của VibeScanner |
|---|---|---|---|
| Semgrep | Rules mạnh, nhiều ngôn ngữ | Output kỹ thuật, khó hiểu | Wrap Semgrep + giải thích bằng AI |
| SonarQube | Toàn diện | Cài đặt phức tạp, cần server | Chạy local, zero config |
| Snyk | Tốt cho dependency | Tập trung vào packages, không phân tích logic | Phân tích cả code logic |
| GitHub Copilot | Tích hợp IDE | Không có báo cáo tổng thể | Báo cáo toàn bộ dự án |
| CodeRabbit | AI code review | Tốn tiền, phụ thuộc cloud | Free, local, không gửi code ra ngoài |

---

## 3. Kiến trúc hệ thống

```
┌─────────────────────────────────────────────────────────┐
│                     INPUT LAYER                         │
│  Local folder  │  Git repo URL  │  File upload (ZIP)    │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                   INGESTION LAYER                       │
│  • Language detection (linguist)                        │
│  • File tree traversal                                  │
│  • Git history analysis (nếu có)                       │
│  • Dependency extraction (package.json, requirements...)│
└────────────────────────┬────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
┌────────▼───────┐ ┌─────▼─────┐ ┌─────▼──────────┐
│ SECURITY       │ │  QUALITY  │ │  ARCHITECTURE  │
│ ENGINE         │ │  ENGINE   │ │  ENGINE        │
│                │ │           │ │                │
│ • Semgrep      │ │ • Radon   │ │ • Madge        │
│ • Gitleaks     │ │ • ESLint  │ │ • Dep-cruiser  │
│ • Bandit       │ │ • Pylint  │ │ • Custom rules │
│ • TruffleHog   │ │ • Custom  │ │                │
└────────┬───────┘ └─────┬─────┘ └─────┬──────────┘
         │               │             │
┌────────▼───────────────▼─────────────▼────────────────┐
│                  AGGREGATION LAYER                      │
│  • Deduplicate findings                                 │
│  • Score severity (Critical / High / Medium / Low)      │
│  • Group by file / category / severity                  │
│  • Build context cho AI layer                           │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                    AI SYNTHESIS LAYER                   │
│  Ollama (local) — Llama 3.1 / Qwen2.5-Coder            │
│                                                         │
│  • Giải thích lỗi bằng ngôn ngữ đơn giản               │
│  • Đề xuất code fix cụ thể                             │
│  • Tổng hợp "executive summary"                         │
│  • Trả lời câu hỏi follow-up về report                  │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                   OUTPUT LAYER                          │
│  CLI summary  │  HTML report  │  JSON (CI/CD)  │  PDF   │
└─────────────────────────────────────────────────────────┘
```

### Nguyên tắc thiết kế

- **Privacy first:** Code không bao giờ rời máy. Mọi phân tích chạy local.
- **Zero config:** Chạy được ngay với một lệnh, không cần file config.
- **Progressive disclosure:** Người không chuyên thấy summary đơn giản, developer thấy chi tiết kỹ thuật.
- **Actionable:** Mỗi issue đi kèm đề xuất fix cụ thể, không chỉ mô tả vấn đề.
- **Funnel Strategy — nguyên tắc sống còn về tài nguyên:** "Tuyệt đối không dùng AI để làm những việc mà code thường có thể làm tốt hơn và nhanh hơn." Dùng static analysis siêu nhẹ để lọc 90% findings, chỉ đánh thức AI khi user chủ động yêu cầu giải thích một issue cụ thể (on-demand). AI không bao giờ chạy ngầm.
- **Lightweight core:** Engine quét viết bằng Go — biên dịch ra binary, goroutines quét file song song, không cần runtime. Đảm bảo chạy mượt từ máy văn phòng CPU-only đến laptop phổ thông (ASUS TUF A15).

---

## 4. Các module phân tích chi tiết

### 4.1 Security Engine

#### 4.1.1 Injection vulnerabilities
- SQL injection (parameterized queries, ORM misuse)
- NoSQL injection (MongoDB query injection)
- Command injection (`exec`, `system`, `eval`)
- Path traversal (`../../../etc/passwd`)
- Server-Side Template Injection (SSTI)
- LDAP injection

**Tool:** Semgrep với ruleset `p/owasp-top-ten`, `p/sql-injection`, `p/javascript`

#### 4.1.2 Authentication & Authorization
- Hardcoded credentials trong source code
- JWT secret yếu (`"secret"`, `"password"`, `"123456"`)
- Thiếu authentication middleware trên route nhạy cảm
- IDOR (Insecure Direct Object Reference) — user A truy cập data user B
- Session không expire
- Password không được hash (lưu plain text)
- Bcrypt cost factor quá thấp (< 10)

#### 4.1.3 Secrets & Credentials Detection
Sử dụng entropy analysis + pattern matching:
- API keys (AWS, Stripe, OpenAI, Google, Twilio...)
- Private keys (RSA, SSH, PGP)
- Database connection strings có password
- OAuth tokens
- `.env` file bị commit vào git

**Tool:** Gitleaks, TruffleHog

#### 4.1.4 Web Security (nếu là web app)
- CORS mở wildcard (`Access-Control-Allow-Origin: *`)
- CSP không được set hoặc quá lỏng
- XSS — `innerHTML`, `dangerouslySetInnerHTML` với user input
- CSRF — thiếu token hoặc SameSite cookie
- Clickjacking — thiếu `X-Frame-Options`
- Rate limiting không có
- Input không được validate/sanitize

#### 4.1.5 Insecure Dependencies
- Package có CVE đã biết
- Package version quá cũ (> 2 major versions)
- Package deprecated

**Tool:** `npm audit`, `pip-audit`, `safety`

### 4.2 Code Quality Engine

#### 4.2.1 Complexity metrics
- **Cyclomatic complexity:** Số lượng independent paths qua một function. > 10 là cần refactor.
- **Cognitive complexity:** Mức độ khó hiểu của code. > 15 là critical.
- **Function length:** > 50 dòng cần xem xét, > 100 dòng là vấn đề.
- **File length:** > 300 dòng thường là dấu hiệu God Object.
- **Nesting depth:** > 4 level là quá sâu.
- **Parameter count:** > 5 parameter là code smell.

**Tool:** Radon (Python), ESLint `complexity` rule, code-complexity (JS)

#### 4.2.2 Code smells
- **Dead code:** Functions, variables được định nghĩa nhưng không dùng
- **Duplicate code:** Đoạn code giống nhau ở nhiều nơi (clone detection)
- **Magic numbers/strings:** `if (status === 3)` thay vì `if (status === ORDER_SHIPPED)`
- **Long parameter lists:** Cần gom vào object
- **God class:** Class có quá nhiều responsibilities
- **Feature envy:** Method dùng data của class khác nhiều hơn class mình
- **Shotgun surgery:** Một thay đổi nhỏ phải sửa nhiều file

#### 4.2.3 Error handling
- Empty catch blocks: `catch(e) {}` — nuốt lỗi im lặng
- `console.log` thay vì proper logging
- Không handle promise rejection
- Unhandled async errors
- Error messages lộ thông tin nhạy cảm (stack trace, database schema)

#### 4.2.4 Testing
- Test coverage < 50% (critical), < 80% (warning)
- Test files nhưng không có assertions
- Hardcoded test data thay vì fixtures/factories

### 4.3 Architecture Engine

#### 4.3.1 Dependency analysis
- **Circular dependencies:** Module A import B, B import C, C import A
- **Coupling score:** Module phụ thuộc vào quá nhiều module khác
- **Cohesion score:** Module có quá nhiều chức năng không liên quan
- **Dependency direction violations:** UI layer import database layer trực tiếp

**Tool:** Madge (JS/TS), dependency-cruiser

#### 4.3.2 Architecture patterns detection
Phát hiện vi phạm các pattern phổ biến:
- **MVC violations:** Business logic trong Controller/View
- **Repository pattern:** Data access code rải rác khắp codebase
- **SOLID violations:**
  - Single Responsibility: Class/function làm quá nhiều thứ
  - Open/Closed: Code cần sửa thay vì extend khi thêm feature
  - Dependency Inversion: High-level module phụ thuộc implementation details

#### 4.3.3 API design
- REST inconsistencies (mix GET/POST cho cùng operation)
- Thiếu versioning (`/api/users` thay vì `/api/v1/users`)
- Response schema không nhất quán
- Missing pagination trên list endpoints
- Trả về 200 cho errors thay vì đúng HTTP status code

### 4.4 Performance Engine

#### 4.4.1 Database
- **N+1 queries:** Query trong vòng lặp
- **Missing indexes:** Foreign keys không có index, cột thường filter không có index
- **Select *:** Query tất cả columns khi chỉ cần vài cột
- **Large data sets:** Load toàn bộ table vào memory

#### 4.4.2 JavaScript/Node.js đặc thù
- **Blocking event loop:** Sync operations (fs.readFileSync) trong async context
- **Memory leaks:** Event listeners không được cleanup, closures giữ reference
- **Bundle size:** Import cả thư viện khi chỉ cần một function (`import _ from 'lodash'`)
- **Render blocking:** Script không có `async`/`defer`

#### 4.4.3 Vibe-coding specific patterns
Các lỗi đặc trưng của code được tạo bởi AI:
- Regenerate session ID sau mỗi request (vì AI không hiểu stateful sessions)
- Fetch toàn bộ data rồi filter ở application layer thay vì filter ở DB
- Duplicate middleware (cùng một middleware được đăng ký nhiều lần)
- Environment check sai (`if (env === 'production')` thay vì `if (process.env.NODE_ENV === 'production')`)

---

## 5. Stack kỹ thuật

### 5.1 Core Technologies

#### Core engine / CLI
```
Language:      Go 1.22+
CLI framework: Cobra (standard Go CLI)
Concurrency:   Goroutines (quét file song song, không overhead)
Embed UI:      go:embed (nuốt toàn bộ Vue build vào binary)
Build/release: goreleaser (cross-compile Windows/Linux/macOS 1 lệnh)
```

**Lý do chọn Go thay vì Python:**
- Biên dịch ra native binary — không cần user cài Python, Node, hay bất cứ runtime nào
- Goroutines quét hàng ngàn file song song mà không làm nặng máy
- `go:embed` cho phép nhúng toàn bộ Vue dashboard vào 1 file thực thi duy nhất
- Cross-compile trivial: `GOOS=windows GOARCH=amd64 go build` → ra `.exe`
- Semgrep và Gitleaks gọi qua subprocess (`exec.Command`) — hoạt động tốt, không cần Python API

#### Frontend Dashboard
```
Framework:  Vue 3 + Vite
Styling:    Tailwind CSS
Charts:     Chart.js
State:      Pinia
Đóng gói:  go:embed (build Vue → static files → Go nuốt vào binary)
```

Vue được build ra `dist/` tĩnh, Go embed toàn bộ folder đó. Kết quả: user chạy `vibescanner scan .` → Go tự mở local web server → trình duyệt hiển thị dashboard — không cần cài Node.js hay bất kỳ thứ gì.

#### AI Layer (On-Demand, không chạy ngầm)
```
Runtime:    Ollama (subprocess, không nhúng llama.cpp trực tiếp)
Giao tiếp: REST API localhost:11434 (streaming SSE)
Model mặc định: Qwen2.5-Coder:3B Q4_K_M  ← điểm ngọt nhất
Quantization:   GGUF 4-bit bắt buộc
Context window: 2048–4096 tokens (chỉ chunk function bị lỗi ± 20 dòng)
```

**Lý do không nhúng llama.cpp trực tiếp vào Go:**
Nhúng llama.cpp đòi hỏi CGo bridge (C++ ↔ Go), phá vỡ khả năng cross-compile đơn giản và thêm ~3–4 tuần complexity không cần thiết. Ollama subprocess cho kết quả tương đương, API sạch hơn, và Ollama đã xử lý toàn bộ CUDA/Metal/CPU fallback.

### 5.2 Chiến lược đóng gói — 3 tier

Đây là quyết định kiến trúc quan trọng nhất, giải quyết bài toán "nặng không?":

```
TIER 1 — Core (bắt buộc, tải 1 lần):          ~30 MB
├── vibescanner binary (Go + Vue embedded)      ~10 MB
└── Semgrep rules YAML                          ~20 MB
    → Đủ để chạy toàn bộ static analysis
    → AI chưa cần, máy hoàn toàn mát

TIER 2 — AI Pack (tùy chọn, tải 1 lần):       ~1–4 GB
├── Ollama binary                               ~68 MB
└── Model GGUF (user chọn theo máy):
    • Qwen2.5-Coder:1.5B Q4  → 930 MB  (CPU yếu, RAM < 4GB)
    • Qwen2.5-Coder:3B Q4    → 1.9 GB  (khuyến nghị, điểm ngọt)
    • Qwen2.5-Coder:7B Q4    → 4.1 GB  (máy mạnh, chất lượng tốt hơn)
    → Lưu mãi tại ~/.vibescanner/ai/
    → Chỉ tải 1 lần, dùng offline mãi mãi

TIER 3 — Detect sẵn có (zero download):
    Nếu user đã cài Ollama → dùng luôn, không tải gì thêm
    Nếu chưa có → hướng dẫn chạy: vibescanner ai-setup
```

**Luồng cài đặt thực tế:**

```bash
# Bước 1: Tải core (~30MB), chạy được ngay
curl -L https://github.com/you/vibescanner/releases/latest/download/vibescanner-linux-amd64 -o vibescanner
chmod +x vibescanner
./vibescanner scan ./my-project    # ← AI chưa cần, scan tĩnh đã chạy

# Bước 2 (tùy chọn): Kích hoạt AI
./vibescanner ai-setup
# Tool tự detect Ollama → nếu chưa có thì tải + cài tự động
# Hỏi user chọn model theo RAM máy
# Pull model về ~/.vibescanner/ai/models/
# Từ đây: nút "Hỏi bác sĩ AI" trong dashboard hoạt động
```

### 5.3 Scanner Dependencies — gọi qua subprocess

Không cần user cài bất kỳ tool nào. VibeScanner tự bundle hoặc tự tải:

| Tool | Cách bundle | Dùng cho |
|---|---|---|
| Semgrep binary | Bundle vào release package | Security rules chính |
| Gitleaks binary | Bundle vào release package | Secret detection |
| Semgrep rules YAML | Bundle vào binary (go:embed) | Rule definitions |
| Ollama | Tải qua `ai-setup` (nếu chưa có) | AI explanations |
| GGUF model | Tải qua `ai-setup` (user chọn) | LLM inference |

Semgrep và Gitleaks binary được đặt trong `~/.vibescanner/bin/` sau lần chạy đầu tiên, không cần quyền admin.

### 5.4 Data Flow & Storage

Tất cả local, không có database server, không có cloud:

```
~/.vibescanner/
├── bin/            # Semgrep, Gitleaks binaries (tự tải lần đầu)
├── cache/          # Cache kết quả scan theo hash file (incremental scan)
├── reports/        # Lịch sử các lần scan (SQLite)
├── rules/          # Custom Semgrep rules của user
└── ai/
    ├── ollama      # Ollama binary (nếu dùng ai-setup)
    └── models/     # GGUF models (lưu mãi, không tải lại)
```

Kết quả scan lưu SQLite cho fast querying, export ra JSON/HTML/PDF.

---

## 6. Cấu trúc dự án

```
vibescanner/
├── README.md
├── go.mod
├── go.sum
├── main.go                         # Entry point
├── Makefile                        # build, release, dev tasks
├── .goreleaser.yaml                # Cross-compile config
│
├── cmd/                            # CLI commands (Cobra)
│   ├── root.go                     # Root command, global flags
│   ├── scan.go                     # `vibescanner scan <path>`
│   ├── serve.go                    # `vibescanner serve` → mở dashboard
│   └── ai_setup.go                 # `vibescanner ai-setup` → cài AI pack
│
├── internal/
│   ├── ingestion/
│   │   ├── walker.go               # File tree traversal, .vibescannerignore
│   │   ├── detector.go             # Language detection (by extension + content)
│   │   └── git.go                  # Git history, check file committed
│   │
│   ├── engines/
│   │   ├── semgrep.go              # Gọi Semgrep subprocess, parse JSON output
│   │   ├── gitleaks.go             # Gọi Gitleaks subprocess, parse output
│   │   ├── complexity.go           # Đo complexity bằng Go (tree-sitter bindings)
│   │   └── deps.go                 # Parse package.json, requirements.txt → audit
│   │
│   ├── aggregation/
│   │   ├── dedup.go                # Loại bỏ findings trùng lặp
│   │   ├── scorer.go               # Tính health score
│   │   └── grouper.go              # Group theo file/severity/category
│   │
│   ├── ai/
│   │   ├── client.go               # Ollama REST API client (streaming)
│   │   ├── chunker.go              # Trích xuất đúng function ± 20 dòng context
│   │   ├── prompts.go              # Prompt templates theo issue type
│   │   └── setup.go                # Logic tải Ollama + model (ai-setup)
│   │
│   ├── output/
│   │   ├── terminal.go             # Colored terminal output
│   │   ├── json.go                 # JSON export cho CI/CD
│   │   └── pdf.go                  # PDF export
│   │
│   └── store/
│       └── sqlite.go               # Lưu lịch sử scan, cache findings
│
├── web/                            # Vue 3 dashboard source
│   ├── src/
│   │   ├── App.vue
│   │   ├── components/
│   │   │   ├── FindingCard.vue     # Card từng issue, nút "Hỏi bác sĩ AI"
│   │   │   ├── HealthScore.vue     # Gauge 0-100
│   │   │   └── AiStream.vue        # Hiển thị AI response streaming
│   │   └── stores/
│   │       └── scan.ts             # Pinia store
│   └── dist/                       # Build output → go:embed nuốt vào binary
│
├── rules/                          # Custom Semgrep rules (go:embed)
│   ├── vibe_coding_patterns.yaml
│   ├── ai_generated_antipatterns.yaml
│   └── secrets_extended.yaml
│
└── tests/
    ├── fixtures/                   # Sample code với known vulnerabilities
    └── engines/
```

---

## 7. Lộ trình phát triển

### Phase 0 — Validation (1–2 tuần)
**Mục tiêu:** Kiểm chứng concept trước khi đầu tư nhiều thời gian.

**Làm:**
- Viết script Go đơn giản gọi Semgrep subprocess + Gitleaks trên một dự án mẫu
- Parse JSON output, in ra terminal có màu sắc (dùng `github.com/fatih/color`)
- Gọi Ollama REST API để giải thích 5-10 loại issue phổ biến nhất
- Test trên 3-5 project thực từ Lovable/v0

**Thành công nếu:** Output tìm được ít nhất 5 issue thực sự có giá trị trong mỗi project test.

---

### Phase 1 — MVP CLI (4–6 tuần)
**Target:** Developers kỹ thuật, open-source release.

**Features:**
- `vibescanner scan <path>` — quét và xuất kết quả terminal
- `vibescanner scan <path> --report html` — xuất HTML report standalone
- `vibescanner ai-setup` — cài AI pack (Ollama + model, tự động)
- Hỗ trợ: JavaScript/TypeScript, Python
- 4 categories: Security, Quality, Architecture, Secrets
- AI on-demand: chỉ gọi khi user chạy `--explain <finding-id>`, không chạy ngầm
- Severity scoring: Critical / High / Medium / Low / Info
- `.vibescannerignore` để skip folders

**Tech stack MVP:**
```go
// Core dependencies (go.mod)
github.com/spf13/cobra         // CLI framework
github.com/fatih/color         // Terminal colors
github.com/mattn/go-sqlite3    // Lưu scan history
github.com/gin-gonic/gin       // Web server cho dashboard
```

**Đóng gói release:**
```yaml
# .goreleaser.yaml
builds:
  - env: [CGO_ENABLED=0]
    goos: [linux, windows, darwin]
    goarch: [amd64, arm64]
```
Một lệnh `goreleaser release` → ra 6 binary cho đủ platform.

**Milestone:** Release lên GitHub, post lên HackerNews/Reddit r/programming.

---

### Phase 2 — Web Dashboard on-demand (6–8 tuần sau Phase 1)
**Target:** Non-technical vibe coders.

**Features:**
- `vibescanner scan .` → tự động mở http://localhost:7420 trong trình duyệt
- Toàn bộ Vue dashboard đã được `go:embed` vào binary — không cần Node, không cần server riêng
- Drag & drop folder để scan (qua web UI)
- Interactive dashboard: filter, search, sort findings
- Nút "Hỏi bác sĩ AI" trên từng finding card — gọi Ollama on-demand, stream response
- Trend view: so sánh với lần scan trước
- Export PDF / HTML report standalone

**Luồng on-demand AI (quan trọng):**
```
User thấy 50 findings trong dashboard
           │
           │  AI ĐANG NGỦ — máy hoàn toàn mát
           │
User bấm "Hỏi bác sĩ AI" trên finding #7
           │
Go gọi Ollama REST API với:
  - Chỉ đoạn function bị lỗi (~30 dòng)
  - Context window 2048-4096 tokens
           │
Response stream từng chữ về Vue (SSE)
           │
AI ngủ lại — dù project có 10.000 findings
```

**Mở rộng ngôn ngữ:**
- PHP, Go, Ruby (thêm Semgrep rulesets)
- SQL files (schema review)

**UX principles:**
- Dùng ngôn ngữ đời thường, không jargon
- Mỗi issue có: "Tại sao nguy hiểm" + "Cách sửa" + "Code example"
- Health score 0-100 như khám sức khỏe tổng quát

---

### Phase 3 — Ecosystem (3–4 tháng sau Phase 2)
**Target:** Developer teams, CI/CD integration.

**Features:**
- VS Code extension: highlight issues inline
- GitHub Action: auto-scan trên PR
- Pre-commit hook integration
- Baseline support: chỉ report issues mới, không report issues đã biết
- Custom rules builder (GUI)
- Team dashboard: aggregate reports từ nhiều project
- Webhook notifications (Slack, Discord)

**AI nâng cao:**
- So sánh code trước/sau: "Bản này tệ hơn bản cũ ở những điểm này"
- Học từ feedback: user đánh dấu false positive, model học để không báo lại
- Generate PR description tóm tắt security fixes

---

### Phase 4 — Monetization (song song với Phase 3)
**Model đề xuất:** Open-core

| Tier | Giá | Features |
|---|---|---|
| Community | Free | CLI, cơ bản, Ollama local |
| Pro | $9/tháng | Web dashboard, PDF reports, history |
| Team | $29/tháng | Multi-project, CI/CD, custom rules |
| Enterprise | Custom | Self-hosted, SSO, audit logs |

---

## 8. Thiết kế Report Output

### 8.1 Terminal Output (CLI)

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  🔍 VibeScanner — Kết quả khám bệnh
  Dự án: my-saas-app    Thời gian: 2m 34s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  🏥 ĐIỂM SỨC KHỎE TỔNG QUÁT: 34/100 ⚠️ CẦN ĐIỀU TRỊ KHẨN CẤP

  ┌─────────────┬─────────┬──────────────────────────┐
  │ Hạng mục   │  Điểm  │ Tình trạng               │
  ├─────────────┼─────────┼──────────────────────────┤
  │ 🔴 Bảo mật │  12/100 │ NGUY HIỂM                │
  │ 🟡 Chất lượng│ 45/100 │ CẦN CẢI THIỆN            │
  │ 🟡 Kiến trúc│ 55/100 │ CẦN CẢI THIỆN            │
  │ 🟢 Hiệu năng│ 70/100 │ KHÁ TỐT                  │
  └─────────────┴─────────┴──────────────────────────┘

  PHÁT HIỆN: 3 Critical · 12 High · 28 Medium · 45 Low

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  🚨 CRITICAL — Cần xử lý NGAY trước khi deploy
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  [C-001] 🔑 API Key OpenAI lộ trong source code
  File: src/utils/ai-helper.js:15
  
  Vấn đề: API key của bạn đang nằm thẳng trong code.
  Bất kỳ ai xem source code (kể cả trên GitHub) đều có thể
  lấy key này và dùng tốn tiền tài khoản của bạn.
  
  Sửa ngay:
  ❌  const openai = new OpenAI({ apiKey: "sk-abc123..." })
  ✅  const openai = new OpenAI({ apiKey: process.env.OPENAI_API_KEY })
  
  Và thêm vào file .env (đừng commit file này lên git!)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 8.2 HTML Report Structure

Report HTML gồm các section:

**Executive Summary:** Health score, số issues theo severity, top 3 vấn đề cần ưu tiên, estimated fix time.

**Hồ sơ bệnh án chi tiết:** Mỗi issue có card với đầy đủ: vị trí file, mô tả đời thường, tại sao nguy hiểm, hướng dẫn sửa với code diff, links tham khảo.

**Dependency Report:** Danh sách packages có CVE, package quá cũ.

**Architecture Map:** Visualize dependency graph, highlight circular deps.

**Lịch sử:** So sánh với lần scan trước, trend chart.

### 8.3 JSON Output (cho CI/CD)

```json
{
  "scan_id": "uuid-...",
  "timestamp": "2025-01-15T10:30:00Z",
  "project": {
    "name": "my-saas-app",
    "path": "/home/user/projects/my-saas-app",
    "languages": ["javascript", "typescript"],
    "files_scanned": 234
  },
  "health_score": {
    "overall": 34,
    "security": 12,
    "quality": 45,
    "architecture": 55,
    "performance": 70
  },
  "summary": {
    "critical": 3,
    "high": 12,
    "medium": 28,
    "low": 45,
    "info": 18
  },
  "findings": [
    {
      "id": "C-001",
      "severity": "critical",
      "category": "security",
      "subcategory": "secret_exposure",
      "title": "Hardcoded API key detected",
      "file": "src/utils/ai-helper.js",
      "line": 15,
      "code_snippet": "const openai = new OpenAI({ apiKey: \"sk-abc123...\" })",
      "explanation": "...",
      "fix_suggestion": "...",
      "references": ["https://owasp.org/..."],
      "false_positive_likelihood": "low"
    }
  ]
}
```

---

## 9. AI Layer — Cách tích hợp LLM

### 9.1 Triết lý: AI là bác sĩ chuyên khoa, không phải y tá trực 24/7

Static analysis (Semgrep, Gitleaks) làm 90% công việc nhanh và rẻ. AI chỉ được gọi khi user chủ động bấm "Hỏi bác sĩ AI" trên một finding cụ thể. Không có batch AI, không có auto-explain, không chạy ngầm.

AI layer làm 3 việc mà static analysis không làm được:
1. **Dịch** từ jargon kỹ thuật sang ngôn ngữ đời thường cho người không chuyên
2. **Contextualize** — giải thích tại sao lỗi này nguy hiểm với *project cụ thể* của user
3. **Generate fix** — tạo code fix phù hợp với framework và style hiện tại

### 9.2 Kiến trúc On-Demand AI

```
User bấm "Hỏi bác sĩ AI" trên finding
              │
              ▼
    Chunker (internal/ai/chunker.go)
    - Đọc file tại finding.File
    - Trích xuất đúng function bị lỗi
    - Lấy ±20 dòng context xung quanh
    - Giới hạn ~2048 tokens tổng
              │
              ▼
    Prompt Builder (internal/ai/prompts.go)
    - Chọn template theo issue type
    - Inject: code chunk, framework, language
              │
              ▼
    Ollama Client (internal/ai/client.go)
    - POST localhost:11434/api/generate
    - Stream=true → nhận từng token
              │
              ▼
    SSE stream về Vue frontend
    - Từng chữ hiện ra như ChatGPT
    - User thấy response ngay lập tức
```

### 9.3 Model Selection — chốt theo tier máy

| Model | File GGUF | RAM dùng | Chất lượng | Dành cho |
|---|---|---|---|---|
| Qwen2.5-Coder:1.5B Q4_K_M | 930 MB | ~1.2 GB | Tạm — hay hallucinate với context phức tạp | CPU yếu, RAM < 6GB |
| **Qwen2.5-Coder:3B Q4_K_M** | **1.9 GB** | **~2.2 GB** | **Tốt — điểm ngọt nhất** | **Khuyến nghị mặc định** |
| Qwen2.5-Coder:7B Q4_K_M | 4.1 GB | ~4.8 GB | Rất tốt | Laptop >= 16GB RAM |
| Qwen2.5-Coder:7B Q8 | 7.2 GB | ~8 GB | Xuất sắc | Máy mạnh, muốn chất lượng cao |

`ai-setup` hỏi user RAM máy → tự chọn model phù hợp. User có thể đổi sau bằng `vibescanner ai-setup --model qwen2.5-coder:7b`.

### 9.4 Prompt Strategy

**Nguyên tắc:**
- Prompt ngắn, tập trung một issue — không nhét cả file
- Chunker đảm bảo không vượt 2048 tokens ngay cả với function dài
- Output có cấu trúc rõ để Go parse được
- Dùng few-shot examples cho từng loại issue

**Ví dụ prompt template:**

```go
// internal/ai/prompts.go
var SecurityPrompt = `Bạn là security expert giải thích cho người không có kiến thức kỹ thuật.

Dự án: {{.ProjectType}} viết bằng {{.Language}}, framework {{.Framework}}
File: {{.FilePath}}
Loại lỗi: {{.IssueType}}

Code có vấn đề:
` + "```" + `{{.Language}}
{{.CodeChunk}}
` + "```" + `

Trả lời ngắn gọn theo đúng format này:

NGUY_HIỂM: [1-2 câu, ví dụ cụ thể hacker có thể làm gì với lỗi này]
SỬA_NHƯ_NÀY:
` + "```" + `{{.Language}}
[code đã sửa]
` + "```" + `
ƯU_TIÊN: [NGAY_BÂY_GIỜ / TUẦN_NÀY / KHI_CÓ_THỜI_GIAN]`
```

### 9.5 Ollama Client trong Go

```go
// internal/ai/client.go
package ai

import (
    "bufio"
    "encoding/json"
    "net/http"
)

type OllamaClient struct {
    BaseURL string  // default: http://localhost:11434
    Model   string  // default: qwen2.5-coder:3b
}

func (c *OllamaClient) StreamExplain(prompt string, out chan<- string) error {
    body := map[string]any{
        "model":  c.Model,
        "prompt": prompt,
        "stream": true,
        "options": map[string]any{
            "temperature":   0.1,    // Thấp để output nhất quán
            "num_ctx":       2048,   // Context window
            "num_predict":   512,    // Max output tokens
        },
    }
    // POST → đọc NDJSON stream → push từng token vào channel
    // Vue nhận qua SSE endpoint /api/ai/stream/:findingId
}

func IsOllamaAvailable() bool {
    resp, err := http.Get("http://localhost:11434/api/tags")
    return err == nil && resp.StatusCode == 200
}
```

### 9.6 Caching AI responses

AI calls tốn thời gian (2–10 giây/finding). Cache để không gọi lại:

```go
// Cache key = sha256(model + issue_type + code_chunk)
// Lưu trong ~/.vibescanner/cache/ai/
// TTL: 30 ngày (rules và model không đổi thường xuyên)
```

Lần đầu gọi: ~5 giây, hiện streaming. Lần sau cùng finding: instant từ cache.

---

## 10. Hướng dẫn triển khai MVP

### 10.1 Thiết lập môi trường dev

```bash
# Cài Go 1.22+
brew install go          # macOS
# hoặc https://go.dev/dl/ cho Windows/Linux

# Clone và setup
git clone https://github.com/you/vibescanner
cd vibescanner
go mod download

# Cài goreleaser để build release
brew install goreleaser

# Dev: build và chạy
go build -o vibescanner . && ./vibescanner scan ./testdata/vulnerable-app

# Cài Ollama (cho AI features khi dev)
curl -fsSL https://ollama.ai/install.sh | sh
ollama pull qwen2.5-coder:3b
```

### 10.2 Implementation MVP — Các bước cụ thể

#### Bước 1: CLI Entry Point (Cobra)

```go
// cmd/scan.go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/you/vibescanner/internal/engines"
    "github.com/you/vibescanner/internal/output"
)

var scanCmd = &cobra.Command{
    Use:   "scan [path]",
    Short: "Quét và phân tích toàn bộ codebase",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        reportFormat, _ := cmd.Flags().GetString("report")
        openBrowser, _ := cmd.Flags().GetBool("open")

        results, err := engines.ScanProject(args[0])
        if err != nil {
            return err
        }

        switch reportFormat {
        case "html":
            return output.GenerateHTML(results, args[0], openBrowser)
        case "json":
            return output.WriteJSON(results)
        default:
            output.PrintTerminal(results)
        }
        return nil
    },
}

func init() {
    scanCmd.Flags().String("report", "terminal", "Output format: terminal/html/json")
    scanCmd.Flags().Bool("open", true, "Tự động mở browser sau khi tạo HTML report")
    rootCmd.AddCommand(scanCmd)
}
```

#### Bước 2: Core Scanner — chạy song song bằng goroutines

```go
// internal/engines/scanner.go
package engines

import (
    "sync"
    "github.com/you/vibescanner/internal/aggregation"
)

func ScanProject(path string) (*ScanResult, error) {
    var wg sync.WaitGroup
    var mu sync.Mutex
    var allFindings []Finding

    collect := func(findings []Finding) {
        mu.Lock()
        allFindings = append(allFindings, findings...)
        mu.Unlock()
    }

    // Chạy song song — máy không bị đơ
    wg.Add(3)
    go func() { defer wg.Done(); f, _ := RunSemgrep(path); collect(f) }()
    go func() { defer wg.Done(); f, _ := RunGitleaks(path); collect(f) }()
    go func() { defer wg.Done(); f, _ := RunComplexity(path); collect(f) }()
    wg.Wait()

    findings := aggregation.Deduplicate(allFindings)
    score := aggregation.CalculateHealthScore(findings)

    return &ScanResult{
        ProjectPath:  path,
        Findings:     findings,
        HealthScore:  score,
    }, nil
}
```

#### Bước 3: Gọi Semgrep subprocess

```go
// internal/engines/semgrep.go
package engines

import (
    "encoding/json"
    "os/exec"
)

func RunSemgrep(path string) ([]Finding, error) {
    // Semgrep binary nằm tại ~/.vibescanner/bin/semgrep
    semgrepBin := getSemgrepBin() // auto-download nếu chưa có

    cmd := exec.Command(semgrepBin,
        "--config", getRulesPath(),   // rules được go:embed sẵn
        "--json",
        "--quiet",
        path,
    )
    out, err := cmd.Output()
    if err != nil {
        // exit code 1 = found issues — không phải lỗi thật
        if _, ok := err.(*exec.ExitError); !ok {
            return nil, err
        }
    }

    var result struct {
        Results []struct {
            RuleID  string `json:"check_id"`
            Path    string `json:"path"`
            Start   struct{ Line int } `json:"start"`
            Extra   struct{ Message string; Severity string } `json:"extra"`
        } `json:"results"`
    }
    json.Unmarshal(out, &result)

    var findings []Finding
    for _, r := range result.Results {
        findings = append(findings, Finding{
            RuleID:   r.RuleID,
            File:     r.Path,
            Line:     r.Start.Line,
            Message:  r.Extra.Message,
            Severity: parseSeverity(r.Extra.Severity),
            Category: "security",
        })
    }
    return findings, nil
}
```

#### Bước 4: go:embed Vue Dashboard

```go
// internal/output/server.go
package output

import (
    "embed"
    "net/http"
    "github.com/gin-gonic/gin"
)

//go:embed ../../web/dist
var webDist embed.FS

func ServeDashboard(results *ScanResult, port int) error {
    r := gin.New()

    // Serve Vue SPA
    r.StaticFS("/app", http.FS(webDist))

    // API endpoints cho Vue gọi
    r.GET("/api/scan", func(c *gin.Context) {
        c.JSON(200, results)
    })

    // AI streaming endpoint
    r.GET("/api/ai/explain/:findingId", func(c *gin.Context) {
        findingId := c.Param("findingId")
        finding := results.FindByID(findingId)

        c.Header("Content-Type", "text/event-stream")
        c.Header("Cache-Control", "no-cache")

        // Stream Ollama response về client
        ai.StreamExplain(finding, c.Writer)
    })

    return r.Run(fmt.Sprintf(":%d", port))
}
```

#### Bước 5: ai-setup command

```go
// cmd/ai_setup.go
package cmd

var aiSetupCmd = &cobra.Command{
    Use:   "ai-setup",
    Short: "Cài đặt AI Pack (Ollama + model)",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Check Ollama đã có chưa
        if ai.IsOllamaAvailable() {
            fmt.Println("Ollama đã được cài. Bỏ qua bước tải Ollama.")
        } else {
            fmt.Println("Đang tải Ollama...")
            if err := ai.DownloadOllama(); err != nil {
                return err
            }
        }

        // 2. Hỏi user chọn model
        model := promptModelChoice() // hiển thị bảng model theo RAM
        
        // 3. Pull model
        fmt.Printf("Đang tải model %s (có thể mất vài phút)...\n", model)
        return ai.PullModel(model)
    },
}
```

### 10.3 Custom Semgrep Rules cho Vibe Coding

```yaml
# rules/vibe_coding_patterns.yaml
rules:
  - id: vibescan-cors-wildcard
    patterns:
      - pattern: |
          app.use(cors({ origin: "*" }))
    message: |
      CORS wildcard cho phép BẤT KỲ website nào gọi API của bạn.
      Chỉ định domain cụ thể: cors({ origin: "https://yourdomain.com" })
    severity: ERROR
    languages: [javascript, typescript]
    
  - id: vibescan-jwt-weak-secret
    patterns:
      - pattern: |
          jwt.sign($PAYLOAD, "secret", ...)
      - pattern: |
          jwt.sign($PAYLOAD, "password", ...)
      - pattern: |
          jwt.sign($PAYLOAD, "123456", ...)
    message: |
      JWT secret quá yếu! Ai cũng có thể đoán được và tạo token giả.
      Dùng secret ngẫu nhiên dài ít nhất 32 ký tự:
      const secret = require('crypto').randomBytes(64).toString('hex')
    severity: ERROR
    languages: [javascript, typescript]
    
  - id: vibescan-no-rate-limit
    patterns:
      - pattern: |
          app.post("/login", ...)
      - pattern-not: |
          app.post("/login", rateLimit(...), ...)
    message: |
      Route /login không có rate limiting. Hacker có thể thử
      hàng triệu mật khẩu tự động (brute force attack).
      Thêm: const limiter = rateLimit({ windowMs: 15*60*1000, max: 5 })
    severity: WARNING
    languages: [javascript, typescript]
    
  - id: vibescan-dotenv-committed
    pattern: |
      DB_PASSWORD=$VALUE
    paths:
      include:
        - "**/.env"
        - "**/.env.production"
    message: |
      File .env đang được commit! Xóa ngay và thêm vào .gitignore.
      Nếu đã push lên GitHub: rotate tất cả credentials ngay lập tức.
    severity: ERROR
    languages: [generic]
```

---

## 11. Chiến lược mở rộng & kinh doanh

### 11.1 Go-to-market

**Giai đoạn 1 — Community (tháng 1-3):**
- Release open-source lên GitHub
- Post lên: HackerNews, Reddit r/webdev r/node r/Python, Product Hunt
- Tạo content: "Tôi quét 100 project từ Lovable và tìm thấy..."
- Target: 500 GitHub stars, 50 active users

**Giai đoạn 2 — Traction (tháng 3-6):**
- Demo video quét một project thực, live
- Partnership với Lovable/Cursor communities
- Blog series: "Top 10 lỗi bảo mật trong code AI tạo ra"
- Target: 2000 stars, 200 users, $500 MRR

**Giai đoạn 3 — Scale (tháng 6-12):**
- Paid tiers
- VS Code extension
- GitHub Marketplace integration
- Target: 5000 stars, $5000 MRR

### 11.2 Positioning

Tránh cạnh tranh trực tiếp với Snyk hay SonarQube (enterprise market). Thay vào đó:

**"The code doctor for vibe coders"** — không phải security tool cho enterprise, mà là người bạn tech-savvy giúp indie hacker kiểm tra code trước khi launch.

Messaging:
- "Cursor builds fast. VibeScanner makes it safe."
- "One command. Know if your vibe code is production-ready."
- "Your AI wrote it. We check if it's safe to ship."

### 11.3 Metrics cần theo dõi

- **Activation:** % users chạy scan thứ 2 trong 7 ngày
- **Value metric:** Số critical issues được tìm ra per scan
- **Retention:** Weekly active scanners
- **Expansion:** Số projects per user

---

## 12. Rủi ro & cách xử lý

### 12.1 Technical risks

**False positives cao (tool báo nhầm quá nhiều):**
- Giải pháp: Cho user feedback (👍👎 từng finding), dùng data này để tune rules
- Baseline: Nhắm < 15% false positive rate

**Performance — scan chậm với project lớn:**
- Giải pháp: Incremental scanning (chỉ scan files đã thay đổi), caching aggressive
- Target: < 60 giây cho project 10k files

**Ollama không available:**
- Giải pháp: Graceful fallback về static explanations, tool vẫn hoạt động không cần AI

**Semgrep rules lỗi thời:**
- Giải pháp: Auto-update rules weekly, pin version, test suite với known vulnerabilities

### 12.2 Business risks

**Anthropic/OpenAI hoặc Cursor tự build feature này:**
- Giải pháp: Focus vào privacy (local-only) và vibe-coder UX — enterprise tools không ưu tiên điều này

**User không dùng tool dù miễn phí:**
- Nguyên nhân thường: Friction khi cài đặt quá cao
- Giải pháp: Làm web version (upload ZIP) để không cần cài gì, convert sau

**False sense of security:**
- Risk: User nghĩ đã scan là an toàn hoàn toàn
- Giải pháp: Luôn note rõ tool không thể thay thế manual security review cho production

### 12.3 Legal/Compliance

- Không lưu code của user, không gửi ra ngoài → GDPR safe
- Open-source tools được dùng (Semgrep, Gitleaks) có license tương thích
- Cần review nếu build SaaS version (code gửi lên server)

---

## Phụ lục — Resources & References

### Tools hữu ích để build

```
go.dev/doc                    — Go documentation
github.com/spf13/cobra        — Go CLI framework
github.com/gin-gonic/gin      — Go web server
github.com/goreleaser/goreleaser — Cross-compile & release
semgrep.dev/docs              — Semgrep documentation & rule registry
github.com/gitleaks/gitleaks  — Gitleaks secret detection
ollama.ai/library             — Available LLM models (GGUF)
vuejs.org/guide               — Vue 3 documentation
vitejs.dev                    — Vite build tool
```

### Semgrep Rule Registries để tích hợp
- `p/owasp-top-ten` — OWASP Top 10 vulnerabilities
- `p/sql-injection` — SQL injection patterns
- `p/javascript` — JavaScript best practices
- `p/typescript` — TypeScript specific
- `p/python` — Python security
- `p/secrets` — Secret/credential detection
- `p/nodejsscan` — Node.js security
- `p/react` — React security patterns

### Benchmark — "Điểm sức khỏe" tham chiếu

| Score | Trạng thái | Ý nghĩa |
|---|---|---|
| 80-100 | 🟢 Tốt | Có thể deploy, monitor regularly |
| 60-79 | 🟡 Trung bình | Fix Critical/High trước khi scale |
| 40-59 | 🟠 Cần cải thiện | Cần sprint fix issues |
| 0-39 | 🔴 Nguy hiểm | Không nên deploy production |

---

*Tài liệu này là living document — cập nhật khi có thông tin mới.*
*Version 2.0 — Tháng 3/2026 — Cập nhật: chuyển core sang Go, AI layer on-demand qua Ollama subprocess, đóng gói 3-tier (core ~30MB + AI Pack tùy chọn)*
