// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	chatmod "aipivot/internal/modules/chat"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取会话列表
func NewListConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListConversationLogic {
	return &ListConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListConversationLogic) ListConversation(req *types.ListConversationRequest) (resp *types.ConversationListResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	list, total, err := l.svcCtx.ConversationRepo.GetList(l.ctx, tenantID, req.Page, req.PageSize, req.Status)
	if err != nil {
		l.Logger.Errorf("ListConversation err: %v", err)
		return nil, errorx.NewInternalError("查询会话列表失败")
	}

	return &types.ConversationListResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data: types.ConversationListData{
			Total: total,
			List:  chatmod.ToShowConversationList(list),
		},
	}, nil
}
