package engines

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/nhh0718/vibe-scanner-/internal/models"
)

var (
	scannableExtensions = map[string]bool{
		".js": true, ".jsx": true, ".ts": true, ".tsx": true,
		".go": true, ".py": true, ".php": true, ".java": true,
		".rb": true, ".cs": true, ".sql": true, ".env": true,
	}

	sqlConcatPattern    = regexp.MustCompile(`(?i)(select|insert|update|delete)[^\n]{0,180}(\+|fmt\.s?printf|%s|\$\{|f"|f')`)
	sqlQueryCallPattern = regexp.MustCompile(`(?i)(query|exec|raw|queryrow)\s*\([^\n]{0,220}(\+|fmt\.s?printf|%s|\$\{|f"|f')`)
	secretAssignPattern = regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|passwd|jwt[_-]?secret|access[_-]?key|private[_-]?key)\s*[:=]\s*["'][^"'\n]{8,}["']`)
	corsWildcardPattern = regexp.MustCompile(`(?i)(access-control-allow-origin[^\n]*\*|origin\s*[:=]\s*["']\*["']|allow_origins\s*=\s*\[[^\]]*["']\*["'])`)
	jwtWeakSecretPattern = regexp.MustCompile(`(?i)(jwt[_-]?secret|secret[_-]?key)\s*[:=]\s*["'](secret|changeme|default|test|123456|admin)["']`)
	privateKeyPattern = regexp.MustCompile(`(?s)-----BEGIN (RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----`)
)

func RunNativeSecurity(path string) ([]models.Finding, error) {
	var findings []models.Finding

	err := filepath.Walk(path, func(current string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := strings.ToLower(info.Name())
			if name == ".git" || name == "node_modules" || name == "vendor" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > 1024*1024 {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(current))
		if !scannableExtensions[ext] && filepath.Base(current) != ".env" {
			return nil
		}

		content, readErr := os.ReadFile(current)
		if readErr != nil {
			return nil
		}
		if bytes.IndexByte(content, 0) >= 0 {
			return nil
		}

		relPath, _ := filepath.Rel(path, current)
		if relPath == "" {
			relPath = current
		}

		fileFindings := scanNativeSecurityFile(relPath, content)
		findings = append(findings, fileFindings...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return findings, nil
}

func scanNativeSecurityFile(relPath string, content []byte) []models.Finding {
	var findings []models.Finding
	seen := map[string]bool{}

	if loc := privateKeyPattern.FindIndex(content); loc != nil {
		findings = append(findings, models.Finding{
			RuleID:      "native.private-key.exposed",
			Severity:    models.Critical,
			Category:    models.Secrets,
			Subcategory: "private_key",
			Title:       "Lộ private key trong source",
			Message:     "Phát hiện private key được lưu trực tiếp trong mã nguồn hoặc file cấu hình.",
			File:        relPath,
			Line:        1,
			Column:      1,
			CodeSnippet: firstSnippet(string(content), 160),
			Engine:      "native-security",
			Timestamp:   time.Now(),
		})
		seen["native.private-key.exposed:1"] = true
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		add := func(ruleID string, severity models.Severity, category models.Category, subcategory, title, message string) {
			key := fmt.Sprintf("%s:%d:%s", ruleID, lineNo, trimmed)
			if seen[key] {
				return
			}
			seen[key] = true
			findings = append(findings, models.Finding{
				RuleID:      ruleID,
				Severity:    severity,
				Category:    category,
				Subcategory: subcategory,
				Title:       title,
				Message:     message,
				File:        relPath,
				Line:        lineNo,
				Column:      1,
				CodeSnippet: trimmed,
				Engine:      "native-security",
				Timestamp:   time.Now(),
			})
		}

		if sqlConcatPattern.MatchString(line) || sqlQueryCallPattern.MatchString(line) {
			add(
				"native.sql-injection.dynamic-query",
				models.High,
				models.Security,
				"sql_injection",
				"Query SQL được ghép động",
				"Phát hiện query SQL có dấu hiệu ghép chuỗi hoặc format trực tiếp với input, có nguy cơ SQL injection.",
			)
		}

		if secretAssignPattern.MatchString(line) {
			add(
				"native.secret.hardcoded",
				models.High,
				models.Secrets,
				"hardcoded_secret",
				"Secret được hardcode",
				"Phát hiện giá trị nhạy cảm được hardcode trực tiếp trong mã nguồn hoặc file cấu hình.",
			)
		}

		if jwtWeakSecretPattern.MatchString(line) {
			add(
				"native.jwt.weak-secret",
				models.High,
				models.Security,
				"weak_secret",
				"JWT secret yếu hoặc mặc định",
				"Secret dùng cho JWT đang là giá trị mặc định/yếu, rất dễ bị đoán hoặc brute-force.",
			)
		}

		if corsWildcardPattern.MatchString(line) {
			add(
				"native.cors.wildcard",
				models.Medium,
				models.Security,
				"cors",
				"CORS mở wildcard",
				"Cấu hình CORS cho phép mọi origin, có thể làm lộ API cho các domain không tin cậy.",
			)
		}
	}

	return findings
}

func firstSnippet(content string, maxLen int) string {
	content = strings.TrimSpace(content)
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen]
}
