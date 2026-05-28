package etcd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WatchEvent watch 收到的事件
type WatchEvent struct {
	Type  string // "PUT" 或 "DELETE"
	Value string // PUT 时的新值（已解码）；DELETE 时为空
}

// watchResponse 是 etcd watch 流中每行 JSON 的结构
type watchResponse struct {
	Result struct {
		Events []struct {
			Type string `json:"type"`
			Kv   struct {
				Value string `json:"value"`
			} `json:"kv"`
		} `json:"events"`
	} `json:"result"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Watch 监听 key 变化，通过 events/errs channel 通知，ctx 取消时停止。
// 连接断开后会自动重连，重连间隔从 1s 指数退避至 30s。
func (c *Client) Watch(ctx context.Context, key string) (<-chan WatchEvent, <-chan error) {
	events := make(chan WatchEvent, 8)
	errs := make(chan error, 4)

	go func() {
		defer close(events)
		defer close(errs)

		const (
			initBackoff = time.Second
			maxBackoff  = 30 * time.Second
		)
		backoff := initBackoff

		for {
			// 检查 ctx 是否已取消
			select {
			case <-ctx.Done():
				return
			default:
			}

			err, connected := c.watchOnce(ctx, key, events, errs)
			if err == nil {
				// ctx 取消导致的正常退出
				return
			}
			// 连接曾经成功过，说明 etcd 是健康的，重置退避时间
			if connected {
				backoff = initBackoff
			}

			// 发送错误，非阻塞
			select {
			case errs <- fmt.Errorf("etcd watch 断开，准备重连: %w", err):
			default:
			}

			// 退避等待后重连
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			if backoff < maxBackoff {
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
		}
	}()

	return events, errs
}

// watchOnce 建立一次 watch 连接并持续读取事件，直到连接断开或 ctx 取消。
// 返回值：err=nil 表示 ctx 取消的正常退出；connected 表示本次连接是否曾成功收到过事件。
func (c *Client) watchOnce(ctx context.Context, key string, events chan<- WatchEvent, errs chan<- error) (err error, connected bool) {
	body, _ := json.Marshal(map[string]any{
		"key":             b64(key),
		"progress_notify": false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.endpoint()+"/v3/watch", bytes.NewReader(body))
	if err != nil {
		return err, false
	}
	c.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// ctx 取消导致的错误，视为正常退出
		if ctx.Err() != nil {
			return nil, false
		}
		return err, false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("etcd watch 失败 %d: %s", resp.StatusCode, raw), false
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, connected
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var msg watchResponse
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}
		if msg.Error.Message != "" {
			select {
			case errs <- fmt.Errorf("etcd watch 服务端错误: %s", msg.Error.Message):
			default:
			}
			continue
		}

		for _, ev := range msg.Result.Events {
			var val string
			if ev.Type == "PUT" {
				decoded, decErr := base64.StdEncoding.DecodeString(ev.Kv.Value)
				if decErr != nil {
					select {
					case errs <- fmt.Errorf("etcd watch value 解码失败: %w", decErr):
					default:
					}
					continue
				}
				val = string(decoded)
			}
			connected = true // 成功收到事件，标记连接曾经健康
			select {
			case events <- WatchEvent{Type: ev.Type, Value: val}:
			case <-ctx.Done():
				return nil, connected
			}
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		if ctx.Err() != nil {
			return nil, connected
		}
		return err, connected
	}
	return fmt.Errorf("watch 连接已关闭"), connected
}
