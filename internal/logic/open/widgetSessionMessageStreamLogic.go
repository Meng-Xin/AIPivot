// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package open

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/channel"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/shared/sse"
	"aipivot/internal/svc"
	"aipivot/internal/types"
	"aipivot/pkg/llm"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type WidgetSessionMessageStreamLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Widget — 流式发送消息（SSE，持久化 user + assistant 消息）
func NewWidgetSessionMessageStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WidgetSessionMessageStreamLogic {
	return &WidgetSessionMessageStreamLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// WidgetSessionMessageStream 处理 Widget 流式发消息：鉴权 → 加载会话 → 保存 user 消息 →
// 调用 RAG.AnswerStream 推送 SSE delta → 持久化 assistant 消息。
// 融合 webhookInbound（消息持久化模式）与 chatCompletionStream（SSE 推送模式）。
func (l *WidgetSessionMessageStreamLogic) WidgetSessionMessageStream(w http.ResponseWriter, req *types.WidgetMessageSendRequest) {
	tenantID := auth.TenantIDFromAPIKeyContext(l.ctx)
	if tenantID == 0 {
		http.Error(w, `{"code":401,"msg":"invalid API key context"}`, http.StatusUnauthorized)
		return
	}
	if !auth.IsPublicApiKey(l.ctx) {
		http.Error(w, `{"code":400,"msg":"public key required"}`, http.StatusBadRequest)
		return
	}

	sseWriter, err := sse.NewWriter(w)
	if err != nil {
		l.Logger.Errorf("WidgetStream SSE init err: %v", err)
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// 加载会话；必须同租户且为 widget 渠道
	conv, err := l.svcCtx.ConversationRepo.GetByUUID(l.ctx, req.SessionToken)
	if err != nil || conv == nil {
		l.Logger.Errorf("WidgetStream GetByUUID err: %v", err)
		sseWriter.WriteError(errorx.CodeNotFound, "会话不存在")
		return
	}
	if conv.TenantID != tenantID || conv.Channel != channel.Widget.String() {
		sseWriter.WriteError(errorx.CodeForbid, "无权访问该会话")
		return
	}

	contentType := req.ContentType
	if contentType == "" {
		contentType = "text"
	}
	if req.Content == "" {
		sseWriter.WriteError(errorx.CodeBadRequest, "消息内容不能为空")
		return
	}

	// 访客维度滑窗限流（防止恶意刷量），key 同时包含 visitorId + conversationId 防止越权绕过
	visitorKey := "widget:msg:" + conv.ExternalUserID + ":" + conv.UUID
	if l.svcCtx.WidgetRateLimiter != nil && !l.svcCtx.WidgetRateLimiter.Allow(l.ctx, visitorKey) {
		sseWriter.WriteError(errorx.CodeTooManyRequests, "发送太频繁，请稍后再试")
		return
	}

	// 租户维度日 Token 配额检查（Widget 与其他渠道共享租户配额）
	if err := l.svcCtx.TokenLimiter.CheckByApiKey(l.ctx, tenantID); err != nil {
		sseWriter.WriteError(errorx.CodeTokenExceeded, err.Error())
		return
	}

	// 保存用户消息
	userMsgUUID := uuid.New().String()
	userMsg := &po.Message{
		UUID:           userMsgUUID,
		ConversationID: conv.ID,
		TenantID:       tenantID,
		Role:           "user",
		Content:        req.Content,
		ContentType:    contentType,
	}
	if err = l.svcCtx.MessageRepo.Create(l.ctx, userMsg); err != nil {
		l.Logger.Errorf("WidgetStream CreateUserMsg err: %v", err)
		sseWriter.WriteError(errorx.CodeFailed, "保存消息失败")
		return
	}
	_ = l.svcCtx.ConversationRepo.IncrMessageCount(l.ctx, conv.ID)

	// 构建 LLM 上下文（最近 10 条历史消息，已按时间正序）
	history := l.buildHistory(conv.ID)
	kbID, _ := auth.KnowledgeBaseIDFromApiKeyContext(l.ctx)

	// 流式生成
	stream, meta, err := l.svcCtx.RAGService.AnswerStream(l.ctx, kbID, req.Content, history, "", nil)
	if err != nil {
		l.Logger.Errorf("WidgetStream RAG.AnswerStream err: %v", err)
		sseWriter.WriteError(errorx.CodeLLMUnavailable, "AI 回复生成失败")
		return
	}

	assistantUUID := uuid.New().String()
	_ = sseWriter.WriteEvent("message_start", sse.MessageStart{
		MessageID:      assistantUUID,
		ConversationID: conv.ID,
	})

	var contentBuf strings.Builder
	var finalModel string
	var finalUsage *llm.ChatUsage
	startedAt := time.Now()

	for evt := range stream {
		if evt.Err != nil {
			l.Logger.Errorf("WidgetStream stream err: %v", evt.Err)
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
	latencyMs := int(time.Since(startedAt).Milliseconds())

	// 持久化 assistant 消息
	sourcesJSON := "[]"
	if meta != nil && len(meta.Sources) > 0 {
		if buf, mErr := json.Marshal(meta.Sources); mErr == nil {
			sourcesJSON = string(buf)
		}
	}
	aiMsg := &po.Message{
		UUID:           assistantUUID,
		ConversationID: conv.ID,
		TenantID:       tenantID,
		Role:           "assistant",
		Content:        contentBuf.String(),
		ContentType:    "text",
		TokenCount:     tokenCount,
		Model:          finalModel,
		LatencyMs:      latencyMs,
		Sources:        sourcesJSON,
	}
	if err = l.svcCtx.MessageRepo.Create(l.ctx, aiMsg); err != nil {
		l.Logger.Errorf("WidgetStream CreateAIMsg err: %v", err)
	}
	_ = l.svcCtx.ConversationRepo.IncrMessageCount(l.ctx, conv.ID)

	endSources := []string{}
	if meta != nil {
		endSources = meta.Sources
	}
	_ = sseWriter.WriteEvent("message_end", sse.MessageEnd{
		MessageID:  assistantUUID,
		Model:      finalModel,
		TokenCount: tokenCount,
		LatencyMs:  latencyMs,
		Sources:    endSources,
	})

	// 累加租户当日 Token 用量（fire-and-forget，不阻塞响应）
	l.svcCtx.TokenLimiter.IncrByApiKey(l.ctx, tenantID, tokenCount)

	_ = sseWriter.WriteDone()
}

// buildHistory 加载最近 N 条消息并转为 llm.ChatMessage（剔除刚保存的当前 user 消息由调用方传入）。
func (l *WidgetSessionMessageStreamLogic) buildHistory(convID int64) []llm.ChatMessage {
	msgs, err := l.svcCtx.MessageRepo.GetRecentMessages(l.ctx, convID, 11)
	if err != nil {
		return nil
	}
	// 跳过最后一条（刚保存的当前 user 消息），保留前 N 条历史
	history := make([]llm.ChatMessage, 0, len(msgs))
	limit := len(msgs) - 1
	if limit < 0 {
		limit = 0
	}
	for i := 0; i < limit; i++ {
		m := msgs[i]
		if m.Role != "user" && m.Role != "assistant" {
			continue
		}
		history = append(history, llm.ChatMessage{Role: m.Role, Content: m.Content})
	}
	return history
}
