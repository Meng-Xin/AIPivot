package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ========== Streaming Chat Completion ==========

// StreamDelta 单个流式 chunk 中的增量内容。
type StreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// StreamChoice 流式 chunk 中的单个选择。
type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        StreamDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

// StreamChunk 对应 OpenAI streaming API 的单个 SSE data payload。
type StreamChunk struct {
	ID      string         `json:"id"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
	Usage   *ChatUsage     `json:"usage,omitempty"` // 部分 provider 在最后一个 chunk 返回 usage
}

// StreamEvent 从流中读取的事件，Content 为增量文本，Done 标记流结束。
type StreamEvent struct {
	Content      string     // 增量文本（可能为空字符串）
	Done         bool       // 流结束标记
	Model        string     // 模型名称（仅最后一个 chunk 有意义）
	Usage        *ChatUsage // token 统计（仅部分 provider 在末尾返回）
	FinishReason string     // stop / length / tool_calls 等
	Err          error      // 解析错误
}

// ChatCompletionStream 流式调用 LLM，通过 channel 逐步返回增量 token。
// 调用方应持续读取 channel 直到关闭或收到 Done=true 的事件。
func (c *Client) ChatCompletionStream(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error) {
	if c.useResponsesAPI {
		return c.ResponsesCompletionStream(ctx, req)
	}

	req.Stream = true
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal stream chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create stream chat request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("stream chat request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("stream chat failed (status=%d): %s", resp.StatusCode, string(respBody))
	}

	ch := make(chan StreamEvent, 16)
	go c.readSSEStream(ctx, resp, ch)
	return ch, nil
}

// readSSEStream 解析 OpenAI SSE 协议的流式响应，将增量内容推入 channel。
func (c *Client) readSSEStream(ctx context.Context, resp *http.Response, ch chan<- StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var lastModel string
	var lastUsage *ChatUsage

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			ch <- StreamEvent{Err: ctx.Err()}
			return
		default:
		}

		line := scanner.Text()

		// SSE 协议：空行分隔事件，只处理 "data: " 前缀的行
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		// "[DONE]" 标记流结束
		if data == "[DONE]" {
			ch <- StreamEvent{Done: true, Model: lastModel, Usage: lastUsage}
			return
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			ch <- StreamEvent{Err: fmt.Errorf("decode stream chunk: %w", err)}
			return
		}

		if chunk.Model != "" {
			lastModel = chunk.Model
		}
		if chunk.Usage != nil {
			lastUsage = chunk.Usage
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		evt := StreamEvent{
			Content: choice.Delta.Content,
			Model:   chunk.Model,
		}
		if choice.FinishReason != nil {
			evt.FinishReason = *choice.FinishReason
		}

		ch <- evt
	}

	if err := scanner.Err(); err != nil {
		ch <- StreamEvent{Err: fmt.Errorf("read stream: %w", err)}
	}
}
