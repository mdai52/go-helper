package websocket

import (
	"net/http"
	"strings"

	gorilla "github.com/gorilla/websocket"
	"github.com/gin-gonic/gin"
)

// ServerConn WebSocket 服务端连接
type ServerConn struct {
	*Conn
}

// Handler 创建 WebSocket 处理器（gin HandlerFunc）
// 使用方式：router.GET("/ws", config.Handler(func(conn *ServerConn) { ... }))
func (c *ServerConfig) Handler(handler func(*ServerConn)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !c.CheckOrigin(ctx.GetHeader("Origin")) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		upgrader := gorilla.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			return
		}
		conn := &ServerConn{newConn(ws)}
		handler(conn)
	}
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
