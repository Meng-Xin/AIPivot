package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"aipivot/pkg/llm"
)

// EscalationTool 人工客服转接工具。
// 当 AI 判断无法回答用户问题、用户明确要求转人工、或对话情绪激烈时，Agent 调用此工具触发转接。
type EscalationTool struct{}

type escalationArgs struct {
	Reason string `json:"reason"` // 转接原因
}

type escalationResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewEscalationTool() *EscalationTool {
	return &EscalationTool{}
}

func (t *EscalationTool) Name() string { return "escalate_to_human" }
func (t *EscalationTool) Description() string {
	return "将对话转接给人工客服。当你无法回答用户的问题、用户明确要求转人工、或用户情绪激动时调用此工具"
}

func (t *EscalationTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Type: "function",
		Function: llm.FunctionDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"reason": map[string]interface{}{
						"type":        "string",
						"description": "转接原因，简要说明为何需要人工介入",
					},
				},
				"required": []string{"reason"},
			},
		},
	}
}

// Execute 返回转接确认。实际状态变更由 Logic 层根据 ToolUseRecord 处理。
func (t *EscalationTool) Execute(_ context.Context, arguments string) (string, error) {
	var args escalationArgs
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("parse escalation args: %w", err)
	}

	result := escalationResult{
		Status:  "escalated",
		Message: fmt.Sprintf("已为您转接人工客服，原因：%s。请稍候，客服人员将尽快接入。", args.Reason),
	}

	data, _ := json.Marshal(result)
	return string(data), nil
}
