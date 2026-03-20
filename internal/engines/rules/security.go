package rules

import (
	"regexp"
	"strings"
)

// ========== SQL INJECTION RULE ==========

type SQLInjectionRule struct{}

func (r *SQLInjectionRule) ID() string      { return "VS-SEC-001" }
func (r *SQLInjectionRule) Title() string   { return "SQL Injection — nối chuỗi trực tiếp vào câu query" }
func (r *SQLInjectionRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "php", "go"}
}

func (r *SQLInjectionRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	// Regex patterns cho SQL injection
	sqlPatterns := []*regexp.Regexp{
		// JS/TS: db.query("SELECT..." + userInput)
		regexp.MustCompile(`(?i)(query|execute|run|all|get|prepare)\s*\(\s*["'].*(SELECT|INSERT|UPDATE|DELETE|FROM|WHERE).*["']\s*\+\s*([^"\)]+)`),
		// Python: cursor.execute("SELECT..." + user_input)
		regexp.MustCompile(`(?i)(execute|executemany|fetchall|fetchone)\s*\(\s*["'].*(SELECT|INSERT|UPDATE|DELETE|FROM|WHERE).*["']\s*\+\s*([^"\)]+)`),
		// Python f-string: f"SELECT...{var}"
		regexp.MustCompile(`(?i)f["'].*(SELECT|INSERT|UPDATE|DELETE|FROM|WHERE).*\{[^}]+\}`),
		// Template string: `SELECT...${var}`
		regexp.MustCompile(`(?i)` + "`" + `.*(SELECT|INSERT|UPDATE|DELETE|FROM|WHERE).*` + "`" + `\s*\+\s*`),
	}

	for i, line := range lines {
		for _, pattern := range sqlPatterns {
			if pattern.MatchString(line) {
				// Skip if in comment
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") {
					continue
				}
				if strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "*/") {
					continue
				}

				findings = append(findings, Finding{
					RuleID:   r.ID(),
					Title:    r.Title(),
					Description: "Dòng này có thể bị SQL Injection. Hacker có thể nhập ' OR 1=1 -- vào form và xem/xóa toàn bộ dữ liệu.",
					Fix: `// ❌ Nguy hiểm:
db.query("SELECT * FROM users WHERE id = " + userId)

// ✅ An toàn — dùng parameterized query:
db.query("SELECT * FROM users WHERE id = ?", [userId])`,
					File:     file.Path,
					Line:     i + 1,
					Col:      1,
					Snippet:  strings.TrimSpace(line),
					Severity: Critical,
					Category: "security",
					Tags:     []string{"injection", "sql", "owasp-a03"},
				})
				break // Chỉ báo một lần cho mỗi dòng
			}
		}
	}

	return findings
}

// ========== COMMAND INJECTION RULE ==========

type CommandInjectionRule struct{}

func (r *CommandInjectionRule) ID() string    { return "VS-SEC-002" }
func (r *CommandInjectionRule) Title() string { return "Command Injection — thực thi lệnh shell với user input" }
func (r *CommandInjectionRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "php", "go"}
}

func (r *CommandInjectionRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	cmdPatterns := []*regexp.Regexp{
		// JS: exec(), spawn(), execSync() với string concat
		regexp.MustCompile(`(?i)(exec|execSync|spawn)\s*\(\s*["'].*["']\s*\+\s*([^"\)]+)`),
		// Python: os.system(), subprocess.call(), subprocess.Popen()
		regexp.MustCompile(`(?i)(os\.system|subprocess\.call|subprocess\.Popen|subprocess\.run)\s*\(\s*["'].*["']\s*\+\s*([^"\)]+)`),
		// PHP: exec(), system(), shell_exec(), passthru()
		regexp.MustCompile(`(?i)(exec|system|shell_exec|passthru)\s*\(\s*["'].*["']\s*\+\s*([^"\)]+)`),
		// Go: os/exec với fmt.Sprintf
		regexp.MustCompile(`(?i)(exec\.Command|exec\.CommandContext)\s*\(\s*["'].*["']\s*,\s*fmt\.Sprintf\s*\(`),
	}

	for i, line := range lines {
		for _, pattern := range cmdPatterns {
			if pattern.MatchString(line) {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
					continue
				}

				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "User input được nối trực tiếp vào lệnh shell. Hacker có thể chạy lệnh tùy ý trên server.",
					Fix: `// ❌ Nguy hiểm:
exec("ping " + userInput)

// ✅ An toàn:
// Không bao giờ ghép user input vào lệnh shell
// Dùng thư viện có sẵn thay vì shell commands`,
					File:        file.Path,
					Line:        i + 1,
					Col:         1,
					Snippet:     strings.TrimSpace(line),
					Severity:    Critical,
					Category:    "security",
					Tags:        []string{"injection", "command", "rce"},
				})
				break
			}
		}
	}

	return findings
}

// ========== HARDCODED SECRET RULE ==========

type HardcodedSecretRule struct{}

func (r *HardcodedSecretRule) ID() string    { return "VS-SEC-003" }
func (r *HardcodedSecretRule) Title() string { return "Hardcoded Secret — API key/password trong source code" }
func (r *HardcodedSecretRule) Languages() []string { return []string{"*"} }

func (r *HardcodedSecretRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	// Secret patterns
	secretPatterns := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{"AWS Access Key", regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`)},
		{"AWS Secret Key", regexp.MustCompile(`(?i)aws_secret_access_key\s*=\s*["'][A-Za-z0-9/+=]{40}["']`)},
		{"OpenAI API Key", regexp.MustCompile(`sk-[a-zA-Z0-9]{48}`)},
		{"Stripe Live Key", regexp.MustCompile(`sk_live_[a-zA-Z0-9]{24,}`)},
		{"Stripe Test Key", regexp.MustCompile(`sk_test_[a-zA-Z0-9]{24,}`)},
		{"GitHub Token", regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`)},
		{"Generic API Key", regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*["'][a-zA-Z0-9_\-]{16,}["']`)},
		{"Generic Secret", regexp.MustCompile(`(?i)(secret|password|passwd|pwd)\s*[:=]\s*["'][^"']{8,}["']`)},
		{"Private Key", regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`)},
		{"JWT Secret", regexp.MustCompile(`(?i)(jwt[_-]?secret|jwt[_-]?key)\s*[:=]\s*["'][^"']{8,}["']`)},
	}

	for i, line := range lines {
		for _, sp := range secretPatterns {
			if sp.pattern.MatchString(line) {
				// Skip if in comment or string that looks like placeholder
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") {
					continue
				}
				if strings.Contains(line, "example") || strings.Contains(line, "placeholder") || strings.Contains(line, "YOUR_") {
					continue
				}

				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Phát hiện " + sp.name + " trong code. Secrets không được commit vào git vì sẽ bị lộ vĩnh viễn.",
					Fix: `// ❌ Không được:
const API_KEY = "sk-abc123..."

// ✅ Dùng biến môi trường:
const API_KEY = process.env.API_KEY
// Và thêm .env vào .gitignore`,
					File:        file.Path,
					Line:        i + 1,
					Col:         1,
					Snippet:     strings.TrimSpace(line),
					Severity:    Critical,
					Category:    "security",
					Tags:        []string{"secret", "credentials", "owasp-a07"},
				})
				break
			}
		}
	}

	return findings
}

// ========== WEAK JWT SECRET RULE ==========

type WeakJWTSecretRule struct{}

func (r *WeakJWTSecretRule) ID() string    { return "VS-SEC-004" }
func (r *WeakJWTSecretRule) Title() string { return "JWT Secret yếu — dễ bị brute force" }
func (r *WeakJWTSecretRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go"}
}

func (r *WeakJWTSecretRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	weakSecrets := []string{"secret", "password", "123456", "qwerty", "admin", "jwtsecret", "mysecret", "test", "dev"}
	pattern := regexp.MustCompile(`(?i)(jwt[_-]?secret|secret[_-]?key|jwt[_-]?key)\s*[:=]\s*["']([^"']+)["']`)

	for i, line := range lines {
		matches := pattern.FindStringSubmatch(line)
		if len(matches) > 2 {
			secretValue := strings.ToLower(matches[2])
			for _, weak := range weakSecrets {
				if strings.Contains(secretValue, weak) {
					findings = append(findings, Finding{
						RuleID:      r.ID(),
						Title:       r.Title(),
						Description: "JWT secret quá yếu và dễ đoán. Hacker có thể brute force và giả mã token.",
						Fix: `// ❌ Yếu:
jwt.sign(payload, "mysecret")

// ✅ Mạnh — dùng 256-bit random:
jwt.sign(payload, process.env.JWT_SECRET)
// JWT_SECRET trong .env: 256-bit random hex`,
						File:        file.Path,
						Line:        i + 1,
						Col:         1,
						Snippet:     strings.TrimSpace(line),
						Severity:    Critical,
						Category:    "security",
						Tags:        []string{"jwt", "crypto", "weak-secret"},
					})
					break
				}
			}
		}
	}

	return findings
}

// ========== JWT NO VERIFY RULE ==========

type JWTNoVerifyRule struct{}

func (r *JWTNoVerifyRule) ID() string    { return "VS-SEC-005" }
func (r *JWTNoVerifyRule) Title() string { return "JWT không verify signature" }
func (r *JWTNoVerifyRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *JWTNoVerifyRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	pattern := regexp.MustCompile(`(?i)jwt\.decode\s*\(`)

	for i, line := range lines {
		if pattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Dùng jwt.decode() thay vì jwt.verify() là nguy hiểm. Hacker có thể tự tạo token giả mạo.",
					Fix: `// ❌ Không verify signature:
const payload = jwt.decode(token)

// ✅ Verify signature:
const payload = jwt.verify(token, process.env.JWT_SECRET)`,
					File:        file.Path,
					Line:        i + 1,
					Col:         1,
					Snippet:     strings.TrimSpace(line),
					Severity:    High,
					Category:    "security",
					Tags:        []string{"jwt", "auth", "verification"},
			})
		}
	}

	return findings
}

// ========== PLAIN PASSWORD RULE ==========

type PlainPasswordRule struct{}

func (r *PlainPasswordRule) ID() string    { return "VS-SEC-006" }
func (r *PlainPasswordRule) Title() string { return "Password lưu plain text — không hash" }
func (r *PlainPasswordRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "php", "go"}
}

func (r *PlainPasswordRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	// Look for database insert/update with password field
	// Note: Go regex doesn't support negative lookahead, so we filter in code
	pattern := regexp.MustCompile(`(?i)(insert|update|save|create)\s*.*password\s*[:=]\s*`)

	for i, line := range lines {
		if pattern.MatchString(line) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
				continue
			}
			// Skip if already hashed
			if strings.Contains(line, "bcrypt") || strings.Contains(line, "hash") || strings.Contains(line, "argon") {
				continue
			}

			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Password đang được lưu dạng plain text vào database. Nếu DB bị hack, tất cả password bị lộ.",
				Fix: `// ❌ Không được:
await db.users.insert({ email, password: userInput })

// ✅ Hash trước khi lưu:
const hashedPassword = await bcrypt.hash(userInput, 12)
await db.users.insert({ email, password: hashedPassword })`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Critical,
				Category:    "security",
				Tags:        []string{"password", "crypto", "storage"},
			})
		}
	}

	return findings
}

// ========== WEAK BCRYPT RULE ==========

type WeakBcryptRule struct{}

func (r *WeakBcryptRule) ID() string    { return "VS-SEC-007" }
func (r *WeakBcryptRule) Title() string { return "Bcrypt cost factor < 10 — yếu" }
func (r *WeakBcryptRule) Languages() []string {
	return []string{"javascript", "typescript", "python"}
}

func (r *WeakBcryptRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	pattern := regexp.MustCompile(`(?i)bcrypt\.hash\s*\([^,]+,\s*(\d+)\s*\)`)

	for i, line := range lines {
		matches := pattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			// Simple parse - check if cost is < 10
			if matches[1] == "1" || matches[1] == "2" || matches[1] == "3" || matches[1] == "4" ||
			   matches[1] == "5" || matches[1] == "6" || matches[1] == "7" || matches[1] == "8" || matches[1] == "9" {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Bcrypt cost factor quá thấp (< 10). Hash dễ bị crack bằng GPU.",
					Fix: `// ❌ Yếu (dễ crack):
bcrypt.hash(password, 8)

// ✅ Khuyến nghị (OWASP):
bcrypt.hash(password, 12)`,
					File:        file.Path,
					Line:        i + 1,
					Col:         1,
					Snippet:     strings.TrimSpace(line),
					Severity:    Medium,
					Category:    "security",
					Tags:        []string{"bcrypt", "crypto", "hash"},
				})
			}
		}
	}

	return findings
}

// ========== EVAL USER INPUT RULE ==========

type EvalUserInputRule struct{}

func (r *EvalUserInputRule) ID() string    { return "VS-SEC-008" }
func (r *EvalUserInputRule) Title() string { return "eval() với user input — RCE" }
func (r *EvalUserInputRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *EvalUserInputRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	pattern := regexp.MustCompile(`(?i)eval\s*\(\s*.*(req\.|request\.|body|query|params|input)`)

	for i, line := range lines {
		if pattern.MatchString(line) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}

			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "eval() với user input cho phép hacker chạy code tùy ý trên server (Remote Code Execution).",
				Fix: `// ❌ Nguy hiểm:
eval(req.body.code)

// ✅ Dùng sandbox hoặc parser:
// JSON.parse() cho data, hoặc thư viện vm2 (cũng cẩn thận)`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Critical,
				Category:    "security",
				Tags:        []string{"eval", "rce", "injection"},
			})
		}
	}

	return findings
}

// ========== PATH TRAVERSAL RULE ==========

type PathTraversalRule struct{}

func (r *PathTraversalRule) ID() string    { return "VS-SEC-009" }
func (r *PathTraversalRule) Title() string { return "Path Traversal — đọc file tùy ý" }
func (r *PathTraversalRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go", "php"}
}

func (r *PathTraversalRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	pattern := regexp.MustCompile(`(?i)(readFile|readFileSync|open|createReadStream|sendFile)\s*\(\s*.*(req\.|request\.|body|query|params)`)

	for i, line := range lines {
		if pattern.MatchString(line) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
				continue
			}

			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Đọc file với user input cho phép attacker đọc /etc/passwd, source code, hoặc file nhạy cảm.",
				Fix: `// ❌ Nguy hiểm:
app.get('/file', (req, res) => {
  res.sendFile(req.query.path)
})

// ✅ Validate path:
const safePath = path.join(__dirname, 'public', path.basename(req.query.path))
res.sendFile(safePath)`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"path-traversal", "file", "lfd"},
			})
		}
	}

	return findings
}

// ========== ENV IN GITIGNORE RULE ==========

type EnvInGitignoreRule struct{}

func (r *EnvInGitignoreRule) ID() string    { return "VS-SEC-010" }
func (r *EnvInGitignoreRule) Title() string { return ".env file không trong .gitignore" }
func (r *EnvInGitignoreRule) Languages() []string { return []string{"*"} }

func (r *EnvInGitignoreRule) Check(file *ParsedFile) []Finding {
	var findings []Finding

	// Chỉ chạy khi file là .gitignore
	if !strings.HasSuffix(file.Path, ".gitignore") {
		return findings
	}

	content := string(file.Content)

	// Check if .env is ignored
	if !strings.Contains(content, ".env") {
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "File .env chứa secrets nhưng không có trong .gitignore. Secrets sẽ bị commit lên git.",
			Fix:         "Thêm `.env` vào file .gitignore",
			File:        file.Path,
			Line:        1,
			Col:         1,
			Snippet:     "(thiếu .env trong .gitignore)",
			Severity:    High,
			Category:    "security",
			Tags:        []string{"git", "env", "secrets"},
		})
	}

	return findings
}
