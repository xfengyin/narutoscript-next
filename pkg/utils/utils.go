package utils

import (
	"os"
	"path/filepath"
)

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// EnsureDir 确保目录存在
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// GetExeDir 获取可执行文件目录
func GetExeDir() string {
	exe, _ := os.Executable()
	return filepath.Dir(exe)
}

// Clamp 限制数值范围
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}