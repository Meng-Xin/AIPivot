// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package webhook

import (
	"context"
	"time"

	whPkg "aipivot/internal/modules/channel/webhook"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 Webhook 详情
func NewGetWebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWebhookLogic {
	return &GetWebhookLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetWebhookLogic) GetWebhook(req *types.GetWebhookRequest) (resp *types.WebhookDetailResponse, err error) {
	wh, err := l.svcCtx.WebhookRepo.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Logger.Errorf("GetWebhook err: %v", err)
		return nil, errorx.NewInternalError("获取 Webhook 失败")
	}
	if wh == nil {
		return nil, errorx.NewBusinessError(errorx.CodeNotFound, "Webhook 不存在")
	}

	show := whPkg.WebhookPoToShow(wh)
	return &types.WebhookDetailResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      show,
	}, nil
}
