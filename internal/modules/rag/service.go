package rag

import (
	"context"
	"fmt"
	"strings"

	"aipivot/internal/modules/knowledge/repo"
	"aipivot/pkg/llm"

	"github.com/zeromicro/go-zero/core/logx"
)

// Service 是 RAG 编排服务：检索相关切块 → 组装 prompt → 调用 LLM 生成回复。
type Service struct {
	llmClient *llm.Client
	chunkRepo repo.DocumentChunkRepository

	// LLM 配置
	chatModel      string
	embeddingModel string
	maxTokens      int
	temperature    float64
}

// Config RAG 服务配置。
type Config struct {
	ChatModel      string
	EmbeddingModel string
	MaxTokens      int
	Temperature    float64
}

func NewService(llmClient *llm.Client, chunkRepo repo.DocumentChunkRepository, cfg Config) *Service {
	return &Service{
		llmClient:      llmClient,
		chunkRepo:      chunkRepo,
		chatModel:      cfg.ChatModel,
		embeddingModel: cfg.EmbeddingModel,
		maxTokens:      cfg.MaxTokens,
		temperature:    cfg.Temperature,
	}
}

// Answer 执行 RAG 问答：retrieve → prompt → generate。
// kbID 为 0 时跳过检索，直接调用 LLM。
func (s *Service) Answer(ctx context.Context, kbID int64, question string, history []llm.ChatMessage) (*AnswerResult, error) {
	var contexts []RetrievedChunk

	// 1. 检索相关切块（知识库 ID > 0 时执行）
	if kbID > 0 {
		retrieved, err := s.retrieve(ctx, kbID, question)
		if err != nil {
			logx.WithContext(ctx).Errorf("RAG retrieve err: %v", err)
			// 检索失败不阻断，降级为纯 LLM 回复
		} else {
			contexts = retrieved
		}
	}

	// 2. 组装 prompt
	messages := s.buildPrompt(question, contexts, history)

	// 3. 调用 LLM 生成回复
	resp, err := s.llmClient.ChatCompletion(ctx, &llm.ChatRequest{
		Model:       s.chatModel,
		Messages:    messages,
		MaxTokens:   s.maxTokens,
		Temperature: s.temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned empty choices")
	}

	// 4. 构建来源引用列表
	sources := make([]string, 0, len(contexts))
	for _, c := range contexts {
		sources = append(sources, fmt.Sprintf("chunk_%d(score=%.2f)", c.ChunkIndex, c.Score))
	}

	return &AnswerResult{
		Content:    resp.Choices[0].Message.Content,
		Model:      resp.Model,
		TokenCount: resp.Usage.TotalTokens,
		Sources:    sources,
	}, nil
}

// retrieve 对用户问题做 embedding 后在 pgvector 中搜索 Top-K 相关切块。
func (s *Service) retrieve(ctx context.Context, kbID int64, question string) ([]RetrievedChunk, error) {
	queryVec, err := s.llmClient.EmbedSingle(ctx, s.embeddingModel, question)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	const topK = 5
	results, err := s.chunkRepo.SimilaritySearch(ctx, kbID, queryVec, topK)
	if err != nil {
		return nil, fmt.Errorf("similarity search: %w", err)
	}

	chunks := make([]RetrievedChunk, 0, len(results))
	for _, r := range results {
		chunks = append(chunks, RetrievedChunk{
			Content:    r.Content,
			ChunkIndex: r.ChunkIndex,
			Score:      r.Score,
			DocumentID: r.DocumentID,
		})
	}

	logx.WithContext(ctx).Infof("RAG retrieved %d chunks for kbID=%d", len(chunks), kbID)
	return chunks, nil
}

// buildPrompt 组装 RAG prompt：system instruction + 检索上下文 + 对话历史 + 当前问题。
func (s *Service) buildPrompt(question string, contexts []RetrievedChunk, history []llm.ChatMessage) []llm.ChatMessage {
	var systemPrompt strings.Builder
	systemPrompt.WriteString("你是 AIPivot AI 智能客服助手。请根据提供的知识库内容准确回答用户问题。")
	systemPrompt.WriteString("如果知识库中没有相关信息，请诚实说明你不确定，不要编造答案。")

	// 注入检索到的上下文
	if len(contexts) > 0 {
		systemPrompt.WriteString("\n\n以下是从知识库中检索到的相关内容，请基于这些内容回答：\n")
		for i, c := range contexts {
			systemPrompt.WriteString(fmt.Sprintf("\n--- 参考片段 %d (相关度: %.0f%%) ---\n%s\n", i+1, c.Score*100, c.Content))
		}
	}

	messages := make([]llm.ChatMessage, 0, len(history)+2)
	messages = append(messages, llm.ChatMessage{Role: "system", Content: systemPrompt.String()})

	// 保留最近对话历史（避免超出 context window）
	maxHistory := 10
	if len(history) > maxHistory {
		history = history[len(history)-maxHistory:]
	}
	messages = append(messages, history...)

	messages = append(messages, llm.ChatMessage{Role: "user", Content: question})

	return messages
}

// ========== 结果类型 ==========

// AnswerResult RAG 回答结果。
type AnswerResult struct {
	Content    string   // AI 回复内容
	Model      string   // 使用的 LLM 模型
	TokenCount int      // 总 token 消耗
	Sources    []string // 来源引用
}

// RetrievedChunk 检索到的切块信息。
type RetrievedChunk struct {
	Content    string
	ChunkIndex int
	Score      float64
	DocumentID int64
}
