package vision

import (
	"bytes"
	"image"
	_ "image/png"

	"github.com/xfengyin/narutoscript-next/pkg/logger"
)

type Matcher struct {
	log *logger.Logger
}

func NewMatcher(log *logger.Logger) *Matcher {
	return &Matcher{log: log}
}

// MatchTemplate 模板匹配
// 返回匹配位置和相似度
func (m *Matcher) MatchTemplate(screen, template []byte) (x, y int, confidence float64, err error) {
	// 解码图像
	screenImg, _, err := image.Decode(bytes.NewReader(screen))
	if err != nil {
		return 0, 0, 0, err
	}

	tmplImg, _, err := image.Decode(bytes.NewReader(template))
	if err != nil {
		return 0, 0, 0, err
	}

	// TODO: 实现实际的模板匹配算法
	// 这里可以使用 gocv 或纯 Go 实现
	screenBounds := screenImg.Bounds()
	tmplBounds := tmplImg.Bounds()

	// 简单返回中心位置
	return screenBounds.Dx()/2, screenBounds.Dy()/2, 0.95, nil
}

// FindAll 查找所有匹配位置
func (m *Matcher) FindAll(screen, template []byte, threshold float64) ([]image.Point, error) {
	// TODO: 实现多目标匹配
	return []image.Point{}, nil
}

// IsOnScreen 检查模板是否在屏幕上
func (m *Matcher) IsOnScreen(screen, template []byte, threshold float64) bool {
	_, _, confidence, err := m.MatchTemplate(screen, template)
	if err != nil {
		return false
	}
	return confidence >= threshold
}

// WaitForImage 等待图像出现
func (m *Matcher) WaitForImage(screenshotFunc func() ([]byte, error), template []byte, timeout int) (int, int, error) {
	// TODO: 实现等待逻辑
	return 0, 0, nil
}