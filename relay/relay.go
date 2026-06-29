package relay

import "io"

// Endpoint 表示一个可读写的转发端点。
// Closer 可选；Bridge 在任一方向结束后会关闭两端以解除另一方向的阻塞。
type Endpoint struct {
	Reader io.Reader
	Writer io.Writer
	Closer io.Closer
}

// NewEndpoint 使用独立的读写端和可选关闭器创建端点。
func NewEndpoint(reader io.Reader, writer io.Writer, closer io.Closer) Endpoint {
	return Endpoint{Reader: reader, Writer: writer, Closer: closer}
}

// NewReadWriter 使用同一个 io.ReadWriter 创建端点；若 rw 实现 io.Closer，Bridge 会负责关闭。
func NewReadWriter(rw io.ReadWriter) Endpoint {
	endpoint := Endpoint{Reader: rw, Writer: rw}
	if closer, ok := rw.(io.Closer); ok {
		endpoint.Closer = closer
	}
	return endpoint
}

// Bridge 在两个端点之间做双向数据转发，直到任一方向结束。
func Bridge(a, b Endpoint) error {
	ch := make(chan error, 2)
	go copyData(a.Writer, b.Reader, ch)
	go copyData(b.Writer, a.Reader, ch)

	err := <-ch
	closeEndpoint(a)
	closeEndpoint(b)
	<-ch

	if err == io.EOF {
		return nil
	}
	return err
}

func copyData(dst io.Writer, src io.Reader, ch chan<- error) {
	_, err := io.Copy(dst, src)
	ch <- err
}

func closeEndpoint(endpoint Endpoint) {
	if endpoint.Closer != nil {
		_ = endpoint.Closer.Close()
	}
}
