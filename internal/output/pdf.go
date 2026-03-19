package output

import (
	"fmt"

	"github.com/vibescanner/vibescanner/internal/models"
)

// GeneratePDF tạo PDF report (placeholder implementation)
func GeneratePDF(results *models.ScanResult, path string) error {
	// TODO: Implement PDF generation using a library like gofpdf or unidoc
	// For now, return an informative message
	fmt.Println("⚠️ PDF generation chưa được implement.")
	fmt.Println("💡 Sử dụng --report html để xem báo cáo trong browser.")
	return nil
}
