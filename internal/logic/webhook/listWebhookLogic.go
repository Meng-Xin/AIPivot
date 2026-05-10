// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package webhook

import (
	"context"
	"time"

	"aipivot/internal/modules/auth"
	whPkg "aipivot/internal/modules/channel/webhook"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListWebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 Webhook 列表
func NewListWebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWebhookLogic {
	return &ListWebhookLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListWebhookLogic) ListWebhook() (resp *types.WebhookListResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	list, err := l.svcCtx.WebhookRepo.GetListByTenant(l.ctx, tenantID)
	if err != nil {
		l.Logger.Errorf("ListWebhook err: %v", err)
		return nil, errorx.NewInternalError("获取 Webhook 列表失败")
	}

	return &types.WebhookListResponse{
		Code:      0,
		Msg:       "OK",
		Timestamp: time.Now().Unix(),
		Data:      whPkg.WebhookPoListToShowList(list),
	}, nil
}
