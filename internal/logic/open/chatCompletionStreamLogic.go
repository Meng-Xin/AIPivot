package open

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/sse"
	"aipivot/internal/svc"
	"aipivot/internal/types"
	"aipivot/pkg/llm"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type ChatCompletionStreamLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatCompletionStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatCompletionStreamLogic {
	return &ChatCompletionStreamLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ChatCompletionStream SSE 流式 Chat Completion：直接操作 ResponseWriter 推送增量 token。
func (l *ChatCompletionStreamLogic) ChatCompletionStream(w http.ResponseWriter, req *types.ChatCompletionRequest) {
	tenantID := auth.TenantIDFromAPIKeyContext(l.ctx)
	if tenantID == 0 {
		http.Error(w, `{"code":401,"msg":"invalid API key context"}`, http.StatusUnauthorized)
		return
	}

	sseWriter, err := sse.NewWriter(w)
	if err != nil {
		l.Logger.Errorf("ChatCompletionStream SSE init err: %v", err)
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	if len(req.Messages) == 0 {
		sseWriter.WriteError(errorx.CodeBadRequest, "messages cannot be empty")
		return
	}

	question, history := extractQuestionAndHistory(req.Messages)
	if question == "" {
		sseWriter.WriteError(errorx.CodeBadRequest, "last message must have role=user")
		return
	}

	// 调用 RAG 流式生成
	stream, meta, err := l.svcCtx.RAGService.AnswerStream(l.ctx, req.KnowledgeBaseID, question, history, req.Model, nil)
	if err != nil {
		l.Logger.Errorf("ChatCompletionStream RAG.AnswerStream err: %v", err)
		sseWriter.WriteError(errorx.CodeLLMUnavailable, "AI 回复生成失败")
		return
	}

	completionID := fmt.Sprintf("chatcmpl-%s", uuid.New().String()[:8])
	_ = sseWriter.WriteEvent("message_start", sse.MessageStart{
		MessageID:      completionID,
		ConversationID: 0, // Open API 无状态，不关联 conversation
	})

	// 消费流式 token，逐个推送 delta 事件
	var contentBuf strings.Builder
	var finalModel string
	var finalUsage *llm.ChatUsage

	for evt := range stream {
		if evt.Err != nil {
			l.Logger.Errorf("ChatCompletionStream stream err: %v", evt.Err)
			sseWriter.WriteError(errorx.CodeLLMUnavailable, "AI 回复生成中断")
			break
		}
		if evt.Content != "" {
			contentBuf.WriteString(evt.Content)
			_ = sseWriter.WriteEvent("delta", sse.Delta{Content: evt.Content})
		}
		if evt.Model != "" {
			finalModel = evt.Model
		}
		if evt.Usage != nil {
			finalUsage = evt.Usage
		}
		if evt.Done {
			break
		}
	}

	tokenCount := 0
	if finalUsage != nil {
		tokenCount = finalUsage.TotalTokens
	}

	_ = sseWriter.WriteEvent("message_end", sse.MessageEnd{
		MessageID:  completionID,
		Model:      finalModel,
		TokenCount: tokenCount,
		Sources:    meta.Sources,
	})
	_ = sseWriter.WriteDone()
}
