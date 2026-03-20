package rules

import (
	"regexp"
	"strings"

	"github.com/odvcencio/gotreesitter"
)

// ========== ENV LOAD ORDER RULE ==========
type EnvLoadOrderRule struct{}

func (r *EnvLoadOrderRule) ID() string    { return "VS-VIBE-001" }
func (r *EnvLoadOrderRule) Title() string { return "ENV không được load trước khi dùng" }
func (r *EnvLoadOrderRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *EnvLoadOrderRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		var firstEnvUsage uint32
		var firstDotenvConfig uint32
		foundEnv := false
		foundDotenv := false
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeSize := node.EndByte() - node.StartByte()
			if nodeSize > 500 {
				return // Skip large parent nodes to avoid false matches
			}
			text := getNodeText(node, file.Content)
			_ = lang
			if !foundEnv && strings.Contains(text, "process.env.") {
				firstEnvUsage = node.StartByte()
				foundEnv = true
			}
			if !foundDotenv && (strings.Contains(text, "dotenv.config") || (strings.Contains(text, "require") && strings.Contains(text, "dotenv"))) {
				firstDotenvConfig = node.StartByte()
				foundDotenv = true
			}
		})
		if foundEnv && foundDotenv && firstEnvUsage < firstDotenvConfig {
			line := getLineNumber(file.Content, firstEnvUsage)
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "process.env được dùng trước khi dotenv.config() được gọi - env variables sẽ undefined.",
				Fix: `// ❌ Sai thứ tự:
const dbUrl = process.env.DATABASE_URL // undefined!
require('dotenv').config()

// ✅ Đúng thứ tự:
require('dotenv').config()
const dbUrl = process.env.DATABASE_URL // OK`,
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"env", "dotenv", "order"},
			})
		}
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	// Kiểm tra có dùng dotenv không
	hasDotenvConfig := regexp.MustCompile(`(?i)dotenv\.config\(\)`)
	if !hasDotenvConfig.MatchString(content) {
		return findings
	}

	envUsageLine := -1
	dotenvLine := -1

	for i, line := range lines {
		if regexp.MustCompile(`(?i)process\.env\.`).MatchString(line) && envUsageLine == -1 {
			envUsageLine = i
		}
		if hasDotenvConfig.MatchString(line) && dotenvLine == -1 {
			dotenvLine = i
		}
	}

	if envUsageLine != -1 && dotenvLine != -1 && envUsageLine < dotenvLine {
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "process.env được dùng trước khi dotenv.config() được gọi - env variables sẽ undefined.",
			Fix: `// ❌ Sai thứ tự:
const dbUrl = process.env.DATABASE_URL // undefined!
require('dotenv').config()

// ✅ Đúng thứ tự:
require('dotenv').config()
const dbUrl = process.env.DATABASE_URL // OK`,
			File:        file.Path,
			Line:        envUsageLine + 1,
			Col:         1,
			Snippet:     strings.TrimSpace(lines[envUsageLine]),
			Severity:    High,
			Category:    "security",
			Tags:        []string{"env", "dotenv", "order"},
		})
	}
	return findings
}

// ========== DUPLICATE MIDDLEWARE RULE ==========
type DuplicateMiddlewareRule struct{}

func (r *DuplicateMiddlewareRule) ID() string    { return "VS-VIBE-002" }
func (r *DuplicateMiddlewareRule) Title() string { return "Middleware đăng ký nhiều lần" }
func (r *DuplicateMiddlewareRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *DuplicateMiddlewareRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		middlewareCounts := make(map[string][]int)
		middlewares := []string{"cors", "helmet", "morgan", "bodyParser", "json", "urlencoded", "cookieParser"}
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "call_expression" && node.Type(lang) != "call" {
				return
			}
			text := getNodeText(node, file.Content)
			if !strings.Contains(text, "app.use") {
				return
			}
			for _, mw := range middlewares {
				if strings.Contains(text, mw) {
					line := getLineNumber(file.Content, node.StartByte())
					middlewareCounts[mw] = append(middlewareCounts[mw], line)
				}
			}
		})
		for mw, lineNums := range middlewareCounts {
			if len(lineNums) > 1 {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: mw + "() được đăng ký nhiều lần - có thể gây conflict hoặc overhead.",
					Fix: `// ❌ Duplicate:
app.use(cors())
// ... other middlewares
app.use(cors()) // Không cần thiết

// ✅ Chỉ một lần:
app.use(cors())`,
					File:        file.Path,
					Line:        lineNums[1],
					Col:         1,
					Snippet:     "app.use(" + mw + "())",
					Severity:    Low,
					Category:    "quality",
					Tags:        []string{"middleware", "duplicate", "express"},
				})
			}
		}
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	middlewareCounts := make(map[string][]int)

	middlewares := []string{"cors", "helmet", "morgan", "bodyParser", "json", "urlencoded", "cookieParser"}

	for i, line := range lines {
		for _, mw := range middlewares {
			pattern := regexp.MustCompile(`(?i)app\.use\s*\(\s*` + mw)
			if pattern.MatchString(line) {
				middlewareCounts[mw] = append(middlewareCounts[mw], i+1)
			}
		}
	}

	for mw, lineNums := range middlewareCounts {
		if len(lineNums) > 1 {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: mw + "() được đăng ký " + string(rune('0'+len(lineNums))) + " lần - có thể gây conflict hoặc overhead.",
				Fix: `// ❌ Duplicate:
app.use(cors())
// ... other middlewares
app.use(cors()) // Không cần thiết

// ✅ Chỉ một lần:
app.use(cors())`,
				File:        file.Path,
				Line:        lineNums[1],
				Col:         1,
				Snippet:     "app.use(" + mw + "())",
				Severity:    Low,
				Category:    "quality",
				Tags:        []string{"middleware", "duplicate", "express"},
			})
		}
	}
	return findings
}

// ========== SESSION SECRET RULE ==========
type SessionSecretRule struct{}

func (r *SessionSecretRule) ID() string    { return "VS-VIBE-003" }
func (r *SessionSecretRule) Title() string { return "Session không có secret mạnh" }
func (r *SessionSecretRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *SessionSecretRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)

	hasSession := regexp.MustCompile(`(?i)(express-session|session\s*\()`)
	if !hasSession.MatchString(content) {
		return findings
	}
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "call_expression" && node.Type(lang) != "call" {
				return
			}
			text := getNodeText(node, file.Content)
			if !strings.Contains(text, "session") {
				return
			}
			if !strings.Contains(text, "secret") {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "express-session không có secret - session dễ bị giả mạo.",
					Fix: `// ❌ Không secret:
app.use(session({
  resave: false,
  saveUninitialized: false
}))

// ✅ Có secret từ env:
app.use(session({
  secret: process.env.SESSION_SECRET,
  resave: false,
  saveUninitialized: false
}))`,
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    High,
					Category:    "security",
					Tags:        []string{"session", "secret", "express"},
				})
				return
			}
			weakPatterns := []string{"'secret'", "'password'", "'123456'", "'keyboard cat'"}
			for _, wp := range weakPatterns {
				if strings.Contains(text, wp) {
					line := getLineNumber(file.Content, node.StartByte())
					findings = append(findings, Finding{
						RuleID:      r.ID(),
						Title:       r.Title(),
						Description: "Session secret quá yếu - dễ bị brute force.",
						Fix: `// ❌ Secret yếu:
secret: 'keyboard cat'

// ✅ Secret từ env:
secret: process.env.SESSION_SECRET`,
						File:        file.Path,
						Line:        line,
						Col:         1,
						Snippet:     getSnippet(string(file.Content), line),
						Severity:    High,
						Category:    "security",
						Tags:        []string{"session", "weak-secret"},
					})
					break
				}
			}
		})
		return findings
	}
	lines := strings.Split(content, "\n")

	weakSecrets := []string{
		`secret:\s*['"]secret['"]`,
		`secret:\s*['"]password['"]`,
		`secret:\s*['"]123456['"]`,
		`secret:\s*['"]keyboard cat['"]`,
		`secret:\s*process\.env\.SESSION_SECRET`, // OK nếu dùng env
	}

	for i, line := range lines {
		if hasSession.MatchString(line) {
			// Kiểm tra có secret trong options không
			if !regexp.MustCompile(`(?i)secret\s*:`).MatchString(content) {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "express-session không có secret - session dễ bị giả mạo.",
					Fix: `// ❌ Không secret:
app.use(session({
  resave: false,
  saveUninitialized: false
}))

// ✅ Có secret từ env:
app.use(session({
  secret: process.env.SESSION_SECRET,
  resave: false,
  saveUninitialized: false
}))`,
					File:        file.Path,
					Line:        i + 1,
					Col:         1,
					Snippet:     strings.TrimSpace(line),
					Severity:    High,
					Category:    "security",
					Tags:        []string{"session", "secret", "express"},
				})
			}

			// Kiểm tra weak secret
			for _, weak := range weakSecrets[:4] {
				if regexp.MustCompile(`(?i)`+weak).MatchString(line) {
					findings = append(findings, Finding{
						RuleID:      r.ID(),
						Title:       r.Title(),
						Description: "Session secret quá yếu - dễ bị brute force.",
						Fix: `// ❌ Secret yếu:
secret: 'keyboard cat'

// ✅ Secret từ env:
secret: process.env.SESSION_SECRET`,
						File:        file.Path,
						Line:        i + 1,
						Col:         1,
						Snippet:     strings.TrimSpace(line),
						Severity:    High,
						Category:    "security",
						Tags:        []string{"session", "weak-secret"},
					})
				}
			}
		}
	}
	return findings
}

// ========== MONGOOSE VALIDATE RULE ==========
type MongooseValidateRule struct{}

func (r *MongooseValidateRule) ID() string    { return "VS-VIBE-004" }
func (r *MongooseValidateRule) Title() string { return "Mongoose schema thiếu validation" }
func (r *MongooseValidateRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *MongooseValidateRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)

	hasMongoose := regexp.MustCompile(`(?i)(mongoose|new\s+Schema)`)
	if !hasMongoose.MatchString(content) {
		return findings
	}
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "new_expression" {
				return
			}
			text := getNodeText(node, file.Content)
			if !strings.Contains(text, "Schema") {
				return
			}
			hasValidation := strings.Contains(text, "required") ||
				strings.Contains(text, "validate") ||
				strings.Contains(text, "minlength") ||
				strings.Contains(text, "maxlength") ||
				strings.Contains(text, "match")
			if !hasValidation {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Mongoose schema không có validation - dữ liệu không được kiểm tra trước khi lưu.",
					Fix: `// ❌ Thiếu validation:
const userSchema = new Schema({
  email: String,
  age: Number
})

// ✅ Có validation:
const userSchema = new Schema({
  email: {
    type: String,
    required: true,
    validate: [validator.isEmail, 'Invalid email']
  },
  age: {
    type: Number,
    min: 0,
    max: 150
  }
})`,
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    Medium,
					Category:    "quality",
					Tags:        []string{"mongoose", "validation", "schema"},
				})
			}
		})
		return findings
	}
	lines := strings.Split(content, "\n")
	inSchema := false
	braceCount := 0
	schemaHasValidation := false
	schemaStart := 0

	for i, line := range lines {
		// Detect schema start
		if regexp.MustCompile(`(?i)new\s+Schema\s*\(`).MatchString(line) {
			inSchema = true
			schemaStart = i
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")
			continue
		}

		if inSchema {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")

			// Check for validation
			if regexp.MustCompile(`(?i)(required\s*:\s*true|validate|minlength|maxlength|match)`).MatchString(line) {
				schemaHasValidation = true
			}

			// Schema ends
			if braceCount <= 0 {
				if !schemaHasValidation {
					findings = append(findings, Finding{
						RuleID:      r.ID(),
						Title:       r.Title(),
						Description: "Mongoose schema không có validation - dữ liệu không được kiểm tra trước khi lưu.",
						Fix: `// ❌ Thiếu validation:
const userSchema = new Schema({
  email: String,
  age: Number
})

// ✅ Có validation:
const userSchema = new Schema({
  email: {
    type: String,
    required: true,
    validate: [validator.isEmail, 'Invalid email']
  },
  age: {
    type: Number,
    min: 0,
    max: 150
  }
})`,
						File:        file.Path,
						Line:        schemaStart + 1,
						Col:         1,
						Snippet:     "Schema without validation",
						Severity:    Medium,
						Category:    "quality",
						Tags:        []string{"mongoose", "validation", "schema"},
					})
				}
				inSchema = false
				schemaHasValidation = false
			}
		}
	}
	return findings
}

// ========== RESPONSE STATUS CODE RULE ==========
type ResponseStatusCodeRule struct{}

func (r *ResponseStatusCodeRule) ID() string    { return "VS-VIBE-005" }
func (r *ResponseStatusCodeRule) Title() string { return "Response không có status code" }
func (r *ResponseStatusCodeRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *ResponseStatusCodeRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		lines := strings.Split(string(file.Content), "\n")
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "call_expression" && node.Type(lang) != "call" {
				return
			}
			funcNode := node.ChildByFieldName("function", lang)
			if funcNode == nil {
				return
			}
			funcText := getNodeText(funcNode, file.Content)
			if funcText != "res.json" {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			hasStatus := false
			for j := line - 2; j < line && j >= 0 && j < len(lines); j++ {
				if strings.Contains(lines[j], ".status(") {
					hasStatus = true
					break
				}
			}
			if strings.Contains(getNodeText(node, file.Content), ".status(") {
				hasStatus = true
			}
			if !hasStatus {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "res.json() không có status code - client không biết success hay error.",
					Fix: `// ❌ Thiếu status:
res.json({ data: users })

// ✅ Có status:
res.status(200).json({ data: users })

// Hoặc error:
res.status(400).json({ error: 'Invalid input' })`,
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    Low,
					Category:    "quality",
					Tags:        []string{"http", "status", "api"},
				})
			}
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	pattern := regexp.MustCompile(`(?i)res\.json\s*\(`)

	for i, line := range lines {
		if pattern.MatchString(line) {
			// Kiểm tra có .status() trước đó không
			foundStatus := false
			for j := i - 1; j >= max(0, i-3); j-- {
				if regexp.MustCompile(`(?i)\.status\s*\(`).MatchString(lines[j]) {
					foundStatus = true
					break
				}
			}

			if !foundStatus {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "res.json() không có status code - client không biết success hay error.",
					Fix: `// ❌ Thiếu status:
res.json({ data: users })

// ✅ Có status:
res.status(200).json({ data: users })

// Hoặc error:
res.status(400).json({ error: 'Invalid input' })`,
					File:        file.Path,
					Line:        i + 1,
					Col:         1,
					Snippet:     strings.TrimSpace(line),
					Severity:    Low,
					Category:    "quality",
					Tags:        []string{"http", "status", "api"},
				})
			}
		}
	}
	return findings
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ========== SELECT STAR RULE ==========
type SelectStarRule struct{}

func (r *SelectStarRule) ID() string    { return "VS-VIBE-006" }
func (r *SelectStarRule) Title() string { return "SELECT * trong raw SQL" }
func (r *SelectStarRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go", "php"}
}

func (r *SelectStarRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "string" && node.Type(lang) != "template_string" && node.Type(lang) != "string_fragment" {
				return
			}
			text := getNodeText(node, file.Content)
			if !regexp.MustCompile(`(?i)SELECT\s+\*`).MatchString(text) {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "SELECT * lấy tất cả columns - tốn băng thông và memory, dễ break khi schema thay đổi.",
				Fix: `// ❌ Không tốt:
SELECT * FROM users WHERE id = ?

// ✅ Chỉ lấy cần thiết:
SELECT id, name, email FROM users WHERE id = ?`,
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    Low,
				Category:    "performance",
				Tags:        []string{"sql", "select", "performance"},
			})
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	pattern := regexp.MustCompile(`(?i)(SELECT\s+\*|select\s*\*)`)

	for i, line := range lines {
		if pattern.MatchString(line) {
			// Bỏ qua nếu trong comment
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") {
				continue
			}

			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "SELECT * lấy tất cả columns - tốn băng thông và memory, dễ break khi schema thay đổi.",
				Fix: `// ❌ Không tốt:
SELECT * FROM users WHERE id = ?

// ✅ Chỉ lấy cần thiết:
SELECT id, name, email FROM users WHERE id = ?`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Low,
				Category:    "performance",
				Tags:        []string{"sql", "select", "performance"},
			})
		}
	}
	return findings
}

// ========== API KEY FRONTEND RULE ==========
type APIKeyFrontendRule struct{}

func (r *APIKeyFrontendRule) ID() string    { return "VS-VIBE-007" }
func (r *APIKeyFrontendRule) Title() string { return "API key để ở frontend (VITE_ prefix)" }
func (r *APIKeyFrontendRule) Languages() []string {
	return []string{"javascript", "typescript", "javascriptreact", "typescriptreact"}
}

func (r *APIKeyFrontendRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "member_expression" {
				return
			}
			text := getNodeText(node, file.Content)
			matched := false
			if strings.Contains(text, "import.meta.env.VITE_") {
				for _, s := range []string{"API_KEY", "SECRET", "PASSWORD", "PRIVATE"} {
					if strings.Contains(strings.ToUpper(text), s) {
						matched = true
						break
					}
				}
			}
			if strings.Contains(text, "process.env.REACT_APP_") {
				for _, s := range []string{"API_KEY", "SECRET", "PASSWORD", "PRIVATE"} {
					if strings.Contains(strings.ToUpper(text), s) {
						matched = true
						break
					}
				}
			}
			if !matched {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "API key được expose ở frontend - user có thể lấy key và dùng trực tiếp.",
				Fix: `// ❌ Frontend có secret:
const apiKey = import.meta.env.VITE_API_KEY

// ✅ Chỉ dùng ở backend:
// Trong server.js:
const apiKey = process.env.API_KEY // Không có VITE_ prefix`,
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    Critical,
				Category:    "security",
				Tags:        []string{"api-key", "frontend", "expose"},
			})
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)import\.meta\.env\.VITE_(API_KEY|SECRET|PASSWORD|PRIVATE)`),
		regexp.MustCompile(`(?i)process\.env\.REACT_APP_(API_KEY|SECRET|PASSWORD|PRIVATE)`),
	}

	for i, line := range lines {
		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "API key được expose ở frontend - user có thể lấy key và dùng trực tiếp.",
					Fix: `// ❌ Frontend có secret:
const apiKey = import.meta.env.VITE_API_KEY

// ✅ Chỉ dùng ở backend:
// Trong server.js:
const apiKey = process.env.API_KEY // Không có VITE_ prefix`,
					File:        file.Path,
					Line:        i + 1,
					Col:         1,
					Snippet:     strings.TrimSpace(line),
					Severity:    Critical,
					Category:    "security",
					Tags:        []string{"api-key", "frontend", "expose"},
				})
				break
			}
		}
	}
	return findings
}

// ========== HARDCODE PORT RULE ==========
type HardcodePortRule struct{}

func (r *HardcodePortRule) ID() string    { return "VS-VIBE-008" }
func (r *HardcodePortRule) Title() string { return "app.listen() không bind đúng port" }
func (r *HardcodePortRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *HardcodePortRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	if strings.Contains(content, "process.env.PORT") {
		return findings
	}
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "call_expression" && node.Type(lang) != "call" {
				return
			}
			text := getNodeText(node, file.Content)
			if !strings.Contains(text, "app.listen") {
				return
			}
			if regexp.MustCompile(`app\.listen\s*\(\s*\d{2,5}\s*\)`).MatchString(text) {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Port được hardcode - không flexible cho deployment.",
					Fix: `// ❌ Hardcoded:
app.listen(3000)

// ✅ Dùng env:
app.listen(process.env.PORT || 3000)`,
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    Low,
					Category:    "quality",
					Tags:        []string{"port", "deployment", "config"},
				})
			}
		})
		return findings
	}
	lines := strings.Split(content, "\n")
	hardcodedPort := regexp.MustCompile(`(?i)app\.listen\s*\(\s*\d{2,5}\s*\)`)

	for i, line := range lines {
		if hardcodedPort.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Port được hardcode - không flexible cho deployment.",
				Fix: `// ❌ Hardcoded:
app.listen(3000)

// ✅ Dùng env:
app.listen(process.env.PORT || 3000)`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Low,
				Category:    "quality",
				Tags:        []string{"port", "deployment", "config"},
			})
		}
	}
	return findings
}
