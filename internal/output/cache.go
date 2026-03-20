package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// getCacheDir returns the cache directory for VibeScanner
func getCacheDir() string {
	var cacheDir string
	if runtime.GOOS == "windows" {
		cacheDir = os.Getenv("APPDATA")
		if cacheDir == "" {
			cacheDir = os.Getenv("USERPROFILE")
		}
	} else {
		cacheDir = os.Getenv("HOME")
	}
	return filepath.Join(cacheDir, ".vibescanner")
}

// SaveLastScan saves scan results to cache with timestamp and repo name
func SaveLastScan(results *models.ScanResult) error {
	cacheDir := getCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("cannot create cache dir: %w", err)
	}

	// Generate filename with timestamp and repo name
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	repoName := getRepoName(results.Project.Path)
	filename := fmt.Sprintf("%s_%s.json", timestamp, sanitizeFilename(repoName))
	cacheFile := filepath.Join(cacheDir, filename)
	
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal results: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("cannot write cache file: %w", err)
	}

	// Also save as last-scan.json for backward compatibility
	lastScanFile := filepath.Join(cacheDir, "last-scan.json")
	os.WriteFile(lastScanFile, data, 0644)

	return nil
}

// LoadLastScan loads the most recent scan results using report_manager as single source of truth
func LoadLastScan() (*models.ScanResult, error) {
	// Primary: load from reports directory (managed by report_manager)
	reports, err := ListScanReports()
	if err == nil && len(reports) > 0 {
		// Reports are already sorted newest-first
		return LoadScanReport(reports[0].Filename)
	}

	// Fallback: try last-scan.json in cache dir (backward compatibility)
	cacheDir := getCacheDir()
	lastScanFile := filepath.Join(cacheDir, "last-scan.json")
	data, err := os.ReadFile(lastScanFile)
	if err != nil {
		return nil, fmt.Errorf("chưa có kết quả scan nào. Hãy chạy 'vibescanner scan .' trước")
	}
	var results models.ScanResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("cannot parse cache file: %w", err)
	}
	return &results, nil
}

// getRepoName extracts repo name from project path
func getRepoName(path string) string {
	if path == "" || path == "." {
		wd, _ := os.Getwd()
		path = wd
	}
	return filepath.Base(path)
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}
	return name
}
