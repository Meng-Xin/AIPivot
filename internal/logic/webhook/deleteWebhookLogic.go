// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package webhook

import (
	"context"
	"time"

	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteWebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除 Webhook
func NewDeleteWebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteWebhookLogic {
	return &DeleteWebhookLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteWebhookLogic) DeleteWebhook(req *types.DeleteWebhookRequest) (resp *types.CommResponse, err error) {
	if err = l.svcCtx.WebhookRepo.Delete(l.ctx, req.ID); err != nil {
		l.Logger.Errorf("DeleteWebhook err: %v", err)
		return nil, errorx.NewInternalError("删除 Webhook 失败")
	}

	return &types.CommResponse{
		Code:      0,
		Msg:       "删除成功",
		Timestamp: time.Now().Unix(),
	}, nil
}
