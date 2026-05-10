package agent

import (
	"context"
	"fmt"
	"sync"

	"aipivot/pkg/llm"

	"github.com/zeromicro/go-zero/core/logx"
)

// Registry 管理所有已注册的 Tool，提供按名查找和批量导出定义。
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register 注册一个工具，重复注册同名工具会覆盖。
func (r *Registry) Register(t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[t.Name()] = t
	logx.Infof("agent: registered tool %q", t.Name())
}

// Count 返回已注册工具数量。
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// Definitions 导出所有工具的 OpenAI ToolDefinition 列表，用于 ChatRequest.Tools。
func (r *Registry) Definitions() []llm.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]llm.ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, t.Definition())
	}
	return defs
}

// Execute 按名执行工具调用。工具不存在时返回错误。
func (r *Registry) Execute(ctx context.Context, name, arguments string) (string, error) {
	r.mu.RLock()
	t, ok := r.tools[name]
	r.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("agent: tool %q not found", name)
	}

	result, err := t.Execute(ctx, arguments)
	if err != nil {
		logx.WithContext(ctx).Errorf("agent: tool %q execute err: %v", name, err)
		return "", fmt.Errorf("tool %q execution failed: %w", name, err)
	}

	logx.WithContext(ctx).Infof("agent: tool %q executed, result length=%d", name, len(result))
	return result, nil
}
