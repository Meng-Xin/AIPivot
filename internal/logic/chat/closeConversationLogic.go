// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"
	"time"

	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CloseConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 关闭会话
func NewCloseConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseConversationLogic {
	return &CloseConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CloseConversationLogic) CloseConversation(req *types.CloseConversationRequest) (resp *types.CommResponse, err error) {
	conv, err := l.svcCtx.ConversationRepo.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Logger.Errorf("CloseConversation GetByID err: %v", err)
		return nil, errorx.NewInternalError("关闭会话失败")
	}
	if conv == nil {
		return nil, errorx.NewNotFoundError("会话不存在")
	}
	if conv.Status == "closed" {
		return nil, errorx.NewBusinessError(errorx.CodeFailed, "会话已关闭")
	}

	if err = l.svcCtx.ConversationRepo.Close(l.ctx, req.ID); err != nil {
		l.Logger.Errorf("CloseConversation err: %v", err)
		return nil, errorx.NewInternalError("关闭会话失败")
	}

	return &types.CommResponse{
		Code:      0,
		Msg:       "会话已关闭",
		Timestamp: time.Now().Unix(),
	}, nil
}
