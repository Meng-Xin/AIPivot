package rag

import (
	"context"
	"fmt"

	"aipivot/pkg/llm"

	"github.com/zeromicro/go-zero/core/logx"
)

// StreamMeta 流式问答的同步可用元数据（检索来源等，在流开始前即可获得）。
type StreamMeta struct {
	Sources []string
}

// AnswerStream 执行 RAG 流式问答：retrieve → prompt → stream generate。
// 返回 channel 逐步推送增量 token，以及检索来源列表（同步可用）。
// kbID 为 0 时跳过检索，直接流式调用 LLM。
func (s *Service) AnswerStream(ctx context.Context, kbID int64, question string, history []llm.ChatMessage) (<-chan llm.StreamEvent, *StreamMeta, error) {
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

	// 流式调用 LLM（区别于同步 Answer 使用 ChatCompletion）
	stream, err := s.llmClient.ChatCompletionStream(ctx, &llm.ChatRequest{
		Model:       s.chatModel,
		Messages:    messages,
		MaxTokens:   s.maxTokens,
		Temperature: s.temperature,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("LLM stream chat: %w", err)
	}

	sources := make([]string, 0, len(contexts))
	for _, c := range contexts {
		sources = append(sources, fmt.Sprintf("chunk_%d(score=%.2f)", c.ChunkIndex, c.Score))
	}

	return stream, &StreamMeta{Sources: sources}, nil
}
