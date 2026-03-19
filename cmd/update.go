package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// version, commit, date are set at build time via ldflags in root.go

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Cập nhật VibeScanner lên phiên bản mới nhất",
	Long:  `Tự động kiểm tra và cài đặt phiên bản mới nhất từ GitHub releases.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate()
	},
}

func runUpdate() error {
	fmt.Println("🔍 Đang kiểm tra phiên bản mới...")

	// Get latest release info from GitHub API
	latestVersion, downloadURL, err := getLatestReleaseInfo()
	if err != nil {
		return fmt.Errorf("không thể kiểm tra phiên bản mới: %w", err)
	}

	// Check if already up to date
	currentVersion := version
	if currentVersion == "dev" {
		fmt.Println("⚠️  Phiên bản dev (build từ source). Không thể tự động update.")
		fmt.Println("   Vui lòng build lại từ source hoặc tải binary từ GitHub.")
		return nil
	}

	if latestVersion == currentVersion {
		color.Green("✅ Bạn đang dùng phiên bản mới nhất (%s)", currentVersion)
		return nil
	}

	fmt.Printf("📦 Phiên bản mới: %s (hiện tại: %s)\n", latestVersion, currentVersion)
	fmt.Println("📥 Đang tải xuống...")

	// Download new binary
	tempFile, err := downloadBinary(downloadURL)
	if err != nil {
		return fmt.Errorf("lỗi khi tải xuống: %w", err)
	}
	defer os.Remove(tempFile)

	fmt.Println("🔧 Đang cài đặt...")

	// Uninstall old version if installed globally
	if isGloballyInstalled() {
		if err := uninstallGlobal(); err != nil {
			// Non-fatal: continue with install
			fmt.Printf("⚠️  Không thể gỡ bản cũ: %v\n", err)
		}
	}

	// Install new version
	if err := installBinary(tempFile); err != nil {
		return fmt.Errorf("lỗi khi cài đặt: %w", err)
	}

	color.Green("✅ Đã cập nhật lên %s!", latestVersion)
	fmt.Println("🚀 Chạy 'vibescanner --version' để kiểm tra.")
	return nil
}

func getLatestReleaseInfo() (version, url string, err error) {
	// Call GitHub API to get latest release
	apiURL := "https://api.github.com/repos/nhh0718/vibe-scanner-/releases/latest"
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", "", fmt.Errorf("không thể kết nối GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API trả về status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("không thể parse response: %w", err)
	}

	latestVersion := release.TagName
	if latestVersion == "" {
		return "", "", fmt.Errorf("không tìm thấy tag name trong release")
	}

	// Find appropriate binary for this OS/arch
	var filename string
	switch runtime.GOOS {
	case "windows":
		filename = "vibescanner-windows-amd64.exe"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			filename = "vibescanner-darwin-arm64"
		} else {
			filename = "vibescanner-darwin-amd64"
		}
	default: // linux
		filename = "vibescanner-linux-amd64"
	}

	// Find the asset URL
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == filename {
			downloadURL = asset.URL
			break
		}
	}

	if downloadURL == "" {
		return "", "", fmt.Errorf("không tìm thấy binary %s trong release %s", filename, latestVersion)
	}

	return latestVersion, downloadURL, nil
}

func downloadBinary(url string) (string, error) {
	// Create temp file
	tempFile, err := os.CreateTemp("", "vibescanner-update-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Download
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Save to temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	// Make executable (Unix)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempFile.Name(), 0755); err != nil {
			return "", err
		}
	}

	return tempFile.Name(), nil
}

func isGloballyInstalled() bool {
	_, err := exec.LookPath("vibescanner")
	return err == nil
}

func installBinary(sourcePath string) error {
	// Find install directory
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Get target path
	targetPath := execPath
	if isGloballyInstalled() {
		// Keep the same global path
		path, _ := exec.LookPath("vibescanner")
		targetPath = path
	}

	// On Windows, we need to handle the running executable
	if runtime.GOOS == "windows" {
		// Create a batch file to replace the binary after exit
		batchFile := sourcePath + ".bat"
		batchContent := fmt.Sprintf(`
@echo off
timeout /t 2 /nobreak >nul
move /Y "%s" "%s"
del "%%~f0"
`, sourcePath, targetPath)

		if err := os.WriteFile(batchFile, []byte(batchContent), 0644); err != nil {
			return err
		}

		// Start the batch file and exit
		exec.Command("cmd", "/c", "start", "", batchFile).Start()
		fmt.Println("🔄 Vui lòng đợi vài giây để hoàn tất cập nhật...")
		return nil
	}

	// Unix: directly replace
	if err := os.Rename(sourcePath, targetPath); err != nil {
		// Try copy instead
		input, err := os.ReadFile(sourcePath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, input, 0755); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
