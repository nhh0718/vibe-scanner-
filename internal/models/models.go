// Package models chứa các data structures core của VibeScanner
package models

import (
	"time"
)

// Severity levels cho findings
type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
	Low      Severity = "low"
	Info     Severity = "info"
)

// Category types cho findings
type Category string

const (
	Security     Category = "security"
	Quality      Category = "quality"
	Architecture Category = "architecture"
	Performance  Category = "performance"
	Secrets      Category = "secrets"
)

// Finding đại diện cho một issue được phát hiện
type Finding struct {
	ID          string    `json:"id"`
	RuleID      string    `json:"rule_id"`
	Severity    Severity  `json:"severity"`
	Category    Category  `json:"category"`
	Subcategory string    `json:"subcategory"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	File        string    `json:"file"`
	Line        int       `json:"line"`
	Column      int       `json:"column"`
	CodeSnippet string    `json:"code_snippet"`
	
	// AI-generated fields
	Explanation      string `json:"explanation,omitempty"`
	FixSuggestion    string `json:"fix_suggestion,omitempty"`
	FixCode          string `json:"fix_code,omitempty"`
	References       []string `json:"references,omitempty"`
	FalsePositiveLikelihood string `json:"false_positive_likelihood,omitempty"`
	
	// Metadata
	Engine    string    `json:"engine"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthScore đại diện cho điểm sức khỏe codebase
type HealthScore struct {
	Overall      int `json:"overall"`
	Security     int `json:"security"`
	Quality      int `json:"quality"`
	Architecture int `json:"architecture"`
	Performance  int `json:"performance"`
}

// ScanSummary chứa thống kê tổng quan
type ScanSummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
	Total    int `json:"total"`
}

// ProjectInfo chứa thông tin về project được scan
type ProjectInfo struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Languages    []string `json:"languages"`
	FilesScanned int      `json:"files_scanned"`
	LinesOfCode  int      `json:"lines_of_code"`
}

// ScanResult là kết quả tổng hợp của một lần scan
type ScanResult struct {
	ScanID      string       `json:"scan_id"`
	Timestamp   time.Time    `json:"timestamp"`
	Duration    time.Duration `json:"duration"`
	Project     ProjectInfo  `json:"project"`
	HealthScore HealthScore  `json:"health_score"`
	Summary     ScanSummary  `json:"summary"`
	Findings    []Finding    `json:"findings"`
}

// FindByID tìm finding theo ID
func (r *ScanResult) FindByID(id string) *Finding {
	for i := range r.Findings {
		if r.Findings[i].ID == id {
			return &r.Findings[i]
		}
	}
	return nil
}

// FindingsBySeverity lọc findings theo severity
func (r *ScanResult) FindingsBySeverity(sev Severity) []Finding {
	var result []Finding
	for _, f := range r.Findings {
		if f.Severity == sev {
			result = append(result, f)
		}
	}
	return result
}

// FindingsByCategory lọc findings theo category
func (r *ScanResult) FindingsByCategory(cat Category) []Finding {
	var result []Finding
	for _, f := range r.Findings {
		if f.Category == cat {
			result = append(result, f)
		}
	}
	return result
}
