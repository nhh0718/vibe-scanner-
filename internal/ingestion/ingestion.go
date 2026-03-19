package ingestion

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// languageExtensions ánh xạ extension sang ngôn ngữ
var languageExtensions = map[string]string{
	".go":    "go",
	".js":    "javascript",
	".jsx":   "javascript",
	".ts":    "typescript",
	".tsx":   "typescript",
	".py":    "python",
	".rb":    "ruby",
	".php":   "php",
	".java":  "java",
	".cs":    "csharp",
	".cpp":   "cpp",
	".c":     "c",
	".h":     "c",
	".hpp":   "cpp",
	".swift": "swift",
	".rs":    "rust",
	".kt":    "kotlin",
	".scala": "scala",
	".r":     "r",
	".sh":    "shell",
	".html":  "html",
	".css":   "css",
	".scss":  "scss",
	".sass":  "sass",
	".less":  "less",
	".sql":   "sql",
	".json":  "json",
	".xml":   "xml",
	".yaml":  "yaml",
	".yml":   "yaml",
	".md":    "markdown",
	".dockerfile": "dockerfile",
}

// ignoredDirectories các thư mục bị bỏ qua
var ignoredDirectories = []string{
	".git", ".svn", ".hg", ".bzr",
	"node_modules", "vendor", "dist", "build",
	"target", "out", "bin", "obj",
	".next", ".nuxt", ".vuepress", ".docusaurus",
	"coverage", ".nyc_output", "__pycache__",
	".venv", "venv", "env", "virtualenv",
	".idea", ".vscode", ".vs",
	".cache", ".temp", ".tmp",
}

// ignoredFiles các file bị bỏ qua
var ignoredFiles = []string{
	".DS_Store", "Thumbs.db",
	"package-lock.json", "yarn.lock", "pnpm-lock.yaml",
	".env", ".env.local", ".env.production", ".env.development",
}

// AnalyzeProject phân tích project và trả về thông tin
func AnalyzeProject(path string) (*models.ProjectInfo, error) {
	info := &models.ProjectInfo{
		Path:      path,
		Name:      filepath.Base(path),
		Languages: []string{},
	}

	languageSet := make(map[string]bool)
	linesOfCode := 0

	err := filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if fileInfo.IsDir() {
			if shouldIgnoreDir(fileInfo.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be ignored
		if shouldIgnoreFile(fileInfo.Name()) {
			return nil
		}

		// Count as scanned file
		info.FilesScanned++

		// Detect language
		ext := strings.ToLower(filepath.Ext(filePath))
		if lang, ok := languageExtensions[ext]; ok {
			languageSet[lang] = true
		}

		// Count lines
		lines, err := countLines(filePath)
		if err == nil {
			linesOfCode += lines
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert language set to slice
	for lang := range languageSet {
		info.Languages = append(info.Languages, lang)
	}

	info.LinesOfCode = linesOfCode

	return info, nil
}

// shouldIgnoreDir kiểm tra xem directory có nên bị ignore không
func shouldIgnoreDir(name string) bool {
	for _, ignored := range ignoredDirectories {
		if strings.EqualFold(name, ignored) {
			return true
		}
	}
	// Ignore hidden directories
	if strings.HasPrefix(name, ".") && len(name) > 1 {
		return true
	}
	return false
}

// shouldIgnoreFile kiểm tra xem file có nên bị ignore không
func shouldIgnoreFile(name string) bool {
	for _, ignored := range ignoredFiles {
		if strings.EqualFold(name, ignored) {
			return true
		}
	}
	// Ignore hidden files (except .env which is interesting for security)
	if strings.HasPrefix(name, ".") && !strings.HasPrefix(name, ".env") {
		return true
	}
	return false
}

// countLines đếm số dòng trong file
func countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}

	return count, scanner.Err()
}

// DetectLanguage phát hiện ngôn ngữ từ file path
func DetectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if lang, ok := languageExtensions[ext]; ok {
		return lang
	}
	
	// Special cases
	base := strings.ToLower(filepath.Base(filePath))
	if base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") {
		return "dockerfile"
	}
	if base == "makefile" {
		return "makefile"
	}
	
	return "unknown"
}

// GetProjectFiles trả về danh sách tất cả files trong project
func GetProjectFiles(path string) ([]string, error) {
	var files []string

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if shouldIgnoreDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldIgnoreFile(info.Name()) {
			return nil
		}

		files = append(files, filePath)
		return nil
	})

	return files, err
}
