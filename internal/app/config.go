package app

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server     ServerConfig     `mapstructure:"server" yaml:"server"`
	Device     DeviceConfig     `mapstructure:"device" yaml:"device"`
	Automation AutomationConfig `mapstructure:"automation" yaml:"automation"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port            int  `mapstructure:"port" yaml:"port"`
	AutoOpenBrowser bool `mapstructure:"auto_open_browser" yaml:"auto_open_browser"`
}

// DeviceConfig 设备配置
type DeviceConfig struct {
	ADBPath      string `mapstructure:"adb_path" yaml:"adb_path"`
	ScreenWidth  int    `mapstructure:"screen_width" yaml:"screen_width"`
	ScreenHeight int    `mapstructure:"screen_height" yaml:"screen_height"`
}

// AutomationConfig 自动化配置
type AutomationConfig struct {
	MaxConcurrency int `mapstructure:"max_concurrency" yaml:"max_concurrency"`
	RetryCount     int `mapstructure:"retry_count" yaml:"retry_count"`
	RetryDelay     int `mapstructure:"retry_delay" yaml:"retry_delay"`
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	// 设置默认值
	setDefaults()

	// 配置文件路径
	configPath := "config.yaml"

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return nil, err
		}
	}

	// 读取配置
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

// setDefaults 设置默认值
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

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(path string) error {
	defaultConfig := Config{
		Server: ServerConfig{
			Port:            8080,
			AutoOpenBrowser: true,
		},
		Device: DeviceConfig{
			ADBPath:      "adb",
			ScreenWidth:  1280,
			ScreenHeight: 720,
		},
		Automation: AutomationConfig{
			MaxConcurrency: 3,
			RetryCount:     3,
			RetryDelay:     1000,
		},
	}

	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return err
	}

	header := `# NarutoScript Next 配置文件
# 配置文档: https://github.com/xfengyin/narutoscript-next#配置说明

`

	return os.WriteFile(path, append([]byte(header), data...), 0644)
}

// SaveConfig 保存配置
func SaveConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile("config.yaml", data, 0644)
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	exePath, _ := os.Executable()
	return filepath.Dir(exePath)
}