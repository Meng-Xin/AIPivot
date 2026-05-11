// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"
	"time"

	chatmod "aipivot/internal/modules/chat"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取消息历史
func NewListMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMessageLogic {
	return &ListMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMessageLogic) ListMessage(req *types.ListMessageRequest) (resp *types.MessageListResponse, err error) {
	list, total, err := l.svcCtx.MessageRepo.GetList(l.ctx, req.ConversationID, req.Page, req.PageSize)
	if err != nil {
		l.Logger.Errorf("ListMessage err: %v", err)
		return nil, errorx.NewInternalError("查询消息列表失败")
	}

	return &types.MessageListResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data: types.MessageListData{
			Total: total,
			List:  chatmod.ToShowMessageList(list),
		},
	}, nil
}
