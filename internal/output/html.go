package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nhh0718/vibe-scanner-/internal/aggregation"
	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// GenerateHTML tạo HTML report và mở browser nếu cần
func GenerateHTML(results *models.ScanResult, path string, openBrowser bool) error {
	filename := fmt.Sprintf("vibescanner-report-%s.html", time.Now().Format("20060102-150405"))

	html := generateHTMLContent(results)

	if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
		return fmt.Errorf("lỗi ghi file HTML: %w", err)
	}

	absPath, _ := filepath.Abs(filename)
	fmt.Printf("✅ Đã tạo báo cáo HTML: %s\n", absPath)

	if openBrowser {
		url := "file://" + absPath
		if err := openBrowserURL(url); err != nil {
			fmt.Printf("⚠️ Không thể mở browser: %v\n", err)
		}
	}

	return nil
}

// generateHTMLContent tạo nội dung HTML
func generateHTMLContent(results *models.ScanResult) string {
	data, _ := json.Marshal(results)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="vi">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>VibeScanner Report - %s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f172a;
            color: #e2e8f0;
            line-height: 1.6;
        }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        header {
            background: linear-gradient(135deg, #1e293b 0%%, #0f172a 100%%);
            padding: 40px;
            border-radius: 16px;
            margin-bottom: 30px;
            border: 1px solid #334155;
        }
        h1 { font-size: 2.5em; margin-bottom: 10px; background: linear-gradient(90deg, #60a5fa, #a78bfa); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
        .subtitle { color: #94a3b8; font-size: 1.1em; }
        .health-score {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .score-card {
            background: #1e293b;
            padding: 24px;
            border-radius: 12px;
            border: 1px solid #334155;
            text-align: center;
        }
        .score-value { font-size: 3em; font-weight: bold; margin-bottom: 8px; }
        .score-label { color: #94a3b8; text-transform: uppercase; font-size: 0.9em; letter-spacing: 1px; }
        .score-good { color: #22c55e; }
        .score-warning { color: #f59e0b; }
        .score-danger { color: #ef4444; }
        .summary-bar {
            display: flex;
            gap: 10px;
            margin-bottom: 30px;
            flex-wrap: wrap;
        }
        .summary-item {
            background: #1e293b;
            padding: 12px 20px;
            border-radius: 8px;
            border: 1px solid #334155;
        }
        .summary-count { font-size: 1.5em; font-weight: bold; }
        .summary-label { color: #94a3b8; font-size: 0.85em; }
        .critical { color: #ef4444; }
        .high { color: #f97316; }
        .medium { color: #eab308; }
        .low { color: #3b82f6; }
        .findings-section { margin-top: 30px; }
        .section-title {
            font-size: 1.5em;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 2px solid #334155;
        }
        .finding-card {
            background: #1e293b;
            border: 1px solid #334155;
            border-radius: 12px;
            padding: 20px;
            margin-bottom: 16px;
            transition: all 0.2s;
        }
        .finding-card:hover { border-color: #60a5fa; }
        .finding-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 12px;
        }
        .finding-id { font-family: monospace; color: #60a5fa; }
        .finding-severity {
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: 600;
        }
        .severity-critical { background: rgba(239, 68, 68, 0.2); color: #ef4444; }
        .severity-high { background: rgba(249, 115, 22, 0.2); color: #f97316; }
        .severity-medium { background: rgba(234, 179, 8, 0.2); color: #eab308; }
        .severity-low { background: rgba(59, 130, 246, 0.2); color: #3b82f6; }
        .finding-title { font-size: 1.1em; margin-bottom: 8px; }
        .finding-file {
            color: #60a5fa;
            font-family: monospace;
            font-size: 0.9em;
            margin-bottom: 12px;
        }
        .finding-message { color: #cbd5e1; margin-bottom: 12px; }
        .code-snippet {
            background: #0f172a;
            border: 1px solid #334155;
            border-radius: 8px;
            padding: 16px;
            font-family: 'Fira Code', monospace;
            font-size: 0.9em;
            overflow-x: auto;
            margin-top: 12px;
        }
        .category-tag {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 0.75em;
            margin-right: 8px;
        }
        .cat-security { background: rgba(239, 68, 68, 0.2); color: #ef4444; }
        .cat-quality { background: rgba(59, 130, 246, 0.2); color: #3b82f6; }
        .cat-architecture { background: rgba(168, 85, 247, 0.2); color: #a855f7; }
        .cat-secrets { background: rgba(236, 72, 153, 0.2); color: #ec4899; }
        footer {
            margin-top: 40px;
            padding: 20px;
            text-align: center;
            color: #64748b;
            border-top: 1px solid #334155;
        }
        .filter-bar {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
            flex-wrap: wrap;
        }
        .filter-btn {
            background: #334155;
            border: none;
            color: #e2e8f0;
            padding: 8px 16px;
            border-radius: 6px;
            cursor: pointer;
            transition: all 0.2s;
        }
        .filter-btn:hover { background: #475569; }
        .filter-btn.active { background: #60a5fa; }
        .hidden { display: none; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🔍 VibeScanner Report</h1>
            <p class="subtitle">%s • %s • %d files scanned</p>
        </header>

        <div class="health-score">
            <div class="score-card">
                <div class="score-value %s">%d</div>
                <div class="score-label">Tổng quát</div>
            </div>
            <div class="score-card">
                <div class="score-value %s">%d</div>
                <div class="score-label">Bảo mật</div>
            </div>
            <div class="score-card">
                <div class="score-value %s">%d</div>
                <div class="score-label">Chất lượng</div>
            </div>
            <div class="score-card">
                <div class="score-value %s">%d</div>
                <div class="score-label">Kiến trúc</div>
            </div>
        </div>

        <div class="summary-bar">
            <div class="summary-item"><span class="summary-count critical">%d</span><div class="summary-label">Critical</div></div>
            <div class="summary-item"><span class="summary-count high">%d</span><div class="summary-label">High</div></div>
            <div class="summary-item"><span class="summary-count medium">%d</span><div class="summary-label">Medium</div></div>
            <div class="summary-item"><span class="summary-count low">%d</span><div class="summary-label">Low</div></div>
            <div class="summary-item"><span class="summary-count">%d</span><div class="summary-label">Info</div></div>
        </div>

        <div class="findings-section">
            <h2 class="section-title">🚨 Phát hiện (%d)</h2>
            <div class="filter-bar">
                <button class="filter-btn active" onclick="filterFindings('all')">Tất cả</button>
                <button class="filter-btn" onclick="filterFindings('critical')">Critical</button>
                <button class="filter-btn" onclick="filterFindings('high')">High</button>
                <button class="filter-btn" onclick="filterFindings('security')">Bảo mật</button>
            </div>
            <div id="findings-list">
                %s
            </div>
        </div>

        <footer>
            <p>Generated by VibeScanner v0.1.0 • %s</p>
            <p>🔒 100%% Local Analysis • Code không rời máy</p>
        </footer>
    </div>

    <script>
        const scanData = %s;

        function filterFindings(filter) {
            document.querySelectorAll('.filter-btn').forEach(btn => btn.classList.remove('active'));
            event.target.classList.add('active');
            
            const cards = document.querySelectorAll('.finding-card');
            cards.forEach(card => {
                if (filter === 'all') {
                    card.classList.remove('hidden');
                } else if (filter === 'critical' || filter === 'high') {
                    card.classList.toggle('hidden', !card.dataset.severity === filter);
                } else {
                    card.classList.toggle('hidden', !card.dataset.category === filter);
                }
            });
        }
    </script>
</body>
</html>`,
		results.Project.Name,
		results.Project.Name,
		results.Timestamp.Format("2006-01-02 15:04"),
		results.Project.FilesScanned,
		getScoreClass(results.HealthScore.Overall),
		results.HealthScore.Overall,
		getScoreClass(results.HealthScore.Security),
		results.HealthScore.Security,
		getScoreClass(results.HealthScore.Quality),
		results.HealthScore.Quality,
		getScoreClass(results.HealthScore.Architecture),
		results.HealthScore.Architecture,
		results.Summary.Critical,
		results.Summary.High,
		results.Summary.Medium,
		results.Summary.Low,
		results.Summary.Info,
		results.Summary.Total,
		generateFindingsHTML(results.Findings),
		time.Now().Format("2006-01-02 15:04:05"),
		string(data),
	)
}

// generateFindingsHTML tạo HTML cho các findings
func generateFindingsHTML(findings []models.Finding) string {
	if len(findings) == 0 {
		return "<p style='text-align:center; color:#64748b;'>Không phát hiện vấn đề nào 🎉</p>"
	}

	html := ""
	for _, f := range findings {
		categoryClass := "cat-" + string(f.Category)
		severityClass := "severity-" + string(f.Severity)

		codeHTML := ""
		if f.CodeSnippet != "" {
			codeHTML = fmt.Sprintf(`<div class="code-snippet">%s</div>`, f.CodeSnippet)
		}

		html += fmt.Sprintf(`
<div class="finding-card" data-severity="%s" data-category="%s">
    <div class="finding-header">
        <span class="finding-id">[%s]</span>
        <span class="finding-severity %s">%s</span>
    </div>
    <span class="category-tag %s">%s</span>
    <h3 class="finding-title">%s</h3>
    <div class="finding-file">%s:%d</div>
    <p class="finding-message">%s</p>
    %s
</div>`,
			f.Severity,
			f.Category,
			f.ID,
			severityClass,
			aggregation.GetSeverityLabel(f.Severity),
			categoryClass,
			aggregation.GetCategoryLabel(f.Category),
			f.Title,
			f.File, f.Line,
			f.Message,
			codeHTML,
		)
	}

	return html
}

// getScoreClass trả về class CSS cho score
func getScoreClass(score int) string {
	if score >= 80 {
		return "score-good"
	}
	if score >= 60 {
		return "score-warning"
	}
	return "score-danger"
}

// openBrowserURL mở browser với URL
func openBrowserURL(url string) error {
	return openBrowser(url)
}
