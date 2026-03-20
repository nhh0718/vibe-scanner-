package rules

import (
	"regexp"
	"strings"

	"github.com/odvcencio/gotreesitter"
)

// ========== RATE LIMITING RULE ==========
type RateLimitingRule struct{}

func (r *RateLimitingRule) ID() string    { return "VS-SEC-012" }
func (r *RateLimitingRule) Title() string { return "Thiếu rate limiting trên /login, /register" }
func (r *RateLimitingRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *RateLimitingRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)

	hasAuthRoute := regexp.MustCompile(`(?i)(['"]/?login['"]|['"]/?register['"]|['"]/?signin['"]|['"]/?signup['"])`)
	if !hasAuthRoute.MatchString(content) {
		return findings
	}
	hasRateLimit := regexp.MustCompile(`(?i)(rateLimit|rate-limit|throttle|express-rate-limit)`)
	if hasRateLimit.MatchString(content) {
		return findings
	}
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "string" && node.Type(lang) != "string_fragment" && node.Type(lang) != "template_string" {
				return
			}
			text := getNodeText(node, file.Content)
			if !hasAuthRoute.MatchString(text) {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Route authentication không có rate limiting - dễ bị brute force attack.",
				Fix: `// ❌ Thiếu protection:
app.post('/login', (req, res) => { ... })

// ✅ Thêm rate limiting:
import rateLimit from 'express-rate-limit'

const loginLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 5,
  message: 'Too many login attempts'
})

app.post('/login', loginLimiter, (req, res) => { ... })`,
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"rate-limit", "brute-force", "auth"},
			})
		})
		return findings
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if hasAuthRoute.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Route authentication không có rate limiting - dễ bị brute force attack.",
				Fix: `// ❌ Thiếu protection:
app.post('/login', (req, res) => { ... })

// ✅ Thêm rate limiting:
import rateLimit from 'express-rate-limit'

const loginLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 phút
  max: 5, // 5 lần thử
  message: 'Too many login attempts'
})

app.post('/login', loginLimiter, (req, res) => { ... })`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"rate-limit", "brute-force", "auth"},
			})
			break
		}
	}
	return findings
}

// ========== CSRF RULE ==========
type CSRFRule struct{}

func (r *CSRFRule) ID() string    { return "VS-SEC-015" }
func (r *CSRFRule) Title() string { return "CSRF — form POST không có token protection" }
func (r *CSRFRule) Languages() []string {
	return []string{"javascript", "typescript", "html"}
}

func (r *CSRFRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)

	hasFormPost := regexp.MustCompile(`(?i)<form[^>]*method=["']?post["']?`)
	if !hasFormPost.MatchString(content) {
		return findings
	}
	hasCSRF := regexp.MustCompile(`(?i)(csrf|_token|xsrf|csrfToken|csrf-token)`)
	if hasCSRF.MatchString(content) {
		return findings
	}
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeType := node.Type(lang)
			if nodeType != "element" && nodeType != "start_tag" && nodeType != "self_closing_tag" &&
				nodeType != "template_string" && nodeType != "string" {
				return
			}
			text := getNodeText(node, file.Content)
			if hasFormPost.MatchString(text) {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Form POST không có CSRF token - dễ bị cross-site request forgery attack.",
					Fix: `// ❌ Thiếu CSRF protection:\n<form method="POST" action="/transfer">\n\n// ✅ Thêm CSRF token:\nimport csrf from 'csurf'\napp.post('/transfer', csrfProtection, (req, res) => { ... })`,
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    High,
					Category:    "security",
					Tags:        []string{"csrf", "form", "token"},
				})
			}
		})
		if len(findings) > 0 {
			return findings
		}
	}
	findings = append(findings, Finding{
		RuleID:      r.ID(),
		Title:       r.Title(),
		Description: "Form POST không có CSRF token - dễ bị cross-site request forgery attack.",
		Fix: `// ❌ Thiếu CSRF protection:
<form method="POST" action="/transfer">
  <input name="to" value="attacker">
  <input name="amount" value="1000">
</form>

// ✅ Thêm CSRF token:
import csrf from 'csurf'
const csrfProtection = csrf({ cookie: true })

app.post('/transfer', csrfProtection, (req, res) => { ... })`,
		File:        file.Path,
		Line:        1,
		Col:         1,
		Snippet:     "Form POST without CSRF protection",
		Severity:    High,
		Category:    "security",
		Tags:        []string{"csrf", "form", "token"},
	})
	return findings
}

// ========== AUTH MIDDLEWARE RULE ==========
type AuthMiddlewareRule struct{}

func (r *AuthMiddlewareRule) ID() string    { return "VS-SEC-016" }
func (r *AuthMiddlewareRule) Title() string { return "Missing auth middleware trên admin routes" }
func (r *AuthMiddlewareRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *AuthMiddlewareRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)

	adminPattern := regexp.MustCompile(`(?i)(['"]/?admin['"]|['"]/?dashboard['"]|['"]/?manage['"])`)
	hasAuthMiddleware := regexp.MustCompile(`(?i)(requireAuth|isAuthenticated|authMiddleware|verifyToken|protect)`)

	if hasAuthMiddleware.MatchString(content) {
		return findings
	}
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeType := node.Type(lang)
			if nodeType != "string" && nodeType != "string_fragment" && nodeType != "template_string" {
				return
			}
			text := getNodeText(node, file.Content)
			if !adminPattern.MatchString(text) {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Route admin không có authentication middleware - ai cũng có thể truy cập.",
				Fix: `// ❌ Không an toàn:\napp.get('/admin/users', (req, res) => { ... })\n\n// ✅ Thêm auth middleware:\napp.get('/admin/users', requireAuth, requireAdmin, (req, res) => { ... })`,
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"auth", "middleware", "admin"},
			})
		})
		return findings
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if adminPattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Route admin không có authentication middleware - ai cũng có thể truy cập.",
				Fix: `// ❌ Không an toàn:
app.get('/admin/users', (req, res) => { ... })

// ✅ Thêm auth middleware:
app.get('/admin/users', requireAuth, requireAdmin, (req, res) => { ... })`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"auth", "middleware", "admin"},
			})
		}
	}
	return findings
}

// ========== PICKLE DESERIALIZATION RULE ==========
type PickleDeserializationRule struct{}

func (r *PickleDeserializationRule) ID() string    { return "VS-SEC-020" }
func (r *PickleDeserializationRule) Title() string { return "Pickle deserialization với user data — RCE" }
func (r *PickleDeserializationRule) Languages() []string {
	return []string{"python"}
}

func (r *PickleDeserializationRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "call" && node.Type(lang) != "call_expression" {
				return
			}
			text := getNodeText(node, file.Content)
			textLower := strings.ToLower(text)
			if !strings.Contains(textLower, "pickle.load") {
				return
			}
			if !strings.Contains(textLower, "request") && !strings.Contains(textLower, "body") &&
				!strings.Contains(textLower, "input") && !strings.Contains(textLower, "data") {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "pickle.loads() với user input cho phép RCE - hacker có thể serialize malicious object.",
				Fix: `// ❌ Nguy hiểm:
data = pickle.loads(request.body)

// ✅ Dùng JSON hoặc validate schema:
import json
data = json.loads(request.body)`,
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    Critical,
				Category:    "security",
				Tags:        []string{"pickle", "deserialization", "rce", "python"},
			})
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	pattern := regexp.MustCompile(`(?i)pickle\.loads?\s*\(\s*.*(request|body|input|data)`)

	for i, line := range lines {
		if pattern.MatchString(line) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "pickle.loads() với user input cho phép RCE - hacker có thể serialize malicious object.",
				Fix: `// ❌ Nguy hiểm:
data = pickle.loads(request.body)

// ✅ Dùng JSON hoặc validate schema:
import json
data = json.loads(request.body)
# Hoặc dùng safer alternatives như marshmallow`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     trimmed,
				Severity:    Critical,
				Category:    "security",
				Tags:        []string{"pickle", "deserialization", "rce", "python"},
			})
		}
	}
	return findings
}

// ========== DEBUG MODE RULE ==========
type DebugModeRule struct{}

func (r *DebugModeRule) ID() string    { return "VS-SEC-022" }
func (r *DebugModeRule) Title() string { return "Debug mode bật trong production" }
func (r *DebugModeRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go"}
}

func (r *DebugModeRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		debugPattern := regexp.MustCompile(`(?i)(debug\s*:\s*true|DEBUG\s*=\s*True|showStack\s*:\s*true|FLASK_ENV\s*=\s*['"]development['"])`)
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeType := node.Type(lang)
			if nodeType != "pair" && nodeType != "property" && nodeType != "assignment_expression" && nodeType != "assignment" {
				return
			}
			text := getNodeText(node, file.Content)
			if debugPattern.MatchString(text) {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Debug mode đang bật trong production - tiết lộ thông tin nhạy cảm.",
					Fix: `// ❌ Không an toàn:
const config = { debug: true }

// ✅ Production:
const config = { debug: process.env.NODE_ENV !== 'production' }`,
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    High,
					Category:    "security",
					Tags:        []string{"debug", "production", "config"},
				})
			}
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	patterns := []struct {
		pattern *regexp.Regexp
		lang    string
	}{
		{regexp.MustCompile(`(?i)(debug\s*:\s*true|DEBUG\s*=\s*True)`), "general"},
		{regexp.MustCompile(`(?i)app\.use\s*\(\s*errorhandler\s*\(\s*\{\s*showStack\s*:\s*true\s*\}\s*\)\s*\)`), "express"},
		{regexp.MustCompile(`(?i)FLASK_ENV\s*=\s*['"]development['"]|DEBUG\s*=\s*True`), "python"},
	}

	for i, line := range lines {
		for _, p := range patterns {
			if p.pattern.MatchString(line) {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Debug mode đang bật trong production - tiết lộ thông tin nhạy cảm.",
					Fix: `// ❌ Không an toàn:
const config = { debug: true }

// ✅ Production:
const config = { debug: process.env.NODE_ENV !== 'production' }`,
					File:        file.Path,
					Line:        i + 1,
					Col:         1,
					Snippet:     strings.TrimSpace(line),
					Severity:    High,
					Category:    "security",
					Tags:        []string{"debug", "production", "config"},
				})
			}
		}
	}
	return findings
}

// ========== CONSOLE LOG SENSITIVE RULE ==========
type ConsoleLogSensitiveRule struct{}

func (r *ConsoleLogSensitiveRule) ID() string    { return "VS-SEC-023" }
func (r *ConsoleLogSensitiveRule) Title() string { return "console.log với sensitive data" }
func (r *ConsoleLogSensitiveRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *ConsoleLogSensitiveRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	sensitiveWords := []string{"password", "token", "secret", "key", "credential", "ssn", "credit"}
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "call_expression" && node.Type(lang) != "call" {
				return
			}
			text := getNodeText(node, file.Content)
			textLower := strings.ToLower(text)
			if !strings.Contains(textLower, "console.log") && !strings.Contains(textLower, "console.debug") && !strings.Contains(textLower, "console.warn") {
				return
			}
			hasSensitive := false
			for _, w := range sensitiveWords {
				if strings.Contains(textLower, w) {
					hasSensitive = true
					break
				}
			}
			if !hasSensitive {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Logging sensitive data có thể lộ thông tin trong logs.",
				Fix: `// ❌ Không an toàn:
console.log("User password:", user.password)

// ✅ Che sensitive data:
console.log("User:", { ...user, password: '[REDACTED]' })`,
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    Medium,
				Category:    "security",
				Tags:        []string{"logging", "sensitive", "leak"},
			})
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	pattern := regexp.MustCompile(`(?i)console\.(log|debug|warn)\s*\([^)]*(password|token|secret|key|credential|ssn|credit)`)

	for i, line := range lines {
		if pattern.MatchString(line) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Logging sensitive data có thể lộ thông tin trong logs.",
				Fix: `// ❌ Không an toàn:
console.log("User password:", user.password)

// ✅ Che sensitive data:
console.log("User:", { ...user, password: '[REDACTED]' })`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     trimmed,
				Severity:    Medium,
				Category:    "security",
				Tags:        []string{"logging", "sensitive", "leak"},
			})
		}
	}
	return findings
}

// ========== ENV CHECK RULE ==========
type EnvCheckRule struct{}

func (r *EnvCheckRule) ID() string    { return "VS-SEC-024" }
func (r *EnvCheckRule) Title() string { return "process.env check thiếu — biến môi trường undefined" }
func (r *EnvCheckRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *EnvCheckRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)

	pattern := regexp.MustCompile(`(?i)process\.env\.([A-Z_]+)`)
	matches := pattern.FindAllStringSubmatch(content, -1)

	checkedPattern := regexp.MustCompile(`(?i)(if\s*\(\s*process\.env|process\.env\.\w+\s*\?\?|process\.env\.\w+\s*\|\|)`)
	hasCheck := checkedPattern.MatchString(content)

	if len(matches) > 0 && !hasCheck {
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "Sử dụng process.env không có fallback hoặc validation - có thể undefined.",
			Fix: `// ❌ Thiếu check:
const dbUrl = process.env.DATABASE_URL

// ✅ Có fallback:
const dbUrl = process.env.DATABASE_URL || 'default-url'
// Hoặc throw nếu required:
if (!process.env.DATABASE_URL) {
  throw new Error('DATABASE_URL is required')
}`,
			File:        file.Path,
			Line:        1,
			Col:         1,
			Snippet:     "process.env usage without check",
			Severity:    Medium,
			Category:    "security",
			Tags:        []string{"env", "config", "validation"},
		})
	}
	return findings
}

// ========== INPUT VALIDATION RULE ==========
type InputValidationRule struct{}

func (r *InputValidationRule) ID() string    { return "VS-SEC-025" }
func (r *InputValidationRule) Title() string { return "Thiếu input validation trên request body" }
func (r *InputValidationRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go"}
}

func (r *InputValidationRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)

	hasRouteHandler := regexp.MustCompile(`(?i)(app\.(get|post|put|delete|patch)|router\.|@Controller)`)
	if !hasRouteHandler.MatchString(content) {
		return findings
	}
	hasValidation := regexp.MustCompile(`(?i)(joi|zod|yup|class-validator|express-validator|validate|schema)`)
	if hasValidation.MatchString(content) {
		return findings
	}
	hasBodyParser := regexp.MustCompile(`(?i)(req\.body|request\.body|@Body\(\))`)
	if !hasBodyParser.MatchString(content) {
		return findings
	}
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "member_expression" {
				return
			}
			text := getNodeText(node, file.Content)
			if text != "req.body" && text != "request.body" {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Request body được dùng trực tiếp không qua validation - dễ bị injection attacks.",
				Fix: `// ❌ Không an toàn:
app.post('/users', (req, res) => {
  const user = req.body
  db.users.create(user)
})

// ✅ Thêm validation:
import { z } from 'zod'

const userSchema = z.object({
  email: z.email(),
  name: z.string().min(1).max(100)
})

app.post('/users', (req, res) => {
  const user = userSchema.parse(req.body)
  db.users.create(user)
})`,
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"validation", "input", "schema"},
			})
		})
		return findings
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if hasBodyParser.MatchString(line) && !hasValidation.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Request body được dùng trực tiếp không qua validation - dễ bị injection attacks.",
				Fix: `// ❌ Không an toàn:
app.post('/users', (req, res) => {
  const user = req.body  // Không validation!
  db.users.create(user)
})

// ✅ Thêm validation:
import { z } from 'zod'

const userSchema = z.object({
  email: z.email(),
  name: z.string().min(1).max(100)
})

app.post('/users', (req, res) => {
  const user = userSchema.parse(req.body)
  db.users.create(user)
})`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"validation", "input", "schema"},
			})
			break
		}
	}
	return findings
}
