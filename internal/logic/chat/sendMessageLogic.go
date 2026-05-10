// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"
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

	// 3. TODO: 调用 RAG/LLM 生成 AI 回复（MVP 阶段使用 stub 回复）
	startTime := time.Now()
	aiContent := "你好！我是 AIPivot AI 助手，目前处于 MVP 阶段。RAG 和 LLM 集成即将上线，敬请期待！"
	latencyMs := int(time.Since(startTime).Milliseconds())

	// 4. 保存 AI 回复消息
	aiMsg := &po.Message{
		ConversationID: req.ConversationID,
		TenantID:       tenantID,
		Role:           "assistant",
		Content:        aiContent,
		ContentType:    "text",
		Model:          "stub-v1",
		LatencyMs:      latencyMs,
		Sources:        "[]",
	}
	if err = l.svcCtx.MessageRepo.Create(l.ctx, aiMsg); err != nil {
		l.Logger.Errorf("SendMessage CreateAIMsg err: %v", err)
		return nil, errorx.NewInternalError("发送消息失败")
	}

	// 5. 更新会话消息计数（+2: 用户消息 + AI 回复）
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
