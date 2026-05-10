package websocket

import "encoding/json"

// Message WebSocket 消息结构
type Message struct {
	Method  string `json:"method"`
	TaskID  uint   `json:"taskId"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Payload any    `json:"payload"`
}

// GetPayload 解析 Payload 到目标类型
func (m *Message) GetPayload(v any) error {
	payload, err := json.Marshal(m.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, v)
}