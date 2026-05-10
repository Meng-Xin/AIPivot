package rag

import (
	"context"
	"fmt"

	"aipivot/internal/modules/agent"
	"aipivot/pkg/llm"

	"github.com/zeromicro/go-zero/core/logx"
)

// StreamMeta 流式问答的同步可用元数据（检索来源、工具调用记录等，在流开始前即可获得）。
type StreamMeta struct {
	Sources  []string
	ToolUses []agent.ToolUseRecord
}

// AnswerStream 执行 RAG 流式问答：retrieve → prompt → stream generate。
// 返回 channel 逐步推送增量 token，以及检索来源列表（同步可用）。
// kbID 为 0 时跳过检索，直接流式调用 LLM。
// model 为空时使用配置中的默认聊天模型。
func (s *Service) AnswerStream(ctx context.Context, kbID int64, question string, history []llm.ChatMessage, model string) (<-chan llm.StreamEvent, *StreamMeta, error) {
	var contexts []RetrievedChunk

	// 检索相关切块（与同步模式共用逻辑）
	if kbID > 0 {
		retrieved, err := s.retrieve(ctx, kbID, question)
		if err != nil {
			logx.WithContext(ctx).Errorf("RAG stream retrieve err: %v", err)
		} else {
			contexts = retrieved
		}
	}

	messages := s.buildPrompt(question, contexts, history)

	// 通过 Agent 流式执行（有工具时同步处理 tool calls，无工具时直接流式）
	chatModel := s.chatModelOrDefault(model)
	var stream <-chan llm.StreamEvent
	var toolUses []agent.ToolUseRecord

	if s.agent != nil && s.agent.HasTools() {
		var agentMeta *agent.StreamMeta
		var agentErr error
		stream, agentMeta, agentErr = s.agent.RunStream(ctx, &agent.RunRequest{
			Model:       chatModel,
			Messages:    messages,
			MaxTokens:   s.maxTokens,
			Temperature: s.temperature,
		})
		if agentErr != nil {
			return nil, nil, fmt.Errorf("agent stream: %w", agentErr)
		}
		if agentMeta != nil {
			toolUses = agentMeta.ToolUses
		}
	} else {
		var streamErr error
		stream, streamErr = s.llmClient.ChatCompletionStream(ctx, &llm.ChatRequest{
			Model:       chatModel,
			Messages:    messages,
			MaxTokens:   s.maxTokens,
			Temperature: s.temperature,
		})
		if streamErr != nil {
			return nil, nil, fmt.Errorf("LLM stream chat: %w", streamErr)
		}
	}

	sources := make([]string, 0, len(contexts)+len(toolUses))
	for _, c := range contexts {
		sources = append(sources, fmt.Sprintf("chunk_%d(score=%.2f)", c.ChunkIndex, c.Score))
	}
	for _, tu := range toolUses {
		sources = append(sources, fmt.Sprintf("tool:%s", tu.Name))
	}

	return stream, &StreamMeta{Sources: sources, ToolUses: toolUses}, nil
}
