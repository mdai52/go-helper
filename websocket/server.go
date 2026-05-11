package websocket

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

// ServerConn WebSocket 服务端连接
type ServerConn struct {
	*websocket.Conn
}

// Handler 创建 WebSocket 处理器（gin HandlerFunc）
// 使用方式：router.GET("/ws", config.Handler(func(conn *ServerConn) { ... }))
func (c *ServerConfig) Handler(handler func(*ServerConn)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 检查 Origin
		if !c.CheckOrigin(ctx.GetHeader("Origin")) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		// 升级 WebSocket
		websocket.Handler(func(ws *websocket.Conn) {
			conn := &ServerConn{ws}
			handler(conn)
		}).ServeHTTP(ctx.Writer, ctx.Request)
	}
}

// Read 读取消息
func (c *ServerConn) Read(v []byte) error {
	return websocket.Message.Receive(c.Conn, v)
}

// ReadJSON 读取 JSON 消息
func (c *ServerConn) ReadJSON(v any) error {
	return websocket.JSON.Receive(c.Conn, v)
}

// Write 写入消息
func (c *ServerConn) Write(p []byte) error {
	return websocket.Message.Send(c.Conn, p)
}

// WriteJSON 写入 JSON 消息
func (c *ServerConn) WriteJSON(v any) error {
	return websocket.JSON.Send(c.Conn, v)
}

// Close 关闭连接，忽略 EOF 错误
func (c *ServerConn) Close() error {
	err := c.Conn.Close()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// Die 发送消息并关闭连接
func (c *ServerConn) Die(msg string) {
	_, _ = c.Conn.Write([]byte(msg))
	_ = c.Close()
}

// ServerConfig WebSocket 服务端配置
type ServerConfig struct {
	AllowedOrigins []string // 允许的 Origin 列表，支持通配符 *
}

// CheckOrigin 检查 Origin 是否允许
func (c *ServerConfig) CheckOrigin(origin string) bool {
	if len(c.AllowedOrigins) == 0 {
		return true
	}
	if origin == "" {
		return true
	}
	for _, allowed := range c.AllowedOrigins {
		if c.matchOrigin(origin, allowed) {
			return true
		}
	}
	return false
}

// matchOrigin 检查 origin 是否匹配允许的模式（支持通配符 *）
func (c *ServerConfig) matchOrigin(origin, pattern string) bool {
	if origin == pattern {
		return true
	}
	if strings.Contains(pattern, "*") {
		parts := strings.SplitN(pattern, "*", 2)
		if len(parts) == 2 {
			return strings.HasPrefix(origin, parts[0]) && strings.HasSuffix(origin, parts[1])
		}
	}
	return false
}

// CorsMiddleware CORS 中间件
func (c *ServerConfig) CorsMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if c.CheckOrigin(origin) {
			ctx.Header("Access-Control-Allow-Origin", origin)
			ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			ctx.Header("Access-Control-Allow-Credentials", "true")
		}
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}
