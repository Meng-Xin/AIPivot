// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package webhook

import (
	"context"
	"encoding/json"
	"time"

	"aipivot/internal/modules/auth"
	whPkg "aipivot/internal/modules/channel/webhook"
	"aipivot/internal/shared/errorx"
	"aipivot/internal/shared/po"
	"aipivot/internal/svc"
	"aipivot/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateWebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建 Webhook
func NewCreateWebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateWebhookLogic {
	return &CreateWebhookLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateWebhookLogic) CreateWebhook(req *types.CreateWebhookRequest) (resp *types.WebhookDetailResponse, err error) {
	tenantID := auth.TenantIDFromContext(l.ctx)

	events := req.Events
	if len(events) == 0 {
		events = []string{"message.created"}
	}
	eventsJSON, _ := json.Marshal(events)

	wh := &po.Webhook{
		UUID:        uuid.New().String(),
		TenantID:    tenantID,
		Name:        req.Name,
		URL:         req.URL,
		Secret:      req.Secret,
		Events:      string(eventsJSON),
		ChannelType: req.ChannelType,
		Status:      "active",
		RetryCount:  req.RetryCount,
		TimeoutMs:   req.TimeoutMs,
	}
	if wh.ChannelType == "" {
		wh.ChannelType = "webhook"
	}

	if err = l.svcCtx.WebhookRepo.Create(l.ctx, wh); err != nil {
		l.Logger.Errorf("CreateWebhook err: %v", err)
		return nil, errorx.NewInternalError("创建 Webhook 失败")
	}

	show := whPkg.WebhookPoToShow(wh)
	return &types.WebhookDetailResponse{
		Code:      0,
		Msg:       "创建成功",
		Timestamp: time.Now().Unix(),
		Data:      show,
	}, nil
}
