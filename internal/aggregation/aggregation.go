package aggregation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// Deduplicate loại bỏ findings trùng lặp
func Deduplicate(findings []models.Finding) []models.Finding {
	seen := make(map[string]bool)
	var result []models.Finding

	for _, f := range findings {
		// Create a unique key for each finding
		key := fmt.Sprintf("%s:%s:%d:%s", f.RuleID, f.File, f.Line, f.Message)
		if !seen[key] {
			seen[key] = true
			result = append(result, f)
		}
	}

	return result
}

// SortBySeverity sắp xếp findings theo severity (critical first)
func SortBySeverity(findings []models.Finding) []models.Finding {
	severityOrder := map[models.Severity]int{
		models.Critical: 0,
		models.High:     1,
		models.Medium:   2,
		models.Low:      3,
		models.Info:     4,
	}

	sort.Slice(findings, func(i, j int) bool {
		orderI := severityOrder[findings[i].Severity]
		orderJ := severityOrder[findings[j].Severity]
		if orderI != orderJ {
			return orderI < orderJ
		}
		// Same severity, sort by file then line
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		return findings[i].Line < findings[j].Line
	})

	return findings
}

// CalculateHealthScore tính điểm sức khỏe codebase
func CalculateHealthScore(findings []models.Finding) models.HealthScore {
	score := models.HealthScore{
		Overall:      100,
		Security:     100,
		Quality:      100,
		Architecture: 100,
		Performance:  100,
	}

	// Penalties for each severity
	penalties := map[models.Severity]int{
		models.Critical: 30,
		models.High:     15,
		models.Medium:   8,
		models.Low:      3,
		models.Info:     0,
	}

	for _, f := range findings {
		penalty := penalties[f.Severity]

		switch f.Category {
		case models.Security:
			score.Security -= penalty
			score.Overall -= penalty * 2 // Security issues affect overall more
		case models.Quality:
			score.Quality -= penalty
			score.Overall -= penalty
		case models.Architecture:
			score.Architecture -= penalty
			score.Overall -= penalty
		case models.Performance:
			score.Performance -= penalty
			score.Overall -= penalty
		case models.Secrets:
			score.Security -= penalty * 2 // Secrets are security issues
			score.Overall -= penalty * 2
		}
	}

	// Clamp scores to 0-100
	score.Overall = clamp(score.Overall, 0, 100)
	score.Security = clamp(score.Security, 0, 100)
	score.Quality = clamp(score.Quality, 0, 100)
	score.Architecture = clamp(score.Architecture, 0, 100)
	score.Performance = clamp(score.Performance, 0, 100)

	return score
}

// CalculateSummary tính toán summary statistics
func CalculateSummary(findings []models.Finding) models.ScanSummary {
	summary := models.ScanSummary{}

	for _, f := range findings {
		summary.Total++
		switch f.Severity {
		case models.Critical:
			summary.Critical++
		case models.High:
			summary.High++
		case models.Medium:
			summary.Medium++
		case models.Low:
			summary.Low++
		case models.Info:
			summary.Info++
		}
	}

	return summary
}

// GroupByCategory nhóm findings theo category
func GroupByCategory(findings []models.Finding) map[models.Category][]models.Finding {
	groups := make(map[models.Category][]models.Finding)

	for _, f := range findings {
		groups[f.Category] = append(groups[f.Category], f)
	}

	return groups
}

// GroupBySeverity nhóm findings theo severity
func GroupBySeverity(findings []models.Finding) map[models.Severity][]models.Finding {
	groups := make(map[models.Severity][]models.Finding)

	for _, f := range findings {
		groups[f.Severity] = append(groups[f.Severity], f)
	}

	return groups
}

// GroupByFile nhóm findings theo file
func GroupByFile(findings []models.Finding) map[string][]models.Finding {
	groups := make(map[string][]models.Finding)

	for _, f := range findings {
		groups[f.File] = append(groups[f.File], f)
	}

	return groups
}

// GetTopFindings lấy N findings quan trọng nhất
func GetTopFindings(findings []models.Finding, n int) []models.Finding {
	if len(findings) <= n {
		return findings
	}
	return findings[:n]
}

// GetSeverityLabel trả về label cho severity với emoji
func GetSeverityLabel(sev models.Severity) string {
	switch sev {
	case models.Critical:
		return "🔴 Critical"
	case models.High:
		return "🟠 High"
	case models.Medium:
		return "🟡 Medium"
	case models.Low:
		return "🔵 Low"
	case models.Info:
		return "⚪ Info"
	default:
		return string(sev)
	}
}

// GetCategoryLabel trả về label cho category
func GetCategoryLabel(cat models.Category) string {
	switch cat {
	case models.Security:
		return "🔐 Bảo mật"
	case models.Quality:
		return "✨ Chất lượng"
	case models.Architecture:
		return "🏗️ Kiến trúc"
	case models.Performance:
		return "⚡ Hiệu năng"
	case models.Secrets:
		return "🔑 Secrets"
	default:
		return string(cat)
	}
}

// GetHealthStatus trả về trạng thái sức khỏe theo điểm
func GetHealthStatus(score int) (string, string) {
	switch {
	case score >= 80:
		return "🟢 Tốt", "Có thể deploy, monitor regularly"
	case score >= 60:
		return "🟡 Trung bình", "Fix Critical/High trước khi scale"
	case score >= 40:
		return "🟠 Cần cải thiện", "Cần sprint fix issues"
	default:
		return "🔴 Nguy hiểm", "Không nên deploy production"
	}
}

// GetCategoryDistribution trả về phân bố findings theo category
func GetCategoryDistribution(findings []models.Finding) map[models.Category]int {
	dist := make(map[models.Category]int)

	for _, f := range findings {
		dist[f.Category]++
	}

	return dist
}

// SearchFindings tìm kiếm findings theo query
func SearchFindings(findings []models.Finding, query string) []models.Finding {
	query = strings.ToLower(query)
	var results []models.Finding

	for _, f := range findings {
		if strings.Contains(strings.ToLower(f.Title), query) ||
		   strings.Contains(strings.ToLower(f.Message), query) ||
		   strings.Contains(strings.ToLower(f.File), query) ||
		   strings.Contains(strings.ToLower(f.RuleID), query) ||
		   strings.Contains(strings.ToLower(string(f.Category)), query) {
			results = append(results, f)
		}
	}

	return results
}

// FilterFindings lọc findings theo criteria
func FilterFindings(findings []models.Finding, severity *models.Severity, category *models.Category, filePrefix string) []models.Finding {
	var results []models.Finding

	for _, f := range findings {
		if severity != nil && f.Severity != *severity {
			continue
		}
		if category != nil && f.Category != *category {
			continue
		}
		if filePrefix != "" && !strings.HasPrefix(f.File, filePrefix) {
			continue
		}
		results = append(results, f)
	}

	return results
}

// clamp giới hạn giá trị trong range
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
