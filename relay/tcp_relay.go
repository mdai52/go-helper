package relay

import (
	"net"

	"github.com/rehiy/libgo/websocket"
)

// TCPParam TCP 转发参数。
type TCPParam struct {
	TargetAddr string `json:"targetAddr" note:"目标地址"`
	BinaryMode bool   `json:"binaryMode" note:"二进制模式"`
}

// TCPRelay 在 WebSocket 与目标 TCP 地址之间做双向数据转发。
func TCPRelay(ws *websocket.Conn, p *TCPParam) error {
	defer ws.Close()

	if p.BinaryMode {
		ws.SetMessageType(websocket.BinaryMessage)
	}

	conn, err := net.Dial("tcp", p.TargetAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	return Bridge(NewReadWriter(ws), NewReadWriter(conn))
}
