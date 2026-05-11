// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package open

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/channel"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"
	"aipivot/pkg/llm"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type WebhookInboundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Webhook 入站消息 — 接收外部平台推送的用户消息并返回 AI 回复
func NewWebhookInboundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WebhookInboundLogic {
	return &WebhookInboundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// WebhookInbound 处理入站 Webhook 消息：验证 → 查找/创建会话 → 保存用户消息 → RAG 生成 → 保存 AI 回复。
func (l *WebhookInboundLogic) WebhookInbound(req *types.WebhookInboundRequest) (resp *types.WebhookInboundResponse, err error) {
	tenantID := auth.TenantIDFromAPIKeyContext(l.ctx)
	if tenantID == 0 {
		return nil, errorx.NewUnauthError("invalid API key context")
	}

	// 验证 Webhook 配置存在
	wh, err := l.svcCtx.WebhookRepo.GetByID(l.ctx, req.WebhookID)
	if err != nil {
		l.Logger.Errorf("WebhookInbound GetWebhook err: %v", err)
		return nil, errorx.NewInternalError("处理入站消息失败")
	}
	if wh == nil || wh.Status != "active" {
		return nil, errorx.NewNotFoundError("Webhook 不存在或已禁用")
	}

	// 查找或创建会话
	var convID int64
	if req.ConversationID > 0 {
		conv, err := l.svcCtx.ConversationRepo.GetByID(l.ctx, req.ConversationID)
		if err != nil || conv == nil {
			return nil, errorx.NewNotFoundError("会话不存在")
		}
		convID = conv.ID
	} else {
		// 为入站消息自动创建新会话
		conv := &po.Conversation{
			UUID:     uuid.New().String(),
			TenantID: tenantID,
			Title:    "Webhook: " + req.ExternalUserID,
			Status:   "active",
			Channel:  channel.Webhook.String(),
		}
		if err = l.svcCtx.ConversationRepo.Create(l.ctx, conv); err != nil {
			l.Logger.Errorf("WebhookInbound CreateConversation err: %v", err)
			return nil, errorx.NewInternalError("创建会话失败")
		}
		convID = conv.ID
	}

	// 保存用户消息
	userMsgUUID := uuid.New().String()
	contentType := req.ContentType
	if contentType == "" {
		contentType = "text"
	}
	userMsg := &po.Message{
		UUID:           userMsgUUID,
		ConversationID: convID,
		TenantID:       tenantID,
		Role:           "user",
		Content:        req.Content,
		ContentType:    contentType,
	}
	if err = l.svcCtx.MessageRepo.Create(l.ctx, userMsg); err != nil {
		l.Logger.Errorf("WebhookInbound CreateUserMsg err: %v", err)
		return nil, errorx.NewInternalError("保存消息失败")
	}
	_ = l.svcCtx.ConversationRepo.IncrMessageCount(l.ctx, convID)

	// 调用 RAG 生成 AI 回复
	result, err := l.svcCtx.RAGService.Answer(l.ctx, 0, req.Content, []llm.ChatMessage{}, "")
	if err != nil {
		l.Logger.Errorf("WebhookInbound RAG.Answer err: %v", err)
		return nil, errorx.NewBusinessError(errorx.CodeLLMUnavailable, "AI 回复生成失败")
	}

	// 保存 AI 回复
	aiMsgUUID := uuid.New().String()
	aiMsg := &po.Message{
		UUID:           aiMsgUUID,
		ConversationID: convID,
		TenantID:       tenantID,
		Role:           "assistant",
		Content:        result.Content,
		ContentType:    "text",
		TokenCount:     result.TokenCount,
		Model:          result.Model,
	}
	if err = l.svcCtx.MessageRepo.Create(l.ctx, aiMsg); err != nil {
		l.Logger.Errorf("WebhookInbound CreateAIMsg err: %v", err)
	}
	_ = l.svcCtx.ConversationRepo.IncrMessageCount(l.ctx, convID)

	return &types.WebhookInboundResponse{
		Code:           0,
		Msg:            "OK",
		Timestamp:      time.Now().Unix(),
		ConversationID: convID,
		MessageID:      aiMsgUUID,
		Content:        result.Content,
	}, nil
}
