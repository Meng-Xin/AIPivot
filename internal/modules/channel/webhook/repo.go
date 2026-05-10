package webhook

import (
	"context"

	"aipivot/internal/shared/po"
)

// Repository Webhook 数据仓储接口。
type Repository interface {
	Create(ctx context.Context, wh *po.Webhook) error
	GetByID(ctx context.Context, id int64) (*po.Webhook, error)
	GetListByTenant(ctx context.Context, tenantID int64) ([]*po.Webhook, error)
	GetActiveByTenantAndEvent(ctx context.Context, tenantID int64, event string) ([]*po.Webhook, error)
	Update(ctx context.Context, wh *po.Webhook) error
	Delete(ctx context.Context, id int64) error
}

type WebhookRepo struct {
	dao *WebhookDao
}

func NewWebhookRepo(dao *WebhookDao) *WebhookRepo {
	return &WebhookRepo{dao: dao}
}

func (r *WebhookRepo) Create(ctx context.Context, wh *po.Webhook) error {
	return r.dao.Create(ctx, wh)
}

func (r *WebhookRepo) GetByID(ctx context.Context, id int64) (*po.Webhook, error) {
	return r.dao.GetByID(ctx, id)
}

func (r *WebhookRepo) GetListByTenant(ctx context.Context, tenantID int64) ([]*po.Webhook, error) {
	return r.dao.GetListByTenant(ctx, tenantID)
}

func (r *WebhookRepo) GetActiveByTenantAndEvent(ctx context.Context, tenantID int64, event string) ([]*po.Webhook, error) {
	return r.dao.GetActiveByTenantAndEvent(ctx, tenantID, event)
}

func (r *WebhookRepo) Update(ctx context.Context, wh *po.Webhook) error {
	return r.dao.Update(ctx, wh)
}

func (r *WebhookRepo) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}
