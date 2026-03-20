package engines

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// RunSemgrep chạy Semgrep và trả về findings
func RunSemgrep(path string) ([]models.Finding, error) {
	// External Semgrep is optional; native security engine handles core coverage.
	semgrepBin, err := getSemgrepBinary()
	if err != nil {
		return []models.Finding{}, nil
	}

	// Run semgrep with JSON output
	cmd := exec.Command(semgrepBin,
		"--config", "p/owasp-top-ten",
		"--config", "p/sql-injection",
		"--config", "p/javascript",
		"--config", "p/secrets",
		"--json",
		"--quiet",
		"--max-target-bytes", "1048576", // 1MB limit per file
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		// Exit code 1 means findings detected, not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Continue processing
		} else {
			return nil, fmt.Errorf("semgrep lỗi: %w", err)
		}
	}

	return parseSemgrepOutput(output, path)
}

// parseSemgrepOutput parse JSON output từ Semgrep
func parseSemgrepOutput(output []byte, basePath string) ([]models.Finding, error) {
	var result struct {
		Results []struct {
			CheckID string `json:"check_id"`
			Path    string `json:"path"`
			Start   struct {
				Line   int `json:"line"`
				Column int `json:"col"`
			} `json:"start"`
			Extra struct {
				Message   string `json:"message"`
				Severity  string `json:"severity"`
				Code      string `json:"lines"`
				Metadata  map[string]interface{} `json:"metadata"`
			} `json:"extra"`
		} `json:"results"`
		Errors []interface{} `json:"errors"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("lỗi parse semgrep output: %w", err)
	}

	var findings []models.Finding
	for _, r := range result.Results {
		category := determineCategory(r.CheckID, r.Extra.Metadata)
		severity := parseSeverity(r.Extra.Severity)

		// Make path relative
		relPath, _ := filepath.Rel(basePath, r.Path)
		if relPath == "" {
			relPath = r.Path
		}

		findings = append(findings, models.Finding{
			RuleID:      r.CheckID,
			Severity:    severity,
			Category:    category,
			Title:       extractTitle(r.CheckID, r.Extra.Metadata),
			Message:     r.Extra.Message,
			File:        relPath,
			Line:        r.Start.Line,
			Column:      r.Start.Column,
			CodeSnippet: strings.TrimSpace(r.Extra.Code),
			Engine:      "semgrep",
			Timestamp:   time.Now(),
		})
	}

	return findings, nil
}

// getSemgrepBinary trả về đường dẫn đến semgrep binary, tự cài qua pip nếu chưa có
func getSemgrepBinary() (string, error) {
	// First, check if semgrep is in PATH
	if path, err := exec.LookPath("semgrep"); err == nil {
		return path, nil
	}

	// Try to install via pip
	fmt.Println("⬇️  Đang cài đặt Semgrep qua pip...")
	if err := installSemgrepViaPip(); err != nil {
		return "", fmt.Errorf("không thể cài Semgrep: %w", err)
	}

	// Try again after install
	if path, err := exec.LookPath("semgrep"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("đã cài Semgrep nhưng không tìm thấy trong PATH")
}

// installSemgrepViaPip cài Semgrep qua pip
func installSemgrepViaPip() error {
	// Tìm python
	pythonCmd := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err := exec.LookPath("python"); err == nil {
			pythonCmd = "python"
		} else {
			return fmt.Errorf("không tìm thấy Python. Vui lòng cài Python từ https://python.org")
		}
	}

	fmt.Println("📦 Đang cài Semgrep qua pip (có thể mất vài phút)...")
	
	cmd := exec.Command(pythonCmd, "-m", "pip", "install", "--user", "semgrep")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lỗi cài đặt: %w\n%s", err, string(output))
	}

	fmt.Println("✅ Semgrep đã được cài đặt")
	return nil
}

// determineCategory xác định category dựa trên rule ID và metadata
func determineCategory(ruleID string, metadata map[string]interface{}) models.Category {
	ruleLower := strings.ToLower(ruleID)

	// Check metadata cwe first
	if cwe, ok := metadata["cwe"]; ok {
		cweStr := fmt.Sprintf("%v", cwe)
		if strings.Contains(cweStr, "798") || strings.Contains(cweStr, " hardcoded") {
			return models.Secrets
		}
	}

	// Check rule ID patterns
	if strings.Contains(ruleLower, "secret") || strings.Contains(ruleLower, "password") ||
	   strings.Contains(ruleLower, "api-key") || strings.Contains(ruleLower, "token") {
		return models.Secrets
	}

	if strings.Contains(ruleLower, "sql") || strings.Contains(ruleLower, "inject") ||
	   strings.Contains(ruleLower, "xss") || strings.Contains(ruleLower, "cors") ||
	   strings.Contains(ruleLower, "jwt") || strings.Contains(ruleLower, "auth") {
		return models.Security
	}

	if strings.Contains(ruleLower, "complexity") || strings.Contains(ruleLower, "deadcode") ||
	   strings.Contains(ruleLower, "duplicate") {
		return models.Quality
	}

	// Default to security for OWASP rules
	if strings.HasPrefix(ruleLower, "owasp") {
		return models.Security
	}

	return models.Security
}

// parseSeverity chuyển đổi severity string sang model
func parseSeverity(sev string) models.Severity {
	switch strings.ToLower(sev) {
	case "error", "critical":
		return models.Critical
	case "warning", "high":
		return models.High
	case "info", "medium":
		return models.Medium
	case "low":
		return models.Low
	default:
		return models.Medium
	}
}

// extractTitle trích xuất title từ rule ID hoặc metadata
func extractTitle(ruleID string, metadata map[string]interface{}) string {
	if message, ok := metadata["message"]; ok {
		return fmt.Sprintf("%v", message)
	}

	// Convert rule ID to readable title
	parts := strings.Split(ruleID, ".")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		// Convert camelCase/snake_case to Title Case
		last = strings.ReplaceAll(last, "-", " ")
		last = strings.ReplaceAll(last, "_", " ")
		return strings.Title(last)
	}

	return ruleID
}

// getPlatformSuffix trả về suffix cho binary theo platform
func getPlatformSuffix() string {
	switch runtime.GOOS {
	case "windows":
		return "windows.exe"
	case "darwin":
		return "macos"
	default:
		return "linux"
	}
}
