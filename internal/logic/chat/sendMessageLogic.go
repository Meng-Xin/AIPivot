// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"
	"encoding/json"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/modules/chat/domain/assembler"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发送消息（同步模式，MVP 阶段）
func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendMessageLogic) SendMessage(req *types.SendMessageRequest) (resp *types.SendMessageResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	// 1. 校验会话存在且未关闭
	conv, err := l.svcCtx.ConversationRepo.GetByID(l.ctx, req.ConversationID)
	if err != nil {
		l.Logger.Errorf("SendMessage GetConv err: %v", err)
		return nil, errorx.NewInternalError("发送消息失败")
	}
	if conv == nil {
		return nil, errorx.NewNotFoundError("会话不存在")
	}
	if conv.Status == "closed" {
		return nil, errorx.NewBusinessError(errorx.CodeFailed, "会话已关闭，无法发送消息")
	}

	// 2. 保存用户消息
	userMsg := &po.Message{
		ConversationID: req.ConversationID,
		TenantID:       tenantID,
		Role:           "user",
		Content:        req.Content,
		ContentType:    req.ContentType,
	}
	if err = l.svcCtx.MessageRepo.Create(l.ctx, userMsg); err != nil {
		l.Logger.Errorf("SendMessage CreateUserMsg err: %v", err)
		return nil, errorx.NewInternalError("发送消息失败")
	}

	// 3. 获取最近对话历史，构建 LLM context
	recentMsgs, _ := l.svcCtx.MessageRepo.GetRecentMessages(l.ctx, req.ConversationID, 10)
	history := buildChatHistory(recentMsgs)

	// 4. 调用 RAG 服务生成 AI 回复
	startTime := time.Now()
	var kbID int64
	if conv.KnowledgeBaseID != nil {
		kbID = *conv.KnowledgeBaseID
	}

	result, err := l.svcCtx.RAGService.Answer(l.ctx, kbID, req.Content, history)
	if err != nil {
		l.Logger.Errorf("SendMessage RAG.Answer err: %v", err)
		return nil, errorx.NewBusinessError(errorx.CodeLLMUnavailable, "AI 回复生成失败，请稍后重试")
	}
	latencyMs := int(time.Since(startTime).Milliseconds())

	// 5. 保存 AI 回复消息
	sourcesJSON, _ := json.Marshal(result.Sources)
	aiMsg := &po.Message{
		ConversationID: req.ConversationID,
		TenantID:       tenantID,
		Role:           "assistant",
		Content:        result.Content,
		ContentType:    "text",
		TokenCount:     result.TokenCount,
		Model:          result.Model,
		LatencyMs:      latencyMs,
		Sources:        string(sourcesJSON),
	}
	if err = l.svcCtx.MessageRepo.Create(l.ctx, aiMsg); err != nil {
		l.Logger.Errorf("SendMessage CreateAIMsg err: %v", err)
		return nil, errorx.NewInternalError("发送消息失败")
	}

	// 6. 更新会话消息计数（+2: 用户消息 + AI 回复）
	_ = l.svcCtx.ConversationRepo.IncrMessageCount(l.ctx, req.ConversationID)
	_ = l.svcCtx.ConversationRepo.IncrMessageCount(l.ctx, req.ConversationID)

	show := assembler.MessagePoToShow(aiMsg)
	return &types.SendMessageResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      show,
	}, nil
}
