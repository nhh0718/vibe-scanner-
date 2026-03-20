package rules

import (
	"regexp"
	"strings"

	"github.com/odvcencio/gotreesitter"
)

// ========== CORS WILDCARD RULE ==========
type CORSWildcardRule struct{}

func (r *CORSWildcardRule) ID() string    { return "VS-SEC-011" }
func (r *CORSWildcardRule) Title() string { return "CORS wildcard origin" }
func (r *CORSWildcardRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *CORSWildcardRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeType := node.Type(lang)
			if nodeType != "pair" && nodeType != "property" && nodeType != "assignment_expression" {
				return
			}
			text := getNodeText(node, file.Content)
			if strings.Contains(strings.ToLower(text), "origin") && (strings.Contains(text, "'*'") || strings.Contains(text, "\"*\"")) {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "CORS wildcard cho phép mọi origin truy cập.",
					Fix:         "Chỉ whitelist các origin hợp lệ thay vì dùng *.",
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    High,
					Category:    "security",
					Tags:        []string{"cors", "origin", "security"},
				})
			}
		})
		return findings
	}
	lines := strings.Split(string(file.Content), "\n")
	pattern := regexp.MustCompile(`(?i)origin\s*:\s*["']?\*["']?`)
	for i, line := range lines {
		if pattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "CORS wildcard cho phép mọi origin truy cập.",
				Fix:         "Chỉ whitelist các origin hợp lệ thay vì dùng *.",
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"cors", "origin", "security"},
			})
		}
	}
	return findings
}

// ========== DANGEROUS HTML RULE ==========
type DangerousHTMLRule struct{}

func (r *DangerousHTMLRule) ID() string    { return "VS-SEC-013" }
func (r *DangerousHTMLRule) Title() string { return "dangerouslySetInnerHTML với user input" }
func (r *DangerousHTMLRule) Languages() []string {
	return []string{"javascript", "typescript", "jsx", "tsx"}
}

func (r *DangerousHTMLRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeType := node.Type(lang)
			if nodeType != "jsx_attribute" && nodeType != "property_identifier" {
				return
			}
			text := getNodeText(node, file.Content)
			if strings.Contains(text, "dangerouslySetInnerHTML") {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "dangerouslySetInnerHTML có thể dẫn đến XSS nếu dữ liệu chưa được sanitize.",
					Fix:         "Sanitize HTML hoặc render bằng component an toàn thay vì inject trực tiếp.",
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    High,
					Category:    "security",
					Tags:        []string{"xss", "react", "html"},
				})
			}
		})
		return findings
	}
	lines := strings.Split(string(file.Content), "\n")
	pattern := regexp.MustCompile(`(?i)dangerouslySetInnerHTML`)
	for i, line := range lines {
		if pattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "dangerouslySetInnerHTML có thể dẫn đến XSS nếu dữ liệu chưa được sanitize.",
				Fix:         "Sanitize HTML hoặc render bằng component an toàn thay vì inject trực tiếp.",
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"xss", "react", "html"},
			})
		}
	}
	return findings
}

// ========== XSS INNERHTML RULE ==========
type XSSInnerHTMLRule struct{}

func (r *XSSInnerHTMLRule) ID() string    { return "VS-SEC-014" }
func (r *XSSInnerHTMLRule) Title() string { return "innerHTML với user data - XSS" }
func (r *XSSInnerHTMLRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *XSSInnerHTMLRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "assignment_expression" {
				return
			}
			text := getNodeText(node, file.Content)
			if strings.Contains(text, "innerHTML") {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Gán trực tiếp vào innerHTML có thể gây XSS.",
					Fix:         "Dùng textContent hoặc sanitize trước khi render HTML.",
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    High,
					Category:    "security",
					Tags:        []string{"xss", "dom", "html"},
				})
			}
		})
		return findings
	}
	lines := strings.Split(string(file.Content), "\n")
	pattern := regexp.MustCompile(`(?i)innerHTML\s*=`)
	for i, line := range lines {
		if pattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Gán trực tiếp vào innerHTML có thể gây XSS.",
				Fix:         "Dùng textContent hoặc sanitize trước khi render HTML.",
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    High,
				Category:    "security",
				Tags:        []string{"xss", "dom", "html"},
			})
		}
	}
	return findings
}

// ========== HTTP NOT HTTPS RULE ==========
type HTTPNotHTTPSRule struct{}

func (r *HTTPNotHTTPSRule) ID() string    { return "VS-SEC-018" }
func (r *HTTPNotHTTPSRule) Title() string { return "HTTP thay vì HTTPS trong production" }
func (r *HTTPNotHTTPSRule) Languages() []string { return []string{"*"} }

func (r *HTTPNotHTTPSRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	httpPattern := regexp.MustCompile(`(?i)http://[^"]+`)
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeType := node.Type(lang)
			if nodeType != "string" && nodeType != "string_fragment" && nodeType != "template_string" {
				return
			}
			text := getNodeText(node, file.Content)
			if !httpPattern.MatchString(text) {
				return
			}
			if strings.Contains(text, "localhost") || strings.Contains(text, "127.0.0.1") {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Sử dụng HTTP trong production có thể làm lộ dữ liệu khi truyền tải.",
				Fix:         "Chuyển endpoint sang HTTPS.",
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    Medium,
				Category:    "security",
				Tags:        []string{"http", "https", "transport"},
			})
		})
		return findings
	}
	lines := strings.Split(string(file.Content), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !httpPattern.MatchString(line) {
			continue
		}
		if strings.Contains(line, "localhost") || strings.Contains(line, "127.0.0.1") {
			continue
		}
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "Sử dụng HTTP trong production có thể làm lộ dữ liệu khi truyền tải.",
			Fix:         "Chuyển endpoint sang HTTPS.",
			File:        file.Path,
			Line:        i + 1,
			Col:         1,
			Snippet:     trimmed,
			Severity:    Medium,
			Category:    "security",
			Tags:        []string{"http", "https", "transport"},
		})
	}
	return findings
}

// ========== MATH RANDOM SECURITY RULE ==========
type MathRandomSecurityRule struct{}

func (r *MathRandomSecurityRule) ID() string    { return "VS-SEC-019" }
func (r *MathRandomSecurityRule) Title() string { return "Math.random() cho mục đích bảo mật" }
func (r *MathRandomSecurityRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *MathRandomSecurityRule) Check(file *ParsedFile) []Finding {
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
			if getNodeText(funcNode, file.Content) == "Math.random" {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Math.random() không đủ an toàn cho token, secret hoặc ID nhạy cảm.",
					Fix:         "Dùng crypto.randomBytes() hoặc Web Crypto API.",
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    Medium,
					Category:    "security",
					Tags:        []string{"random", "crypto", "security"},
				})
			}
		})
		return findings
	}
	lines := strings.Split(string(file.Content), "\n")
	pattern := regexp.MustCompile(`(?i)Math\.random\s*\(`)
	for i, line := range lines {
		if pattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Math.random() không đủ an toàn cho token, secret hoặc ID nhạy cảm.",
				Fix:         "Dùng crypto.randomBytes() hoặc Web Crypto API.",
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Medium,
				Category:    "security",
				Tags:        []string{"random", "crypto", "security"},
			})
		}
	}
	return findings
}

// ========== EXPOSE STACK TRACE RULE ==========
type ExposeStackTraceRule struct{}

func (r *ExposeStackTraceRule) ID() string    { return "VS-SEC-017" }
func (r *ExposeStackTraceRule) Title() string { return "Expose stack trace trong error response" }
func (r *ExposeStackTraceRule) Languages() []string {
	return []string{"javascript", "typescript", "python", "go"}
}

func (r *ExposeStackTraceRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if node.Type(lang) != "member_expression" && node.Type(lang) != "call_expression" {
				return
			}
			text := getNodeText(node, file.Content)
			textLower := strings.ToLower(text)
			if strings.Contains(textLower, "err.stack") || strings.Contains(textLower, "stacktrace") ||
				strings.Contains(textLower, "printstacktrace") || strings.Contains(textLower, "debug.stack(") {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Stack trace bị expose ra response hoặc output, làm lộ cấu trúc hệ thống.",
					Fix:         "Ẩn stack trace ở production, chỉ log nội bộ.",
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    Medium,
					Category:    "security",
					Tags:        []string{"stacktrace", "error", "security"},
				})
			}
		})
		return findings
	}
	lines := strings.Split(string(file.Content), "\n")
	pattern := regexp.MustCompile(`(?i)(err\.stack|stacktrace|printStackTrace|debug\.Stack\()`) 
	for i, line := range lines {
		if pattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Stack trace bị expose ra response hoặc output, làm lộ cấu trúc hệ thống.",
				Fix:         "Ẩn stack trace ở production, chỉ log nội bộ.",
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Medium,
				Category:    "security",
				Tags:        []string{"stacktrace", "error", "security"},
			})
		}
	}
	return findings
}

// ========== COMPLEXITY RULE ==========
type ComplexityRule struct {
	Threshold int
}

func (r *ComplexityRule) ID() string    { return "VS-QUAL-001" }
func (r *ComplexityRule) Title() string { return "Cyclomatic complexity cao" }
func (r *ComplexityRule) Languages() []string { return []string{"*"} }

func (r *ComplexityRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if r.Threshold == 0 {
		r.Threshold = 10
	}
	if file.Tree == nil {
		return findings
	}
	lang := file.Tree.Language()
	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if !qualityIsFunctionLikeNode(node, lang) {
			return
		}
		branchCount := qualityCountBranches(node, file, true)
		if branchCount <= r.Threshold {
			return
		}
		line := getLineNumber(file.Content, node.StartByte())
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "Hàm này có quá nhiều nhánh logic, khó đọc và bảo trì.",
			Fix:         "Tách thành các hàm nhỏ hơn, mỗi hàm làm một việc.",
			File:        file.Path,
			Line:        line,
			Col:         1,
			Snippet:     getSnippet(string(file.Content), line),
			Severity:    Medium,
			Category:    "quality",
			Tags:        []string{"complexity", "refactor"},
		})
	})
	return findings
}

// ========== LONG FUNCTION RULE ==========
type LongFunctionRule struct{}

func (r *LongFunctionRule) ID() string    { return "VS-QUAL-002" }
func (r *LongFunctionRule) Title() string { return "Function dài > 50 dòng" }
func (r *LongFunctionRule) Languages() []string { return []string{"*"} }

func (r *LongFunctionRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	lang := file.Tree.Language()
	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if !qualityIsFunctionLikeNode(node, lang) {
			return
		}
		startLine := getLineNumber(file.Content, node.StartByte())
		endLine := getLineNumber(file.Content, node.EndByte())
		lineCount := endLine - startLine + 1
		if lineCount <= 50 {
			return
		}
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "Hàm quá dài (>50 dòng), nên tách thành các hàm nhỏ hơn.",
			Fix:         "Tách logic thành các helper functions riêng.",
			File:        file.Path,
			Line:        startLine,
			Col:         1,
			Snippet:     getSnippet(string(file.Content), startLine),
			Severity:    Low,
			Category:    "quality",
			Tags:        []string{"length", "refactor"},
		})
	})
	return findings
}

// ========== LONG FILE RULE ==========
type LongFileRule struct{}

func (r *LongFileRule) ID() string    { return "VS-QUAL-003" }
func (r *LongFileRule) Title() string { return "File dài > 300 dòng" }
func (r *LongFileRule) Languages() []string { return []string{"*"} }

func (r *LongFileRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	lines := strings.Split(string(file.Content), "\n")
	if len(lines) > 300 {
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "File quá dài (>300 dòng), khó đọc và bảo trì.",
			Fix:         "Tách thành các modules/files nhỏ hơn.",
			File:        file.Path,
			Line:        1,
			Col:         1,
			Snippet:     "total lines > 300",
			Severity:    Low,
			Category:    "quality",
			Tags:        []string{"length", "modularity"},
		})
	}
	return findings
}

// ========== DEEP NESTING RULE ==========
type DeepNestingRule struct{}

func (r *DeepNestingRule) ID() string    { return "VS-QUAL-004" }
func (r *DeepNestingRule) Title() string { return "Nesting depth > 4" }
func (r *DeepNestingRule) Languages() []string { return []string{"*"} }

func (r *DeepNestingRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	maxDepth, maxDepthLine := qualityMaxControlNesting(file.Tree.RootNode(), file, 0)
	if maxDepth > 4 {
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "Code có nesting sâu (>4 levels), khó theo dõi logic.",
			Fix:         "Tách nested blocks thành các hàm riêng.",
			File:        file.Path,
			Line:        maxDepthLine,
			Col:         1,
			Snippet:     getSnippet(string(file.Content), maxDepthLine),
			Severity:    Low,
			Category:    "quality",
			Tags:        []string{"nesting", "complexity"},
		})
	}
	return findings
}

// ========== TOO MANY PARAMS RULE ==========
type TooManyParamsRule struct{}

func (r *TooManyParamsRule) ID() string    { return "VS-QUAL-005" }
func (r *TooManyParamsRule) Title() string { return "Parameter count > 5" }
func (r *TooManyParamsRule) Languages() []string { return []string{"*"} }

func (r *TooManyParamsRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	lang := file.Tree.Language()
	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if !qualityIsFunctionLikeNode(node, lang) {
			return
		}
		params := qualityFunctionParameterCount(node, lang)
		if params <= 5 {
			return
		}
		line := getLineNumber(file.Content, node.StartByte())
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "Hàm có quá nhiều tham số (>5), nên dùng object/struct.",
			Fix:         "Gom params thành một config object.",
			File:        file.Path,
			Line:        line,
			Col:         1,
			Snippet:     getSnippet(string(file.Content), line),
			Severity:    Low,
			Category:    "quality",
			Tags:        []string{"params", "api"},
		})
	})
	return findings
}

// ========== EMPTY CATCH RULE ==========
type EmptyCatchRule struct{}

func (r *EmptyCatchRule) ID() string    { return "VS-QUAL-006" }
func (r *EmptyCatchRule) Title() string { return "Empty catch block" }
func (r *EmptyCatchRule) Languages() []string {
	return []string{"javascript", "typescript", "java"}
}

func (r *EmptyCatchRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	lang := file.Tree.Language()
	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if node.Type(lang) != "catch_clause" {
			return
		}
		if !qualityIsEmptyCatchClause(node, file) {
			return
		}
		line := getLineNumber(file.Content, node.StartByte())
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "Catch block rỗng - lỗi bị nuốt, khó debug.",
			Fix: `// ❌ Không tốt:
catch(e) {}

// ✅ Nên làm:
catch(e) {
  console.error(e)
  // hoặc throw lại, hoặc xử lý lỗi`,
			File:        file.Path,
			Line:        line,
			Col:         1,
			Snippet:     getSnippet(string(file.Content), line),
			Severity:    Medium,
			Category:    "quality",
			Tags:        []string{"error-handling"},
		})
	})
	return findings
}

// ========== CONSOLE LOG RULE ==========
type ConsoleLogRule struct{}

func (r *ConsoleLogRule) ID() string    { return "VS-QUAL-007" }
func (r *ConsoleLogRule) Title() string { return "console.log còn trong code" }
func (r *ConsoleLogRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *ConsoleLogRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree == nil {
		return findings
	}
	walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
		if !qualityIsConsoleLogCall(node, file) {
			return
		}
		line := getLineNumber(file.Content, node.StartByte())
		findings = append(findings, Finding{
			RuleID:      r.ID(),
			Title:       r.Title(),
			Description: "console.log không nên commit vào production.",
			Fix:         "Xóa console.log hoặc dùng logger chuyên nghiệp.",
			File:        file.Path,
			Line:        line,
			Col:         1,
			Snippet:     getSnippet(string(file.Content), line),
			Severity:    Low,
			Category:    "quality",
			Tags:        []string{"debug", "cleanup"},
		})
	})
	return findings
}

// ========== UNUSED VAR RULE ==========
type UnusedVarRule struct{}

func (r *UnusedVarRule) ID() string    { return "VS-QUAL-010" }
func (r *UnusedVarRule) Title() string { return "Biến khai báo nhưng không dùng" }
func (r *UnusedVarRule) Languages() []string {
	return []string{"javascript", "typescript", "go"}
}

func (r *UnusedVarRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	content := string(file.Content)
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeType := node.Type(lang)
			if nodeType != "variable_declarator" && nodeType != "lexical_declaration" {
				return
			}
			if nodeType == "lexical_declaration" {
				for i := 0; i < node.NamedChildCount(); i++ {
					child := node.NamedChild(i)
					if child != nil && child.Type(lang) == "variable_declarator" {
						nameNode := child.ChildByFieldName("name", lang)
						if nameNode == nil {
							continue
						}
						varName := getNodeText(nameNode, file.Content)
						if len(varName) == 0 || varName == "_" {
							continue
						}
						varPattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(varName) + `\b`)
						if len(varPattern.FindAllString(content, -1)) == 1 {
							line := getLineNumber(file.Content, child.StartByte())
							findings = append(findings, Finding{
								RuleID:      r.ID(),
								Title:       r.Title(),
								Description: "Biến được khai báo nhưng không sử dụng.",
								Fix:         "Xóa biến không dùng hoặc sử dụng nó.",
								File:        file.Path,
								Line:        line,
								Col:         1,
								Snippet:     getSnippet(content, line),
								Severity:    Low,
								Category:    "quality",
								Tags:        []string{"unused", "cleanup"},
							})
						}
					}
				}
				return
			}
		})
		return findings
	}
	pattern := regexp.MustCompile(`(?i)\b(const|let|var)\s+(\w+)\s*=\s*[^;]+;?`)
	matches := pattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) <= 2 {
			continue
		}
		varName := match[2]
		varPattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(varName) + `\b`)
		if len(varPattern.FindAllString(content, -1)) == 1 {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Biến được khai báo nhưng không sử dụng.",
				Fix:         "Xóa biến không dùng hoặc sử dụng nó.",
				File:        file.Path,
				Line:        1,
				Col:         1,
				Snippet:     varName,
				Severity:    Low,
				Category:    "quality",
				Tags:        []string{"unused", "cleanup"},
			})
		}
	}
	return findings
}

// ========== DEAD CODE RULE ==========
type DeadCodeRule struct{}

func (r *DeadCodeRule) ID() string    { return "VS-QUAL-011" }
func (r *DeadCodeRule) Title() string { return "Dead code sau return" }
func (r *DeadCodeRule) Languages() []string { return []string{"*"} }

func (r *DeadCodeRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeType := node.Type(lang)
			if nodeType != "statement_block" && nodeType != "block" {
				return
			}
			foundTerminator := false
			for i := 0; i < node.NamedChildCount(); i++ {
				child := node.NamedChild(i)
				if child == nil {
					continue
				}
				if foundTerminator {
					line := getLineNumber(file.Content, child.StartByte())
					findings = append(findings, Finding{
						RuleID:      r.ID(),
						Title:       r.Title(),
						Description: "Code sau return statement không bao giờ chạy.",
						Fix:         "Xóa dead code.",
						File:        file.Path,
						Line:        line,
						Col:         1,
						Snippet:     getSnippet(string(file.Content), line),
						Severity:    Low,
						Category:    "quality",
						Tags:        []string{"dead-code"},
					})
					return
				}
				if qualityIsTerminatingStatement(child.Type(lang)) {
					foundTerminator = true
				}
			}
		})
		if len(findings) > 0 {
			return findings
		}
	}
	lines := strings.Split(string(file.Content), "\n")
	foundReturn := false
	deadCodeStart := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if foundReturn {
			if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") && !strings.HasPrefix(trimmed, "*") && trimmed != "}" && !strings.HasPrefix(trimmed, "return") {
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Code sau return statement không bao giờ chạy.",
					Fix:         "Xóa dead code.",
					File:        file.Path,
					Line:        deadCodeStart + 1,
					Col:         1,
					Snippet:     trimmed,
					Severity:    Low,
					Category:    "quality",
					Tags:        []string{"dead-code"},
				})
				foundReturn = false
			}
			if trimmed == "}" {
				foundReturn = false
			}
		}
		if strings.HasPrefix(trimmed, "return") && !strings.Contains(trimmed, "function") {
			foundReturn = true
			deadCodeStart = i + 1
		}
	}
	return findings
}

// ========== MAGIC NUMBER RULE ==========
type MagicNumberRule struct{}

func (r *MagicNumberRule) ID() string    { return "VS-QUAL-012" }
func (r *MagicNumberRule) Title() string { return "Magic numbers" }
func (r *MagicNumberRule) Languages() []string { return []string{"*"} }

func (r *MagicNumberRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	magicPattern := regexp.MustCompile(`(?i)[^\w]([2-9]\d{2,}|\d{4,})[^\w]`)
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeType := node.Type(lang)
			if nodeType != "number" && nodeType != "integer" && nodeType != "float" {
				return
			}
			text := getNodeText(node, file.Content)
			if !magicPattern.MatchString(" " + text + " ") {
				return
			}
			parent := node.Parent()
			if parent != nil {
				pt := parent.Type(lang)
				if pt == "variable_declarator" || pt == "assignment_expression" || pt == "const_declaration" {
					return
				}
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Số cứng không có context, khó hiểu.",
				Fix:         "Đặt tên constant: const MAX_ITEMS = 100",
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    Info,
				Category:    "quality",
				Tags:        []string{"magic-number"},
			})
		})
		return findings
	}
	lines := strings.Split(string(file.Content), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if magicPattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Số cứng không có context, khó hiểu.",
				Fix:         "Đặt tên constant: const MAX_ITEMS = 100",
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     trimmed,
				Severity:    Info,
				Category:    "quality",
				Tags:        []string{"magic-number"},
			})
		}
	}
	return findings
}

// ========== TODO COMMENT RULE ==========
type TodoCommentRule struct{}

func (r *TodoCommentRule) ID() string    { return "VS-QUAL-013" }
func (r *TodoCommentRule) Title() string { return "TODO/FIXME comment" }
func (r *TodoCommentRule) Languages() []string { return []string{"*"} }

func (r *TodoCommentRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	todoPattern := regexp.MustCompile(`(?i)(TODO|FIXME|HACK|XXX|BUG):?\s*(.+)`)
	if file.Tree != nil {
		lang := file.Tree.Language()
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			nodeType := node.Type(lang)
			if nodeType != "comment" && nodeType != "line_comment" && nodeType != "block_comment" {
				return
			}
			text := getNodeText(node, file.Content)
			matches := todoPattern.FindStringSubmatch(text)
			if len(matches) > 2 {
				line := getLineNumber(file.Content, node.StartByte())
				findings = append(findings, Finding{
					RuleID:      r.ID(),
					Title:       r.Title(),
					Description: "Comment đánh dấu công việc chưa xong: " + matches[2],
					Fix:         "Hoàn thành TODO hoặc chuyển vào issue tracker.",
					File:        file.Path,
					Line:        line,
					Col:         1,
					Snippet:     getSnippet(string(file.Content), line),
					Severity:    Info,
					Category:    "quality",
					Tags:        []string{"todo", "technical-debt"},
				})
			}
		})
		return findings
	}
	lines := strings.Split(string(file.Content), "\n")
	for i, line := range lines {
		matches := todoPattern.FindStringSubmatch(line)
		if len(matches) > 2 {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "Comment đánh dấu công việc chưa xong: " + matches[2],
				Fix:         "Hoàn thành TODO hoặc chuyển vào issue tracker.",
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Info,
				Category:    "quality",
				Tags:        []string{"todo", "technical-debt"},
			})
		}
	}
	return findings
}

// ========== VAR INSTEAD OF CONST RULE ==========
type VarInsteadOfConstRule struct{}

func (r *VarInsteadOfConstRule) ID() string    { return "VS-QUAL-015" }
func (r *VarInsteadOfConstRule) Title() string { return "Dùng var thay vì const/let" }
func (r *VarInsteadOfConstRule) Languages() []string {
	return []string{"javascript", "typescript"}
}

func (r *VarInsteadOfConstRule) Check(file *ParsedFile) []Finding {
	var findings []Finding
	if file.Tree != nil {
		walkTree(file.Tree.RootNode(), func(node *gotreesitter.Node) {
			if !qualityIsVarDeclaration(node, file) {
				return
			}
			line := getLineNumber(file.Content, node.StartByte())
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "var có scope không rõ ràng, nên dùng const hoặc let.",
				Fix: `// ❌ Cũ:
var x = 5

// ✅ Mới:
const x = 5 // hoặc let nếu cần thay đổi`,
				File:        file.Path,
				Line:        line,
				Col:         1,
				Snippet:     getSnippet(string(file.Content), line),
				Severity:    Low,
				Category:    "quality",
				Tags:        []string{"es6", "best-practice"},
			})
		})
		return findings
	}
	lines := strings.Split(string(file.Content), "\n")
	pattern := regexp.MustCompile(`(?i)^\s*var\s+\w+\s*=\s*`)
	for i, line := range lines {
		if pattern.MatchString(line) {
			findings = append(findings, Finding{
				RuleID:      r.ID(),
				Title:       r.Title(),
				Description: "var có scope không rõ ràng, nên dùng const hoặc let.",
				Fix: `// ❌ Cũ:
var x = 5

// ✅ Mới:
const x = 5 // hoặc let nếu cần thay đổi`,
				File:        file.Path,
				Line:        i + 1,
				Col:         1,
				Snippet:     strings.TrimSpace(line),
				Severity:    Low,
				Category:    "quality",
				Tags:        []string{"es6", "best-practice"},
			})
		}
	}
	return findings
}

// Helper functions
func qualityIsFunctionLikeNode(node *gotreesitter.Node, lang *gotreesitter.Language) bool {
	if node == nil {
		return false
	}
	nodeType := node.Type(lang)
	return nodeType == "function_declaration" ||
		nodeType == "function_expression" ||
		nodeType == "arrow_function" ||
		nodeType == "method_definition" ||
		nodeType == "function_definition" ||
		nodeType == "func_literal" ||
		nodeType == "method_declaration" ||
		strings.Contains(nodeType, "function") ||
		strings.Contains(nodeType, "method")
}

func qualityIsBranchNode(nodeType string) bool {
	return nodeType == "if_statement" ||
		nodeType == "for_statement" ||
		nodeType == "for_in_statement" ||
		nodeType == "while_statement" ||
		nodeType == "do_statement" ||
		nodeType == "switch_statement" ||
		nodeType == "switch_case" ||
		nodeType == "case_statement" ||
		nodeType == "catch_clause" ||
		nodeType == "conditional_expression"
}

func qualityIsTerminatingStatement(nodeType string) bool {
	return nodeType == "return_statement" ||
		nodeType == "throw_statement" ||
		nodeType == "break_statement" ||
		nodeType == "continue_statement"
}

func qualityIsVarDeclaration(node *gotreesitter.Node, file *ParsedFile) bool {
	lang := file.Tree.Language()
	if node == nil {
		return false
	}
	nodeType := node.Type(lang)
	if nodeType != "variable_declaration" && nodeType != "local_variable_declaration" {
		return false
	}
	text := strings.TrimSpace(getNodeText(node, file.Content))
	return strings.HasPrefix(text, "var ")
}

func qualityCountBranches(node *gotreesitter.Node, file *ParsedFile, isRoot bool) int {
	if node == nil {
		return 0
	}
	lang := file.Tree.Language()
	if !isRoot && qualityIsFunctionLikeNode(node, lang) {
		return 0
	}
	count := 0
	if qualityIsBranchNode(node.Type(lang)) {
		count++
	}
	for i := 0; i < node.ChildCount(); i++ {
		count += qualityCountBranches(node.Child(i), file, false)
	}
	return count
}

func qualityFunctionParameterCount(node *gotreesitter.Node, lang *gotreesitter.Language) int {
	for _, field := range []string{"parameters", "parameter_list"} {
		params := node.ChildByFieldName(field, lang)
		if params != nil {
			return params.NamedChildCount()
		}
	}
	return 0
}

func qualityMaxControlNesting(node *gotreesitter.Node, file *ParsedFile, depth int) (int, int) {
	if node == nil {
		return depth, 1
	}
	lang := file.Tree.Language()
	currentDepth := depth
	if qualityIsBranchNode(node.Type(lang)) {
		currentDepth++
	}
	maxDepth := currentDepth
	maxLine := getLineNumber(file.Content, node.StartByte())
	for i := 0; i < node.ChildCount(); i++ {
		childDepth, childLine := qualityMaxControlNesting(node.Child(i), file, currentDepth)
		if childDepth > maxDepth {
			maxDepth = childDepth
			maxLine = childLine
		}
	}
	return maxDepth, maxLine
}

func qualityIsEmptyCatchClause(node *gotreesitter.Node, file *ParsedFile) bool {
	lang := file.Tree.Language()
	for _, field := range []string{"body", "block"} {
		body := node.ChildByFieldName(field, lang)
		if body != nil {
			return body.NamedChildCount() == 0
		}
	}
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type(lang) == "statement_block" || child.Type(lang) == "block" {
			return child.NamedChildCount() == 0
		}
	}
	return strings.Contains(strings.TrimSpace(getNodeText(node, file.Content)), "{}")
}

func qualityIsConsoleLogCall(node *gotreesitter.Node, file *ParsedFile) bool {
	lang := file.Tree.Language()
	if node == nil || (node.Type(lang) != "call_expression" && node.Type(lang) != "call") {
		return false
	}
	funcNode := node.ChildByFieldName("function", lang)
	if funcNode == nil {
		return false
	}
	name := strings.ToLower(getNodeText(funcNode, file.Content))
	return strings.Contains(name, "console.log") || strings.Contains(name, "console.debug") || strings.Contains(name, "console.warn")
}
