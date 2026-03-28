package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Service OCR 服务
type Service struct {
	apiKey     string
	apiUrl     string
	useCloud   bool
	mu         sync.Mutex
	lastResult string
}

// NewService 创建 OCR 服务
func NewService() *Service {
	return &Service{
		// 默认使用百度 OCR API（需要配置 API Key）
		apiUrl: "https://aip.baidubce.com/rest/2.0/ocr/v1/general_basic",
		useCloud: false,
	}
}

// SetAPIKey 设置 API Key
func (s *Service) SetAPIKey(key, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apiKey = key
	s.apiUrl = url
	s.useCloud = true
}

// Recognize 识别图片中的文字
func (s *Service) Recognize(img []byte) (string, error) {
	// 如果启用了云 OCR
	if s.useCloud && s.apiKey != "" {
		return s.recognizeCloud(img)
	}

	// 否则使用本地简化方案
	return s.recognizeLocal(img)
}

// recognizeCloud 云 OCR 识别
func (s *Service) recognizeCloud(img []byte) (string, error) {
	// 解码图片获取尺寸
	imgDecoded, format, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return "", fmt.Errorf("解码图片失败: %w", err)
	}

	// 确保 PNG 格式
	if format != "png" {
		buf := &bytes.Buffer{}
		if err := png.Encode(buf, imgDecoded); err != nil {
			return "", err
		}
		img = buf.Bytes()
	}

	// 构建请求
	reqBody := map[string]string{
		"image": encodeBase64(img),
	}
	jsonBody, _ := json.Marshal(reqBody)

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s?access_token=%s", s.apiUrl, s.apiKey)

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("OCR API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析响应
	var result OCRResult
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.ErrorCode != 0 {
		return "", fmt.Errorf("OCR 错误: %s", result.ErrorMsg)
	}

	// 合并所有文字
	texts := make([]string, 0, len(result.WordsResult))
	for _, word := range result.WordsResult {
		texts = append(texts, word.Words)
	}

	return strings.Join(texts, ""), nil
}

// recognizeLocal 本地简化识别（用于数字）
func (s *Service) recognizeLocal(img []byte) (string, error) {
	// 解码图片
	imgDecoded, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return "", err
	}

	bounds := imgDecoded.Bounds()

	// 简化：分析图片中的数字区域
	// 这是一个非常简化的实现，实际项目中应使用专业 OCR 库

	// 统计像素亮度分布
	var brightPixels, darkPixels int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := imgDecoded.At(x, y)
			r, g, b, _ := c.RGBA()
			brightness := (r + g + b) / 3
			if brightness > 30000 {
				brightPixels++
			} else {
				darkPixels++
			}
		}
	}

	// 返回占位符
	// 实际项目中应该：
	// 1. 使用 gosseract (Tesseract Go 绑定)
	// 2. 或使用 PaddleOCR
	// 3. 或使用云 OCR API

	_ = brightPixels
	_ = darkPixels

	return "", nil
}

// RecognizeNumber 识别数字（用于金币、铜币等）
func (s *Service) RecognizeNumber(img []byte) (int64, error) {
	text, err := s.Recognize(img)
	if err != nil {
		return 0, err
	}

	// 清理文本，提取数字
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, ",", "")

	// 处理特殊格式（如 "1.5k", "2M"）
	if strings.HasSuffix(text, "k") || strings.HasSuffix(text, "K") {
		numStr := strings.TrimSuffix(text, "k")
		numStr = strings.TrimSuffix(numStr, "K")
		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, err
		}
		return int64(num * 1000), nil
	}

	if strings.HasSuffix(text, "m") || strings.HasSuffix(text, "M") {
		numStr := strings.TrimSuffix(text, "m")
		numStr = strings.TrimSuffix(numStr, "M")
		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, err
		}
		return int64(num * 1000000), nil
	}

	// 直接解析
	num, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("无法解析数字: %s", text)
	}

	return num, nil
}

// RecognizeRegion 识别图片指定区域
func (s *Service) RecognizeRegion(img []byte, x, y, w, h int) (string, error) {
	// 解码图片
	imgDecoded, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return "", err
	}

	bounds := imgDecoded.Bounds()
	if x < 0 || y < 0 || x+w > bounds.Dx() || y+h > bounds.Dy() {
		return "", fmt.Errorf("区域超出图片边界")
	}

	// 裁剪区域
	subImg := imgDecoded.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(x, y, x+w, y+h))

	// 编码裁剪后的图片
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, subImg); err != nil {
		return "", err
	}

	return s.Recognize(buf.Bytes())
}

// RecognizeRegionNumber 识别区域数字
func (s *Service) RecognizeRegionNumber(img []byte, x, y, w, h int) (int64, error) {
	text, err := s.RecognizeRegion(img, x, y, w, h)
	if err != nil {
		return 0, err
	}
	return s.parseNumber(text)
}

// parseNumber 解析数字
func (s *Service) parseNumber(text string) (int64, error) {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, ",", "")

	num, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}

// encodeBase64 Base64 编码
func encodeBase64(data []byte) string {
	return fmt.Sprintf("data:image/png;base64,%s", encodeBytes(data))
}

func encodeBytes(data []byte) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	result := make([]byte, 0, (len(data)+2)/3*4)

	for i := 0; i < len(data); i += 3 {
		remaining := len(data) - i
		var n uint32
		var pad int

		switch remaining {
		case 1:
			n = uint32(data[i]) << 16
			pad = 2
		case 2:
			n = uint32(data[i]) << 16 | uint32(data[i+1]) << 8
			pad = 1
		default:
			n = uint32(data[i]) << 16 | uint32(data[i+1]) << 8 | uint32(data[i+2])
			pad = 0
		}

		result = append(result,
			base64Chars[n>>18&0x3F],
			base64Chars[n>>12&0x3F],
			base64Chars[n>>6&0x3F],
			base64Chars[n&0x3F],
		)

		for j := 0; j < pad; j++ {
			result[len(result)-1-j] = '='
		}
	}

	return string(result)
}

// OCRResult OCR API 响应结构
type OCRResult struct {
	WordsResultNum int          `json:"words_result_num"`
	WordsResult    []OCRWord    `json:"words_result"`
	ErrorCode      int          `json:"error_code"`
	ErrorMsg       string       `json:"error_msg"`
}

// OCRWord 单个识别结果
type OCRWord struct {
	Words string `json:"words"`
}