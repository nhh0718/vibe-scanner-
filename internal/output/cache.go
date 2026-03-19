package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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

// SaveLastScan saves scan results to cache
func SaveLastScan(results *models.ScanResult) error {
	cacheDir := getCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("cannot create cache dir: %w", err)
	}

	cacheFile := filepath.Join(cacheDir, "last-scan.json")
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal results: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("cannot write cache file: %w", err)
	}

	return nil
}

// LoadLastScan loads the last scan results from cache
func LoadLastScan() (*models.ScanResult, error) {
	cacheFile := filepath.Join(getCacheDir(), "last-scan.json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("chưa có kết quả scan nào được lưu. Hãy chạy 'vibescanner scan .' trước")
		}
		return nil, fmt.Errorf("cannot read cache file: %w", err)
	}

	var results models.ScanResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("cannot parse cache file: %w", err)
	}

	return &results, nil
}
