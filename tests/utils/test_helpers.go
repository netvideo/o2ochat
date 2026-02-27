package utils

import (
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// CreateTestFile 创建测试文件
func CreateTestFile(t *testing.T, path string, size int64) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("创建目录失败：%v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("创建文件失败：%v", err)
	}
	defer file.Close()

	_, err = io.CopyN(file, rand.Reader, size)
	if err != nil {
		t.Fatalf("写入文件失败：%v", err)
	}
}

// CleanupTestDir 清理测试目录
func CleanupTestDir(t *testing.T, path string) {
	t.Helper()

	if err := os.RemoveAll(path); err != nil {
		t.Logf("清理目录失败：%v", err)
	}
}

// WaitForCondition 等待条件成立
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}

	return false
}

// GetFreePort 获取空闲端口
func GetFreePort(t *testing.T) int {
	t.Helper()

	// 简化实现，返回测试端口
	return 18080 + int(time.Now().Unix()%1000)
}

// CreateTestDirectory 创建临时测试目录
func CreateTestDirectory(t *testing.T, name string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", name)
	if err != nil {
		t.Fatalf("创建临时目录失败：%v", err)
	}

	return dir
}

// GenerateRandomBytes 生成随机字节
func GenerateRandomBytes(size int) []byte {
	bytes := make([]byte, size)
	rand.Read(bytes)
	return bytes
}
