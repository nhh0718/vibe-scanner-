package ast

import (
	"path/filepath"
	"strings"

	"github.com/odvcencio/gotreesitter/grammars"
)

// GetLanguage lấy language parser cho file extension
func GetLanguage(path string) (string, *grammars.LangEntry) {
	ext := strings.ToLower(filepath.Ext(path))
	baseName := strings.ToLower(filepath.Base(path))

	lang := ""
	switch ext {
	case ".js", ".mjs", ".cjs":
		lang = "javascript"
	case ".ts":
		lang = "typescript"
	case ".jsx":
		lang = "jsx"
	case ".tsx":
		lang = "tsx"
	case ".py":
		lang = "python"
	case ".go":
		lang = "go"
	case ".php":
		lang = "php"
	case ".rb":
		lang = "ruby"
	case ".rs":
		lang = "rust"
	case ".java":
		lang = "java"
	case ".html", ".htm":
		lang = "html"
	default:
		// Special files matched by base name
		if baseName == ".gitignore" || baseName == ".env" {
			return "*", nil
		}
		return "", nil
	}

	entry := grammars.DetectLanguage(path)
	return lang, entry
}

// SupportedLanguages trả về danh sách ngôn ngữ được hỗ trợ
func SupportedLanguages() []string {
	return []string{"javascript", "typescript", "jsx", "tsx", "python", "go", "php", "java", "ruby", "rust"}
}
