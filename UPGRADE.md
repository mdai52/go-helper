# Libgo 升级指南

本文档指导如何从旧版本升级到最新版本。

## 版本变更

### v0.14.0（当前版本）

本次重构涉及模块重命名、包重组、拆分和重命名，主要变更如下：

---

## 模块名变更

**模块名从 `pango` 改为 `libgo`**

```go
// 旧版本
import "github.com/rehiy/pango/logman"

// 新版本
import "github.com/rehiy/libgo/logman"
```

**所有导入路径需批量替换**：

```bash
# 批量替换
find . -name "*.go" -exec sed -i 's|github.com/rehiy/pango|github.com/rehiy/libgo|g' {} +
```

---

## 包导入路径变更

### 必须更新的导入路径

| 旧导入路径 | 新导入路径 | 说明 |
|-----------|-----------|------|
| `github.com/rehiy/pango/onquit` | `github.com/rehiy/libgo/signal` | 退出信号处理 |
| `github.com/rehiy/pango/socket` | `github.com/rehiy/libgo/websocket` | WebSocket 连接封装 |
| `github.com/rehiy/pango/socket` | `github.com/rehiy/libgo/tcprelay` | TCP 转发功能 |
| `github.com/rehiy/pango/cache` | `github.com/rehiy/libgo/ttlcache` | 带过期时间的缓存 |
| `github.com/rehiy/pango/alibaba` | `github.com/rehiy/libgo/cloud/alibaba` | 阿里云 API |
| `github.com/rehiy/pango/tencent` | `github.com/rehiy/libgo/cloud/tencent` | 腾讯云 API |
| `github.com/rehiy/pango/cloudflare` | `github.com/rehiy/libgo/cloud/cloudflare` | Cloudflare API |

### 删除的包

| 旧导入路径 | 替代方案 | 说明 |
|-----------|---------|------|
| `github.com/rehiy/pango/recovery` | `github.com/rehiy/libgo/httpd.Recovery()` | Gin panic 恢复中间件 |

---

## API 变更详情

### 1. signal 包（原 onquit）

```go
// 旧版本
import "github.com/rehiy/pango/onquit"
onquit.Register(func() { ... })
onquit.CallQuitFuncs()

// 新版本
import "github.com/rehiy/libgo/signal"
signal.OnQuit(func() { ... })
signal.CallQuitFuncs()
```

### 2. websocket 包（从 socket 拆分）

```go
// 旧版本
import "github.com/rehiy/pango/socket"
socket.WsConn{}
socket.NewWsClient(url, protocol, origin)
socket.PlainData{}

// 新版本
import "github.com/rehiy/libgo/websocket"
websocket.Conn{}
websocket.NewClient(url, protocol, origin)
websocket.Message{}  // PlainData 重命名为 Message
```

### 3. tcprelay 包（从 socket 拆分）

```go
// 旧版本
import "github.com/rehiy/pango/socket"
socket.TcpRelay(ws, &socket.TcpRelayParam{...})

// 新版本
import "github.com/rehiy/libgo/tcprelay"
tcprelay.Relay(ws, &tcprelay.Param{...})  // TcpRelayParam 重命名为 Param
```

### 4. ttlcache 包（原 cache）

```go
// 旧版本
import "github.com/rehiy/pango/cache"
cache.NewTimedCache(interval)

// 新版本
import "github.com/rehiy/libgo/ttlcache"
ttlcache.NewTimedCache(interval)  // API 不变
```

### 5. httpd 包（合并 recovery）

```go
// 旧版本
import "github.com/rehiy/pango/recovery"
func handler() {
    defer recovery.Handler()
    ...
}

// 新版本 - 作为 Gin 中间件使用
import "github.com/rehiy/libgo/httpd"
engine.Use(httpd.Recovery())

// 或在路由中单独使用
func handler(c *gin.Context) {
    defer httpd.Recovery()(c)
    ...
}
```

### 6. cloud 包（云服务整合）

```go
// 旧版本
import "github.com/rehiy/pango/alibaba"
import "github.com/rehiy/pango/tencent"
import "github.com/rehiy/pango/cloudflare"

// 新版本
import "github.com/rehiy/libgo/cloud/alibaba"
import "github.com/rehiy/libgo/cloud/tencent"
import "github.com/rehiy/libgo/cloud/cloudflare"

// API 不变，仅路径调整
alibaba.Request(&alibaba.ReqeustParam{...})
tencent.Request(&tencent.ReqeustParam{...})
cloudflare.Request(&cloudflare.ReqeustParam{...})
```

---

## 升级步骤

### 步骤 1：更新依赖版本

```bash
go get github.com/rehiy/libgo@v0.14.0
```

### 步骤 2：批量替换模块名

```bash
# 替换所有 Go 文件中的模块名
find . -name "*.go" -exec sed -i 's|github.com/rehiy/pango|github.com/rehiy/libgo|g' {} +

# 或使用 gofmt（推荐）
gofmt -w -r 'github.com/rehiy/pango -> github.com/rehiy/libgo' .
```

### 步骤 3：更新 go.mod

```bash
# 删除旧依赖
go mod edit -droprequire github.com/rehiy/pango

# 添加新依赖
go mod edit -require github.com/rehiy/libgo@v0.14.0

# 整理依赖
go mod tidy
```

### 步骤 4：处理包路径变更

```bash
# Linux/macOS - 批量替换旧包路径
find . -name "*.go" -exec sed -i '' \
  -e 's|github.com/rehiy/libgo/onquit|github.com/rehiy/libgo/signal|g' \
  -e 's|github.com/rehiy/libgo/cache|github.com/rehiy/libgo/ttlcache|g' \
  -e 's|github.com/rehiy/libgo/alibaba|github.com/rehiy/libgo/cloud/alibaba|g' \
  -e 's|github.com/rehiy/libgo/tencent|github.com/rehiy/libgo/cloud/tencent|g' \
  -e 's|github.com/rehiy/libgo/cloudflare|github.com/rehiy/libgo/cloud/cloudflare|g' \
  {} +
```

### 步骤 5：处理 socket 包拆分

socket 包拆分为 websocket 和 tcprelay，需手动处理：

```bash
# 查找使用 socket 包的文件
grep -r "github.com/rehiy/libgo/socket" . --include="*.go"
```

根据实际使用情况更新：
- 使用 `WsConn`、`NewWsClient`、`PlainData` → 改为 `websocket` 包
- 使用 `TcpRelay`、`TcpRelayParam` → 改为 `tcprelay` 包

### 步骤 6：处理 recovery 包删除

```bash
# 查找使用 recovery 包的文件
grep -r "github.com/rehiy/libgo/recovery" . --include="*.go"
```

删除导入，改用 `httpd.Recovery()` 中间件。

### 步骤 7：验证编译

```bash
go build ./...
go test ./...
```

---

## 类型重命名

| 旧类型名 | 新类型名 | 所在包 |
|---------|---------|--------|
| `socket.WsConn` | `websocket.Conn` | websocket |
| `socket.PlainData` | `websocket.Message` | websocket |
| `socket.TcpRelayParam` | `tcprelay.Param` | tcprelay |
| `alibaba.ReqeustParam` | `alibaba.RequestParam` | cloud/alibaba |
| `tencent.ReqeustParam` | `tencent.RequestParam` | cloud/tencent |
| `cloudflare.ReqeustParam` | `cloudflare.RequestParam` | cloud/cloudflare |
| `certmagic.ReqeustParam` | `certmagic.RequestParam` | certmagic |

---

## 字段重命名

以下结构体字段名称已规范化（遵循 Go 缩写命名规范）：

| 所在包 | 结构体 | 旧字段名 | 新字段名 |
|--------|--------|----------|----------|
| certify | Manager | DirectoryUrl | DirectoryURL |
| psutil | SummaryStat | PublicIpv4 | PublicIPv4 |
| psutil | SummaryStat | PublicIpv6 | PublicIPv6 |
| dborm | - | Db | DB |

---

## JSON 标签添加

以下结构体已添加 JSON 标签，确保序列化时字段名一致：

| 所在包 | 结构体 |
|--------|--------|
| websocket | Message |
| tcprelay | Param |
| certmagic | RequestParam, Certificate |
| cloud/alibaba | RequestParam |
| cloud/tencent | RequestParam |
| cloud/cloudflare | RequestParam |
| webssh | SSHClientOption |
| filer | FileInfo |
| psutil | GoMemoryStat, SummaryStat, DetailStat, DiskPartition, NetInterface |
| gpu | DeviceStat |
| dborm | Config |
| logman | Config |

---

## 兼容性说明

- **upgrade 包**：保持不变，API 未调整
- **logman 包**：保持不变，仅内部依赖调整
- **httpd 包**：新增 `Recovery()` 中间件，其他 API 不变
- **其他包**：API 保持兼容，仅路径调整

---

## 性能优化

### 并发安全改进

| 所在包 | 改进内容 |
|--------|----------|
| certmagic | `magicPool` 添加 `sync.RWMutex` 保护 |
| psutil | `publicIPv4/publicIPv6` 添加 `sync.RWMutex` 保护 |
| ttlcache | `Get()` 方法优化为读锁优先，减少锁竞争 |

### 内存优化

| 所在包 | 改进内容 |
|--------|----------|
| psutil | `InterfaceAddrs()` 预分配切片容量 |

### 逻辑优化

| 所在包 | 改进内容 |
|--------|----------|
| certify | 包重命名：`certman` → `certify` |
| certify | 新增 HTTP-01 验证支持 |
| certify | 新增 `ChallengeType` 配置项（`ChallengeDNS01`/`ChallengeHTTP01`） |
| certify | `GetCertificate()` 异步保存证书到缓存 |
| certify | `fulfill()` 支持上下文取消，避免资源泄漏 |
| httpd | `Recovery()` 添加请求方法和堆栈信息 |
| websocket | `Close()` 忽略已关闭错误 |
| logman | `replaceAttr` 提取为包级函数，减少闭包开销 |

---

## 完整替换脚本

```bash
#!/bin/bash

# 1. 替换模块名
find . -name "*.go" -exec sed -i 's|github.com/rehiy/pango|github.com/rehiy/libgo|g' {} +

# 2. 替换包路径
find . -name "*.go" -exec sed -i \
  -e 's|github.com/rehiy/libgo/onquit|github.com/rehiy/libgo/signal|g' \
  -e 's|github.com/rehiy/libgo/cache|github.com/rehiy/libgo/ttlcache|g' \
  -e 's|github.com/rehiy/libgo/alibaba|github.com/rehiy/libgo/cloud/alibaba|g' \
  -e 's|github.com/rehiy/libgo/tencent|github.com/rehiy/libgo/cloud/tencent|g' \
  -e 's|github.com/rehiy/libgo/cloudflare|github.com/rehiy/libgo/cloud/cloudflare|g' \
  {} +

# 3. 更新 go.mod
go mod edit -droprequire github.com/rehiy/pango
go mod edit -require github.com/rehiy/libgo@v0.14.0
go mod tidy

# 4. 验证编译
go build ./...

echo "升级完成！请检查 socket 包拆分和 recovery 包删除的手动处理。"
```

---

## 常见问题

### Q: 编译报错 "package not found"

检查 go.mod 中的版本是否已更新到 v0.14.0，并确认模块名已改为 `github.com/rehiy/libgo`。

### Q: socket 包找不到

socket 包已拆分，根据功能选择 websocket 或 tcprelay。

### Q: recovery.Handler() 找不到

recovery 包已删除，使用 `httpd.Recovery()` 作为 Gin 中间件。

### Q: 类型名 PlainData 找不到

已重命名为 `websocket.Message`。

### Q: 导入路径已替换但仍报错

确保同时更新了模块名（`pango` → `libgo`）和包路径（如 `onquit` → `signal`）。

---

## AI Agent 升级提示

当检测到项目使用旧版本 pango 时，按以下流程处理：

1. **识别旧模块名**：搜索 `github.com/rehiy/pango`
2. **批量替换模块名**：`pango` → `libgo`
3. **替换旧包路径**：
   - `onquit` → `signal`
   - `cache` → `ttlcache`
   - `alibaba` → `cloud/alibaba`
   - `tencent` → `cloud/tencent`
   - `cloudflare` → `cloud/cloudflare`
4. **手动处理 socket**：根据实际使用拆分为 websocket 或 tcprelay
5. **删除 recovery**：改用 httpd.Recovery() 中间件
6. **更新类型引用**：WsConn → Conn，PlainData → Message，TcpRelayParam → Param
7. **更新 go.mod**：删除旧依赖，添加新依赖
8. **验证编译**：执行 go build 确认无错误