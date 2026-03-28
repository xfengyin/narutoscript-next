package ocr

import (
	"bytes"
	"image"
	_ "image/png"
)

// Service OCR 服务
type Service struct {
	// 可以后续添加模型路径等配置
}

func NewService() *Service {
	return &Service{}
}

// Recognize 识别图片中的文字
func (s *Service) Recognize(img []byte) (string, error) {
	// TODO: 实现 OCR 识别
	// 可选方案：
	// 1. gosseract (Tesseract Go 绑定)
	// 2. PaddleOCR Go 绑定
	// 3. 云 OCR API
	
	return "", nil
}

// RecognizeNumber 识别数字（用于金币、铜币等）
func (s *Service) RecognizeNumber(img []byte) (int64, error) {
	text, err := s.Recognize(img)
	if err != nil {
		return 0, err
	}
	
	// TODO: 解析数字
	_ = text
	return 0, nil
}

// RecognizeRegion 识别图片指定区域
func (s *Service) RecognizeRegion(img []byte, x, y, w, h int) (string, error) {
	// 解码图片
	imgDecoded, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return "", err
	}
	
	// 裁剪区域
	bounds := imgDecoded.Bounds()
	if x < 0 || y < 0 || x+w > bounds.Dx() || y+h > bounds.Dy() {
		return "", nil
	}
	
	// TODO: 裁剪并识别
	_ = imgDecoded
	return "", nil
}