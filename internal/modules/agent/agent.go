package agent

import (
	"context"
	"fmt"

	"aipivot/pkg/llm"

	"github.com/zeromicro/go-zero/core/logx"
)

const defaultMaxRounds = 5

// Agent 基于 Function Calling 的 ReAct 循环编排器。
// 每轮：调用 LLM → 若返回 tool_calls 则执行工具 → 将结果回传 LLM → 循环直到获得文本回复。
type Agent struct {
	llmClient *llm.Client
	registry  *Registry
	maxRounds int
}

// RunRequest 单次 Agent 执行的输入参数。
type RunRequest struct {
	Model       string
	Messages    []llm.ChatMessage
	MaxTokens   int
	Temperature float64
}

// RunResult Agent 执行的最终结果。
type RunResult struct {
	Content    string          // 最终文本回复
	Model      string          // 使用的模型
	Usage      *llm.ChatUsage  // token 统计（仅最后一轮可用）
	ToolUses   []ToolUseRecord // 本次调用中使用的所有工具记录
	TotalRound int             // 实际执行轮数
}

// StreamMeta Agent 流式执行的同步可用元数据。
type StreamMeta struct {
	ToolUses []ToolUseRecord
}

func NewAgent(llmClient *llm.Client, registry *Registry, maxRounds int) *Agent {
	if maxRounds <= 0 {
		maxRounds = defaultMaxRounds
	}
	return &Agent{
		llmClient: llmClient,
		registry:  registry,
		maxRounds: maxRounds,
	}
}

// Run 同步执行 Agent ReAct 循环，返回最终文本回复。
func (a *Agent) Run(ctx context.Context, req *RunRequest) (*RunResult, error) {
	messages := make([]llm.ChatMessage, len(req.Messages))
	copy(messages, req.Messages)

	defs := a.registry.Definitions()
	var toolUses []ToolUseRecord

	for round := 0; round < a.maxRounds; round++ {
		chatReq := &llm.ChatRequest{
			Model:       req.Model,
			Messages:    messages,
			MaxTokens:   req.MaxTokens,
			Temperature: req.Temperature,
		}
		// 仅在有注册工具时附加 tools 参数
		if len(defs) > 0 {
			chatReq.Tools = defs
			chatReq.ToolChoice = "auto"
		}

		resp, err := a.llmClient.ChatCompletion(ctx, chatReq)
		if err != nil {
			return nil, fmt.Errorf("agent round %d LLM call: %w", round, err)
		}
		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("agent round %d: LLM returned empty choices", round)
		}

		choice := resp.Choices[0]

		// 无 tool_calls → 最终文本回复
		if len(choice.Message.ToolCalls) == 0 {
			return &RunResult{
				Content:    choice.Message.Content,
				Model:      resp.Model,
				Usage:      &resp.Usage,
				ToolUses:   toolUses,
				TotalRound: round + 1,
			}, nil
		}

		// 处理 tool_calls：执行每个工具并将结果追加到消息列表
		logx.WithContext(ctx).Infof("agent round %d: %d tool call(s)", round, len(choice.Message.ToolCalls))

		messages = append(messages, choice.Message)
		for _, tc := range choice.Message.ToolCalls {
			result, execErr := a.registry.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			if execErr != nil {
				// 工具执行失败时将错误信息回传给 LLM，由模型决定如何处理
				result = fmt.Sprintf("Error: %s", execErr.Error())
			}

			toolUses = append(toolUses, ToolUseRecord{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
				Result:    result,
			})

			messages = append(messages, llm.ChatMessage{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    result,
			})
		}
	}

	return nil, fmt.Errorf("agent: exceeded max rounds (%d), possible infinite tool loop", a.maxRounds)
}

// RunStream 流式执行 Agent。
// 无工具注册时直接走 LLM 流式；有工具时先完成 tool-calling 同步循环，最终回复以 fake stream 返回。
func (a *Agent) RunStream(ctx context.Context, req *RunRequest) (<-chan llm.StreamEvent, *StreamMeta, error) {
	defs := a.registry.Definitions()

	// 快速路径：无工具 → 直接流式调用 LLM
	if len(defs) == 0 {
		stream, err := a.llmClient.ChatCompletionStream(ctx, &llm.ChatRequest{
			Model:       req.Model,
			Messages:    req.Messages,
			MaxTokens:   req.MaxTokens,
			Temperature: req.Temperature,
		})
		if err != nil {
			return nil, nil, err
		}
		return stream, &StreamMeta{}, nil
	}

	// 有工具：同步 ReAct 循环获取最终回复，然后包装为 stream channel
	result, err := a.Run(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan llm.StreamEvent, 2)
	go func() {
		defer close(ch)
		ch <- llm.StreamEvent{
			Content: result.Content,
			Model:   result.Model,
		}
		ch <- llm.StreamEvent{
			Done:  true,
			Model: result.Model,
			Usage: result.Usage,
		}
	}()

	return ch, &StreamMeta{ToolUses: result.ToolUses}, nil
}

// HasTools 返回 Agent 是否注册了任何工具。
func (a *Agent) HasTools() bool {
	return a.registry.Count() > 0
}
