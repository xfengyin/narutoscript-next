package app

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/xfengyin/narutoscript-next/internal/config"
	"gopkg.in/yaml.v3"
)

// LoadConfig 加载配置
func LoadConfig() (*config.Config, error) {
	setDefaults()

	configPath := "config.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return nil, err
		}
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.auto_open_browser", true)
	viper.SetDefault("device.adb_path", "adb")
	viper.SetDefault("device.screen_width", 1280)
	viper.SetDefault("device.screen_height", 720)
	viper.SetDefault("automation.max_concurrency", 3)
	viper.SetDefault("automation.retry_count", 3)
	viper.SetDefault("automation.retry_delay", 1000)
}

func createDefaultConfig(path string) error {
	defaultConfig := config.Config{
		Server: config.ServerConfig{Port: 8080, AutoOpenBrowser: true},
		Device: config.DeviceConfig{ADBPath: "adb", ScreenWidth: 1280, ScreenHeight: 720},
		Automation: config.AutomationConfig{MaxConcurrency: 3, RetryCount: 3, RetryDelay: 1000},
	}

	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return err
	}

	header := `# NarutoScript Next 配置文件

`

	return os.WriteFile(path, append([]byte(header), data...), 0644)
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	exePath, _ := os.Executable()
	return filepath.Dir(exePath)
}