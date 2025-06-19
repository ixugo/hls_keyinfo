package hlskeyinfo

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestKeyInfoExample(t *testing.T) {
	// 创建 KeyInfo 实例
	k, err := NewKeyInfo("http://localhost:4123/keyinfo")
	if err != nil {
		t.Fatalf("创建 KeyInfo 失败: %v", err)
	}
	defer k.Dispose()

	// 验证密钥已生成
	key := k.GetKey()
	if len(key) != 16 {
		t.Errorf("期望密钥长度为 16 字节，实际: %d", len(key))
	}

	// 验证临时文件已创建
	if k.KeyFile == "" {
		t.Error("密钥文件路径为空")
	}

	// 验证文件存在
	if _, err := os.Stat(k.KeyFile); os.IsNotExist(err) {
		t.Error("密钥文件不存在")
	}

	// 测试链式调用
	k.SetIV("12345678901234567890123456789012")

	// 测试 WriteTo
	var buf bytes.Buffer
	n, err := k.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo 失败: %v", err)
	}

	if n == 0 {
		t.Error("WriteTo 返回的字节数为 0")
	}

	// 验证输出格式
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 3 {
		t.Errorf("期望输出 3 行，实际: %d", len(lines))
	}

	if lines[0] != "http://localhost:4123/keyinfo" {
		t.Errorf("第一行不匹配，期望: http://localhost:4123/keyinfo，实际: %s", lines[0])
	}

	if lines[1] != k.KeyFile {
		t.Errorf("第二行不匹配，期望: %s，实际: %s", k.KeyFile, lines[1])
	}

	if lines[2] != "12345678901234567890123456789012" {
		t.Errorf("第三行不匹配，期望: 12345678901234567890123456789012，实际: %s", lines[2])
	}
}

func TestRandIV(t *testing.T) {
	k, err := NewKeyInfo("http://localhost:4123/keyinfo")
	if err != nil {
		t.Fatalf("创建 KeyInfo 失败: %v", err)
	}
	defer k.Dispose()

	// 测试随机 IV 生成
	k.RandIV()

	if len(k.IV) != 32 {
		t.Errorf("期望 IV 长度为 32 个字符，实际: %d", len(k.IV))
	}

	// 验证是否为十六进制
	for _, c := range k.IV {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("IV 包含非十六进制字符: %c", c)
		}
	}
}

func TestDispose(t *testing.T) {
	k, err := NewKeyInfo("http://localhost:4123/keyinfo")
	if err != nil {
		t.Fatalf("创建 KeyInfo 失败: %v", err)
	}

	keyFile := k.KeyFile

	// 验证文件存在
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		t.Error("密钥文件不存在")
	}

	// 清理文件
	err = k.Dispose()
	if err != nil {
		t.Fatalf("Dispose 失败: %v", err)
	}

	// 验证文件已删除
	if _, err := os.Stat(keyFile); !os.IsNotExist(err) {
		t.Error("密钥文件未被删除")
	}

	// 验证 KeyFile 字段已清空
	if k.KeyFile != "" {
		t.Error("KeyFile 字段未被清空")
	}
}

func TestWriteToFile(t *testing.T) {
	k, err := NewKeyInfo("http://localhost:4123/keyinfo")
	if err != nil {
		t.Fatalf("创建 KeyInfo 失败: %v", err)
	}
	defer k.Dispose()

	// 设置固定 IV 便于测试
	k.SetIV("abcdef1234567890abcdef1234567890")

	// 写入到临时文件
	tempFile := "test_keyinfo.txt"
	defer os.Remove(tempFile) // 测试完成后清理

	err = k.WriteToFile(tempFile)
	if err != nil {
		t.Fatalf("WriteToFile 失败: %v", err)
	}

	// 验证文件是否存在
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("文件未创建")
	}

	// 读取文件内容并验证
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Errorf("期望输出 3 行，实际: %d", len(lines))
	}

	if lines[0] != "http://localhost:4123/keyinfo" {
		t.Errorf("第一行不匹配，期望: http://localhost:4123/keyinfo，实际: %s", lines[0])
	}

	if lines[2] != "abcdef1234567890abcdef1234567890" {
		t.Errorf("第三行不匹配，期望: abcdef1234567890abcdef1234567890，实际: %s", lines[2])
	}

	// 测试覆盖写入
	k.SetIV("111222333444555666777888999000aa")
	err = k.WriteToFile(tempFile)
	if err != nil {
		t.Fatalf("第二次 WriteToFile 失败: %v", err)
	}

	// 验证文件被覆盖
	content, err = os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("读取覆盖后的文件失败: %v", err)
	}

	lines = strings.Split(strings.TrimSpace(string(content)), "\n")
	if lines[2] != "111222333444555666777888999000aa" {
		t.Errorf("覆盖后第三行不匹配，期望: 111222333444555666777888999000aa，实际: %s", lines[2])
	}
}
