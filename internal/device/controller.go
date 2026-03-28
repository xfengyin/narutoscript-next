package device

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xfengyin/narutoscript-next/internal/config"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

// Controller 设备控制器
type Controller struct {
	config    config.DeviceConfig
	log       *logger.Logger
	device    string
	mu        sync.RWMutex
	connected bool
	adbPath   string
}

// NewController 创建设备控制器
func NewController(cfg config.DeviceConfig, log *logger.Logger) *Controller {
	c := &Controller{
		config: cfg,
		log:    log,
	}
	c.adbPath = c.findADB()
	return c
}

// findADB 查找 ADB 路径
func (c *Controller) findADB() string {
	// 1. 配置中的路径
	if c.config.ADBPath != "" && c.config.ADBPath != "adb" {
		if _, err := os.Stat(c.config.ADBPath); err == nil {
			return c.config.ADBPath
		}
	}

	// 2. PATH 环境变量
	if path, err := exec.LookPath("adb.exe"); err == nil {
		return path
	}
	if path, err := exec.LookPath("adb"); err == nil {
		return path
	}

	// 3. 常见模拟器和 SDK 路径
	searchPaths := []string{
		// MuMu 12
		`C:\Program Files\Netease\MuMuPlayer-12.0\shell\adb.exe`,
		`D:\Program Files\Netease\MuMuPlayer-12.0\shell\adb.exe`,
		`E:\Program Files\Netease\MuMuPlayer-12.0\shell\adb.exe`,
		// MuMu 老版本
		`C:\Program Files\Netease\MuMu Player\platform-tools\adb.exe`,
		`D:\Program Files\Netease\MuMu Player\platform-tools\adb.exe`,
		// 雷电模拟器
		`C:\Program Files\LDPlayer\adb.exe`,
		`D:\Program Files\LDPlayer\adb.exe`,
		`C:\leidian\LDPlayer\adb.exe`,
		// 夜神模拟器
		`C:\Program Files\Nox\bin\adb.exe`,
		`D:\Program Files\Nox\bin\adb.exe`,
		// BlueStacks
		`C:\Program Files\BlueStacks\HD-Adb.exe`,
		// Android SDK
		`C:\Android\platform-tools\adb.exe`,
		`C:\platform-tools\adb.exe`,
		`C:\adb\adb.exe`,
	}

	// 添加用户目录搜索
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		searchPaths = append(searchPaths,
			filepath.Join(homeDir, "AppData", "Local", "Android", "Sdk", "platform-tools", "adb.exe"),
			filepath.Join(homeDir, "Android", "platform-tools", "adb.exe"),
		)
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			c.log.Info("找到 ADB", "path", p)
			return p
		}
	}

	// 4. 搜索 Program Files
	for _, drive := range []string{"C:", "D:", "E:"} {
		for _, pf := range []string{"Program Files", "Program Files (x86)"} {
			base := filepath.Join(drive, pf)
			if _, err := os.Stat(base); err != nil {
				continue
			}
			// 搜索模拟器目录
			filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if strings.Contains(strings.ToLower(path), "mumu") ||
					strings.Contains(strings.ToLower(path), "leidian") ||
					strings.Contains(strings.ToLower(path), "nox") {
					adbPath := filepath.Join(filepath.Dir(path), "adb.exe")
					if _, err := os.Stat(adbPath); err == nil {
						return nil
					}
					adbPath = filepath.Join(path, "shell", "adb.exe")
					if _, err := os.Stat(adbPath); err == nil {
						c.log.Info("找到 ADB", "path", adbPath)
						return nil
					}
				}
				return nil
			})
		}
	}

	return ""
}

// Connect 连接设备
func (c *Controller) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查 ADB
	if c.adbPath == "" {
		c.adbPath = c.findADB()
	}
	if c.adbPath == "" {
		return fmt.Errorf("未找到 ADB，请安装 MuMu 模拟器或配置 adb_path")
	}

	// 尝试连接 MuMu 模拟器
	mumuPorts := []string{"7555", "16384", "16416", "16448", "16480"}
	for _, port := range mumuPorts {
		addr := fmt.Sprintf("127.0.0.1:%s", port)
		cmd := exec.CommandContext(context.Background(), c.adbPath, "connect", addr)
		output, _ := cmd.CombinedOutput()
		if bytes.Contains(output, []byte("connected")) || bytes.Contains(output, []byte("already connected")) {
			c.device = addr
			c.connected = true
			c.log.Info("MuMu 模拟器连接成功", "address", addr)
			return nil
		}
	}

	// 检查已连接设备
	cmd := exec.CommandContext(context.Background(), c.adbPath, "devices")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ADB 执行失败: %w", err)
	}

	lines := bytes.Split(output, []byte("\n"))
	for i, line := range lines {
		if i > 0 && len(line) > 0 {
			fields := bytes.Fields(line)
			if len(fields) >= 2 && string(fields[1]) == "device" {
				c.device = string(fields[0])
				c.connected = true
				c.log.Info("已连接设备", "device", c.device)
				return nil
			}
		}
	}

	return fmt.Errorf("未找到设备。请确保：\n1. MuMu 模拟器已启动\n2. 或手机已连接并开启 USB 调试")
}

// Reconnect 重连
func (c *Controller) Reconnect() error {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	return c.Connect()
}

// IsConnected 检查连接状态
func (c *Controller) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetDeviceName 获取设备名称
func (c *Controller) GetDeviceName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.device != "" {
		return c.device
	}
	return "MuMu 模拟器"
}

// Screenshot 截图
func (c *Controller) Screenshot() ([]byte, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("设备未连接")
	}

	args := []string{}
	if c.device != "" {
		args = append(args, "-s", c.device)
	}
	args = append(args, "exec-out", "screencap", "-p")

	cmd := exec.CommandContext(context.Background(), c.adbPath, args...)
	return cmd.Output()
}

// Tap 点击
func (c *Controller) Tap(x, y int) error {
	if !c.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	sx, sy := c.ScaleCoord(x, y)
	args := []string{}
	if c.device != "" {
		args = append(args, "-s", c.device)
	}
	args = append(args, "shell", "input", "tap", fmt.Sprintf("%d", sx), fmt.Sprintf("%d", sy))

	return exec.CommandContext(context.Background(), c.adbPath, args...).Run()
}

// Swipe 滑动
func (c *Controller) Swipe(x1, y1, x2, y2 int, duration time.Duration) error {
	if !c.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	sx1, sy1 := c.ScaleCoord(x1, y1)
	sx2, sy2 := c.ScaleCoord(x2, y2)
	args := []string{}
	if c.device != "" {
		args = append(args, "-s", c.device)
	}
	args = append(args, "shell", "input", "swipe",
		fmt.Sprintf("%d", sx1), fmt.Sprintf("%d", sy1),
		fmt.Sprintf("%d", sx2), fmt.Sprintf("%d", sy2),
		fmt.Sprintf("%d", duration.Milliseconds()))

	return exec.CommandContext(context.Background(), c.adbPath, args...).Run()
}

// Back 返回
func (c *Controller) Back() error {
	return c.PressKey(4)
}

// Home 主页
func (c *Controller) Home() error {
	return c.PressKey(3)
}

// PressKey 按键
func (c *Controller) PressKey(keyCode int) error {
	if !c.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	args := []string{}
	if c.device != "" {
		args = append(args, "-s", c.device)
	}
	args = append(args, "shell", "input", "keyevent", fmt.Sprintf("%d", keyCode))

	return exec.CommandContext(context.Background(), c.adbPath, args...).Run()
}

// ScaleCoord 坐标缩放
func (c *Controller) ScaleCoord(x, y int) (int, int) {
	baseWidth := 1280.0
	baseHeight := 720.0
	scaleX := float64(c.config.ScreenWidth) / baseWidth
	scaleY := float64(c.config.ScreenHeight) / baseHeight
	return int(float64(x) * scaleX), int(float64(y) * scaleY)
}

// GetADBPath 获取 ADB 路径
func (c *Controller) GetADBPath() string {
	return c.adbPath
}

// SetADBPath 设置 ADB 路径
func (c *Controller) SetADBPath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.adbPath = path
}