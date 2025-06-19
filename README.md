# HLS KeyInfo

HLS（HTTP Live Streaming）加密密钥信息文件生成器。

## 项目介绍

当使用 FFmpeg 生成 HLS 加密流时，需要通过 `-hls_key_info_file` 参数指定一个包含加密信息的文件。此库用于生成符合 FFmpeg 要求的 keyinfo 文件格式。

## 功能特性

- 🔑 自动生成 16 字节随机密钥
- 📁 自动创建临时密钥文件
- 🔄 支持链式调用设置参数
- 🗑️ 自动清理临时文件
- ✍️ 实现 `io.WriterTo` 接口

## 安装

```bash
go get github.com/ixugo/hls_keyinfo
```

## 使用方法

### 基本用法

```go
package main

import (
    "os"
    "github.com/ixugo/hls_keyinfo"
)

func main() {
    // 创建 KeyInfo 实例
    k, err := hlskeyinfo.NewKeyInfo("http://localhost:4123/keyinfo")
    if err != nil {
        panic(err)
    }
    defer k.Dispose() // 清理临时文件

    // 方法1: 使用 WriteTo 写入到文件
    f, err := os.OpenFile("keyinfo.txt", os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(err)
    }
    defer f.Close()
    _, err = k.WriteTo(f)
    if err != nil {
        panic(err)
    }

    // 方法2: 使用 WriteToTempFile 写入临时文件（推荐）
    tempFilePath, err := k.WriteToTempFile()
    if err != nil {
        panic(err)
    }
    defer os.Remove(tempFilePath) // 清理临时文件
}
```

### 链式调用设置参数

```go
k, err := hlskeyinfo.NewKeyInfo("http://localhost:4123/keyinfo")
if err != nil {
    panic(err)
}
defer k.Dispose()

// 链式调用设置 IV
k.SetIV("12345678901234567890123456789012").WriteTo(writer)

// 或者生成随机 IV
k.RandIV().WriteTo(writer)

// 自定义密钥文件路径
k.SetKeyFile("/path/to/custom/key.bin").WriteTo(writer)
```

### 获取密钥

```go
k, _ := hlskeyinfo.NewKeyInfo("http://localhost:4123/keyinfo")
defer k.Dispose()

// 获取密钥字节数组
key := k.GetKey()
fmt.Printf("密钥: %x\n", key)
```

## KeyInfo 文件格式

生成的 keyinfo 文件包含三行内容：

```
http://localhost:4123/keyinfo
/tmp/hls_key_1234567890.bin
12345678901234567890123456789012
```

1. **第一行**: 密钥获取 URL
2. **第二行**: 密钥文件的本地路径
3. **第三行**: 初始化向量 (IV)，可选

## 文件命名规则

- **密钥文件**: `hls_key_*.bin` - 存储实际的 16 字节密钥，`*` 为随机数字
- **keyinfo 文件**: `hls_keyinfo_*.txt` - 存储 keyinfo 配置信息

## API 文档

### 函数

#### `NewKeyInfo(url string) (*KeyInfo, error)`
创建新的 KeyInfo 实例，自动生成随机密钥并创建临时密钥文件。

#### `GetKey() []byte`
获取密钥字节数组的副本。

#### `SetIV(iv string) *KeyInfo`
设置初始化向量，返回自身以支持链式调用。

#### `SetKeyFile(keyFile string) *KeyInfo`
设置密钥文件路径，返回自身以支持链式调用。

#### `RandIV() *KeyInfo`
生成随机初始化向量，返回自身以支持链式调用。

#### `Dispose() error`
清理临时密钥文件。

#### `WriteTo(w io.Writer) (n int64, err error)`
实现 `io.WriterTo` 接口，将 keyinfo 内容写入到 Writer。

#### `WriteToTempFile() (string, error)`
将 keyinfo 信息写入临时文件，返回临时文件路径。

## FFmpeg 集成示例

```bash
# 使用生成的 keyinfo 文件进行 HLS 加密
ffmpeg -i input.mp4 \
    -hls_key_info_file keyinfo.txt \
    -hls_time 10 \
    -hls_list_size 0 \
    -hls_segment_filename segment_%d.ts \
    playlist.m3u8
```

## 许可证

MIT License
