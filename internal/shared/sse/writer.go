package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Writer 封装 SSE（Server-Sent Events）写入，管理 header 设置和事件格式化。
type Writer struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewWriter 创建 SSE Writer 并设置必要的响应头。
// 如果 ResponseWriter 不支持 Flush（极少见），返回 error。
func NewWriter(w http.ResponseWriter) (*Writer, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("response writer does not support streaming")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	return &Writer{w: w, flusher: flusher}, nil
}

// WriteEvent 写入一个具名 SSE 事件，data 会被 JSON 序列化。
func (s *Writer) WriteEvent(event string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal SSE data: %w", err)
	}

	// SSE 格式：event: <name>\ndata: <json>\n\n
	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, jsonData); err != nil {
		return fmt.Errorf("write SSE event: %w", err)
	}
	s.flusher.Flush()
	return nil
}

// WriteDone 写入流结束标记。
func (s *Writer) WriteDone() error {
	if _, err := fmt.Fprint(s.w, "event: done\ndata: [DONE]\n\n"); err != nil {
		return fmt.Errorf("write SSE done: %w", err)
	}
	s.flusher.Flush()
	return nil
}

// WriteError 写入错误事件。
func (s *Writer) WriteError(code int, msg string) {
	_ = s.WriteEvent("error", map[string]any{"code": code, "msg": msg})
}

// ========== 标准 SSE 事件数据结构 ==========

// MessageStart 流开始时发送，包含消息 ID 等前置信息。
type MessageStart struct {
	MessageID      string `json:"messageId"`
	ConversationID int64  `json:"conversationId"`
}

// Delta 增量内容事件。
type Delta struct {
	Content string `json:"content"`
}

// MessageEnd 流结束时发送，包含完整的元信息。
type MessageEnd struct {
	MessageID  string   `json:"messageId"`
	Model      string   `json:"model"`
	TokenCount int      `json:"tokenCount"`
	LatencyMs  int      `json:"latencyMs"`
	Sources    []string `json:"sources,omitempty"`
}
