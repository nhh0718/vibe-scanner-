package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/vibescanner/vibescanner/internal/output"
)

// Config holds application configuration
type Config struct {
	OllamaURL      string            `json:"ollama_url"`
	DefaultModel   string            `json:"default_model"`
	InstalledModels []string          `json:"installed_models"`
	Theme          string            `json:"theme"`
	AutoOpen       bool              `json:"auto_open"`
	CustomRules    []string          `json:"custom_rules"`
	IgnorePaths    []string          `json:"ignore_paths"`
}

var (
	configPath string
	appConfig  *Config
)

func init() {
	configPath = getConfigPath()
}

func getConfigPath() string {
	var configDir string
	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	default:
		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("HOME"), ".config")
		}
	}
	return filepath.Join(configDir, "vibescanner", "config.json")
}

// LoadConfig loads configuration from disk
func LoadConfig() (*Config, error) {
	if appConfig != nil {
		return appConfig, nil
	}

	config := &Config{
		OllamaURL:       "http://localhost:11434",
		DefaultModel:    "qwen2.5-coder:3b",
		InstalledModels: []string{},
		Theme:           "dark",
		AutoOpen:        true,
		CustomRules:     []string{},
		IgnorePaths:     []string{"node_modules", ".git", "vendor", "dist", "build"},
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config doesn't exist yet, return defaults
		return config, nil
	}

	if err := json.Unmarshal(data, config); err != nil {
		return config, fmt.Errorf("failed to parse config: %w", err)
	}

	appConfig = config
	return config, nil
}

// SaveConfig saves configuration to disk
func SaveConfig(config *Config) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	appConfig = config
	return nil
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Quản lý cấu hình VibeScanner",
	Long:  `Xem và chỉnh sửa cấu hình của VibeScanner.`,
}

// configInitCmd initializes default config
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Tạo file cấu hình mặc định",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := &Config{
			OllamaURL:       "http://localhost:11434",
			DefaultModel:    "qwen2.5-coder:3b",
			InstalledModels: []string{},
			Theme:           "dark",
			AutoOpen:        true,
			CustomRules:     []string{},
			IgnorePaths:     []string{"node_modules", ".git", "vendor", "dist", "build"},
		}

		if err := SaveConfig(config); err != nil {
			return err
		}

		output.PrintSuccess("Đã tạo config tại: %s", configPath)
		return nil
	},
}

// configGetCmd gets a config value
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Lấy giá trị cấu hình",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig()
		if err != nil {
			return err
		}

		key := args[0]
		switch key {
		case "ollama_url":
			fmt.Println(config.OllamaURL)
		case "default_model":
			fmt.Println(config.DefaultModel)
		case "theme":
			fmt.Println(config.Theme)
		case "auto_open":
			fmt.Println(config.AutoOpen)
		default:
			return fmt.Errorf("unknown key: %s", key)
		}
		return nil
	},
}

// configSetCmd sets a config value
var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Thiết lập giá trị cấu hình",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig()
		if err != nil {
			return err
		}

		key, value := args[0], args[1]
		switch key {
		case "ollama_url":
			config.OllamaURL = value
		case "default_model":
			config.DefaultModel = value
		case "theme":
			config.Theme = value
		case "auto_open":
			config.AutoOpen = value == "true"
		default:
			return fmt.Errorf("unknown key: %s", key)
		}

		if err := SaveConfig(config); err != nil {
			return err
		}

		output.PrintSuccess("Đã cập nhật %s = %v", key, value)
		return nil
	},
}

// configListCmd lists all config
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Liệt kê tất cả cấu hình",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigList()
	},
}

// runConfigList prints all configuration
func runConfigList() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	fmt.Printf("📁 Config file: %s\n\n", configPath)
	fmt.Printf("  ollama_url:      %s\n", config.OllamaURL)
	fmt.Printf("  default_model:   %s\n", config.DefaultModel)
	fmt.Printf("  theme:           %s\n", config.Theme)
	fmt.Printf("  auto_open:       %v\n", config.AutoOpen)
	fmt.Printf("  ignore_paths:    %v\n", config.IgnorePaths)
	fmt.Printf("  custom_rules:    %v\n", config.CustomRules)
	fmt.Printf("  installed_models: %v\n", config.InstalledModels)
	return nil
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	rootCmd.AddCommand(configCmd)
}
