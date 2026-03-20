package engines

import (
	"fmt"
	"time"

	"github.com/nhh0718/vibe-scanner-/internal/engines/ast"
	"github.com/nhh0718/vibe-scanner-/internal/engines/rules"
	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// RunTreeSitterEngine chạy AST-based analysis engine
func RunTreeSitterEngine(path string) []models.Finding {
	fmt.Println("🌳 Chạy AST Engine (100% AST-based, không dùng regex)...")

	// Chạy scan
	astFindings, err := ast.ScanProject(path, nil)
	if err != nil {
		fmt.Printf("⚠️  Lỗi scan: %v\n", err)
		return nil
	}

	// Convert rules.Finding sang models.Finding
	var findings []models.Finding
	for _, f := range astFindings {
		findings = append(findings, convertFinding(f))
	}

	fmt.Printf("✅ VibeSecurity Engine tìm thấy %d vấn đề\n", len(findings))
	return findings
}

// convertFinding chuyển đổi từ rules.Finding sang models.Finding
func convertFinding(f rules.Finding) models.Finding {
	// Map severity
	var severity models.Severity
	switch f.Severity {
	case rules.Critical:
		severity = models.Critical
	case rules.High:
		severity = models.High
	case rules.Medium:
		severity = models.Medium
	case rules.Low:
		severity = models.Low
	default:
		severity = models.Info
	}

	// Map category
	var category models.Category
	switch f.Category {
	case "security":
		category = models.Security
	case "quality":
		category = models.Quality
	case "performance":
		category = models.Performance
	case "architecture":
		category = models.Architecture
	case "secrets":
		category = models.Secrets
	default:
		category = models.Security
	}

	return models.Finding{
		RuleID:        f.RuleID,
		Title:         f.Title,
		Message:       f.Description,
		FixSuggestion: f.Fix,
		File:          f.File,
		Line:          f.Line,
		Column:        f.Col,
		CodeSnippet:   f.Snippet,
		Severity:      severity,
		Category:      category,
		Engine:        "ast",
		Timestamp:     time.Now(),
	}
}
