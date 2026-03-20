package output

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nhh0718/vibe-scanner-/internal/ai"
	"github.com/nhh0718/vibe-scanner-/internal/models"
	"github.com/nhh0718/vibe-scanner-/internal/ui"
)

// GetWebFSFunc is set by main package to provide embedded web files
var GetWebFSFunc func() (fs.FS, error)

// ServeDashboard khởi động web server để hiển thị dashboard
func ServeDashboard(results *models.ScanResult, port int) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Current active scan result (can be switched via API)
	currentResult := results

	// Serve embedded web dashboard - MUST be before API routes
	staticFS, err := GetWebFSFunc()
	if err != nil {
		return fmt.Errorf("không thể tạo sub filesystem: %w", err)
	}

	// API endpoints first
	api := r.Group("/api")
	{
		// Get current scan result
		api.GET("/scan", func(c *gin.Context) {
			c.JSON(http.StatusOK, currentResult)
		})

		// List all saved reports (for history selector)
		api.GET("/reports", func(c *gin.Context) {
			reports, err := ListScanReports()
			if err != nil {
				c.JSON(http.StatusOK, []ScanReportInfo{})
				return
			}
			c.JSON(http.StatusOK, reports)
		})

		// Load a specific report by filename or scan ID
		api.GET("/reports/:id", func(c *gin.Context) {
			id := c.Param("id")
			report, err := LoadScanReport(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, report)
		})

		// Switch active report (used by frontend report selector)
		api.POST("/reports/:id/activate", func(c *gin.Context) {
			id := c.Param("id")
			report, err := LoadScanReport(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			currentResult = report
			c.JSON(http.StatusOK, gin.H{"status": "ok", "scan_id": report.ScanID})
		})

		// Reload latest scan result
		api.POST("/refresh", func(c *gin.Context) {
			latest, err := LoadLastScan()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			currentResult = latest
			c.JSON(http.StatusOK, currentResult)
		})

		// Delete a report
		api.DELETE("/reports/:id", func(c *gin.Context) {
			id := c.Param("id")
			if err := DeleteScanReport(id); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "deleted"})
		})

		api.GET("/findings", func(c *gin.Context) {
			c.JSON(http.StatusOK, currentResult.Findings)
		})

		api.GET("/finding/:id", func(c *gin.Context) {
			id := c.Param("id")
			finding := currentResult.FindByID(id)
			if finding == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Finding not found"})
				return
			}
			c.JSON(http.StatusOK, finding)
		})

		api.GET("/ai/explain/:id", func(c *gin.Context) {
			id := c.Param("id")
			finding := currentResult.FindByID(id)
			if finding == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Finding not found"})
				return
			}

			// Check if AI is available
			if !ai.IsOllamaAvailable() {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "AI không khả dụng. Chạy 'vibescanner ai-setup' để cài đặt.",
				})
				return
			}

			// Return AI explanation
			explanation, err := generateAIExplanation(finding)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"explanation": explanation,
				"fix_code":    "",
			})
		})
	}

	// Serve static files for all other routes (SPA fallback)
	staticServer := http.FileServer(http.FS(staticFS))
	r.NoRoute(func(c *gin.Context) {
		staticServer.ServeHTTP(c.Writer, c.Request)
	})

	addr := fmt.Sprintf("localhost:%d", port)

	// Create HTTP server
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Channel to listen for errors from server
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		// Build dashboard info for styled banner
		timestamp := ""
		if results != nil && !results.Timestamp.IsZero() {
			timestamp = results.Timestamp.Format("02/01/2006 15:04")
		}
		projectName := "Unknown"
		findingCount := 0
		healthScore := 0
		if results != nil {
			if results.Project.Name != "" {
				projectName = results.Project.Name
			}
			findingCount = len(results.Findings)
			healthScore = results.HealthScore.Overall
		}

		url := fmt.Sprintf("http://%s", addr)
		banner := ui.GetDashboardBanner(ui.DashboardInfo{
			URL:          url,
			ProjectName:  projectName,
			FindingCount: findingCount,
			Timestamp:    timestamp,
			HealthScore:  healthScore,
		})
		fmt.Println(banner)

		// Auto-open browser after a short delay
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowser(url)
		}()

		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive signal or error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case <-shutdown:
		fmt.Println("\n" + ui.Muted("Đang dừng server..."))

		// Give outstanding requests 5 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			return fmt.Errorf("could not gracefully shutdown: %w", err)
		}
	}

	return nil
}

// generateAIExplanation tạo giải thích từ AI cho finding
func generateAIExplanation(finding *models.Finding) (string, error) {
	client := ai.NewOllamaClient()

	prompt := fmt.Sprintf(`Bạn là security expert giải thích cho người không có kiến thức kỹ thuật.

Loại lỗi: %s
File: %s:%d
Message: %s

Giải thích ngắn gọn (2-3 câu):
1. Tại sao lỗi này nguy hiểm?
2. Hacker có thể khai thác như thế nào?
3. Cách sửa cơ bản?

Trả lời bằng tiếng Việt, ngắn gọn, dễ hiểu.`,
		finding.Title,
		finding.File,
		finding.Line,
		finding.Message,
	)

	return client.Generate(prompt)
}

