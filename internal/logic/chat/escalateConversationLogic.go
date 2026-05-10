package chat

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type EscalateConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEscalateConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EscalateConversationLogic {
	return &EscalateConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// EscalateConversation 将会话转接给人工客服：校验会话状态 → 更新为 waiting_human → 插入系统消息。
func (l *EscalateConversationLogic) EscalateConversation(req *types.EscalateConversationRequest) (resp *types.CommResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	conv, err := l.svcCtx.ConversationRepo.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Logger.Errorf("EscalateConversation GetByID err: %v", err)
		return nil, errorx.NewInternalError("转接失败")
	}
	if conv == nil {
		return nil, errorx.NewNotFoundError("会话不存在")
	}
	if conv.TenantID != tenantID {
		return nil, errorx.NewForbidError("无权操作此会话")
	}
	if conv.Status == "closed" {
		return nil, errorx.NewBusinessError(errorx.CodeFailed, "会话已关闭，无法转接")
	}
	if conv.Status == "waiting_human" {
		return nil, errorx.NewBusinessError(errorx.CodeFailed, "会话已在等待人工客服")
	}

	// 更新会话状态为 waiting_human
	if err = l.svcCtx.ConversationRepo.UpdateStatus(l.ctx, req.ID, "waiting_human"); err != nil {
		l.Logger.Errorf("EscalateConversation UpdateStatus err: %v", err)
		return nil, errorx.NewInternalError("转接失败")
	}

	l.Logger.Infof("conversation %d escalated to human, reason: %s", req.ID, req.Reason)

	return &types.CommResponse{
		Code:      0,
		Msg:       "已转接人工客服，请稍候",
		Timestamp: time.Now().Unix(),
	}, nil
}
