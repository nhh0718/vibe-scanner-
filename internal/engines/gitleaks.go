package engines

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/vibescanner/vibescanner/internal/models"
)

// RunGitleaks chạy Gitleaks để detect secrets
func RunGitleaks(path string) ([]models.Finding, error) {
	// Check if gitleaks is available
	gitleaksBin, err := getGitleaksBinary()
	if err != nil {
		// Gitleaks is optional, return empty if not available
		return []models.Finding{}, nil
	}

	// Run gitleaks with JSON output
	cmd := exec.Command(gitleaksBin,
		"detect",
		"--source", path,
		"--no-banner",
		"--verbose",
		"-f", "json",
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	// Gitleaks exits with code 1 when findings are detected
	if err != nil && cmd.ProcessState.ExitCode() != 1 {
		return nil, fmt.Errorf("gitleaks lỗi: %w (stderr: %s)", err, stderr.String())
	}

	return parseGitleaksOutput(out.Bytes(), path)
}

// parseGitleaksOutput parse JSON output từ Gitleaks
func parseGitleaksOutput(output []byte, basePath string) ([]models.Finding, error) {
	var results []struct {
		RuleID      string `json:"RuleID"`
		Description string `json:"Description"`
		StartLine   int    `json:"StartLine"`
		EndLine     int    `json:"EndLine"`
		StartColumn int    `json:"StartColumn"`
		EndColumn   int    `json:"EndColumn"`
		Match       string `json:"Match"`
		Secret      string `json:"Secret"`
		File        string `json:"File"`
		SymlinkFile string `json:"SymlinkFile"`
		Commit      string `json:"Commit"`
		Author      string `json:"Author"`
		Email       string `json:"Email"`
		Date        string `json:"Date"`
		Message     string `json:"Message"`
		Fingerprint string `json:"Fingerprint"`
	}

	// Gitleaks may output multiple JSON objects, one per line
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var finding struct {
			RuleID      string `json:"RuleID"`
			Description string `json:"Description"`
			StartLine   int    `json:"StartLine"`
			EndLine     int    `json:"EndLine"`
			StartColumn int    `json:"StartColumn"`
			EndColumn   int    `json:"EndColumn"`
			Match       string `json:"Match"`
			Secret      string `json:"Secret"`
			File        string `json:"File"`
			SymlinkFile string `json:"SymlinkFile"`
			Commit      string `json:"Commit"`
			Author      string `json:"Author"`
			Email       string `json:"Email"`
			Date        string `json:"Date"`
			Message     string `json:"Message"`
			Fingerprint string `json:"Fingerprint"`
		}

		if err := json.Unmarshal(line, &finding); err != nil {
			// Try parsing as array
			var arr []json.RawMessage
			if err := json.Unmarshal(line, &arr); err == nil {
				for _, item := range arr {
					var f struct {
						RuleID      string `json:"RuleID"`
						Description string `json:"Description"`
						StartLine   int    `json:"StartLine"`
						EndLine     int    `json:"EndLine"`
						StartColumn int    `json:"StartColumn"`
						EndColumn   int    `json:"EndColumn"`
						Match       string `json:"Match"`
						Secret      string `json:"Secret"`
						File        string `json:"File"`
						SymlinkFile string `json:"SymlinkFile"`
						Commit      string `json:"Commit"`
						Author      string `json:"Author"`
						Email       string `json:"Email"`
						Date        string `json:"Date"`
						Message     string `json:"Message"`
						Fingerprint string `json:"Fingerprint"`
					}
					if err := json.Unmarshal(item, &f); err == nil {
						results = append(results, f)
					}
				}
			}
			continue
		}

		results = append(results, finding)
	}

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

// getGitleaksBinary trả về đường dẫn đến gitleaks binary
func getGitleaksBinary() (string, error) {
	// First, check if gitleaks is in PATH
	if path, err := exec.LookPath("gitleaks"); err == nil {
		return path, nil
	}

	// TODO: Implement auto-download
	return "", fmt.Errorf("gitleaks không tìm thấy")
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
