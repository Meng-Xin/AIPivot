// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"
	"time"

	"aipivot/internal/modules/chat/domain/assembler"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取会话详情
func NewGetConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationLogic {
	return &GetConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConversationLogic) GetConversation(req *types.GetConversationRequest) (resp *types.ConversationDetailResponse, err error) {
	conv, err := l.svcCtx.ConversationRepo.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Logger.Errorf("GetConversation err: %v", err)
		return nil, errorx.NewInternalError("查询会话失败")
	}
	if conv == nil {
		return nil, errorx.NewNotFoundError("会话不存在")
	}

	show := assembler.ConversationPoToShow(conv)
	return &types.ConversationDetailResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      show,
	}, nil
}
