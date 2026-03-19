package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

const (
	defaultOllamaURL = "http://localhost:11434"
	defaultModel     = "qwen2.5-coder:3b"
)

// OllamaClient client để giao tiếp với Ollama
type OllamaClient struct {
	BaseURL string
	Model   string
}

// NewOllamaClient tạo client mới
func NewOllamaClient() *OllamaClient {
	return &OllamaClient{
		BaseURL: defaultOllamaURL,
		Model:   defaultModel,
	}
}

// IsOllamaAvailable kiểm tra Ollama có đang chạy không
func IsOllamaAvailable() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(defaultOllamaURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Generate gửi prompt và nhận response
func (c *OllamaClient) Generate(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":  c.Model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.1,
			"num_ctx":     2048,
			"num_predict": 512,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("lỗi gọi Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama trả về status %d", resp.StatusCode)
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("lỗi decode response: %w", err)
	}

	return result.Response, nil
}

// StreamGenerate gửi prompt và nhận streaming response
func (c *OllamaClient) StreamGenerate(prompt string, callback func(string)) error {
	reqBody := map[string]interface{}{
		"model":  c.Model,
		"prompt": prompt,
		"stream": true,
		"options": map[string]interface{}{
			"temperature": 0.1,
			"num_ctx":     2048,
			"num_predict": 512,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("lỗi gọi Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama trả về status %d", resp.StatusCode)
	}

	// Read streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		var chunk struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := json.Unmarshal([]byte(line), &chunk); err == nil {
			callback(chunk.Response)
			if chunk.Done {
				break
			}
		}
	}

	return scanner.Err()
}

// DownloadOllama tải và cài đặt Ollama
func DownloadOllama() error {
	switch runtime.GOOS {
	case "darwin":
		return downloadOllamaMacOS()
	case "linux":
		return downloadOllamaLinux()
	case "windows":
		return downloadOllamaWindows()
	default:
		return fmt.Errorf("hệ điều hành không được hỗ trợ: %s", runtime.GOOS)
	}
}

// downloadOllamaMacOS tải Ollama cho macOS
func downloadOllamaMacOS() error {
	fmt.Println("📥 Đang tải Ollama cho macOS...")
	fmt.Println("💡 Vui lòng cài đặt thủ công từ: https://ollama.ai/download")
	return fmt.Errorf("tự động tải chưa được hỗ trợ trên macOS")
}

// downloadOllamaLinux tải Ollama cho Linux
func downloadOllamaLinux() error {
	fmt.Println("📥 Đang cài đặt Ollama cho Linux...")
	cmd := exec.Command("sh", "-c", "curl -fsSL https://ollama.ai/install.sh | sh")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// downloadOllamaWindows tải Ollama cho Windows
func downloadOllamaWindows() error {
	fmt.Println("📥 Đang tải Ollama cho Windows...")
	fmt.Println("💡 Vui lòng cài đặt thủ công từ: https://ollama.ai/download")
	return fmt.Errorf("tự động tải chưa được hỗ trợ trên Windows")
}

// PullModel tải model về
func PullModel(model string) error {
	if !IsOllamaAvailable() {
		return fmt.Errorf("Ollama không khả dụng. Vui lòng khởi động Ollama trước.")
	}

	fmt.Printf("📥 Đang tải model %s (có thể mất vài phút)...\n", model)

	reqBody := map[string]string{
		"name": model,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(defaultOllamaURL+"/api/pull", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("lỗi pull model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pull model thất bại với status %d", resp.StatusCode)
	}

	// Read progress
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var progress struct {
			Status    string `json:"status"`
			Completed bool   `json:"completed"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &progress); err == nil {
			if progress.Completed {
				break
			}
		}
	}

	return scanner.Err()
}

// ListInstalledModels liệt kê các model đã cài đặt
func ListInstalledModels() ([]string, error) {
	if !IsOllamaAvailable() {
		return nil, fmt.Errorf("Ollama không khả dụng")
	}

	resp, err := http.Get(defaultOllamaURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy danh sách models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API trả về status %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("lỗi decode response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}

	return models, nil
}

// RemoveModel gỡ bỏ một model
func RemoveModel(model string) error {
	if !IsOllamaAvailable() {
		return fmt.Errorf("Ollama không khả dụng")
	}

	reqBody := map[string]string{
		"name": model,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("DELETE", defaultOllamaURL+"/api/delete", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("lỗi tạo request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("lỗi xóa model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("xóa model thất bại với status %d", resp.StatusCode)
	}

	return nil
}
