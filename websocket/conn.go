package websocket

import (
	"errors"
	"io"

	"golang.org/x/net/websocket"

	"github.com/rehiy/libgo/logman"
)

// Conn WebSocket 连接封装
type Conn struct {
	*websocket.Conn
}

// Read 读取消息
func (c *Conn) Read(v []byte) error {
	return websocket.Message.Receive(c.Conn, v)
}

// ReadJSON 读取 JSON 消息
func (c *Conn) ReadJSON(v any) error {
	return websocket.JSON.Receive(c.Conn, v)
}

// Write 写入消息
func (c *Conn) Write(p []byte) error {
	return websocket.Message.Send(c.Conn, p)
}

// WriteJSON 写入 JSON 消息
func (c *Conn) WriteJSON(v any) error {
	return websocket.JSON.Send(c.Conn, v)
}

// Close 关闭连接，忽略已关闭的错误
func (c *Conn) Close() error {
	err := c.Conn.Close()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// Die 发送消息并关闭连接
func (c *Conn) Die(msg string) {
	_, _ = c.Conn.Write([]byte(msg))
	_ = c.Conn.Close()
}

// NewClient 创建 WebSocket 客户端连接
func NewClient(url, protocol, origin string) (*Conn, error) {
	logman.Info("connect to server", "url", url)

	ws, err := websocket.Dial(url, protocol, origin)
	if err != nil {
		logman.Error("connect failed", "error", err)
		return nil, err
	}

	return &Conn{ws}, nil
}