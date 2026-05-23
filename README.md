# Libgo - Go 辅助类库

一个功能丰富、模块化的 Go 语言辅助类库，封装了常用功能和第三方服务接口，帮助开发者快速构建应用程序。

## 特性

- 🚀 **开箱即用** - 简洁直观的 API 设计
- 🧩 **模块化** - 按需引入，无冗余依赖
- 📦 **功能丰富** - 涵盖数据库、日志、加密、系统监控、网络通信等
- ☁️ **云服务集成** - 支持阿里云、腾讯云、Cloudflare DNS API
- 📝 **代码完善** - 详细注释和类型定义

## 安装

```shell
go get -u github.com/rehiy/libgo
```

## 文档

详细 API 文档：[pkg.go.dev](https://pkg.go.dev/github.com/rehiy/libgo)

升级指南：[UPGRADE.md](./UPGRADE.md)

## 模块概览

### 云服务集成

| 模块 | 说明 |
|------|------|
| `cloud/alibaba` | 里云 API 客户端，支持 RPC 风格接口调用 |
| `cloud/tencent` | 腾讯云 API 客户端集成 |
| `cloud/cloudflare` | Cloudflare API 集成 |

### 数据库操作

| 模块 | 说明 |
|------|------|
| `dborm` | GORM 数据库封装，支持 MySQL、PostgreSQL、SQLite |

### 安全与加密

| 模块 | 说明 |
|------|------|
| `secure` | 加密工具集：Bcrypt 密码哈希、DES3 加密、MD5、SSH 密钥生成 |
| `certify` | SSL 证书自动管理（支持 DNS-01/HTTP-01 验证，通配符域名） |

### 网络与通信

| 模块 | 说明 |
|------|------|
| `request` | HTTP 客户端，支持 JSON/表单请求、文件下载 |
| `httpd` | HTTP 服务器（基于 Gin），含 Recovery 中间件、静态文件服务 |
| `websocket` | WebSocket 客户端/服务器封装 |
| `webssh` | SSH 客户端，支持密码和私钥认证 |
| `tcprelay` | TCP 数据转发 |

### 系统监控

| 模块 | 说明 |
|------|------|
| `psutil` | 系统信息采集：CPU、内存、硬盘、网络、主机信息 |
| `gpu` | GPU 信息采集，支持 NVIDIA、AMD、Intel、Apple GPU |
| `command` | 系统命令执行工具，支持 Shell 和直接执行 |

### 日志与文件

| 模块 | 说明 |
|------|------|
| `logman` | 结构化日志管理（基于 slog），支持日志轮转 |
| `filer` | 文件操作：读写、列表、权限管理、嵌入文件支持 |
| `archive` | 归档工具：ZIP 压缩/解压 |

### 工具函数

| 模块 | 说明 |
|------|------|
| `strutil` | 字符串处理：随机生成、编码转换、文本清理 |
| `ttlcache` | 带过期时间的内存缓存 |
| `signal` | 程序退出信号处理（SIGTERM/SIGINT） |
| `upgrade` | 应用自更新服务 |

---

## 使用示例

### 数据库连接

```go
import "github.com/rehiy/libgo/dborm"

db := dborm.Connect(&dborm.Config{
    Type:     "mysql",      // 支持 mysql、pgsql、sqlite
    Host:     "localhost:3306",
    User:     "root",
    Password: "password",
    DbName:   "mydb",
})

// 使用 GORM 操作
db.Create(&User{Name: "test"})
```

### HTTP 请求

```go
import "github.com/rehiy/libgo/request"

client := &request.Client{
    Method:  "POST",
    Url:     "https://api.example.com/data",
    Data:    `{"key": "value"}`,
    Headers: map[string]string{"Authorization": "Bearer token"},
    Timeout: 10 * time.Second,
}

// JSON 请求（自动设置 Content-Type）
body, err := client.JsonRequest()

// 表单请求
text, err := client.TextRequest()
```

### 日志记录

```go
import "github.com/rehiy/libgo/logman"

// 创建命名日志器
logger := logman.Named("myapp")
logger.Info("Application started")
logger.Error("Operation failed", "error", err, "code", 500)

// 格式化输出
logger.Infof("User %s logged in", username)

// 设置输出文件和日志级别
logman.SetOutput("/var/log/myapp.log")
logman.SetLevel(logman.LevelDebug)
```

### 系统监控

```go
import "github.com/rehiy/libgo/psutil"

// 系统概要信息
stat := psutil.Summary(true)  // true 表示获取公网 IP
fmt.Printf("主机: %s\nCPU: %.2f%%\n内存: %d/%d MB\n",
    stat.HostName,
    stat.CpuPercent[0],
    stat.MemoryUsed/1024/1024,
    stat.MemoryTotal/1024/1024)

// 系统详细信息（网络、磁盘分区等）
detail := psutil.Detail(true)
for _, iface := range detail.NetInterface {
    fmt.Printf("网卡 %s: 接收 %d 字节, 发送 %d 字节\n",
        iface.Name, iface.BytesRecv, iface.BytesSent)
}

// Go 运行时内存统计
memStat := psutil.GoMemory()
fmt.Printf("堆内存: %d MB, GC次数: %d\n",
    memStat.HeapInuse/1024/1024, memStat.NumGC)
```

### GPU 监控

```go
import "github.com/rehiy/libgo/gpu"

stats, err := gpu.GetGPUStats(ctx)
for _, stat := range stats {
    fmt.Printf("%s (%s): 利用率 %.1f%%, 显存 %d/%d MB, 温度 %d°C\n",
        stat.Name, stat.Vendor,
        stat.Utilization,
        stat.MemoryUsed/1024/1024, stat.MemoryTotal/1024/1024,
        stat.Temperature)
}
```

### SSH 连接

```go
import "github.com/rehiy/libgo/webssh"

// 密码认证
client, err := webssh.NewSSHClient(&webssh.SSHClientOption{
    Addr:     "example.com:22",
    User:     "username",
    Password: "password",
})

// 私钥认证
client, err := webssh.NewSSHClient(&webssh.SSHClientOption{
    Addr:       "example.com:22",
    User:       "username",
    PrivateKey: "-----BEGIN RSA PRIVATE KEY-----...",
})
```

### WebSocket 客户端

```go
import "github.com/rehiy/libgo/websocket"

conn, err := websocket.NewClient("ws://example.com/ws", "", "http://example.com")
if err != nil {
    panic(err)
}

// 写入数据
conn.Write([]byte("hello"))

// 读取数据
buf := make([]byte, 1024)
n, err := conn.Read(buf)

// JSON 消息
conn.WriteJSON(map[string]string{"type": "ping"})
var msg Message
conn.ReadJSON(&msg)

conn.Close()
```

### HTTP 服务器

```go
import "github.com/rehiy/libgo/httpd"

// 创建路由引擎
httpd.Engine(false)  // false 为非调试模式

// 添加路由
httpd.GET("/api/hello", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "hello"})
})

// 启动服务器（自动注册退出信号处理）
httpd.Server(":8080",
    httpd.WithReadTimeout(30*time.Second),
    httpd.WithWriteTimeout(30*time.Second),
)
```

### 退出信号处理

```go
import "github.com/rehiy/libgo/signal"

// 注册退出回调（支持多个）
signal.OnQuit(func() {
    fmt.Println("清理资源...")
    db.Close()
})

signal.OnQuit(func() {
    fmt.Println("保存状态...")
})
```

### SSL 证书管理

```go
import "github.com/rehiy/libgo/certify"

// 创建证书管理器
manager := certify.New("admin@example.com", certify.DirCache("/var/lib/certs"))

// 指定 ACME 目录（如使用 Buypass）
manager = certify.NewWithDirectory("admin@example.com", 
    certify.DirCache("/var/lib/certs"),
    certify.Buypass)

// 获取证书（自动申请/续期）
cert, err := manager.GetCertificateForDomains([]string{"example.com", "*.example.com"})

// TLS 配置（用于 HTTP 服务器）
tlsConfig := manager.TLSConfig()

// HTTP-01 验证处理器
handler := manager.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("ACME verification"))
}))
```

### 文件操作

```go
import "github.com/rehiy/libgo/filer"

// 列出目录文件
files, err := filer.List("/path/to/dir")
for _, f := range files {
    fmt.Printf("%s: %d bytes, owner: %s\n", f.Name, f.Size, f.Owner)
}

// 获取文件信息（含内容）
info, err := filer.Info("/path/to/file", true)
fmt.Printf("内容: %s\n", string(info.Data))

// 写入文件（自动创建目录）
filer.Write("/path/to/file.txt", []byte("content"))

// 追加内容
filer.Append("/path/to/log.txt", []byte("new line\n"))
```

### 密码哈希

```go
import "github.com/rehiy/libgo/secure"

// 生成密码哈希
hash, err := secure.BcryptHash("password123")

// 验证密码
valid := secure.BcryptVerify("password123", hash)

// 检查是否为 bcrypt 格式
isBcrypt := secure.IsBcrypt(hash)
```

### 随机字符串

```go
import "github.com/rehiy/libgo/strutil"

// 生成随机字符串（字母数字混合）
randStr := strutil.Rand(16)  // 如 "aB3xY9zK2mN7pQ1"

// 生成 UUID v7（时间排序）
uuid := strutil.NewString()
```

---

## 依赖项

| 库 | 用途 |
|----|------|
| [Gin](https://github.com/gin-gonic/gin) | Web 框架 |
| [GORM](https://gorm.io/) | ORM 库 |
| [gopsutil](https://github.com/shirou/gopsutil) | 系统监控 |
| [ghw](https://github.com/jaypipes/ghw) | GPU 硬件信息 |
| [Lumberjack](https://github.com/natefinch/lumberjack) | 日志轮转 |
| [libdns](https://github.com/libdns/libdns) | DNS Provider 抽象 |

---

## 版本要求

- Go 1.25+

---

## 许可证

Copyright (c) 2022 - 2026 OpenTDP