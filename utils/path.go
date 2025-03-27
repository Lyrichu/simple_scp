package utils

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// ParsePath 解析路径，返回主机名（如果有）和实际路径
func ParsePath(path string) (host string, actualPath string, err error) {
    if strings.Contains(path, ":") {
        parts := strings.SplitN(path, ":", 2)
        if len(parts) != 2 {
            return "", "", fmt.Errorf("invalid path format: %s", path)
        }
        return parts[0], parts[1], nil
    }
    return "", path, nil
}

// ValidatePath 验证本地路径是否有效
func ValidatePath(path string) error {
    // 检查路径是否存在
    _, err := os.Stat(path)
    if err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("path does not exist: %s", path)
        }
        return fmt.Errorf("error accessing path: %s", err)
    }
    return nil
}

// EnsureDirectory 确保目录存在，如果不存在则创建
func EnsureDirectory(path string) error {
    dir := filepath.Dir(path)
    return os.MkdirAll(dir, 0755)
}

// IsRemotePath 检查是否为远程路径
func IsRemotePath(path string) bool {
    return strings.Contains(path, ":")
}

// GetFileSize 获取文件大小
func GetFileSize(path string) (int64, error) {
    info, err := os.Stat(path)
    if err != nil {
        return 0, err
    }
    return info.Size(), nil
}