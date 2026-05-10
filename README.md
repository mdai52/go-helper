# Libgo - Go 辅助类库

一个功能丰富、开箱即用的 Go 语言辅助类库，封装了常用的功能和第三方服务接口，帮助开发者快速构建应用程序。

## 特性

- 🚀 开箱即用 - 简单易用的 API 设计
- 🔧 模块化设计 - 按需引入，减少依赖
- 📦 丰富的功能 - 涵盖数据库、日志、加密、系统监控等
- ☁️ 云服务集成 - 支持阿里云、腾讯云、Cloudflare 等
- 📝 完善的文档 - 详细的代码注释和使用示例

## 安装

```shell
go get -u github.com/rehiy/libgo
```

## 文档

详细文档请访问：[pkg.go.dev](https://pkg.go.dev/github.com/rehiy/libgo)

升级指南请查看：[UPGRADE.md](./UPGRADE.md)

## 模块概览

### 🔌 云服务集成

| 模块 | 说明 |
|------|------|
| `cloud/alibaba` | 阿里云 API 集成，支持 RPC 风格的接口调用 |
| `cloud/tencent` | 腾讯云 API 客户端集成 |
| `cloud/cloudflare` | Cloudflare API 集成 |

### 💾 数据库操作

| 模块 | 说明 |
|------|------|
| `dborm` | GORM 数据库操作封装，支持 MySQL、PostgreSQL、SQLite |

### 🔐 安全与加密

| 模块 | 说明 |
|------|------|
| `secure` | 加密解密工具，支持 DES3、MD5、SSH 密钥生成 |
| `certmagic` | SSL 证书自动化管理（基于 Certmagic） |
| `certman` | 证书管理与缓存 |

### 🌐 网络与通信

| 模块 | 说明 |
|------|------|
| `request` | 简洁的 HTTP 客户端，支持 JSON 和表单请求 |
| `websocket` | WebSocket 连接封装 |
| `tcprelay` | TCP 数据转发 |
| `webssh` | SSH 客户端封装，支持密码和密钥认证 |
| `httpd` | HTTP 服务器工具（含 Recovery 中间件） |

### 📊 系统监控

| 模块 | 说明 |
|------|------|
| `psutil` | 系统信息采集，包括 CPU、内存、硬盘、网络等 |
| `gpu` | GPU 信息采集，支持 NVIDIA、AMD、Intel |
| `command` | 系统命令执行工具 |

### 📝 日志与文件

| 模块 | 说明 |
|------|------|
| `logman` | 结构化日志管理，基于 slog |
| `filer` | 文件操作工具，支持读写、列表、权限管理等 |
| `archive` | 归档工具，支持 zip 压缩/解压 |

### 🛠️ 工具函数

| 模块 | 说明 |
|------|------|
| `strutil` | 字符串处理工具集 |
| `ttlcache` | 带过期时间的缓存工具 |
| `signal` | 程序退出信号处理 |
| `upgrade` | 应用自更新服务 |

## 使用示例

### 数据库连接

```go
import "github.com/rehiy/libgo/dborm"

db := dborm.Connect(&dborm.Config{
    Type:     "mysql",
    Host:     "localhost:3306",
    User:     "root",
    Password: "password",
    DbName:   "mydb",
})
```

### HTTP 请求

```go
import "github.com/rehiy/libgo/request"

client := &request.Client{
    Method:  "GET",
    Url:     "https://api.example.com/data",
    Headers: map[string]string{"Authorization": "Bearer token"},
    Timeout: 10 * time.Second,
}
body, _ := client.JsonRequest()
```

### 日志记录

```go
import "github.com/rehiy/libgo/logman"

logger := logman.Named("myapp")
logger.Info("Application started")
logger.Error("Something went wrong", "error", err)
```

### 系统监控

```go
import "github.com/rehiy/libgo/psutil"

// 获取系统摘要信息
stat := psutil.Summary(true)
fmt.Printf("CPU: %.2f%%, Memory: %d/%d MB\n",
    stat.CpuPercent[0],
    stat.MemoryUsed/1024/1024,
    stat.MemoryTotal/1024/1024)
```

### GPU 监控

```go
import "github.com/rehiy/libgo/gpu"

// 获取 GPU 统计信息
stats, _ := gpu.GetGPUStats(ctx)
for _, stat := range stats {
    fmt.Printf("%s: %.1f%%, %d/%d MB\n",
        stat.Name, stat.Utilization,
        stat.MemoryUsed/1024/1024, stat.MemoryTotal/1024/1024)
}
```

### SSH 连接

```go
import "github.com/rehiy/libgo/webssh"

client, _ := webssh.NewSSHClient(&webssh.SSHClientOption{
    Addr:     "example.com:22",
    User:     "username",
    Password: "password",
})
```

### WebSocket 客户端

```go
import "github.com/rehiy/libgo/websocket"

conn, _ := websocket.NewClient("ws://example.com/ws", "", "http://example.com")
conn.Write([]byte("hello"))
```

### 退出信号处理

```go
import "github.com/rehiy/libgo/signal"

signal.OnQuit(func() {
    fmt.Println("Application exiting...")
    // 清理资源
})
```

### SSL 证书管理

```go
import "github.com/rehiy/libgo/certmagic"

err := certmagic.Manage(&certmagic.ReqeustParam{
    Domain:     "example.com,www.example.com",
    Email:      "admin@example.com",
    SecretKey:  "cloudflare_api_key",
    CaType:     "cloudflare",
    StoragePath: "/var/lib/certmagic",
})
```

## 依赖项

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [GORM](https://gorm.io/) - ORM 库
- [gopsutil](https://github.com/shirou/gopsutil) - 系统监控
- [ghw](https://github.com/jaypipes/ghw) - GPU 信息采集
- [Certmagic](https://github.com/caddyserver/certmagic) - SSL 证书管理
- [Lumberjack](https://github.com/natefinch/lumberjack) - 日志轮转

## 版本要求

- Go 1.25+

## 许可证

Copyright (c) 2022 - 2026 OpenTDP