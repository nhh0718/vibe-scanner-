package engines

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func locateEngineBinary(name string) (string, error) {
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	var candidates []string
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		platform := runtime.GOOS + "-" + runtime.GOARCH
		candidates = append(candidates,
			filepath.Join(exeDir, name+ext),
			filepath.Join(exeDir, "tools", name+ext),
			filepath.Join(exeDir, "tools", platform, name+ext),
		)
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(homeDir, ".vibescanner", "bin", name+ext))
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("%s không có trong PATH hoặc bundle tools", name)
}
