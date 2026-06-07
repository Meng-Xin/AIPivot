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

type responsesRequest struct {
	Model           string                  `json:"model"`
	Stream          bool                    `json:"stream,omitempty"`
	Instructions    string                  `json:"instructions,omitempty"`
	MaxOutputTokens int                     `json:"max_output_tokens,omitempty"`
	Tools           []responsesTool         `json:"tools,omitempty"`
	Input           []responsesInputMessage `json:"input"`
}

type responsesTool struct {
	Type       string `json:"type"`
	MaxKeyword int    `json:"max_keyword,omitempty"`
}

type responsesInputMessage struct {
	Role    string                  `json:"role"`
	Content []responsesInputContent `json:"content"`
}

type responsesInputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responsesResponse struct {
	ID         string `json:"id"`
	Model      string `json:"model"`
	OutputText string `json:"output_text"`
	Output     []struct {
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *Client) healthCheckResponses(ctx context.Context) error {
	model := c.healthModel
	if model == "" {
		return fmt.Errorf("llm health model is empty")
	}

	req := &ChatRequest{
		Model:       model,
		Messages:    []ChatMessage{{Role: "user", Content: "ping"}},
		MaxTokens:   1,
		Temperature: 0,
	}
	if _, err := c.responsesCompletion(ctx, req, false); err != nil {
		return fmt.Errorf("responses health failed: %w", err)
	}
	return nil
}

func (c *Client) ResponsesCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return c.responsesCompletion(ctx, req, true)
}

func (c *Client) responsesCompletion(ctx context.Context, req *ChatRequest, includeTools bool) (*ChatResponse, error) {
	body, err := json.Marshal(c.toResponsesRequest(req, false, includeTools))
	if err != nil {
		return nil, fmt.Errorf("marshal responses request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create responses request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("responses request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("responses failed (status=%d): %s", resp.StatusCode, string(respBody))
	}

	var responsesResp responsesResponse
	if err := json.NewDecoder(resp.Body).Decode(&responsesResp); err != nil {
		return nil, fmt.Errorf("decode responses response: %w", err)
	}

	content := responsesResp.text()
	if content == "" {
		return nil, fmt.Errorf("responses returned empty output")
	}
	model := responsesResp.Model
	if model == "" {
		model = req.Model
	}
	return &ChatResponse{
		ID:    responsesResp.ID,
		Model: model,
		Choices: []ChatChoice{
			{
				Index:   0,
				Message: ChatMessage{Role: "assistant", Content: content},
			},
		},
		Usage: ChatUsage{
			PromptTokens:     responsesResp.Usage.InputTokens,
			CompletionTokens: responsesResp.Usage.OutputTokens,
			TotalTokens:      responsesResp.Usage.TotalTokens,
		},
	}, nil
}

func (c *Client) ResponsesCompletionStream(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error) {
	body, err := json.Marshal(c.toResponsesRequest(req, true, true))
	if err != nil {
		return nil, fmt.Errorf("marshal stream responses request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create stream responses request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("stream responses request: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("stream responses failed (status=%d): %s", resp.StatusCode, string(respBody))
	}

	ch := make(chan StreamEvent, 16)
	go c.readResponsesSSE(ctx, resp, ch)
	return ch, nil
}

func (c *Client) toResponsesRequest(req *ChatRequest, stream bool, includeTools bool) *responsesRequest {
	out := &responsesRequest{
		Model:           req.Model,
		Stream:          stream,
		MaxOutputTokens: req.MaxTokens,
		Input:           make([]responsesInputMessage, 0, len(req.Messages)),
	}
	if includeTools && c.enableWebSearch {
		out.Tools = append(out.Tools, responsesTool{
			Type:       "web_search",
			MaxKeyword: c.webSearchMaxKeyword,
		})
	}

	var instructions []string
	for _, msg := range req.Messages {
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}
		role := msg.Role
		if role == "system" {
			instructions = append(instructions, content)
			continue
		}
		if role != "assistant" && role != "user" {
			role = "user"
		}
		out.Input = append(out.Input, responsesInputMessage{
			Role: role,
			Content: []responsesInputContent{
				{Type: "input_text", Text: content},
			},
		})
	}
	out.Instructions = strings.Join(instructions, "\n\n")
	return out
}

func (r *responsesResponse) text() string {
	if r.OutputText != "" {
		return r.OutputText
	}
	var b strings.Builder
	for _, item := range r.Output {
		for _, content := range item.Content {
			if content.Text != "" {
				b.WriteString(content.Text)
			}
		}
	}
	return b.String()
}

func (c *Client) readResponsesSSE(ctx context.Context, resp *http.Response, ch chan<- StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var eventName string
	var dataLines []string
	var lastModel string
	var lastUsage *ChatUsage

	flush := func() bool {
		if len(dataLines) == 0 {
			eventName = ""
			return true
		}
		data := strings.Join(dataLines, "\n")
		dataLines = nil
		if data == "[DONE]" {
			ch <- StreamEvent{Done: true, Model: lastModel, Usage: lastUsage}
			return false
		}

		evt, done := parseResponsesStreamEvent(eventName, data)
		if evt.Err != nil || evt.Content != "" || evt.Done {
			ch <- evt
		}
		if evt.Model != "" {
			lastModel = evt.Model
		}
		if evt.Usage != nil {
			lastUsage = evt.Usage
		}
		eventName = ""
		return !done
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			ch <- StreamEvent{Err: ctx.Err()}
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			if !flush() {
				return
			}
			continue
		}
		if strings.HasPrefix(line, "event: ") {
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event: "))
			continue
		}
		if strings.HasPrefix(line, "data: ") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
		}
	}
	if len(dataLines) > 0 {
		_ = flush()
	}
	if err := scanner.Err(); err != nil {
		ch <- StreamEvent{Err: fmt.Errorf("read responses stream: %w", err)}
	}
}

func parseResponsesStreamEvent(eventName, data string) (StreamEvent, bool) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return StreamEvent{Err: fmt.Errorf("decode responses event: %w", err)}, true
	}

	eventType := eventName
	if value, ok := raw["type"].(string); ok && value != "" {
		eventType = value
	}
	evt := StreamEvent{Model: findString(raw, "model")}
	if usage := parseResponsesUsage(raw); usage != nil {
		evt.Usage = usage
	}
	isCompletedEvent := eventType == "response.completed" || eventType == "response.done" || eventType == "done"
	if !strings.Contains(eventType, "output_text") {
		if isCompletedEvent {
			evt.Done = true
			return evt, true
		}
		return evt, false
	}
	if delta, ok := raw["delta"].(string); ok {
		evt.Content = delta
		return evt, false
	}
	if text, ok := raw["text"].(string); ok && strings.Contains(eventType, "delta") {
		evt.Content = text
		return evt, false
	}
	if isCompletedEvent {
		evt.Done = true
		return evt, true
	}
	return evt, false
}

func parseResponsesUsage(raw map[string]interface{}) *ChatUsage {
	usageValue, ok := raw["usage"]
	if !ok {
		if response, ok := raw["response"].(map[string]interface{}); ok {
			usageValue = response["usage"]
		}
	}
	usageMap, ok := usageValue.(map[string]interface{})
	if !ok {
		return nil
	}
	return &ChatUsage{
		PromptTokens:     numberAsInt(usageMap["input_tokens"]),
		CompletionTokens: numberAsInt(usageMap["output_tokens"]),
		TotalTokens:      numberAsInt(usageMap["total_tokens"]),
	}
}

func findString(raw map[string]interface{}, key string) string {
	if value, ok := raw[key].(string); ok {
		return value
	}
	if response, ok := raw["response"].(map[string]interface{}); ok {
		if value, ok := response[key].(string); ok {
			return value
		}
	}
	return ""
}

func numberAsInt(value interface{}) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}
