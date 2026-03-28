package device

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/xfengyin/narutoscript-next/internal/app"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

// Controller 设备控制器
type Controller struct {
	config      app.DeviceConfig
	log         *logger.Logger
	device      string
	mu          sync.RWMutex
	connected   bool
	lastConnect time.Time
}

// NewController 创建设备控制器
func NewController(cfg app.DeviceConfig, log *logger.Logger) *Controller {
	return &Controller{
		config: cfg,
		log:    log,
	}
}

// Connect 连接设备
func (c *Controller) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 执行 ADB devices
	cmd := exec.CommandContext(context.Background(), c.config.ADBPath, "devices")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ADB 不可用: %w", err)
	}

	// 解析设备列表
	lines := bytes.Split(output, []byte("\n"))
	for i, line := range lines {
		if i > 0 && len(line) > 0 {
			fields := bytes.Fields(line)
			if len(fields) >= 2 && string(fields[1]) == "device" {
				c.device = string(fields[0])
				c.connected = true
				c.lastConnect = time.Now()
				c.log.Info("已连接设备", "device", c.device)
				return nil
			}
		}
	}

	return fmt.Errorf("未找到可用设备，请确保设备已连接并开启 USB 调试")
}

// Reconnect 重连设备
func (c *Controller) Reconnect() error {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	return c.Connect()
}

// IsConnected 检查是否已连接
func (c *Controller) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetDeviceName 获取设备名称
func (c *Controller) GetDeviceName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.device
}

// Screenshot 截图
func (c *Controller) Screenshot() ([]byte, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("设备未连接")
	}

	cmd := exec.CommandContext(context.Background(),
		c.config.ADBPath, "-s", c.device,
		"exec-out", "screencap", "-p")

	output, err := cmd.Output()
	if err != nil {
		c.log.Warn("截图失败", "error", err)
		return nil, err
	}

	return output, nil
}

// Tap 点击
func (c *Controller) Tap(x, y int) error {
	if !c.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	// 坐标缩放
	sx, sy := c.ScaleCoord(x, y)

	cmd := exec.CommandContext(context.Background(),
		c.config.ADBPath, "-s", c.device,
		"shell", "input", "tap",
		fmt.Sprintf("%d", sx), fmt.Sprintf("%d", sy))

	if err := cmd.Run(); err != nil {
		c.log.Warn("点击失败", "x", sx, "y", sy, "error", err)
		return err
	}

	c.log.Debug("点击", "x", sx, "y", sy)
	return nil
}

// Swipe 滑动
func (c *Controller) Swipe(x1, y1, x2, y2 int, duration time.Duration) error {
	if !c.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	sx1, sy1 := c.ScaleCoord(x1, y1)
	sx2, sy2 := c.ScaleCoord(x2, y2)

	cmd := exec.CommandContext(context.Background(),
		c.config.ADBPath, "-s", c.device,
		"shell", "input", "swipe",
		fmt.Sprintf("%d", sx1), fmt.Sprintf("%d", sy1),
		fmt.Sprintf("%d", sx2), fmt.Sprintf("%d", sy2),
		fmt.Sprintf("%d", duration.Milliseconds()))

	if err := cmd.Run(); err != nil {
		c.log.Warn("滑动失败", "error", err)
		return err
	}

	return nil
}

// LongPress 长按
func (c *Controller) LongPress(x, y int, duration time.Duration) error {
	return c.Swipe(x, y, x, y, duration)
}

// InputText 输入文字
func (c *Controller) InputText(text string) error {
	if !c.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	cmd := exec.CommandContext(context.Background(),
		c.config.ADBPath, "-s", c.device,
		"shell", "input", "text", text)

	return cmd.Run()
}

// PressKey 按键
func (c *Controller) PressKey(keyCode int) error {
	if !c.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	cmd := exec.CommandContext(context.Background(),
		c.config.ADBPath, "-s", c.device,
		"shell", "input", "keyevent",
		fmt.Sprintf("%d", keyCode))

	return cmd.Run()
}

// Back 返回
func (c *Controller) Back() error {
	return c.PressKey(4)
}

// Home 主页
func (c *Controller) Home() error {
	return c.PressKey(3)
}

// RecentApps 最近应用
func (c *Controller) RecentApps() error {
	return c.PressKey(187)
}

// PowerMenu 电源菜单
func (c *Controller) PowerMenu() error {
	return c.PressKey(26)
}

// VolumeUp 音量加
func (c *Controller) VolumeUp() error {
	return c.PressKey(24)
}

// VolumeDown 音量减
func (c *Controller) VolumeDown() error {
	return c.PressKey(25)
}

// ScaleCoord 坐标缩放（适配不同分辨率）
func (c *Controller) ScaleCoord(x, y int) (int, int) {
	baseWidth := 1280.0
	baseHeight := 720.0

	scaleX := float64(c.config.ScreenWidth) / baseWidth
	scaleY := float64(c.config.ScreenHeight) / baseHeight

	return int(float64(x) * scaleX), int(float64(y) * scaleY)
}

// ScaleBack 反向坐标缩放（从实际坐标转基准坐标）
func (c *Controller) ScaleBack(x, y int) (int, int) {
	baseWidth := 1280.0
	baseHeight := 720.0

	scaleX := float64(c.config.ScreenWidth) / baseWidth
	scaleY := float64(c.config.ScreenHeight) / baseHeight

	return int(float64(x) / scaleX), int(float64(y) / scaleY)
}

// StartApp 启动应用
func (c *Controller) StartApp(packageName string) error {
	if !c.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	cmd := exec.CommandContext(context.Background(),
		c.config.ADBPath, "-s", c.device,
		"shell", "monkey", "-p", packageName,
		"-c", "android.intent.category.LAUNCHER", "1")

	return cmd.Run()
}

// StopApp 停止应用
func (c *Controller) StopApp(packageName string) error {
	if !c.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	cmd := exec.CommandContext(context.Background(),
		c.config.ADBPath, "-s", c.device,
		"shell", "am", "force-stop", packageName)

	return cmd.Run()
}