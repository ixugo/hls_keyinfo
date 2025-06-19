package hlskeyinfo

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
)

var _ io.WriterTo = &KeyInfo{}

// KeyInfo HLS加密信息结构
type KeyInfo struct {
	URL      string // 密钥获取URL
	KeyFile  string // 密钥文件路径
	IV       string // 初始化向量
	key      []byte // 密钥字节数组（小写私有属性）
	infoFile string // 临时 keyinfo 文件路径（小写私有属性）
}

// NewKeyInfo 创建新的KeyInfo实例
func NewKeyInfo(url string) (*KeyInfo, error) {
	k := &KeyInfo{
		URL: url,
	}

	// 生成16字节的随机密钥
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("生成密钥失败: %w", err)
	}
	k.key = key

	// 在系统临时目录创建密钥文件
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "hls_key_*.bin")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer tempFile.Close()

	// 写入密钥到临时文件
	if _, err := tempFile.Write(key); err != nil {
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("写入密钥文件失败: %w", err)
	}

	k.KeyFile = tempFile.Name()

	return k, nil
}

// GetKey 获取密钥字节数组
func (k *KeyInfo) GetKey() []byte {
	if k.key == nil {
		return nil
	}
	return slices.Clone(k.key)
}

// SetIV 设置初始化向量
func (k *KeyInfo) SetIV(iv string) *KeyInfo {
	k.IV = iv
	return k
}

// SetKeyFile 设置密钥文件路径
func (k *KeyInfo) SetKeyFile(keyFile string) *KeyInfo {
	k.KeyFile = keyFile
	return k
}

// RandIV 生成随机初始化向量
func (k *KeyInfo) RandIV() *KeyInfo {
	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		// 如果生成失败，使用默认值
		k.IV = "00000000000000000000000000000000"
		return k
	}
	// 转换为十六进制字符串
	k.IV = fmt.Sprintf("%032x", iv)
	return k
}

// Dispose 清理临时文件
func (k *KeyInfo) Dispose() error {
	var errs []error

	// 清理密钥文件
	if k.KeyFile != "" {
		if err := os.Remove(k.KeyFile); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("删除临时密钥文件失败: %w", err))
		}
		k.KeyFile = ""
	}

	// 清理 keyinfo 文件
	if k.infoFile != "" {
		if err := os.Remove(k.infoFile); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("删除临时 keyinfo 文件失败: %w", err))
		}
		k.infoFile = ""
	}

	if len(errs) > 0 {
		return fmt.Errorf("清理临时文件时发生错误: %v", errs)
	}

	return nil
}

// WriteToTempFile 将 keyinfo 信息写入临时文件，返回临时文件路径
// 使用密钥的十六进制表示作为文件名，确保唯一性和可识别性
func (k *KeyInfo) WriteToTempFile() (string, error) {
	if k.key == nil {
		return "", fmt.Errorf("密钥未初始化")
	}

	// 使用密钥的十六进制表示作为文件名
	fileName := "hls_keyinfo_*.txt"
	filePath := filepath.Join(os.TempDir(), fileName)

	// 创建文件（如果已存在则覆盖）
	tempFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer tempFile.Close()

	// 写入 keyinfo 内容
	_, err = k.WriteTo(tempFile)
	if err != nil {
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}

	// 记录临时文件路径
	k.infoFile = filePath
	return filePath, nil
}

// WriteTo 实现io.WriterTo接口，按照ffmpeg hls_key_info_file格式写入三行数据
func (k *KeyInfo) WriteTo(w io.Writer) (n int64, err error) {
	// ffmpeg hls_key_info_file格式：
	// 第一行：密钥获取URL
	// 第二行：密钥文件路径
	// 第三行：初始化向量（可选）

	var written int64

	// 写入URL
	urlLine := k.URL + "\n"
	wrote, err := w.Write([]byte(urlLine))
	if err != nil {
		return written, fmt.Errorf("写入URL失败: %w", err)
	}
	written += int64(wrote)

	// 写入密钥文件路径
	keyFileLine := k.KeyFile + "\n"
	wrote, err = w.Write([]byte(keyFileLine))
	if err != nil {
		return written, fmt.Errorf("写入密钥文件路径失败: %w", err)
	}
	written += int64(wrote)

	// 写入IV（如果存在）
	if k.IV != "" {
		ivLine := k.IV + "\n"
		wrote, err = w.Write([]byte(ivLine))
		if err != nil {
			return written, fmt.Errorf("写入IV失败: %w", err)
		}
		written += int64(wrote)
	}

	return written, nil
}
