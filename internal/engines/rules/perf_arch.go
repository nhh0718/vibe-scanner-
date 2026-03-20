package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/odvcencio/gotreesitter"
)

// ========== N+1 QUERY RULE ==========
type NPlusOneQueryRule struct{}

func (r *NPlusOneQueryRule) ID() string    { return "VS-PERF-001" }
func (r *NPlusOneQueryRule) Title() string { return "Query trong vòng lặp — N+1 pattern" }
func (r *NPlusOneQueryRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go"}
}

func (r *NPlusOneQueryRule) Check(file *ParsedFile) []Finding {
	if file.Tree != nil {
		return r.checkAST(file)
	}
	return r.checkRegex(file)
}

func (r *NPlusOneQueryRule) checkAST(file *ParsedFile) []Finding {
	var findings []Finding
	lang := file.Tree.Language()
	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if !perfIsLoopNode(node.Type(lang)) {
			return
		}
		if perfHasDescendantCall(node, file, []string{"findone", "findbyid", "query(", ".exec("}) {
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "N+1 Query pattern detected",
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    Medium,
				Category:    "performance",
				Tags:        []string{"n-plus-one"},
			})
		}
	})
	return findings
}

func (r *NPlusOneQueryRule) checkRegex(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	// Tìm pattern: for/while với db call bên trong
	inLoop := false
	loopStart := 0
	braceCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect loop start
		if regexp.MustCompile(`(?i)^(for|while)\s*\(`).MatchString(trimmed) {
			inLoop = true
			loopStart = i
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")
			continue
		}

		if inLoop {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")

			// Check for db call
			if regexp.MustCompile(`(?i)(findOne|findById|query)\s*\(`).MatchString(line) {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "N+1 Query pattern detected",
					File:        file.Path,
					Line:        loopStart + 1,
					Severity:    Medium,
					Category:    "performance",
					Tags:        []string{"n-plus-one"},
				})
			}

			// Loop ends
			if braceCount <= 0 {
				inLoop = false
			}
		}
	}
	return findings
}

// ========== READ FILE SYNC IN ASYNC RULE ==========
type ReadFileSyncInAsyncRule struct{}

func (r *ReadFileSyncInAsyncRule) ID() string    { return "VS-PERF-002" }
func (r *ReadFileSyncInAsyncRule) Title() string { return "fs.readFileSync trong async context" }
func (r *ReadFileSyncInAsyncRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *ReadFileSyncInAsyncRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if !perfIsAsyncFunction(node, file) {
				return
			}
			if perfHasDescendantCall(node, file, []string{"readfilesync"}) {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Dùng readFileSync trong async context block event loop.",
					Fix: `// ❌ Block event loop:
async function handler() {
  const data = fs.readFileSync('file.txt')
}

// ✅ Không block:
async function handler() {
  const data = await fs.promises.readFile('file.txt')
}`,
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    Medium,
					Category:    "performance",
					Tags:        []string{"async", "blocking", "fs"},
				})
			}
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	// Kiểm tra có async function không
	hasAsync := regexp.MustCompile(`(?i)(async\s+function|async\s*\(|=>\s*\{)`).MatchString(content)
	if !hasAsync {
		return findings
	}

	pattern := regexp.MustCompile(`(?i)readFileSync`)

	for i, line := range lines {
		if pattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Dùng readFileSync trong async context block event loop.",
				Fix: `// ❌ Block event loop:
async function handler() {
  const data = fs.readFileSync('file.txt')
}

// ✅ Không block:
async function handler() {
  const data = await fs.promises.readFile('file.txt')
}`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Medium,
				Category:    "performance",
				Tags:        []string{"async", "blocking", "fs"},
			})
		}
	}
	return findings
}

// ========== LODASH FULL IMPORT RULE ==========
type LodashFullImportRule struct{}

func (r *LodashFullImportRule) ID() string    { return "VS-PERF-003" }
func (r *LodashFullImportRule) Title() string { return "Import cả lodash thay vì named import" }
func (r *LodashFullImportRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *LodashFullImportRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "import_statement" {
				return
			}
			text := getNodeText(node, file.Content)
			if strings.Contains(text, "lodash") && (strings.Contains(text, "* as") || strings.Contains(text, "import _")) {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Import toàn bộ lodash tăng bundle size đáng kể.",
					Fix: `// ❌ Import cả thư viện (~70KB):
import * as _ from 'lodash'

// ✅ Chỉ import cần thiết (~2KB):
import debounce from 'lodash/debounce'
// Hoặc:
import { debounce, throttle } from 'lodash-es'`,
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    Low,
					Category:    "performance",
					Tags:        []string{"bundle-size", "import", "lodash"},
				})
			}
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	pattern := regexp.MustCompile(`(?i)import\s+\*\s+as\s+_\s+from\s+['"]lodash['"]`)

	for i, line := range lines {
		if pattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Import toàn bộ lodash tăng bundle size đáng kể.",
				Fix: `// ❌ Import cả thư viện (~70KB):
import * as _ from 'lodash'

// ✅ Chỉ import cần thiết (~2KB):
import debounce from 'lodash/debounce'
// Hoặc:
import { debounce, throttle } from 'lodash-es'`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Low,
				Category:    "performance",
				Tags:        []string{"bundle-size", "import", "lodash"},
			})
		}
	}
	return findings
}

// ========== JSON PARSE IN LOOP RULE ==========
type JSONParseInLoopRule struct{}

func (r *JSONParseInLoopRule) ID() string    { return "VS-PERF-004" }
func (r *JSONParseInLoopRule) Title() string { return "JSON.parse/JSON.stringify trong vòng lặp" }
func (r *JSONParseInLoopRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *JSONParseInLoopRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if !perfIsLoopNode(node.Type(lang)) {
				return
			}
			if perfHasDescendantCall(node, file, []string{"json.parse", "json.stringify"}) {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "JSON.parse/stringify trong loop expensive - nên parse 1 lần bên ngoài.",
					Fix: `// ❌ Parse trong mỗi iteration:
for (const item of items) {
  const data = JSON.parse(item.json) // Slow!
}

// ✅ Parse trước:
const parsed = items.map(i => JSON.parse(i.json))
for (const data of parsed) {
  // Dùng data
}`,
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    Low,
					Category:    "performance",
					Tags:        []string{"json", "loop", "performance"},
				})
			}
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	// Tìm JSON.parse/stringify trong for/while
	inLoop := false
	loopStart := 0
	braceCount := 0
	jsonPattern := regexp.MustCompile(`(?i)JSON\.(parse|stringify)\s*\(`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect loop
		if regexp.MustCompile(`(?i)^(for|while)\s*\(`).MatchString(trimmed) {
			inLoop = true
			loopStart = i
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")
			continue
		}

		if inLoop {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")

			if jsonPattern.MatchString(line) {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "JSON.parse/stringify trong loop expensive - nên parse 1 lần bên ngoài.",
					Fix: `// ❌ Parse trong mỗi iteration:
for (const item of items) {
  const data = JSON.parse(item.json) // Slow!
}

// ✅ Parse trước:
const parsed = items.map(i => JSON.parse(i.json))
for (const data of parsed) {
  // Dùng data
}`,
					File:        file.Path,
					Line:        loopStart + 1,
					Col:         1,
					Snippet:     strings.TrimSpace(line),
					Severity:    Low,
					Category:    "performance",
					Tags:        []string{"json", "loop", "performance"},
				})
			}

			if braceCount <= 0 {
				inLoop = false
			}
		}
	}
	return findings
}

// ========== CIRCULAR IMPORT RULE ==========
type CircularImportRule struct{}

func (r *CircularImportRule) ID() string    { return "VS-ARCH-001" }
func (r *CircularImportRule) Title() string { return "Circular imports — A import B, B import A" }
func (r *CircularImportRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *CircularImportRule) Check(file *ParsedFile) []Finding {
	var findings []Finding

	// Detect circular imports cần analyze nhiều file
	// Đơn giản: detect import chính file đang xét
	content := string(file.Content)
	filename := file.Path

	// Extract basename without extension for self-import detection
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	// Check if file imports itself (indirect detection)
	selfImport := regexp.MustCompile(`(?i)from\s+['"].*` + regexp.QuoteMeta(base) + `['"]`)
	if selfImport.MatchString(content) {
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "Circular dependency detected - file import chính nó.",
			Fix:         "Refactor để loại bỏ circular dependency, dùng dependency injection hoặc event-driven.",
			File:        file.Path,
			Line:        1,
			Col:         1,
			Snippet:     "Circular import detected",
			Severity:    Medium,
			Category:    "architecture",
			Tags:        []string{"circular", "import", "dependency"},
		})
	}

	return findings
}

// ========== BUSINESS LOGIC IN ROUTE RULE ==========
type BusinessLogicInRouteRule struct{}

func (r *BusinessLogicInRouteRule) ID() string    { return "VS-ARCH-002" }
func (r *BusinessLogicInRouteRule) Title() string { return "Business logic trong route handler" }
func (r *BusinessLogicInRouteRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *BusinessLogicInRouteRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "call_expression" && node.Type(lang) != "call" {
				return
			}
			funcNode := node.ChildByFieldName("function", lang)
			if funcNode == nil {
				return
			}
			text := strings.ToLower(getNodeText(funcNode, file.Content))
			if !perfIsRouteCall(text) {
				return
			}
			args := node.ChildByFieldName("arguments", lang)
			if args == nil {
				return
			}
			for i := 0; i < args.NamedChildCount(); i++ {
				arg := args.NamedChild(i)
				if arg == nil {
					continue
				}
				argType := arg.Type(lang)
				if argType == "arrow_function" || argType == "function_expression" || argType == "function" {
					startLine := getLineNumber(file.Content, arg.StartByte())
					endLine := getLineNumber(file.Content, arg.EndByte())
					if endLine-startLine > 20 {
						findings = append(findings, Finding{
							RuleID:      r.ID(),
							Title:       r.Title(),
							Description: "Route handler có quá nhiều business logic (>20 dòng) - nên tách ra service layer.",
							Fix: `// ❌ Route handler quá dài:\napp.post('/order', async (req, res) => {\n  // 50 dòng logic ở đây...\n})\n\n// ✅ Tách service:\napp.post('/order', orderController.create)`,
							File:        file.Path,
							Line:        startLine,
							Col:         1,
							Snippet:     getSnippet(string(file.Content), startLine),
							Severity:    Medium,
							Category:    "architecture",
							Tags:        []string{"mvc", "separation", "route"},
						})
					}
				}
			}
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	// Tìm route handlers và đếm dòng trong callback
	inRouteHandler := false
	routeStart := 0
	braceCount := 0
	handlerLineCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect route handler start
		if regexp.MustCompile(`(?i)(app|router)\.(get|post|put|delete|patch)\s*\(`).MatchString(trimmed) {
			inRouteHandler = true
			routeStart = i
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")
			handlerLineCount = 1
			continue
		}

		if inRouteHandler {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			handlerLineCount++

			// Handler ends
			if braceCount <= 0 {
				if handlerLineCount > 20 {
					findings = append(findings, Finding{
						RuleID:      r.ID(),
						Title:       r.Title(),
						Description: "Route handler có quá nhiều business logic (>20 dòng) - nên tách ra service layer.",
						Fix: `// ❌ Route handler quá dài:
app.post('/order', async (req, res) => {
  // 50 dòng logic ở đây...
})

// ✅ Tách service:
app.post('/order', orderController.create)`,
						File:        file.Path,
						Line:        routeStart + 1,
						Col:         1,
						Snippet:     strings.TrimSpace(lines[routeStart]),
						Severity:    Medium,
						Category:    "architecture",
						Tags:        []string{"mvc", "separation", "route"},
					})
				}
				inRouteHandler = false
			}
		}
	}

	return findings
}

// ========== DIRECT DB CALL IN CONTROLLER RULE ==========
type DirectDBCallInControllerRule struct{}

func (r *DirectDBCallInControllerRule) ID() string    { return "VS-ARCH-003" }
func (r *DirectDBCallInControllerRule) Title() string { return "Direct DB call trong controller" }
func (r *DirectDBCallInControllerRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go"}
}

func (r *DirectDBCallInControllerRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	isController := strings.Contains(strings.ToLower(file.Path), "controller") ||
		strings.Contains(strings.ToLower(file.Path), "route") ||
		strings.Contains(strings.ToLower(file.Path), "handler")
	if !isController {
		return findings
	}
	content := string(file.Content)
	if strings.Contains(content, "repository") || strings.Contains(content, "Repository") {
		return findings
	}
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "call_expression" && node.Type(lang) != "call" {
				return
			}
			text := strings.ToLower(getNodeText(node, file.Content))
			if !strings.Contains(text, "db.") && !strings.Contains(text, ".find") && !strings.Contains(text, ".query") && !strings.Contains(text, ".execute") {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Controller gọi trực tiếp database - nên qua Repository/DAO pattern.",
				Fix: `// ❌ Trực tiếp:\napp.get('/users', async (req, res) => {\n  const users = await db.users.findAll()\n  res.json(users)\n})\n\n// ✅ Repository pattern:\napp.get('/users', async (req, res) => {\n  const users = await userRepository.findAll()\n  res.json(users)\n})`,
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    Medium,
				Category:    "architecture",
				Tags:        []string{"repository", "dao", "layer"},
			})
		})
		return findings
	}
	lines := strings.Split(content, "\n")
	dbCallPattern := regexp.MustCompile(`(?i)(db\.|\.find|\.query|\.execute|session\.run)`)

	for i, line := range lines {
		if dbCallPattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Controller gọi trực tiếp database - nên qua Repository/DAO pattern.",
				Fix: `// ❌ Trực tiếp:
app.get('/users', async (req, res) => {
  const users = await db.users.findAll()
  res.json(users)
})

// ✅ Repository pattern:
app.get('/users', async (req, res) => {
  const users = await userRepository.findAll()
  res.json(users)
})`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Medium,
				Category:    "architecture",
				Tags:        []string{"repository", "dao", "layer"},
			})
			break
		}
	}
	return findings
}

// ========== DYNAMIC REQUIRE RULE ==========
type DynamicRequireRule struct{}

func (r *DynamicRequireRule) ID() string    { return "VS-ARCH-004" }
func (r *DynamicRequireRule) Title() string { return "require() trong function body" }
func (r *DynamicRequireRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *DynamicRequireRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if !qualityIsFunctionLikeNode(node, lang) {
				return
			}
			if perfHasDescendantCall(node, file, []string{"require"}) {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Dynamic require() trong function gây khó khăn cho bundler và tree-shaking.",
					Fix: `// ❌ Dynamic require:\nfunction loadModule() {\n  const mod = require('./module')\n}\n\n// ✅ Static import ở top:\nimport { something } from './module'\nfunction loadModule() {\n  // Dùng something\n}`,
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    Low,
					Category:    "architecture",
					Tags:        []string{"require", "dynamic", "import"},
				})
			}
		})
		return findings
	}
	content := string(file.Content)
	lines := strings.Split(content, "\n")

	// Tìm require() trong function body (không phải top-level)
	inFunction := false
	funcBraceCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect function start
		if regexp.MustCompile(`(?i)(function\s*\(|=>\s*\{|async\s*\()`).MatchString(trimmed) ||
			regexp.MustCompile(`(?i)^\s*\{`).MatchString(trimmed) {
			if !inFunction {
				inFunction = true
				funcBraceCount = strings.Count(line, "{") - strings.Count(line, "}")
			}
			continue
		}

		if inFunction {
			funcBraceCount += strings.Count(line, "{") - strings.Count(line, "}")

			if regexp.MustCompile(`(?i)require\s*\(`).MatchString(line) {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Dynamic require() trong function gây khó khăn cho bundler và tree-shaking.",
					Fix: `// ❌ Dynamic require:
function loadModule() {
  const mod = require('./module')
}

// ✅ Static import ở top:
import { something } from './module'
function loadModule() {
  // Dùng something
}`,
					File:        file.Path,
					Line:        i + 1,
					Col:         1,
					Snippet:     trimmed,
					Severity:    Low,
					Category:    "architecture",
					Tags:        []string{"require", "dynamic", "import"},
				})
			}

			if funcBraceCount <= 0 {
				inFunction = false
			}
		}
	}
	return findings
}

// ========== PERF/ARCH HELPER FUNCTIONS ==========

func perfIsLoopNode(nodeType string) bool {
	return nodeType == "for_statement" ||
		nodeType == "for_in_statement" ||
		nodeType == "for_of_statement" ||
		nodeType == "while_statement" ||
		nodeType == "do_statement"
}

func perfHasDescendantCall(node *gotreesitter.Node, file *ParsedFile, patterns []string) bool {
	if node == nil {
		return false
	}
	lang := file.Tree.Language()
	nodeType := node.Type(lang)
	if nodeType == "call_expression" || nodeType == "call" {
		text := strings.ToLower(getNodeText(node, file.Content))
		for _, p := range patterns {
			if strings.Contains(text, p) {
				return true
			}
		}
	}
	for i := 0; i < node.ChildCount(); i++ {
		if perfHasDescendantCall(node.Child(i), file, patterns) {
			return true
		}
	}
	return false
}

func perfIsAsyncFunction(node *gotreesitter.Node, file *ParsedFile) bool {
	if node == nil {
		return false
	}
	lang := file.Tree.Language()
	nodeType := node.Type(lang)
	if nodeType != "function_declaration" && nodeType != "arrow_function" &&
		nodeType != "function_expression" && nodeType != "method_definition" {
		return false
	}
	text := strings.TrimSpace(getNodeText(node, file.Content))
	return strings.HasPrefix(text, "async ")
}

func perfIsRouteCall(text string) bool {
	routePatterns := []string{
		"app.get", "app.post", "app.put", "app.delete", "app.patch",
		"router.get", "router.post", "router.put", "router.delete", "router.patch",
	}
	for _, p := range routePatterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}
