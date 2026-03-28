package vision

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"sync"
	"time"

	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

// Matcher 图像匹配器
type Matcher struct {
	log       *logger.Logger
	templates map[string][]byte
	mu        sync.RWMutex
}

// MatchResult 匹配结果
type MatchResult struct {
	X          int
	Y          int
	Width      int
	Height     int
	Confidence float64
	Found      bool
}

// NewMatcher 创建匹配器
func NewMatcher(log *logger.Logger) *Matcher {
	return &Matcher{
		log:       log,
		templates: make(map[string][]byte),
	}
}

// LoadTemplates 从 embed.FS 加载模板
func (m *Matcher) LoadTemplates(fs embed.FS, dir string) error {
	// TODO: 实现从 embed.FS 加载模板
	return nil
}

// LoadTemplate 加载单个模板
func (m *Matcher) LoadTemplate(name string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates[name] = data
}

// GetTemplate 获取模板
func (m *Matcher) GetTemplate(name string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.templates[name]
	return data, ok
}

// MatchTemplate 模板匹配
func (m *Matcher) MatchTemplate(screen, template []byte, threshold float64) (*MatchResult, error) {
	// 解码图像
	screenImg, _, err := image.Decode(bytes.NewReader(screen))
	if err != nil {
		return nil, fmt.Errorf("解码屏幕图像失败: %w", err)
	}

	tmplImg, _, err := image.Decode(bytes.NewReader(template))
	if err != nil {
		return nil, fmt.Errorf("解码模板图像失败: %w", err)
	}

	screenBounds := screenImg.Bounds()
	tmplBounds := tmplImg.Bounds()

	// 检查模板是否大于屏幕
	if tmplBounds.Dx() > screenBounds.Dx() || tmplBounds.Dy() > screenBounds.Dy() {
		return &MatchResult{Found: false}, nil
	}

	// 实现简化的模板匹配（实际项目应使用 gocv 或专业库）
	result := m.simpleMatch(screenImg, tmplImg, threshold)

	return result, nil
}

// simpleMatch 简化的模板匹配实现
func (m *Matcher) simpleMatch(screen, template image.Image, threshold float64) *MatchResult {
	screenBounds := screen.Bounds()
	tmplBounds := template.Bounds()

	bestX, bestY := 0, 0
	bestScore := 0.0

	// 步长优化（降低精度提升速度）
	step := 4

	// 滑动窗口搜索
	for y := 0; y <= screenBounds.Dy()-tmplBounds.Dy(); y += step {
		for x := 0; x <= screenBounds.Dx()-tmplBounds.Dx(); x += step {
			score := m.compareRegion(screen, template, x, y)
			if score > bestScore {
				bestScore = score
				bestX = x
				bestY = y
			}
		}
	}

	return &MatchResult{
		X:          bestX,
		Y:          bestY,
		Width:      tmplBounds.Dx(),
		Height:     tmplBounds.Dy(),
		Confidence: bestScore,
		Found:      bestScore >= threshold,
	}
}

// compareRegion 比较区域相似度
func (m *Matcher) compareRegion(screen, template image.Image, offsetX, offsetY int) float64 {
	tmplBounds := template.Bounds()
	totalPixels := tmplBounds.Dx() * tmplBounds.Dy()
	matchedPixels := 0

	for y := 0; y < tmplBounds.Dy(); y += 2 { // 步长优化
		for x := 0; x < tmplBounds.Dx(); x += 2 {
			sx := offsetX + x
			sy := offsetY + y

			// 获取颜色
			sc := colorToGray(screen.At(sx, sy))
			tc := colorToGray(template.At(x, y))

			// 允许一定误差
			diff := abs(int(sc) - int(tc))
			if diff < 30 { // 容差阈值
				matchedPixels++
			}
		}
	}

	return float64(matchedPixels*4) / float64(totalPixels) // 乘以步长补偿
}

// colorToGray 颜色转灰度
func colorToGray(c color.Color) uint8 {
	r, g, b, _ := c.RGBA()
	return uint8((299*r + 587*g + 114*b) / 1000 >> 8)
}

// abs 绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// FindAll 查找所有匹配位置
func (m *Matcher) FindAll(screen, template []byte, threshold float64) ([]MatchResult, error) {
	// TODO: 实现多目标匹配
	return []MatchResult{}, nil
}

// IsOnScreen 检查模板是否在屏幕上
func (m *Matcher) IsOnScreen(screen, template []byte, threshold float64) bool {
	result, err := m.MatchTemplate(screen, template, threshold)
	if err != nil {
		return false
	}
	return result.Found
}

// WaitForImage 等待图像出现
func (m *Matcher) WaitForImage(
	ctx context.Context,
	screenshotFunc func() ([]byte, error),
	template []byte,
	threshold float64,
	timeout time.Duration,
) (*MatchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			screen, err := screenshotFunc()
			if err != nil {
				continue
			}

			result, err := m.MatchTemplate(screen, template, threshold)
			if err != nil {
				continue
			}

			if result.Found {
				return result, nil
			}
		}
	}
}

// WaitForImageDisappear 等待图像消失
func (m *Matcher) WaitForImageDisappear(
	ctx context.Context,
	screenshotFunc func() ([]byte, error),
	template []byte,
	threshold float64,
	timeout time.Duration,
) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			screen, err := screenshotFunc()
			if err != nil {
				continue
			}

			if !m.IsOnScreen(screen, template, threshold) {
				return nil
			}
		}
	}
}