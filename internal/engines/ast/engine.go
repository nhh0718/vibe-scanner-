package ast

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/nhh0718/vibe-scanner-/internal/engines/rules"
	"github.com/odvcencio/gotreesitter"
)

// Engine chạy AST-based analysis sử dụng gotreesitter
type Engine struct {
	rules []rules.Rule
}

// NewEngine tạo engine mới với danh sách rules
func NewEngine(rules []rules.Rule) *Engine {
	return &Engine{
		rules: rules,
	}
}

// ScanFile quét một file và trả về findings
func (e *Engine) ScanFile(path string) ([]rules.Finding, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Detect language using gotreesitter grammars
	lang, entry := GetLanguage(path)
	if lang == "" {
		return nil, nil // Không support ngôn ngữ này
	}

	parsedFile := &rules.ParsedFile{
		Path:     path,
		Content:  content,
		Language: lang,
		Tree:     nil,
	}

	// Parse với gotreesitter nếu có grammar
	if entry != nil {
		parser := gotreesitter.NewParser(entry.Language())
		tree, err := parser.Parse(content)
		if err == nil {
			parsedFile.Tree = tree
		}
	}

	// Chạy tất cả rules áp dụng cho ngôn ngữ này
	var findings []rules.Finding
	for _, rule := range e.rules {
		if !rules.RuleApplies(rule, lang) {
			continue
		}
		findings = append(findings, rule.Check(parsedFile)...)
	}

	return findings, nil
}

// ClearCache không cần thiết - parser tạo mới mỗi lần
func (e *Engine) ClearCache() {}

// ScanProject quét toàn bộ project
func ScanProject(projectPath string, ignorePatterns []string) ([]rules.Finding, error) {
	files, err := walkFiles(projectPath, append(defaultIgnore, ignorePatterns...))
	if err != nil {
		return nil, err
	}

	astEngine := NewEngine(rules.AllRules())

	var (
		mu       sync.Mutex
		findings []rules.Finding
		wg       sync.WaitGroup
		sem      = make(chan struct{}, 8) // max 8 files song song
	)

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fileFindings, _ := astEngine.ScanFile(f)

			mu.Lock()
			findings = append(findings, fileFindings...)
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	return deduplicate(findings), nil
}

// Các file/folder luôn bỏ qua
var defaultIgnore = []string{
	"node_modules", ".git", ".svn", "dist", "build",
	"__pycache__", ".venv", "vendor", ".next", ".nuxt",
	"target", "bin", "obj", ".idea", ".vscode",
}

func walkFiles(root string, ignorePatterns []string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())

		// Check ignore patterns (match exact directory/file name, not substrings)
		shouldIgnore := false
		for _, pattern := range ignorePatterns {
			if entry.Name() == pattern {
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}

		if entry.IsDir() {
			subFiles, err := walkFiles(path, ignorePatterns)
			if err != nil {
				continue
			}
			files = append(files, subFiles...)
		} else {
			// Only include code files
			if lang, _ := GetLanguage(path); lang != "" {
				files = append(files, path)
			}
		}
	}

	return files, nil
}

// deduplicate loại bỏ findings trùng lặp
func deduplicate(findings []rules.Finding) []rules.Finding {
	seen := make(map[string]bool)
	var result []rules.Finding
	for _, f := range findings {
		key := f.RuleID + ":" + f.File + ":" + string(rune(f.Line))
		if !seen[key] {
			seen[key] = true
			result = append(result, f)
		}
	}
	return result
}
