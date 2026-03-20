package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// getReportsDir returns the reports directory for VibeScanner
func getReportsDir() string {
	var baseDir string
	if runtime.GOOS == "windows" {
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("USERPROFILE")
		}
	} else {
		baseDir = os.Getenv("HOME")
	}
	return filepath.Join(baseDir, ".vibescanner", "reports")
}

// ScanReportInfo holds metadata about a saved scan report
type ScanReportInfo struct {
	ID           string    `json:"id"`
	ScanID       string    `json:"scan_id"`
	ProjectName  string    `json:"project_name"`
	ProjectPath  string    `json:"project_path"`
	Timestamp    time.Time `json:"timestamp"`
	FilesScanned int       `json:"files_scanned"`
	LinesOfCode  int       `json:"lines_of_code"`
	FindingCount int       `json:"finding_count"`
	HealthScore  int       `json:"health_score"`
	Duration     string    `json:"duration"`
	Filename     string    `json:"filename"`
}

// SaveScanReport saves scan results to reports directory with timestamp and project name
func SaveScanReport(results *models.ScanResult) (*ScanReportInfo, error) {
	reportsDir := getReportsDir()
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return nil, fmt.Errorf("không thể tạo thư mục reports: %w", err)
	}

	// Generate filename with timestamp and project name
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	projectName := getRepoName(results.Project.Path)
	filename := fmt.Sprintf("%s_%s.json", timestamp, sanitizeFilename(projectName))
	reportPath := filepath.Join(reportsDir, filename)

	// Save JSON report
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("không thể marshal results: %w", err)
	}

	if err := os.WriteFile(reportPath, data, 0644); err != nil {
		return nil, fmt.Errorf("không thể ghi file report: %w", err)
	}

	// Create report info
	info := &ScanReportInfo{
		ID:           results.ScanID,
		ScanID:       results.ScanID,
		ProjectName:  projectName,
		ProjectPath:  results.Project.Path,
		Timestamp:    results.Timestamp,
		FilesScanned: results.Project.FilesScanned,
		LinesOfCode:  results.Project.LinesOfCode,
		FindingCount: len(results.Findings),
		HealthScore:  results.HealthScore.Overall,
		Duration:     results.Duration.String(),
		Filename:     filename,
	}

	// Save metadata index
	if err := updateReportsIndex(info); err != nil {
		fmt.Printf("⚠️ Không thể cập nhật index: %v\n", err)
	}

	// Also save as last-scan.json for backward compatibility
	lastScanFile := filepath.Join(reportsDir, "..", "last-scan.json")
	os.WriteFile(lastScanFile, data, 0644)

	return info, nil
}

// ListScanReports returns a list of all saved scan reports
func ListScanReports() ([]ScanReportInfo, error) {
	reportsDir := getReportsDir()

	// Try to load from index first
	indexPath := filepath.Join(reportsDir, "index.json")
	if data, err := os.ReadFile(indexPath); err == nil {
		var index []ScanReportInfo
		if err := json.Unmarshal(data, &index); err == nil {
			// Sort by timestamp (newest first)
			sort.Slice(index, func(i, j int) bool {
				return index[i].Timestamp.After(index[j].Timestamp)
			})
			return index, nil
		}
	}

	// Fallback: scan directory
	files, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil, fmt.Errorf("chưa có báo cáo nào. Hãy chạy 'vibescanner scan .' trước")
	}

	var reports []ScanReportInfo
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") || file.Name() == "index.json" {
			continue
		}

		// Load report to get metadata
		reportPath := filepath.Join(reportsDir, file.Name())
		data, err := os.ReadFile(reportPath)
		if err != nil {
			continue
		}

		var result models.ScanResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		reports = append(reports, ScanReportInfo{
			ID:           result.ScanID,
			ScanID:       result.ScanID,
			ProjectName:  getRepoName(result.Project.Path),
			ProjectPath:  result.Project.Path,
			Timestamp:    result.Timestamp,
			FilesScanned: result.Project.FilesScanned,
			LinesOfCode:  result.Project.LinesOfCode,
			FindingCount: len(result.Findings),
			HealthScore:  result.HealthScore.Overall,
			Duration:     result.Duration.String(),
			Filename:     file.Name(),
		})
	}

	// Sort by timestamp (newest first)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Timestamp.After(reports[j].Timestamp)
	})

	return reports, nil
}

// LoadScanReport loads a specific scan report by filename or scan ID
func LoadScanReport(identifier string) (*models.ScanResult, error) {
	reportsDir := getReportsDir()

	// Try to find by filename
	reportPath := filepath.Join(reportsDir, identifier)
	if _, err := os.Stat(reportPath); err == nil {
		data, err := os.ReadFile(reportPath)
		if err != nil {
			return nil, fmt.Errorf("không thể đọc file: %w", err)
		}

		var result models.ScanResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("không thể parse report: %w", err)
		}
		return &result, nil
	}

	// Try to find by scan ID
	reports, err := ListScanReports()
	if err != nil {
		return nil, err
	}

	for _, report := range reports {
		if report.ScanID == identifier || report.ID == identifier {
			reportPath := filepath.Join(reportsDir, report.Filename)
			data, err := os.ReadFile(reportPath)
			if err != nil {
				return nil, fmt.Errorf("không thể đọc file: %w", err)
			}

			var result models.ScanResult
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, fmt.Errorf("không thể parse report: %w", err)
			}
			return &result, nil
		}
	}

	// Try last-scan.json
	lastScanPath := filepath.Join(reportsDir, "..", "last-scan.json")
	if data, err := os.ReadFile(lastScanPath); err == nil {
		var result models.ScanResult
		if err := json.Unmarshal(data, &result); err == nil {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("không tìm thấy báo cáo với ID: %s", identifier)
}

// DeleteScanReport deletes a specific scan report
func DeleteScanReport(identifier string) error {
	reportsDir := getReportsDir()

	// Try to find by filename first
	reportPath := filepath.Join(reportsDir, identifier)
	if _, err := os.Stat(reportPath); err == nil {
		if err := os.Remove(reportPath); err != nil {
			return fmt.Errorf("không thể xóa file: %w", err)
		}
		// Update index
		rebuildReportsIndex()
		return nil
	}

	// Try to find by scan ID
	reports, err := ListScanReports()
	if err != nil {
		return err
	}

	for _, report := range reports {
		if report.ScanID == identifier || report.ID == identifier {
			reportPath := filepath.Join(reportsDir, report.Filename)
			if err := os.Remove(reportPath); err != nil {
				return fmt.Errorf("không thể xóa file: %w", err)
			}
			// Update index
			rebuildReportsIndex()
			return nil
		}
	}

	return fmt.Errorf("không tìm thấy báo cáo với ID: %s", identifier)
}

// updateReportsIndex updates the reports index file
func updateReportsIndex(newReport *ScanReportInfo) error {
	reportsDir := getReportsDir()
	indexPath := filepath.Join(reportsDir, "index.json")

	// Load existing index
	var index []ScanReportInfo
	if data, err := os.ReadFile(indexPath); err == nil {
		json.Unmarshal(data, &index)
	}

	// Add new report
	index = append([]ScanReportInfo{*newReport}, index...)

	// Keep only last 100 reports
	if len(index) > 100 {
		index = index[:100]
	}

	// Save index
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0644)
}

// rebuildReportsIndex rebuilds the entire reports index by scanning directory directly
func rebuildReportsIndex() error {
	reportsDir := getReportsDir()
	indexPath := filepath.Join(reportsDir, "index.json")

	// Scan directory directly - don't use ListScanReports which reads from stale index
	files, err := os.ReadDir(reportsDir)
	if err != nil {
		return err
	}

	var reports []ScanReportInfo
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") || file.Name() == "index.json" {
			continue
		}

		reportPath := filepath.Join(reportsDir, file.Name())
		data, err := os.ReadFile(reportPath)
		if err != nil {
			continue
		}

		var result models.ScanResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		reports = append(reports, ScanReportInfo{
			ID:           result.ScanID,
			ScanID:       result.ScanID,
			ProjectName:  getRepoName(result.Project.Path),
			ProjectPath:  result.Project.Path,
			Timestamp:    result.Timestamp,
			FilesScanned: result.Project.FilesScanned,
			LinesOfCode:  result.Project.LinesOfCode,
			FindingCount: len(result.Findings),
			HealthScore:  result.HealthScore.Overall,
			Duration:     result.Duration.String(),
			Filename:     file.Name(),
		})
	}

	// Sort by timestamp (newest first)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Timestamp.After(reports[j].Timestamp)
	})

	data, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0644)
}

// GetReportFilePath returns the full path to a report file
func GetReportFilePath(filename string) string {
	return filepath.Join(getReportsDir(), filename)
}
