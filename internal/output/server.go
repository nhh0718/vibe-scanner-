package output

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nhh0718/vibe-scanner-/internal/ai"
	"github.com/nhh0718/vibe-scanner-/internal/models"
)

// GetWebFSFunc is set by main package to provide embedded web files
var GetWebFSFunc func() (fs.FS, error)

// ServeDashboard khởi động web server để hiển thị dashboard
func ServeDashboard(results *models.ScanResult, port int) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Serve embedded web dashboard - MUST be before API routes
	staticFS, err := GetWebFSFunc()
	if err != nil {
		return fmt.Errorf("không thể tạo sub filesystem: %w", err)
	}

	// API endpoints first
	api := r.Group("/api")
	{
		api.GET("/scan", func(c *gin.Context) {
			c.JSON(http.StatusOK, results)
		})

		api.GET("/findings", func(c *gin.Context) {
			c.JSON(http.StatusOK, results.Findings)
		})

		api.GET("/finding/:id", func(c *gin.Context) {
			id := c.Param("id")
			finding := results.FindByID(id)
			if finding == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Finding not found"})
				return
			}
			c.JSON(http.StatusOK, finding)
		})

		api.GET("/ai/explain/:id", func(c *gin.Context) {
			id := c.Param("id")
			finding := results.FindByID(id)
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
	fmt.Printf("🌐 Dashboard running at http://%s\n", addr)
	return r.Run(addr)
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
