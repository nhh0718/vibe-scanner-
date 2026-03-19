package engines

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// ComplexityAnalyzer phân tích độ phức tạp code
type ComplexityAnalyzer struct {
	maxFunctionLines  int
	maxFileLines      int
	maxNestingDepth   int
	maxParams         int
}

// NewComplexityAnalyzer tạo analyzer mới
func NewComplexityAnalyzer() *ComplexityAnalyzer {
	return &ComplexityAnalyzer{
		maxFunctionLines: 50,
		maxFileLines:     300,
		maxNestingDepth:  4,
		maxParams:        5,
	}
}

// RunComplexity chạy phân tích complexity
func RunComplexity(path string) ([]models.Finding, error) {
	analyzer := NewComplexityAnalyzer()
	return analyzer.Analyze(path)
}

// Analyze quét toàn bộ project
func (ca *ComplexityAnalyzer) Analyze(path string) ([]models.Finding, error) {
	var findings []models.Finding

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.IsDir() {
			// Skip common directories
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || 
			   name == "vendor" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only analyze code files
		if !isCodeFile(filePath) {
			return nil
		}

		fileFindings, err := ca.analyzeFile(filePath, path)
		if err != nil {
			return nil // Continue on error
		}
		findings = append(findings, fileFindings...)

		return nil
	})

	return findings, err
}

// analyzeFile phân tích một file
func (ca *ComplexityAnalyzer) analyzeFile(filePath, basePath string) ([]models.Finding, error) {
	var findings []models.Finding

	content, err := os.ReadFile(filePath)
	if err != nil {
		return findings, err
	}

	lines := strings.Split(string(content), "\n")
	relPath, _ := filepath.Rel(basePath, filePath)
	if relPath == "" {
		relPath = filePath
	}

	// Check file length
	if len(lines) > ca.maxFileLines {
		findings = append(findings, models.Finding{
			RuleID:      "complexity-file-too-long",
			Severity:    models.Medium,
			Category:    models.Quality,
			Subcategory: "file_size",
			Title:       "File quá dài",
			Message:     fmt.Sprintf("File có %d dòng (giới hạn: %d). Xem xét tách thành nhiều file nhỏ hơn.", len(lines), ca.maxFileLines),
			File:        relPath,
			Line:        1,
			CodeSnippet: "",
			Engine:      "complexity",
			Timestamp:   time.Now(),
		})
	}

	// Analyze for common issues
	findings = append(findings, ca.findMagicNumbers(lines, relPath)...)
	findings = append(findings, ca.findConsoleLogs(lines, relPath)...)
	findings = append(findings, ca.findTODOs(lines, relPath)...)
	findings = append(findings, ca.findEmptyCatches(lines, relPath)...)
	findings = append(findings, ca.findLongLines(lines, relPath)...)

	return findings, nil
}

// isCodeFile kiểm tra xem file có phải code file không
func isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	codeExts := []string{
		".go", ".js", ".ts", ".jsx", ".tsx",
		".py", ".rb", ".php", ".java", ".cs",
		".cpp", ".c", ".h", ".hpp", ".swift",
		".rs", ".kt", ".scala", ".r",
	}
	for _, codeExt := range codeExts {
		if ext == codeExt {
			return true
		}
	}
	return false
}

// findMagicNumbers tìm magic numbers
func (ca *ComplexityAnalyzer) findMagicNumbers(lines []string, filePath string) []models.Finding {
	var findings []models.Finding
	magicNumberPattern := regexp.MustCompile(`(?i)(if|while|for|return|var|let|const|:=|=)\s+[^;]*\b(\d{3,})\b`)

	for i, line := range lines {
		if matches := magicNumberPattern.FindStringSubmatch(line); matches != nil {
			num := matches[2]
			// Skip common patterns
			if num == "200" || num == "201" || num == "404" || num == "500" || num == "0" || num == "1" || num == "-1" || num == "100" || num == "1000" || num == "1024" || num == "3000" || num == "8080" {
				continue
			}

			findings = append(findings, models.Finding{
				RuleID:      "quality-magic-number",
				Severity:    models.Low,
				Category:    models.Quality,
				Subcategory: "magic_number",
				Title:       "Magic number detected",
				Message:     fmt.Sprintf("Số %s nên được định nghĩa thành constant với tên có ý nghĩa.", num),
				File:        filePath,
				Line:        i + 1,
				CodeSnippet: strings.TrimSpace(line),
				Engine:      "complexity",
				Timestamp:   time.Now(),
			})
		}
	}

	return findings
}

// findConsoleLogs tìm console.log/debug statements
func (ca *ComplexityAnalyzer) findConsoleLogs(lines []string, filePath string) []models.Finding {
	var findings []models.Finding
	consolePattern := regexp.MustCompile(`(?i)console\.(log|debug|warn|error)\s*\(`)

	for i, line := range lines {
		if consolePattern.MatchString(line) {
			findings = append(findings, models.Finding{
				RuleID:      "quality-console-log",
				Severity:    models.Low,
				Category:    models.Quality,
				Subcategory: "debug_code",
				Title:       "Console statement trong production code",
				Message:     "Console statements nên được xóa hoặc thay bằng proper logging framework trong production.",
				File:        filePath,
				Line:        i + 1,
				CodeSnippet: strings.TrimSpace(line),
				Engine:      "complexity",
				Timestamp:   time.Now(),
			})
		}
	}

	return findings
}

// findTODOs tìm TODO/FIXME comments
func (ca *ComplexityAnalyzer) findTODOs(lines []string, filePath string) []models.Finding {
	var findings []models.Finding
	todoPattern := regexp.MustCompile(`(?i)(TODO|FIXME|HACK|XXX|BUG)`)

	for i, line := range lines {
		matches := todoPattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			keyword := matches[0]
			severity := models.Low
			if keyword == "FIXME" || keyword == "BUG" {
				severity = models.Medium
			}

			findings = append(findings, models.Finding{
				RuleID:      "quality-todo",
				Severity:    severity,
				Category:    models.Quality,
				Subcategory: "todo",
				Title:       fmt.Sprintf("%s comment found", keyword),
				Message:     fmt.Sprintf("Comment chứa %s - cần được xử lý trước khi release.", keyword),
				File:        filePath,
				Line:        i + 1,
				CodeSnippet: strings.TrimSpace(line),
				Engine:      "complexity",
				Timestamp:   time.Now(),
			})
		}
	}

	return findings
}

// findEmptyCatches tìm empty catch blocks
func (ca *ComplexityAnalyzer) findEmptyCatches(lines []string, filePath string) []models.Finding {
	var findings []models.Finding
	emptyCatchPattern := regexp.MustCompile(`(?i)catch\s*\([^)]*\)\s*\{\s*\}`)

	for i, line := range lines {
		if emptyCatchPattern.MatchString(line) {
			findings = append(findings, models.Finding{
				RuleID:      "quality-empty-catch",
				Severity:    models.High,
				Category:    models.Quality,
				Subcategory: "error_handling",
				Title:       "Empty catch block",
				Message:     "Catch block rỗng nuốt lỗi im lặng. Nên log lỗi hoặc throw lại.",
				File:        filePath,
				Line:        i + 1,
				CodeSnippet: strings.TrimSpace(line),
				Engine:      "complexity",
				Timestamp:   time.Now(),
			})
		}
	}

	return findings
}

// findLongLines tìm dòng code quá dài
func (ca *ComplexityAnalyzer) findLongLines(lines []string, filePath string) []models.Finding {
	var findings []models.Finding
	maxLineLength := 120

	for i, line := range lines {
		if len(line) > maxLineLength {
			findings = append(findings, models.Finding{
				RuleID:      "quality-long-line",
				Severity:    models.Info,
				Category:    models.Quality,
				Subcategory: "formatting",
				Title:       "Dòng code quá dài",
				Message:     fmt.Sprintf("Dòng có %d ký tự (giới hạn: %d). Nên xuống dòng để dễ đọc.", len(line), maxLineLength),
				File:        filePath,
				Line:        i + 1,
				CodeSnippet: strings.TrimSpace(line),
				Engine:      "complexity",
				Timestamp:   time.Now(),
			})
		}
	}

	return findings
}
