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

// Read 读取消息
func (c *ClientConn) Read(v []byte) error {
	return websocket.Message.Receive(c.Conn, v)
}

// ReadJSON 读取 JSON 消息
func (c *ClientConn) ReadJSON(v any) error {
	return websocket.JSON.Receive(c.Conn, v)
}

// Write 写入消息
func (c *ClientConn) Write(p []byte) error {
	return websocket.Message.Send(c.Conn, p)
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
