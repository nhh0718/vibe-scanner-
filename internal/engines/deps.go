package engines

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/vibescanner/vibescanner/internal/models"
)

// RunDependencyAudit audit dependencies của project
func RunDependencyAudit(path string) ([]models.Finding, error) {
	var findings []models.Finding

	// Check for package.json (Node.js)
	if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
		nodeFindings, _ := auditNodeDependencies(path)
		findings = append(findings, nodeFindings...)
	}

	// Check for requirements.txt (Python)
	if _, err := os.Stat(filepath.Join(path, "requirements.txt")); err == nil {
		pythonFindings, _ := auditPythonDependencies(path)
		findings = append(findings, pythonFindings...)
	}

	// Check for go.mod (Go)
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		goFindings, _ := auditGoDependencies(path)
		findings = append(findings, goFindings...)
	}

	return findings, nil
}

// auditNodeDependencies audit Node.js dependencies
func auditNodeDependencies(path string) ([]models.Finding, error) {
	var findings []models.Finding

	// Check if npm audit is available
	cmd := exec.Command("npm", "audit", "--json")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		// npm audit exits with non-zero when vulnerabilities found
		if _, ok := err.(*exec.ExitError); ok && len(output) > 0 {
			// Continue processing output
		} else {
			return findings, nil // npm not available or error
		}
	}

	var result struct {
		Advisories map[string]struct {
			Severity       string `json:"severity"`
			Overview       string `json:"overview"`
			Recommendation string `json:"recommendation"`
			Findings       []struct {
				Version string `json:"version"`
				Paths   []string `json:"paths"`
			} `json:"findings"`
		} `json:"advisories"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return findings, nil
	}

	for _, advisory := range result.Advisories {
		severity := parseNpmSeverity(advisory.Severity)
		for _, finding := range advisory.Findings {
			for _, depPath := range finding.Paths {
				findings = append(findings, models.Finding{
					RuleID:      "dep-npm-vulnerability",
					Severity:    severity,
					Category:    models.Security,
					Subcategory: "vulnerable_dependency",
					Title:       fmt.Sprintf("NPM vulnerability: %s", advisory.Overview),
					Message:     fmt.Sprintf("%s. %s", advisory.Overview, advisory.Recommendation),
					File:        "package.json",
					Line:        1,
					CodeSnippet: depPath,
					Engine:      "dependency-audit",
					Timestamp:   time.Now(),
				})
			}
		}
	}

	return findings, nil
}

// auditPythonDependencies audit Python dependencies
func auditPythonDependencies(path string) ([]models.Finding, error) {
	var findings []models.Finding

	// Try using safety if available
	cmd := exec.Command("safety", "check", "--json", "-r", "requirements.txt")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		// safety not available or error
		return findings, nil
	}

	var vulnerabilities []struct {
		Package string `json:"package_name"`
		VulnID  string `json:"vulnerability_id"`
		Specs   string `json:"vulnerable_spec"`
		Advisory string `json:"advisory"`
	}

	if err := json.Unmarshal(output, &vulnerabilities); err != nil {
		return findings, nil
	}

	for _, vuln := range vulnerabilities {
		findings = append(findings, models.Finding{
			RuleID:      "dep-python-vulnerability",
			Severity:    models.High,
			Category:    models.Security,
			Subcategory: "vulnerable_dependency",
			Title:       fmt.Sprintf("Python vulnerability: %s", vuln.Package),
			Message:     fmt.Sprintf("%s: %s", vuln.Package, vuln.Advisory),
			File:        "requirements.txt",
			Line:        1,
			CodeSnippet: vuln.Specs,
			Engine:      "dependency-audit",
			Timestamp:   time.Now(),
		})
	}

	return findings, nil
}

// auditGoDependencies audit Go dependencies
func auditGoDependencies(path string) ([]models.Finding, error) {
	var findings []models.Finding

	// Try govulncheck if available
	cmd := exec.Command("govulncheck", "-json", "./...")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		// govulncheck not available or error
		return findings, nil
	}

	var result struct {
		Vulns []struct {
			OSV struct {
				ID       string `json:"id"`
				Details  string `json:"details"`
				Severity string `json:"severity"`
			} `json:"osv"`
			Modules []struct {
				Path    string `json:"path"`
				Version string `json:"version"`
			} `json:"modules"`
		} `json:"Vulns"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return findings, nil
	}

	for _, vuln := range result.Vulns {
		for _, mod := range vuln.Modules {
			findings = append(findings, models.Finding{
				RuleID:      fmt.Sprintf("dep-go-%s", vuln.OSV.ID),
				Severity:    models.High,
				Category:    models.Security,
				Subcategory: "vulnerable_dependency",
				Title:       fmt.Sprintf("Go vulnerability: %s", vuln.OSV.ID),
				Message:     vuln.OSV.Details,
				File:        "go.mod",
				Line:        1,
				CodeSnippet: fmt.Sprintf("%s@%s", mod.Path, mod.Version),
				Engine:      "dependency-audit",
				Timestamp:   time.Now(),
			})
		}
	}

	return findings, nil
}

// parseNpmSeverity chuyển đổi npm severity sang model severity
func parseNpmSeverity(sev string) models.Severity {
	switch strings.ToLower(sev) {
	case "critical":
		return models.Critical
	case "high":
		return models.High
	case "moderate":
		return models.Medium
	case "low":
		return models.Low
	default:
		return models.Medium
	}
}
