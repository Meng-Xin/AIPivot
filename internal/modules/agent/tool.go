package agent

import (
	"context"

	"aipivot/pkg/llm"
)

// Tool 定义一个可供 Agent 调用的工具。
// 每个业务技能（查天气、查订单、计算器等）实现此接口。
type Tool interface {
	// Name 工具唯一标识，对应 Function Calling 中的 function.name。
	Name() string

	// Description 工具描述，帮助 LLM 判断何时调用。
	Description() string

	// Definition 返回 OpenAI Function Calling 格式的工具定义。
	Definition() llm.ToolDefinition

	// Execute 执行工具调用，arguments 为 JSON 编码的参数字符串。
	// 返回结果文本（JSON 或纯文本），将作为 tool role 消息回传给 LLM。
	Execute(ctx context.Context, arguments string) (string, error)
}

// ToolUseRecord 一次工具调用的完整记录（含输入输出）。
type ToolUseRecord struct {
	Name      string `json:"name"`      // 工具名称
	Arguments string `json:"arguments"` // 调用参数（JSON 字符串）
	Result    string `json:"result"`    // 执行结果
}
