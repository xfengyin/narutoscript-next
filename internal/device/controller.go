package device

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/xfengyin/narutoscript-next/internal/app"
	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

type Controller struct {
	config app.DeviceConfig
	log    *logger.Logger
	device string
}

func NewController(cfg app.DeviceConfig, log *logger.Logger) *Controller {
	return &Controller{
		config: cfg,
		log:    log,
	}
}

// Connect 连接设备
func (c *Controller) Connect() error {
	cmd := exec.Command(c.config.ADBPath, "devices")
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
				c.log.Info("已连接设备", "device", c.device)
				return nil
			}
		}
	}

	return fmt.Errorf("未找到可用设备")
}

// Screenshot 截图
func (c *Controller) Screenshot() ([]byte, error) {
	cmd := exec.Command(c.config.ADBPath, "-s", c.device, "exec-out", "screencap", "-p")
	return cmd.Output()
}

// Tap 点击
func (c *Controller) Tap(x, y int) error {
	cmd := exec.Command(c.config.ADBPath, "-s", c.device, "shell", "input", "tap",
		fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
	return cmd.Run()
}

// Swipe 滑动
func (c *Controller) Swipe(x1, y1, x2, y2 int, duration time.Duration) error {
	cmd := exec.Command(c.config.ADBPath, "-s", c.device, "shell", "input", "swipe",
		fmt.Sprintf("%d", x1), fmt.Sprintf("%d", y1),
		fmt.Sprintf("%d", x2), fmt.Sprintf("%d", y2),
		fmt.Sprintf("%d", duration.Milliseconds()))
	return cmd.Run()
}

// InputText 输入文字
func (c *Controller) InputText(text string) error {
	cmd := exec.Command(c.config.ADBPath, "-s", c.device, "shell", "input", "text", text)
	return cmd.Run()
}

// PressKey 按键
func (c *Controller) PressKey(keyCode int) error {
	cmd := exec.Command(c.config.ADBPath, "-s", c.device, "shell", "input", "keyevent",
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

// ScaleCoord 坐标缩放（适配不同分辨率）
func (c *Controller) ScaleCoord(x, y int) (int, int) {
	baseWidth := 1280.0
	baseHeight := 720.0
	
	scaleX := float64(c.config.ScreenWidth) / baseWidth
	scaleY := float64(c.config.ScreenHeight) / baseHeight
	
	return int(float64(x) * scaleX), int(float64(y) * scaleY)
}