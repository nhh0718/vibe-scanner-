// Package rules chứa tất cả rules AST-based
package rules

import (
	"strings"

	"github.com/odvcencio/gotreesitter"
)

// ASTRule interface cho rules sử dụng AST
type ASTRule interface {
	ID() string
	Title() string
	Description() string
	Severity() Severity
	Category() Category
	Languages() []string
	CheckAST(file *ParsedFile) []Finding
}

func walkTree(node *gotreesitter.Node, visit func(*gotreesitter.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for i := 0; i < node.ChildCount(); i++ {
		walkTree(node.Child(i), visit)
	}
}

func getNodeText(node *gotreesitter.Node, content []byte) string {
	if node == nil {
		return ""
	}
	return node.Text(content)
}

func getLineNumber(content []byte, byteOffset uint32) int {
	line := 1
	for i := uint32(0); i < byteOffset && i < uint32(len(content)); i++ {
		if content[i] == '\n' {
			line++
		}
	}
	return line
}

// ========== SQL INJECTION AST RULE ==========

type SQLInjectionASTRule struct{}

func (r *SQLInjectionASTRule) ID() string       { return "AST-SQL-001" }
func (r *SQLInjectionASTRule) Title() string    { return "SQL Injection - String concatenation in query" }
func (r *SQLInjectionASTRule) Description() string {
	return "Phát hiện SQL query được ghép chuỗi với user input - dễ bị SQL Injection"
}
func (r *SQLInjectionASTRule) Severity() Severity { return Critical }
func (r *SQLInjectionASTRule) Category() Category { return Security }
func (r *SQLInjectionASTRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go", "php"}
}

func (r *SQLInjectionASTRule) Check(file *ParsedFile) []Finding {
	return r.CheckAST(file)
}

func (r *SQLInjectionASTRule) CheckAST(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	lang := file.Tree.Language()
	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if node.Type(lang) != "call_expression" && node.Type(lang) != "call" {
			return
		}

		funcNode := node.ChildByFieldName("function", lang)
		if funcNode == nil {
			return
		}

		funcName := strings.ToLower(getNodeText(funcNode, file.Content))
		if !isSQLMethod(funcName) {
			return
		}

		args := node.ChildByFieldName("arguments", lang)
		if args == nil {
			args = node.ChildByFieldName("argument_list", lang)
		}
		if args == nil {
			return
		}

		for i := 0; i < args.ChildCount(); i++ {
			arg := args.Child(i)
			if arg == nil {
				continue
			}

			if !isDynamicSQLNode(arg, file) {
				continue
			}

			line, col := getPosition(string(file.Content), int(node.StartByte()))
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: r.Description(),
				File:        file.Path,
				Line:        line,
				Col:         col,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    r.Severity(),
				Category:    string(r.Category()),
				Tags:        []string{"sql-injection", "owasp-a03", "injection"},
				Fix: `// ❌ Nguy hiểm:
db.query("SELECT * FROM users WHERE id = " + userId)

// ✅ An toàn - dùng parameterized query:
db.query("SELECT * FROM users WHERE id = ?", [userId])`,
			})
			return
		}
	})

	return findings
}

func isSQLMethod(name string) bool {
	sqlMethods := []string{"query", "execute", "run", "all", "get", "prepare", "exec", "fetchall", "fetchone", "executemany"}
	lower := strings.ToLower(name)
	for _, m := range sqlMethods {
		if lower == m || strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

// ========== COMMAND INJECTION AST RULE ==========

type CommandInjectionASTRule struct{}

func (r *CommandInjectionASTRule) ID() string       { return "AST-CMD-001" }
func (r *CommandInjectionASTRule) Title() string    { return "Command Injection - Shell command with user input" }
func (r *CommandInjectionASTRule) Description() string {
	return "Phát hiện shell command được ghép với user input - dễ bị Command Injection"
}
func (r *CommandInjectionASTRule) Severity() Severity { return Critical }
func (r *CommandInjectionASTRule) Category() Category { return Security }
func (r *CommandInjectionASTRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go", "php"}
}

func (r *CommandInjectionASTRule) Check(file *ParsedFile) []Finding {
	return r.CheckAST(file)
}

func (r *CommandInjectionASTRule) CheckAST(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	lang := file.Tree.Language()
	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if node.Type(lang) != "call_expression" && node.Type(lang) != "call" {
			return
		}

		funcNode := node.ChildByFieldName("function", lang)
		if funcNode == nil {
			return
		}

		funcName := strings.ToLower(getNodeText(funcNode, file.Content))
		if !isCommandFunc(funcName) {
			return
		}

		args := node.ChildByFieldName("arguments", lang)
		if args == nil {
			args = node.ChildByFieldName("argument_list", lang)
		}
		if args == nil {
			return
		}

		for i := 0; i < args.ChildCount(); i++ {
			arg := args.Child(i)
			if arg == nil {
				continue
			}

			if !containsUserControlledData(arg, file) {
				continue
			}

			line, col := getPosition(string(file.Content), int(node.StartByte()))
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: r.Description(),
				File:        file.Path,
				Line:        line,
				Col:         col,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    r.Severity(),
				Category:    string(r.Category()),
				Tags:        []string{"command-injection", "owasp-a03", "injection"},
				Fix: `// ❌ Nguy hiểm:
exec("ls " + userInput)

// ✅ An toàn - validate input hoặc dùng array:
const args = [userInput].filter(isValidFilename)
execFile("ls", args)`,
			})
			return
		}
	})

	return findings
}

func isCommandFunc(name string) bool {
	cmdFuncs := []string{"exec", "execSync", "spawn", "execFile", "system", "popen", "shell_exec", "passthru", "proc_open"}
	lower := strings.ToLower(name)
	for _, f := range cmdFuncs {
		if lower == f || strings.Contains(lower, f) {
			return true
		}
	}
	return false
}

// ========== HARDCODED SECRET AST RULE ==========

type HardcodedSecretASTRule struct{}

func (r *HardcodedSecretASTRule) ID() string       { return "AST-SEC-001" }
func (r *HardcodedSecretASTRule) Title() string    { return "Hardcoded Secret - API key/token in source code" }
func (r *HardcodedSecretASTRule) Description() string {
	return "Phát hiện API key, token, password được hardcode trực tiếp trong mã nguồn"
}
func (r *HardcodedSecretASTRule) Severity() Severity { return Critical }
func (r *HardcodedSecretASTRule) Category() Category { return Security }
func (r *HardcodedSecretASTRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go", "php", "ruby"}
}

func (r *HardcodedSecretASTRule) Check(file *ParsedFile) []Finding {
	return r.CheckAST(file)
}

func (r *HardcodedSecretASTRule) CheckAST(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	lang := file.Tree.Language()
	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		nodeType := node.Type(lang)
		if nodeType != "variable_declarator" &&
			nodeType != "assignment_expression" &&
			nodeType != "assignment" &&
			nodeType != "short_var_declaration" &&
			nodeType != "assignment_statement" &&
			nodeType != "var_spec" {
			return
		}

		varNameNode := findSensitiveNameNode(node, file)
		secretNode := findStringLiteralNode(node, file)
		if varNameNode == nil || secretNode == nil {
			return
		}

		varName := getNodeText(varNameNode, file.Content)
		secretValue := strings.Trim(getNodeText(secretNode, file.Content), "`\"'")
		if !isSecretVarName(varName) || len(secretValue) < 8 || isPlaceholder(secretValue) {
			return
		}

		line, col := getPosition(string(file.Content), int(node.StartByte()))
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "Biến '" + varName + "' chứa giá trị nhạy cảm được hardcode",
			File:        file.Path,
			Line:        line,
			Col:         col,
			Snippet:     getSnippet(string(file.Content), line),
			Severity:    r.Severity(),
			Category:    string(r.Category()),
			Tags:        []string{"hardcoded-secret", "cwe-798", "secrets"},
			Fix: `// ❌ Không hardcode secret:
const API_KEY = "sk-1234567890abcdef"

// ✅ Dùng environment variable:
const API_KEY = process.env.API_KEY
// Hoặc dùng secret manager`,
		})
	})

	return findings
}

func isSecretVarName(name string) bool {
	secretPatterns := []string{"api_key", "apikey", "secret", "password", "passwd", "pwd", "token", "auth", "credential", "private_key", "access_key", "api_secret"}
	lower := strings.ToLower(name)
	for _, pattern := range secretPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func isPlaceholder(value string) bool {
	placeholders := []string{"xxx", "yyy", "zzz", "your", "example", "test", "demo", "placeholder", "changeme", "password", "secret"}
	lower := strings.ToLower(value)
	for _, p := range placeholders {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func looksLikeSQL(text string) bool {
	lower := strings.ToLower(text)
	keywords := []string{"select ", "insert ", "update ", "delete ", " from ", " where ", " join ", " into "}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func isDynamicSQLNode(node *gotreesitter.Node, file *ParsedFile) bool {
	lang := file.Tree.Language()
	nodeType := node.Type(lang)
	text := getNodeText(node, file.Content)
	if nodeType == "binary_expression" || nodeType == "concatenation" {
		return true
	}
	if nodeType == "template_string" && (strings.Contains(text, "${") || strings.Contains(text, "#{")) {
		return looksLikeSQL(text)
	}
	if nodeType == "call_expression" || nodeType == "call" {
		funcNode := node.ChildByFieldName("function", lang)
		if funcNode == nil {
			return false
		}
		name := strings.ToLower(getNodeText(funcNode, file.Content))
		return strings.Contains(name, "sprintf") || strings.Contains(name, "format")
	}
	return false
}

func containsUserControlledData(node *gotreesitter.Node, file *ParsedFile) bool {
	lang := file.Tree.Language()
	nodeType := node.Type(lang)
	text := strings.ToLower(getNodeText(node, file.Content))
	if nodeType == "binary_expression" || nodeType == "concatenation" {
		return true
	}
	if nodeType == "template_string" && (strings.Contains(text, "${") || strings.Contains(text, "#{")) {
		return true
	}
	markers := []string{"req", "request", "input", "user", "params", "body", "query", "argv", "form"}
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func findSensitiveNameNode(node *gotreesitter.Node, file *ParsedFile) *gotreesitter.Node {
	lang := file.Tree.Language()
	for _, field := range []string{"name", "left"} {
		if child := node.ChildByFieldName(field, lang); child != nil {
			return child
		}
	}
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		childType := child.Type(lang)
		if childType == "identifier" || childType == "property_identifier" || childType == "field_identifier" || childType == "variable_name" {
			return child
		}
	}
	return nil
}

func findStringLiteralNode(node *gotreesitter.Node, file *ParsedFile) *gotreesitter.Node {
	lang := file.Tree.Language()
	for _, field := range []string{"value", "right"} {
		if child := node.ChildByFieldName(field, lang); child != nil && isStringLiteralType(child.Type(lang)) {
			return child
		}
	}
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if isStringLiteralType(child.Type(lang)) {
			return child
		}
	}
	return nil
}

func isStringLiteralType(nodeType string) bool {
	return nodeType == "string" ||
		nodeType == "string_literal" ||
		nodeType == "raw_string_literal" ||
		nodeType == "interpreted_string_literal" ||
		nodeType == "template_string"
}

// ========== HELPER FUNCTIONS ==========

func queryAST(tree *gotreesitter.Tree, content []byte, query string) []map[string]*gotreesitter.Node {
	// Simplified - in real implementation would use proper tree-sitter queries
	// For now, return empty - this is a placeholder for the actual implementation
	return []map[string]*gotreesitter.Node{}
}

func findNodeByCapture(match map[string]*gotreesitter.Node, capture string) *gotreesitter.Node {
	return match[capture]
}

func getPosition(content string, byteOffset int) (line, col int) {
	line = 1
	col = 1
	for i, c := range content {
		if i >= byteOffset {
			return line, col
		}
		if c == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

func getSnippet(content string, lineNum int) string {
	lines := strings.Split(content, "\n")
	if lineNum > 0 && lineNum <= len(lines) {
		return strings.TrimSpace(lines[lineNum-1])
	}
	return ""
}

// AllASTRules returns all AST-based rules
func AllASTRules() []Rule {
	return []Rule{
		&SQLInjectionASTRule{},
		&CommandInjectionASTRule{},
		&HardcodedSecretASTRule{},
		&JWTWeakSecretASTRule{},
		&PathTraversalASTRule{},
		&EvalInjectionASTRule{},
	}
}

// ========== JWT WEAK SECRET AST RULE ==========

type JWTWeakSecretASTRule struct{}

func (r *JWTWeakSecretASTRule) ID() string       { return "ast-jwt-weak-secret" }
func (r *JWTWeakSecretASTRule) Title() string    { return "JWT Weak Secret" }
func (r *JWTWeakSecretASTRule) Name() string     { return "JWT Weak Secret" }
func (r *JWTWeakSecretASTRule) Description() string {
	return "Phát hiện JWT secret quá ngắn hoặc quá đơn giản"
}
func (r *JWTWeakSecretASTRule) Severity() Severity { return High }
func (r *JWTWeakSecretASTRule) Category() Category { return Security }
func (r *JWTWeakSecretASTRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go", "php", "java"}
}

func (r *JWTWeakSecretASTRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	lang := file.Tree.Language()
 
 	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if node.Type(lang) == "call_expression" || node.Type(lang) == "call" {
			r.checkJWTCall(node, file, &findings)
		}
	})

	return findings
}
 
func (r *JWTWeakSecretASTRule) checkJWTCall(node *gotreesitter.Node, file *ParsedFile, findings *[]Finding) {
	lang := file.Tree.Language()
	funcNode := node.ChildByFieldName("function", lang)
	if funcNode == nil {
		return
	}

	funcName := strings.ToLower(getNodeText(funcNode, file.Content))

	// Check for JWT sign/verify calls
	if !strings.Contains(funcName, "jwt") && !strings.Contains(funcName, "jsonwebtoken") {
		return
 	}
 
 	// Check arguments for weak secret
	args := node.ChildByFieldName("arguments", lang)
	if args == nil {
		args = node.ChildByFieldName("argument_list", lang)
	}
	if args == nil {
		return
	}

	for i := 0; i < int(args.ChildCount()); i++ {
		arg := args.Child(i)
		if arg == nil {
 			continue
		}
 
 		// Check for string literal secret
		if arg.Type(lang) == "string" || arg.Type(lang) == "string_literal" {
			secret := getNodeText(arg, file.Content)
			secret = strings.Trim(secret, `"'`)

			if isWeakSecret(secret) {
				line := getLineNumber(file.Content, node.StartByte())
				*findings = append(*findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Name(),
					Description: "JWT secret quá yếu (ngắn hoặc đơn giản)",
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getNodeText(node, file.Content),
					Severity:    r.Severity(),
					Category:    string(r.Category()),
					Fix: `// ❌ Secret yếu:
jwt.sign(payload, "secret")

// ✅ Secret mạnh từ env:
jwt.sign(payload, process.env.JWT_SECRET)

// ✅ Hoặc dùng RS256 không cần shared secret:
jwt.sign(payload, privateKey, { algorithm: 'RS256' })`,
					Tags: []string{"jwt", "weak-secret", "cwe-798", "security"},
				})
				return
			}
		}
	}
}

func isWeakSecret(secret string) bool {
	if len(secret) < 16 {
		return true
	}

	// Check for common weak secrets
	weakPatterns := []string{
		"secret", "password", "123456", "qwerty", "admin",
		"test", "demo", "key", "token", "jwt",
	}
	lower := strings.ToLower(secret)
	for _, p := range weakPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}

	// Check entropy - if mostly same character
	uniqueChars := make(map[rune]bool)
	for _, c := range secret {
		uniqueChars[c] = true
	}
	if len(uniqueChars) < 8 {
		return true
	}

	return false
}

// ========== PATH TRAVERSAL AST RULE ==========

type PathTraversalASTRule struct{}

func (r *PathTraversalASTRule) ID() string       { return "ast-path-traversal" }
func (r *PathTraversalASTRule) Title() string    { return "Path Traversal" }
func (r *PathTraversalASTRule) Name() string     { return "Path Traversal" }
func (r *PathTraversalASTRule) Description() string {
	return "Phát hiện đường dẫn file được ghép với user input"
}
func (r *PathTraversalASTRule) Severity() Severity { return High }
func (r *PathTraversalASTRule) Category() Category { return Security }
func (r *PathTraversalASTRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go", "php", "java", "ruby"}
}

func (r *PathTraversalASTRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	lang := file.Tree.Language()
 
 	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if node.Type(lang) == "call_expression" || node.Type(lang) == "call" {
			r.checkFileOperation(node, file, &findings)
		}
	})

	return findings
}
 
func (r *PathTraversalASTRule) checkFileOperation(node *gotreesitter.Node, file *ParsedFile, findings *[]Finding) {
	lang := file.Tree.Language()
	funcNode := node.ChildByFieldName("function", lang)
	if funcNode == nil {
		return
	}

	funcName := strings.ToLower(getNodeText(funcNode, file.Content))

	// Check for file system operations
	if !isFileOperation(funcName) {
		return
 	}
 
 	// Check arguments for path concatenation with user input
	args := node.ChildByFieldName("arguments", lang)
	if args == nil {
		args = node.ChildByFieldName("argument_list", lang)
	}
	if args == nil {
		return
	}

	for i := 0; i < int(args.ChildCount()); i++ {
		arg := args.Child(i)
		if arg == nil {
			continue
		}

		if r.isDangerousPath(arg, file) {
			line := getLineNumber(file.Content, node.StartByte())
			*findings = append(*findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Name(),
				Description: "Đường dẫn file chứa user input - có thể bị Path Traversal",
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getNodeText(node, file.Content),
				Severity:    r.Severity(),
				Category:    string(r.Category()),
				Fix: `// ❌ Nguy hiểm - path ghép user input:
fs.readFile("./uploads/" + req.query.filename)

// ✅ An toàn - validate + sanitize:
const filename = path.basename(req.query.filename)
const safePath = path.join("./uploads", filename)
fs.readFile(safePath)

// ✅ Hoặc dùng whitelist:
const allowed = ["report1.pdf", "report2.pdf"]
if (!allowed.includes(filename)) throw new Error("Invalid file")`,
				Tags: []string{"path-traversal", "directory-traversal", "owasp-a01", "security"},
			})
			return
		}
	}
}

func (r *PathTraversalASTRule) isDangerousPath(node *gotreesitter.Node, file *ParsedFile) bool {
	lang := file.Tree.Language()
	nodeType := node.Type(lang)

	// Binary expression with path separators
	if nodeType == "binary_expression" || nodeType == "concatenation" {
		text := getNodeText(node, file.Content)
		if strings.Contains(text, "/") || strings.Contains(text, "\\") ||
		   strings.Contains(text, "path.join") || strings.Contains(text, "__dirname") {
			return true
		}
	}

	// Template string with path
	if nodeType == "template_string" {
		for i := 0; i < int(node.ChildCount()); i++ {
			if child := node.Child(i); child != nil {
				if strings.Contains(child.Type(lang), "substitution") ||
				   strings.Contains(child.Type(lang), "interpolation") {
					return true
				}
			}
		}
	}

	// Call expression like path.join with user input
	if nodeType == "call_expression" || nodeType == "call" {
		funcNode := node.ChildByFieldName("function", lang)
		if funcNode != nil {
			funcName := strings.ToLower(getNodeText(funcNode, file.Content))
			if strings.Contains(funcName, "join") || strings.Contains(funcName, "resolve") {
				return true
			}
		}
	}

	return false
}

func isFileOperation(name string) bool {
	name = strings.ToLower(name)
	ops := []string{
		"readfile", "writefile", "readdir", "unlink", "stat",
		"createReadStream", "createWriteStream", "open", "fopen",
		"file_get_contents", "file_put_contents", "include", "require",
	}
	for _, op := range ops {
		if name == op || strings.HasSuffix(name, "."+op) {
			return true
		}
	}
	return false
}

// ========== EVAL INJECTION AST RULE ==========

type EvalInjectionASTRule struct{}

func (r *EvalInjectionASTRule) ID() string       { return "ast-eval-injection" }
func (r *EvalInjectionASTRule) Title() string    { return "Eval/Code Injection" }
func (r *EvalInjectionASTRule) Name() string     { return "Eval/Code Injection" }
func (r *EvalInjectionASTRule) Description() string {
	return "Phát hiện eval/new Function với user input"
}
func (r *EvalInjectionASTRule) Severity() Severity { return Critical }
func (r *EvalInjectionASTRule) Category() Category { return Security }
func (r *EvalInjectionASTRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "php"}
}

func (r *EvalInjectionASTRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	lang := file.Tree.Language()
 
 	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if node.Type(lang) == "call_expression" || node.Type(lang) == "call" {
			r.checkEvalCall(node, file, &findings)
		}
		// Check for new Function()
		if node.Type(lang) == "new_expression" || node.Type(lang) == "new" {
			r.checkNewFunction(node, file, &findings)
		}
	})

	return findings
}
 
func (r *EvalInjectionASTRule) checkEvalCall(node *gotreesitter.Node, file *ParsedFile, findings *[]Finding) {
	lang := file.Tree.Language()
	funcNode := node.ChildByFieldName("function", lang)
	if funcNode == nil {
		return
	}

	funcName := strings.ToLower(getNodeText(funcNode, file.Content))

	// Check for eval-like functions
	if !isEvalFunction(funcName) {
		return
 	}
 
 	// Check arguments contain user input
	args := node.ChildByFieldName("arguments", lang)
	if args == nil {
		args = node.ChildByFieldName("argument_list", lang)
	}
	if args == nil {
		return
	}

	for i := 0; i < int(args.ChildCount()); i++ {
		arg := args.Child(i)
		if arg == nil {
			continue
		}

		if r.containsUserInput(arg, file) {
			line := getLineNumber(file.Content, node.StartByte())
			*findings = append(*findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Name(),
				Description: "eval/new Function chứa user input - RCE vulnerability",
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getNodeText(node, file.Content),
				Severity:    r.Severity(),
				Category:    string(r.Category()),
				Fix: `// ❌ Cực kỳ nguy hiểm - eval với user input:
eval(req.body.code)

// ❌ Cũng nguy hiểm:
new Function(req.body.code)()

// ✅ Không dùng eval, parse JSON an toàn:
JSON.parse(req.body.data)

// ✅ Hoặc dùng vm2 với sandbox:
const { VM } = require('vm2')
const vm = new VM({ timeout: 1000, sandbox: {} })
vm.run(req.body.code) // vẫn cẩn thận!`,
				Tags: []string{"eval-injection", "rce", "code-injection", "owasp-a03", "security"},
			})
			return
		}
	}
}

func (r *EvalInjectionASTRule) checkNewFunction(node *gotreesitter.Node, file *ParsedFile, findings *[]Finding) {
	lang := file.Tree.Language()
	// Get the constructor name
	constructor := node.ChildByFieldName("constructor", lang)
	if constructor == nil {
		// Try first child
		constructor = node.Child(0)
	}
	if constructor == nil {
		return
	}

	consName := strings.ToLower(getNodeText(constructor, file.Content))
	if consName != "function" {
		return
 	}
 
 	// Check arguments for user input
	args := node.ChildByFieldName("arguments", lang)
	if args == nil {
		args = node.ChildByFieldName("argument_list", lang)
	}
	if args == nil {
		return
	}

	for i := 0; i < int(args.ChildCount()); i++ {
		arg := args.Child(i)
		if arg == nil {
			continue
		}

		if r.containsUserInput(arg, file) {
			line := getLineNumber(file.Content, node.StartByte())
			*findings = append(*findings, Finding{
				RuleID:      r.ID(),
				Title:       "Dynamic Function with User Input",
				Description: "new Function() chứa user input - RCE vulnerability",
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getNodeText(node, file.Content),
				Severity:    Critical,
				Category:    "security",
				Fix: `// ❌ Nguy hiểm:
new Function(req.body.code)()

// ✅ Parse JSON thay vì thực thi code:
const data = JSON.parse(req.body.data)
processData(data)`,
				Tags: []string{"function-constructor", "rce", "code-injection", "security"},
			})
			return
		}
	}
}
 
func (r *EvalInjectionASTRule) containsUserInput(node *gotreesitter.Node, file *ParsedFile) bool {
	lang := file.Tree.Language()
	nodeType := node.Type(lang)

	// Binary expression
	if nodeType == "binary_expression" || nodeType == "concatenation" {
		return true
	}

 	// Template string
	if nodeType == "template_string" {
		for i := 0; i < int(node.ChildCount()); i++ {
			if child := node.Child(i); child != nil {
				if strings.Contains(child.Type(lang), "substitution") ||
				   strings.Contains(child.Type(lang), "interpolation") {
					return true
				}
			}
		}
	}
 
 	// Identifier containing request/input
	// Identifier containing request/input
	if nodeType == "identifier" || nodeType == "member_expression" {
		text := strings.ToLower(getNodeText(node, file.Content))
		dangerous := []string{"req", "request", "input", "user", "params", "body", "query", "data"}
		for _, d := range dangerous {
			if strings.Contains(text, d) {
				return true
			}
		}
	}

	return false
}

func isEvalFunction(name string) bool {
	name = strings.ToLower(name)
	funcs := []string{
		"eval", "settimeout", "setinterval", "function",
		"exec", "system", "passthru", "shell_exec",
	}
	for _, f := range funcs {
		if name == f || strings.HasSuffix(name, "."+f) {
			return true
		}
	}
	return false
}
