package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nhh0718/vibe-scanner-/internal/output"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Cài đặt VibeScanner vào PATH để chạy global",
	Long: `Cài đặt VibeScanner vào system PATH để có thể chạy từ bất kỳ đâu.

Windows: Tạo symlink trong %LOCALAPPDATA%\Microsoft\WindowsApps
macOS/Linux: Tạo symlink trong /usr/local/bin hoặc ~/.local/bin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return installGlobal()
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Gỡ cài đặt VibeScanner khỏi PATH",
	RunE: func(cmd *cobra.Command, args []string) error {
		return uninstallGlobal()
	},
}

func installGlobal() error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("không thể xác định đường dẫn hiện tại: %w", err)
	}

	// Convert to absolute path
	execPath, err = filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("không thể resolve đường dẫn: %w", err)
	}

	// Determine install location based on OS
	var installDir string
	switch runtime.GOOS {
	case "windows":
		installDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "WindowsApps")
		if installDir == "" || installDir == `\Microsoft\WindowsApps` {
			// Fallback
			installDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Microsoft", "WindowsApps")
		}
	case "darwin":
		installDir = "/usr/local/bin"
		// Check if we can write to /usr/local/bin
		if _, err := os.Stat(installDir); os.IsNotExist(err) {
			installDir = filepath.Join(os.Getenv("HOME"), ".local", "bin")
		}
	default: // linux
		installDir = "/usr/local/bin"
		if _, err := os.Stat(installDir); os.IsNotExist(err) {
			installDir = filepath.Join(os.Getenv("HOME"), ".local", "bin")
		}
	}

	// Ensure install directory exists
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục cài đặt: %w", err)
	}

	targetPath := filepath.Join(installDir, "vibescanner")
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	// Check if already exists
	if _, err := os.Stat(targetPath); err == nil {
		// Remove existing symlink/file
		if err := os.Remove(targetPath); err != nil {
			return fmt.Errorf("không thể xóa cài đặt cũ: %w", err)
		}
	}

	// Create symlink
	if err := os.Symlink(execPath, targetPath); err != nil {
		// If symlink fails (Windows often requires admin), try copying
		output.PrintInfo("Không thể tạo symlink, đang copy file...")
		if err := copyFile(execPath, targetPath); err != nil {
			return fmt.Errorf("không thể copy file: %w", err)
		}
	}

	// Verify installation
	if _, err := exec.LookPath("vibescanner"); err != nil {
		output.PrintInfo("Lưu ý: Bạn cần restart terminal hoặc mở terminal mới để dùng lệnh 'vibescanner'")
	}

	output.PrintSuccess("Đã cài đặt VibeScanner!")
	fmt.Printf("   Vị trí: %s\n", targetPath)
	fmt.Println()
	fmt.Println("Bạn có thể chạy từ bất kỳ đâu:")
	fmt.Println("  vibescanner scan ./my-project")
	fmt.Println("  vibescanner --help")
	
	return nil
}

func uninstallGlobal() error {
	var filesToDelete []string

	// 1. Find in PATH manually (WindowsApps, /usr/local/bin, etc.)
	pathEnv := os.Getenv("PATH")
	pathDirs := filepath.SplitList(pathEnv)
	
	for _, dir := range pathDirs {
		candidate := filepath.Join(dir, "vibescanner")
		if runtime.GOOS == "windows" {
			candidate += ".exe"
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			filesToDelete = append(filesToDelete, candidate)
		}
	}

	// 2. Also get current executable (in case running from download folder)
	execPath, _ := os.Executable()
	if execPath != "" {
		execPath, _ = filepath.Abs(execPath)
		// Add if not already in list
		found := false
		for _, f := range filesToDelete {
			if strings.EqualFold(f, execPath) {
				found = true
				break
			}
		}
		if !found {
			filesToDelete = append(filesToDelete, execPath)
		}
	}

	if len(filesToDelete) == 0 {
		return fmt.Errorf("không tìm thấy VibeScanner để gỡ cài đặt")
	}

	// Delete all found files
	var deleted []string
	var errors []string

	for _, path := range filesToDelete {
		// On Windows, use delayed deletion for running executable
		if runtime.GOOS == "windows" {
			batchFile := filepath.Join(os.TempDir(), fmt.Sprintf("vibe-uninstall-%d.bat", len(deleted)))
			batchContent := fmt.Sprintf(`
@echo off
timeout /t 1 /nobreak >nul
del "%s" 2>nul
del "%%~f0"
`, path)
			if err := os.WriteFile(batchFile, []byte(batchContent), 0644); err == nil {
				exec.Command("cmd", "/c", "start", "", batchFile).Start()
				deleted = append(deleted, path)
				continue
			}
		}
		
		// Direct deletion
		if err := os.Remove(path); err == nil {
			deleted = append(deleted, path)
		} else {
			errors = append(errors, fmt.Sprintf("%s: %v", path, err))
		}
	}

	if len(deleted) > 0 {
		output.PrintSuccess("Đã gỡ cài đặt VibeScanner:")
		for _, d := range deleted {
			fmt.Printf("   • %s\n", d)
		}
		if runtime.GOOS == "windows" {
			fmt.Println("🔄 Vui lòng đóng terminal để hoàn tất.")
		}
	}

	if len(errors) > 0 {
		fmt.Println("\n⚠️  Không thể xóa một số file (có thể do đang chạy):")
		for _, e := range errors {
			fmt.Printf("   • %s\n", e)
		}
	}

	if len(deleted) == 0 && len(errors) > 0 {
		return fmt.Errorf("không thể gỡ cài đặt")
	}

	return nil
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0755)
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
}
