package websocket

import (
	"errors"
	"io"

	"golang.org/x/net/websocket"

	"github.com/rehiy/libgo/logman"
)

// ClientConn WebSocket 客户端连接
type ClientConn struct {
	*websocket.Conn
}

// NewClient 创建 WebSocket 客户端连接
func NewClient(url, protocol, origin string) (*ClientConn, error) {
	logman.Info("connect to server", "url", url)

	ws, err := websocket.Dial(url, protocol, origin)
	if err != nil {
		logman.Error("connect failed", "error", err)
		return nil, err
	}

	return &ClientConn{ws}, nil
}

// Read 读取数据（实现 io.Reader 接口，流式读取）
func (c *ClientConn) Read(v []byte) (n int, err error) {
	return c.Conn.Read(v)
}

// ReadJSON 读取 JSON 消息
func (c *ClientConn) ReadJSON(v any) error {
	return websocket.JSON.Receive(c.Conn, v)
}

// Write 写入数据（实现 io.Writer 接口，流式写入）
func (c *ClientConn) Write(p []byte) (n int, err error) {
	return c.Conn.Write(p)
}

// WriteJSON 写入 JSON 消息
func (c *ClientConn) WriteJSON(v any) error {
	return websocket.JSON.Send(c.Conn, v)
}

// Close 关闭连接，忽略 EOF 错误
func (c *ClientConn) Close() error {
	err := c.Conn.Close()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// Die 发送消息并关闭连接
func (c *ClientConn) Die(msg string) {
	_, _ = c.Conn.Write([]byte(msg))
	_ = c.Close()
}
