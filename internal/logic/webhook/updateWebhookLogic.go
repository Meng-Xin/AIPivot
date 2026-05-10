// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package webhook

import (
	"context"
	"encoding/json"
	"time"

	whPkg "aipivot/internal/modules/channel/webhook"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateWebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新 Webhook
func NewUpdateWebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateWebhookLogic {
	return &UpdateWebhookLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateWebhookLogic) UpdateWebhook(req *types.UpdateWebhookRequest) (resp *types.WebhookDetailResponse, err error) {
	wh, err := l.svcCtx.WebhookRepo.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Logger.Errorf("UpdateWebhook GetByID err: %v", err)
		return nil, errorx.NewInternalError("更新 Webhook 失败")
	}
	if wh == nil {
		return nil, errorx.NewBusinessError(errorx.CodeNotFound, "Webhook 不存在")
	}

	// 仅更新非空字段
	if req.Name != "" {
		wh.Name = req.Name
	}
	if req.URL != "" {
		wh.URL = req.URL
	}
	if req.Secret != "" {
		wh.Secret = req.Secret
	}
	if len(req.Events) > 0 {
		eventsJSON, _ := json.Marshal(req.Events)
		wh.Events = string(eventsJSON)
	}
	if req.RetryCount > 0 {
		wh.RetryCount = req.RetryCount
	}
	if req.TimeoutMs > 0 {
		wh.TimeoutMs = req.TimeoutMs
	}
	if req.Status != "" {
		wh.Status = req.Status
	}

	if err = l.svcCtx.WebhookRepo.Update(l.ctx, wh); err != nil {
		l.Logger.Errorf("UpdateWebhook err: %v", err)
		return nil, errorx.NewInternalError("更新 Webhook 失败")
	}

	show := whPkg.WebhookPoToShow(wh)
	return &types.WebhookDetailResponse{
		Code:      0,
		Msg:       "更新成功",
		Timestamp: time.Now().Unix(),
		Data:      show,
	}, nil
}
