package tcprelay

import (
	"io"
	"net"

	"golang.org/x/net/websocket"
)

// Param TCP 转发参数
type Param struct {
	TargetAddr string `note:"目标地址"`
	BinaryMode bool   `note:"二进制模式"`
}

// Relay WebSocket 到 TCP 的数据转发
func Relay(ws *websocket.Conn, p *Param) error {
	defer ws.Close()

	if p.BinaryMode {
		ws.PayloadType = websocket.BinaryFrame
	}

	// 连接远程服务器
	conn, err := net.Dial("tcp", p.TargetAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 双向数据转发
	ch := make(chan error, 2)
	defer close(ch)

	go copyData(conn, ws, ch)
	go copyData(ws, conn, ch)

	return <-ch
}

func copyData(dst io.Writer, src io.Reader, ch chan<- error) {
	_, err := io.Copy(dst, src)
	ch <- err
}