package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vibescanner/vibescanner/internal/output"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Mở Web Dashboard để xem kết quả quét",
	Long:  `Khởi động local web server và mở dashboard trong browser.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")

		// Load last scan results
		results, err := output.LoadLastScan()
		if err != nil {
			return fmt.Errorf("không tìm thấy kết quả scan trước đó: %w", err)
		}

		fmt.Printf("🌐 Mở dashboard tại http://localhost:%d\n", port)
		return output.ServeDashboard(results, port)
	},
}

func init() {
	serveCmd.Flags().Int("port", 7420, "Port để chạy web server")
	rootCmd.AddCommand(serveCmd)
}
