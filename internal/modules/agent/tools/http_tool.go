package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"aipivot/internal/shared/po"
	"aipivot/pkg/llm"
)

// HttpTool 将租户定义的 HTTP 回调端点包装为 Agent Tool 接口。
// Agent 调用该工具时，将 LLM 生成的 JSON 参数作为 request body POST 到配置的 endpoint，
// 并将响应体作为工具结果回传给 LLM。
type HttpTool struct {
	id          int64
	name        string
	description string
	parameters  po.JSONMap // JSON Schema
	endpoint    string
	method      string
	headers     po.JSONMap
	client      *http.Client
}

// NewHttpToolFromSkill 从 Skill PO 创建 HttpTool 实例。
func NewHttpToolFromSkill(s *po.Skill) *HttpTool {
	timeout := time.Duration(s.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	method := strings.ToUpper(s.Method)
	if method == "" {
		method = http.MethodPost
	}
	return &HttpTool{
		id:          s.ID,
		name:        s.Name,
		description: s.Description,
		parameters:  s.Parameters,
		endpoint:    s.Endpoint,
		method:      method,
		headers:     s.Headers,
		client:      &http.Client{Timeout: timeout},
	}
}

func (t *HttpTool) Name() string        { return t.name }
func (t *HttpTool) Description() string { return t.description }

func (t *HttpTool) Definition() llm.ToolDefinition {
	// parameters 存储完整 JSON Schema，直接透传给 LLM
	paramsRaw, _ := json.Marshal(t.parameters)
	var paramsSchema interface{}
	_ = json.Unmarshal(paramsRaw, &paramsSchema)
	if paramsSchema == nil {
		paramsSchema = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
	}

	return llm.ToolDefinition{
		Type: "function",
		Function: llm.FunctionDefinition{
			Name:        t.name,
			Description: t.description,
			Parameters:  paramsSchema,
		},
	}
}

// Execute 调用 HTTP 端点并返回响应体文本。
// arguments 是 LLM 生成的 JSON 参数字符串，作为请求 body 直接发送。
func (t *HttpTool) Execute(ctx context.Context, arguments string) (string, error) {
	if t.endpoint == "" {
		return "", fmt.Errorf("skill %q: endpoint not configured", t.name)
	}

	endpoint := t.endpoint
	var body io.Reader
	switch t.method {
	case "GET":
		u, err := url.Parse(endpoint)
		if err != nil {
			return "", fmt.Errorf("skill %q: parse endpoint: %w", t.name, err)
		}
		if strings.TrimSpace(arguments) != "" {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(arguments), &args); err != nil {
				return "", fmt.Errorf("skill %q: parse GET arguments: %w", t.name, err)
			}
			q := u.Query()
			for k, v := range args {
				q.Set(k, fmt.Sprint(v))
			}
			u.RawQuery = q.Encode()
		}
		endpoint = u.String()
		body = nil
	default:
		// POST/PUT：直接将 JSON 参数作为 request body
		body = bytes.NewBufferString(arguments)
	}

	req, err := http.NewRequestWithContext(ctx, t.method, endpoint, body)
	if err != nil {
		return "", fmt.Errorf("skill %q: create request: %w", t.name, err)
	}

	if t.method != "GET" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 附加租户自定义请求头（如认证 token）
	for k, v := range t.headers {
		if strVal, ok := v.(string); ok {
			req.Header.Set(k, strVal)
		}
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("skill %q: call endpoint: %w", t.name, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024)) // 最多读 64KB
	if err != nil {
		return "", fmt.Errorf("skill %q: read response: %w", t.name, err)
	}

	// 非 2xx 响应时将状态码和 body 都返回给 LLM，由 LLM 决定如何处理
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)), nil
	}

	return string(respBody), nil
}
