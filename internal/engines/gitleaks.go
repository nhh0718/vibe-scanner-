package engines

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// RunGitleaksWithStatus chạy Gitleaks và trả về status rõ ràng
func RunGitleaksWithStatus(path string) ([]models.Finding, string) {
	gitleaksBin, err := getGitleaksBinary()
	if err != nil {
		return []models.Finding{}, fmt.Sprintf("⚠️  Gitleaks: Không khả dụng (%v)", err)
	}

	// Create temp file for JSON report (gitleaks -f json writes to file, not stdout)
	tmpFile, err := os.CreateTemp("", "gitleaks-report-*.json")
	if err != nil {
		return []models.Finding{}, fmt.Sprintf("❌ Gitleaks: Không tạo được temp file - %v", err)
	}
	tmpFile.Close()
	reportPath := tmpFile.Name()
	defer os.Remove(reportPath)

	cmd := exec.Command(gitleaksBin,
		"detect",
		"--source", path,
		"--no-banner",
		"-f", "json",
		"--report-path", reportPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	// Gitleaks exits with code 1 when findings are detected (normal behavior)
	if err != nil && cmd.ProcessState.ExitCode() != 1 {
		return []models.Finding{}, fmt.Sprintf("❌ Gitleaks: Lỗi chạy - %v", err)
	}

	// Read the report file
	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		return []models.Finding{}, fmt.Sprintf("❌ Gitleaks: Không đọc được report - %v", err)
	}

	findings, err := parseGitleaksOutput(reportData, path)
	if err != nil {
		return []models.Finding{}, fmt.Sprintf("❌ Gitleaks: Lỗi parse output - %v", err)
	}

	if len(findings) == 0 {
		return findings, "✅ Gitleaks: Không phát hiện secrets"
	}
	return findings, fmt.Sprintf("🔐 Gitleaks: Phát hiện %d secrets", len(findings))
}

// parseGitleaksOutput parse JSON output từ Gitleaks
func parseGitleaksOutput(output []byte, basePath string) ([]models.Finding, error) {
	type gitleaksFinding struct {
		RuleID      string  `json:"RuleID"`
		Description string  `json:"Description"`
		StartLine   int     `json:"StartLine"`
		EndLine     int     `json:"EndLine"`
		StartColumn int     `json:"StartColumn"`
		EndColumn   int     `json:"EndColumn"`
		Match       string  `json:"Match"`
		Secret      string  `json:"Secret"`
		File        string  `json:"File"`
		SymlinkFile string  `json:"SymlinkFile"`
		Commit      string  `json:"Commit"`
		Author      string  `json:"Author"`
		Email       string  `json:"Email"`
		Date        string  `json:"Date"`
		Message     string  `json:"Message"`
		Fingerprint string  `json:"Fingerprint"`
		Entropy     float64 `json:"Entropy"`
	}

	var results []gitleaksFinding

	// Try parsing entire output as JSON array first (gitleaks v8 format)
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &results); err == nil {
			// Successfully parsed as array
			goto convert
		}
	}

	// Fallback: line-by-line parsing
	{
		scanner := bufio.NewScanner(bytes.NewReader(output))
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}
			var finding gitleaksFinding
			if err := json.Unmarshal(line, &finding); err == nil {
				results = append(results, finding)
			}
		}
	}

convert:

	var findings []models.Finding
	for _, r := range results {
		// Make path relative
		relPath, _ := filepath.Rel(basePath, r.File)
		if relPath == "" {
			relPath = r.File
		}

		// Determine severity based on secret type
		severity := models.High
		if strings.Contains(strings.ToLower(r.RuleID), "test") ||
		   strings.Contains(strings.ToLower(r.Description), "test") {
			severity = models.Medium
		}

		message := fmt.Sprintf("Phát hiện %s", r.Description)
		if r.Secret != "" {
			// Mask the secret
			masked := maskSecret(r.Secret)
			message += fmt.Sprintf(": `%s`", masked)
		}

		findings = append(findings, models.Finding{
			RuleID:      r.RuleID,
			Severity:    severity,
			Category:    models.Secrets,
			Subcategory: determineSecretType(r.RuleID, r.Description),
			Title:       r.Description,
			Message:     message,
			File:        relPath,
			Line:        r.StartLine,
			Column:      r.StartColumn,
			CodeSnippet: r.Match,
			Engine:      "gitleaks",
			Timestamp:   time.Now(),
		})
	}

	return findings, nil
}

// getGitleaksBinary trả về đường dẫn đến gitleaks binary, tự download nếu chưa có
func getGitleaksBinary() (string, error) {
	// First, check if gitleaks is in PATH
	if path, err := exec.LookPath("gitleaks"); err == nil {
		return path, nil
	}

	// Check in vibescanner bin directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		gitleaksBin := filepath.Join(homeDir, ".vibescanner", "bin", "gitleaks")
		if runtime.GOOS == "windows" {
			gitleaksBin += ".exe"
		}
		if _, err := os.Stat(gitleaksBin); err == nil {
			return gitleaksBin, nil
		}
	}

	// Auto-download if not found
	fmt.Println("⬇️  Đang tải Gitleaks...")
	if err := downloadGitleaks(); err != nil {
		return "", fmt.Errorf("không thể tải Gitleaks: %w", err)
	}

	// Return the downloaded binary path
	gitleaksBin := filepath.Join(homeDir, ".vibescanner", "bin", "gitleaks")
	if runtime.GOOS == "windows" {
		gitleaksBin += ".exe"
	}
	return gitleaksBin, nil
}

// downloadGitleaks tải và cài đặt Gitleaks binary từ GitHub releases
func downloadGitleaks() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("không tìm thấy thư mục home: %w", err)
	}

	binDir := filepath.Join(homeDir, ".vibescanner", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục bin: %w", err)
	}

	// Determine correct download URL based on OS/Arch
	var zipName string
	var binaryName string

	switch runtime.GOOS {
	case "windows":
		zipName = "gitleaks_8.18.2_windows_x64.zip"
		binaryName = "gitleaks.exe"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			zipName = "gitleaks_8.18.2_darwin_arm64.zip"
		} else {
			zipName = "gitleaks_8.18.2_darwin_x64.zip"
		}
		binaryName = "gitleaks"
	case "linux":
		if runtime.GOARCH == "arm64" {
			zipName = "gitleaks_8.18.2_linux_arm64.zip"
		} else {
			zipName = "gitleaks_8.18.2_linux_x64.zip"
		}
		binaryName = "gitleaks"
	default:
		return fmt.Errorf("OS không được hỗ trợ: %s", runtime.GOOS)
	}

	url := fmt.Sprintf("https://github.com/gitleaks/gitleaks/releases/download/v8.18.2/%s", zipName)
	zipPath := filepath.Join(binDir, zipName)

	fmt.Printf("📥 Đang tải Gitleaks cho %s/%s...\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("📥 Tải từ: %s\n", url)

	// Download the zip file
	var downloadCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		psCmd := fmt.Sprintf("Invoke-WebRequest -Uri '%s' -OutFile '%s' -UseBasicParsing", url, zipPath)
		downloadCmd = exec.Command("powershell", "-Command", psCmd)
	} else {
		downloadCmd = exec.Command("curl", "-L", "-o", zipPath, url)
	}

	if output, err := downloadCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("lỗi tải: %w\n%s", err, string(output))
	}

	fmt.Println("📦 Đang giải nén Gitleaks...")

	// Extract the zip
	var extractCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		psCmd := fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s' -Force", zipPath, binDir)
		extractCmd = exec.Command("powershell", "-Command", psCmd)
	} else {
		extractCmd = exec.Command("unzip", "-o", "-q", zipPath, "-d", binDir)
	}

	if output, err := extractCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("lỗi giải nén: %w\n%s", err, string(output))
	}

	// Clean up zip
	os.Remove(zipPath)

	// Rename to standard name if needed
	extractedPath := filepath.Join(binDir, binaryName)
	finalPath := filepath.Join(binDir, "gitleaks")
	if runtime.GOOS == "windows" {
		finalPath += ".exe"
	}

	if extractedPath != finalPath {
		if _, err := os.Stat(extractedPath); err == nil {
			os.Rename(extractedPath, finalPath)
		}
	}

	// Make executable
	os.Chmod(finalPath, 0755)

	fmt.Println("✅ Gitleaks đã được cài đặt")
	return nil
}

// maskSecret mask secret với asterisks
func maskSecret(secret string) string {
	if len(secret) <= 4 {
		return "****"
	}
	return secret[:2] + strings.Repeat("*", len(secret)-4) + secret[len(secret)-2:]
}

// determineSecretType xác định loại secret
func determineSecretType(ruleID, description string) string {
	lower := strings.ToLower(ruleID + " " + description)

	if strings.Contains(lower, "aws") {
		return "aws_credential"
	}
	if strings.Contains(lower, "api") || strings.Contains(lower, "key") {
		return "api_key"
	}
	if strings.Contains(lower, "token") {
		return "token"
	}
	if strings.Contains(lower, "password") || strings.Contains(lower, "pwd") {
		return "password"
	}
	if strings.Contains(lower, "private") {
		return "private_key"
	}
	if strings.Contains(lower, "database") || strings.Contains(lower, "db") || strings.Contains(lower, "connection") {
		return "database_connection"
	}
	return "generic_secret"
}
