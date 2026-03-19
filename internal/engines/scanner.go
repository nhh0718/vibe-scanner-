package engines

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nhh0718/vibe-scanner-/internal/aggregation"
	"github.com/nhh0718/vibe-scanner-/internal/ingestion"
	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// ScanProject quét toàn bộ project và trả về kết quả
func ScanProject(path string) (*models.ScanResult, error) {
	startTime := time.Now()

	// Ingestion - phân tích project structure
	projectInfo, err := ingestion.AnalyzeProject(path)
	if err != nil {
		return nil, fmt.Errorf("lỗi phân tích project: %w", err)
	}

	fmt.Printf("📁 Đã tìm thấy %d files (%d dòng code)\n", projectInfo.FilesScanned, projectInfo.LinesOfCode)
	fmt.Println("🔍 Đang chạy các engine phân tích...")

	// Chạy các engine song song
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allFindings []models.Finding

	collect := func(findings []models.Finding) {
		mu.Lock()
		allFindings = append(allFindings, findings...)
		mu.Unlock()
	}

	// Chạy các engine
	wg.Add(4)
	
	go func() {
		defer wg.Done()
		f, _ := RunSemgrep(path)
		collect(f)
	}()
	
	go func() {
		defer wg.Done()
		f, _ := RunGitleaks(path)
		collect(f)
	}()
	
	go func() {
		defer wg.Done()
		f, _ := RunComplexity(path)
		collect(f)
	}()
	
	go func() {
		defer wg.Done()
		f, _ := RunDependencyAudit(path)
		collect(f)
	}()

	wg.Wait()

	// Aggregation
	findings := aggregation.Deduplicate(allFindings)
	findings = aggregation.SortBySeverity(findings)
	score := aggregation.CalculateHealthScore(findings)
	summary := aggregation.CalculateSummary(findings)

	// Generate IDs for findings
	for i := range findings {
		findings[i].ID = fmt.Sprintf("F-%s", uuid.New().String()[:8])
		findings[i].Timestamp = time.Now()
	}

	duration := time.Since(startTime)

	return &models.ScanResult{
		ScanID:      uuid.New().String(),
		Timestamp:   time.Now(),
		Duration:    duration,
		Project:     *projectInfo,
		HealthScore: score,
		Summary:     summary,
		Findings:    findings,
	}, nil
}
