package websocket

import (
	"net/http"

	gorilla "github.com/gorilla/websocket"
	"github.com/rehiy/libgo/logman"
)

// ClientConn WebSocket 客户端连接
type ClientConn struct {
	*Conn
}

// NewClient 创建 WebSocket 客户端连接
func NewClient(url, protocol, origin string) (*ClientConn, error) {
	logman.Info("connect to server", "url", url)

	header := http.Header{}
	if origin != "" {
		header.Set("Origin", origin)
	}

	ws, _, err := gorilla.DefaultDialer.Dial(url, header)
	if err != nil {
		logman.Error("connect failed", "error", err)
		return nil, err
	}

	return &ClientConn{newConn(ws)}, nil
}
