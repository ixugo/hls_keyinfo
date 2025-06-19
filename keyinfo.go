package hlskeyinfo

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"slices"
)

var _ io.WriterTo = &KeyInfo{}

// KeyInfo HLS加密信息结构
type KeyInfo struct {
	URL     string // 密钥获取URL
	KeyFile string // 密钥文件路径
	IV      string // 初始化向量
	key     []byte // 密钥字节数组（小写私有属性）
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
	if k.KeyFile != "" {
		if err := os.Remove(k.KeyFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除临时密钥文件失败: %w", err)
		}
		k.KeyFile = ""
	}
	return nil
}

// WriteToFile 直接写入到指定文件路径，强制覆盖
func (k *KeyInfo) WriteToFile(filePath string) error {
	// 创建或覆盖文件
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入内容
	_, err = k.WriteTo(file)
	return err
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
