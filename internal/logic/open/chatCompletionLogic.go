// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package open

import (
	"context"
	"fmt"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"
	"aipivot/pkg/llm"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type ChatCompletionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Chat Completion（同步）— OpenAI 兼容格式，通过 API Key 认证
func NewChatCompletionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatCompletionLogic {
	return &ChatCompletionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatCompletionLogic) ChatCompletion(req *types.ChatCompletionRequest) (resp *types.ChatCompletionResponse, err error) {
	tenantID := auth.TenantIDFromAPIKeyContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("invalid API key context")
	}

	if len(req.Messages) == 0 {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "messages cannot be empty")
	}

	// 提取最后一条用户消息作为 question，前面的作为 history
	question, history := extractQuestionAndHistory(req.Messages)
	if question == "" {
		return nil, errorx.NewBusinessError(errorx.CodeBadRequest, "last message must have role=user")
	}

	// 调用 RAG 服务（kbID=0 时跳过检索，退化为纯 LLM）
	result, err := l.svcCtx.RAGService.Answer(l.ctx, req.KnowledgeBaseID, question, history, req.Model, nil)
	if err != nil {
		l.Logger.Errorf("ChatCompletion RAG.Answer err: %v", err)
		return nil, errorx.NewBusinessError(errorx.CodeLLMUnavailable, "AI 回复生成失败")
	}

	return &types.ChatCompletionResponse{
		ID:     fmt.Sprintf("chatcmpl-%s", uuid.New().String()[:8]),
		Object: "chat.completion",
		Model:  result.Model,
		Choices: []types.ChatCompletionChoice{
			{
				Index:        0,
				Message:      types.OpenChatMessage{Role: "assistant", Content: result.Content},
				FinishReason: "stop",
			},
		},
		Usage: types.ChatCompletionUsage{
			TotalTokens: result.TokenCount,
		},
		Sources: result.Sources,
	}, nil
}

// extractQuestionAndHistory 从请求消息列表中分离最后的用户问题和对话历史。
// 跳过 system 消息（由 RAG 自行注入），仅保留 user/assistant 作为 history。
func extractQuestionAndHistory(msgs []types.OpenChatMessage) (string, []llm.ChatMessage) {
	if len(msgs) == 0 {
		return "", nil
	}

	// 最后一条必须是 user
	last := msgs[len(msgs)-1]
	if last.Role != "user" {
		return "", nil
	}

	history := make([]llm.ChatMessage, 0, len(msgs)-1)
	for _, m := range msgs[:len(msgs)-1] {
		if m.Role == "user" || m.Role == "assistant" {
			history = append(history, llm.ChatMessage{Role: m.Role, Content: m.Content})
		}
	}

	return last.Content, history
}
