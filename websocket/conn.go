package websocket

import (
	"io"
	"sync"
	"time"

	gorilla "github.com/gorilla/websocket"
)

const (
	TextMessage   = gorilla.TextMessage   // = 1
	BinaryMessage = gorilla.BinaryMessage // = 2
)

// Conn gorilla/websocket 的流式包装，提供并发安全的 Read/Write 和原生 Ping
type Conn struct {
	ws      *gorilla.Conn
	wmu     sync.Mutex // guards WriteMessage / WriteJSON
	rmu     sync.Mutex // guards NextReader / ReadJSON
	rdBuf   io.Reader  // current message reader (stream shim)
	msgType int        // TextMessage or BinaryMessage
}

func newConn(ws *gorilla.Conn) *Conn {
	return &Conn{ws: ws, msgType: TextMessage}
}

// Read 流式读，跨多消息拼接
func (c *Conn) Read(p []byte) (int, error) {
	c.rmu.Lock()
	defer c.rmu.Unlock()
	for {
		if c.rdBuf != nil {
			n, err := c.rdBuf.Read(p)
			if n > 0 {
				if err == io.EOF {
					c.rdBuf = nil
				}
				return n, nil
			}
			if err == io.EOF {
				c.rdBuf = nil
				continue
			}
			return 0, err
		}
		_, r, err := c.ws.NextReader()
		if err != nil {
			return 0, err
		}
		c.rdBuf = r
	}
}

// Write 每次调用发一条 WS 消息，wmu 保证与并发 Write 互斥
func (c *Conn) Write(p []byte) (int, error) {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	if err := c.ws.WriteMessage(c.msgType, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// ReadJSON 读取 JSON 消息
func (c *Conn) ReadJSON(v any) error {
	c.rmu.Lock()
	defer c.rmu.Unlock()
	return c.ws.ReadJSON(v)
}

// WriteJSON 写入 JSON 消息
func (c *Conn) WriteJSON(v any) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	return c.ws.WriteJSON(v)
}

// Ping 发送 WS ping 控制帧（gorilla WriteControl 自身并发安全，无需 wmu）
func (c *Conn) Ping(data []byte) error {
	return c.ws.WriteControl(gorilla.PingMessage, data, time.Now().Add(10*time.Second))
}

// SetMessageType 设置写消息类型（供 tcprelay 切换 BinaryMessage）
func (c *Conn) SetMessageType(t int) { c.msgType = t }

// Close 关闭连接
func (c *Conn) Close() error { return c.ws.Close() }

// Die 发送消息并关闭连接
func (c *Conn) Die(msg string) {
	_, _ = c.Write([]byte(msg))
	_ = c.Close()
}
