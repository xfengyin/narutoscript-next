package app

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Device   DeviceConfig   `mapstructure:"device"`
	Automation AutomationConfig `mapstructure:"automation"`
}

type ServerConfig struct {
	Port            int  `mapstructure:"port"`
	AutoOpenBrowser bool `mapstructure:"auto_open_browser"`
}

type DeviceConfig struct {
	ADBPath     string `mapstructure:"adb_path"`
	ScreenWidth  int    `mapstructure:"screen_width"`
	ScreenHeight int    `mapstructure:"screen_height"`
}

type AutomationConfig struct {
	MaxConcurrency int `mapstructure:"max_concurrency"`
	RetryCount     int `mapstructure:"retry_count"`
	RetryDelay     int `mapstructure:"retry_delay"`
}

func LoadConfig() (*Config, error) {
	// 设置默认值
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.auto_open_browser", true)
	viper.SetDefault("device.adb_path", "adb")
	viper.SetDefault("device.screen_width", 1280)
	viper.SetDefault("device.screen_height", 720)
	viper.SetDefault("automation.max_concurrency", 3)
	viper.SetDefault("automation.retry_count", 3)
	viper.SetDefault("automation.retry_delay", 1000)

	// 配置文件路径
	configPath := "config.yaml"
	
	// 如果配置文件不存在，创建默认配置
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

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func createDefaultConfig(path string) error {
	defaultConfig := `# NarutoScript Next 配置文件

server:
  port: 8080                  # Web 服务端口
  auto_open_browser: true     # 启动时自动打开浏览器

device:
  adb_path: "adb"             # ADB 路径
  screen_width: 1280          # 屏幕宽度
  screen_height: 720          # 屏幕高度

automation:
  max_concurrency: 3          # 最大并发任务数
  retry_count: 3              # 失败重试次数
  retry_delay: 1000           # 重试延迟(ms)
`
	return os.WriteFile(path, []byte(defaultConfig), 0644)
}

func GetConfigPath() string {
	exePath, _ := os.Executable()
	return filepath.Dir(exePath)
}