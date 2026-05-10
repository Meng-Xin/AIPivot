package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/shared/sse"
	"aipivot/internal/svc"
	"aipivot/internal/types"
	"aipivot/pkg/llm"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageStreamLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发送消息（SSE 流式模式）
func NewSendMessageStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageStreamLogic {
	return &SendMessageStreamLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SendMessageStream 流式发送消息：校验 → 保存用户消息 → RAG 流式生成 → SSE 推送 → 保存 AI 消息。
// 直接操作 http.ResponseWriter，不走标准 JSON 响应模式。
func (l *SendMessageStreamLogic) SendMessageStream(w http.ResponseWriter, req *types.SendMessageRequest) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	// 初始化 SSE Writer（必须在任何错误响应之前，因为 SSE 需要特定 header）
	sseWriter, err := sse.NewWriter(w)
	if err != nil {
		l.Logger.Errorf("SendMessageStream SSE init err: %v", err)
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// 校验会话存在且未关闭
	conv, err := l.svcCtx.ConversationRepo.GetByID(l.ctx, req.ConversationID)
	if err != nil {
		l.Logger.Errorf("SendMessageStream GetConv err: %v", err)
		sseWriter.WriteError(errorx.CodeFailed, "发送消息失败")
		return
	}
	if conv == nil {
		sseWriter.WriteError(errorx.CodeNotFound, "会话不存在")
		return
	}
	if conv.Status == "closed" {
		sseWriter.WriteError(errorx.CodeFailed, "会话已关闭，无法发送消息")
		return
	}

	// 人工接管状态：仅保存用户消息，通知前端当前为人工服务模式
	if conv.Status == "waiting_human" {
		userMsg := &po.Message{
			UUID:           uuid.New().String(),
			ConversationID: req.ConversationID,
			TenantID:       tenantID,
			Role:           "user",
			Content:        req.Content,
			ContentType:    req.ContentType,
		}
		if err = l.svcCtx.MessageRepo.Create(l.ctx, userMsg); err != nil {
			l.Logger.Errorf("SendMessageStream CreateUserMsg(waiting_human) err: %v", err)
			sseWriter.WriteError(errorx.CodeFailed, "发送消息失败")
			return
		}
		_ = l.svcCtx.ConversationRepo.IncrMessageCount(l.ctx, req.ConversationID)
		// 告知前端当前是人工服务模式，消息已保存但不会有 AI 回复
		_ = sseWriter.WriteEvent("message_start", sse.MessageStart{
			MessageID:      userMsg.UUID,
			ConversationID: req.ConversationID,
		})
		_ = sseWriter.WriteEvent("delta", sse.Delta{Content: "当前会话已转接人工客服，您的消息已发送，请稍候。"})
		_ = sseWriter.WriteEvent("message_end", sse.MessageEnd{
			MessageID: userMsg.UUID,
		})
		_ = sseWriter.WriteDone()
		return
	}

	// 保存用户消息
	userMsg := &po.Message{
		UUID:           uuid.New().String(),
		ConversationID: req.ConversationID,
		TenantID:       tenantID,
		Role:           "user",
		Content:        req.Content,
		ContentType:    req.ContentType,
	}
	if err = l.svcCtx.MessageRepo.Create(l.ctx, userMsg); err != nil {
		l.Logger.Errorf("SendMessageStream CreateUserMsg err: %v", err)
		sseWriter.WriteError(errorx.CodeFailed, "发送消息失败")
		return
	}

	// 获取最近对话历史（util.go 中定义的共享函数）
	recentMsgs, _ := l.svcCtx.MessageRepo.GetRecentMessages(l.ctx, req.ConversationID, 10)
	history := buildChatHistory(recentMsgs)

	// 调用 RAG 流式生成
	startTime := time.Now()
	var kbID int64
	if conv.KnowledgeBaseID != nil {
		kbID = *conv.KnowledgeBaseID
	}

	stream, meta, err := l.svcCtx.RAGService.AnswerStream(l.ctx, kbID, req.Content, history, conv.Model)
	if err != nil {
		l.Logger.Errorf("SendMessageStream RAG.AnswerStream err: %v", err)
		sseWriter.WriteError(errorx.CodeLLMUnavailable, "AI 回复生成失败，请稍后重试")
		return
	}

	// 发送 message_start 事件
	aiMsgUUID := uuid.New().String()
	_ = sseWriter.WriteEvent("message_start", sse.MessageStart{
		MessageID:      aiMsgUUID,
		ConversationID: req.ConversationID,
	})

	// 消费流式 token，逐个推送 delta 事件
	var contentBuf strings.Builder
	var finalModel string
	var finalUsage *llm.ChatUsage

	for evt := range stream {
		if evt.Err != nil {
			l.Logger.Errorf("SendMessageStream stream err: %v", evt.Err)
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

	latencyMs := int(time.Since(startTime).Milliseconds())

	// 优先使用 LLM 返回的 usage 统计 token
	tokenCount := 0
	if finalUsage != nil {
		tokenCount = finalUsage.TotalTokens
	}

	// 发送 message_end 事件
	_ = sseWriter.WriteEvent("message_end", sse.MessageEnd{
		MessageID:  aiMsgUUID,
		Model:      finalModel,
		TokenCount: tokenCount,
		LatencyMs:  latencyMs,
		Sources:    meta.Sources,
	})
	_ = sseWriter.WriteDone()

	// 持久化 AI 回复消息
	sourcesJSON, _ := json.Marshal(meta.Sources)
	aiMsg := &po.Message{
		UUID:           aiMsgUUID,
		ConversationID: req.ConversationID,
		TenantID:       tenantID,
		Role:           "assistant",
		Content:        contentBuf.String(),
		ContentType:    "text",
		TokenCount:     tokenCount,
		Model:          finalModel,
		LatencyMs:      latencyMs,
		Sources:        string(sourcesJSON),
	}
	if err = l.svcCtx.MessageRepo.Create(l.ctx, aiMsg); err != nil {
		l.Logger.Errorf("SendMessageStream CreateAIMsg err: %v", err)
	}

	// 更新会话消息计数（+2: 用户消息 + AI 回复）
	_ = l.svcCtx.ConversationRepo.IncrMessageCount(l.ctx, req.ConversationID)
	_ = l.svcCtx.ConversationRepo.IncrMessageCount(l.ctx, req.ConversationID)

	// Agent 触发的自动转接：检查工具调用中是否包含 escalate_to_human
	for _, tu := range meta.ToolUses {
		if tu.Name == "escalate_to_human" {
			if statusErr := l.svcCtx.ConversationRepo.UpdateStatus(l.ctx, req.ConversationID, "waiting_human"); statusErr != nil {
				l.Logger.Errorf("SendMessageStream auto-escalate err: %v", statusErr)
			}
			l.Logger.Infof("conversation %d auto-escalated by agent (stream)", req.ConversationID)
			break
		}
	}
}
