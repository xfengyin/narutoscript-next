package config

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